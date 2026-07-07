package main

import (
	"ask-checker/database"
	"ask-checker/migrate"
	structs "ask-checker/stucts"
	"context"
	"log"
	"strings"
)

func migrate_ticker(db database.Database) {

	log.Println("Migrate Ticker from Nordet API")

	selectSQL := "SELECT isin, name, lei FROM ask_checker"
	rows, err := db.Query(selectSQL)
	if err != nil {
		log.Printf("Error: %v", err)
		return
	}
	defer rows.Close()

	results := make([]structs.Ask, 0, 10)
	for rows.Next() {
		var item structs.Ask
		if err := rows.Scan(&item.Isin, &item.Name, &item.Lai); err != nil {
			log.Printf("Scan error: %v", err)
			return
		}
		results = append(results, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Rows error: %v", err)
		return
	}

	log.Println("Results Counts: ", len(results))
	patches := migrate.CreateDatabaseTickerPatches(results)
	log.Println("Patches Counts: ", len(patches))
	patchASK(db, patches)
}

func patchASK(db database.Database, patches []migrate.TickerPatch) {
	if len(patches) == 0 {
		log.Println("No ticker updates to apply")
		return
	}

	if err := ensureTickerColumn(db); err != nil {
		log.Printf("Could not ensure ticker column: %v", err)
		return
	}

	conn, err := db.Connect()
	if err != nil {
		log.Printf("Connection error: %v", err)
		return
	}
	defer conn.Close()

	tx, err := conn.BeginTx(context.Background(), nil)
	if err != nil {
		log.Printf("Transaction start error: %v", err)
		return
	}

	updateSQL := "UPDATE ask_checker SET ticker = ? WHERE isin = ?"
	if strings.EqualFold(Cfg.Database, "postgres") {
		updateSQL = "UPDATE ask_checker SET ticker = $1 WHERE isin = $2"
	}

	var updated int64
	for _, patch := range patches {
		result, execErr := tx.ExecContext(context.Background(), updateSQL, patch.Ticker, patch.Isin)
		if execErr != nil {
			_ = tx.Rollback()
			log.Printf("Update failed for ISIN %s: %v", patch.Isin, execErr)
			return
		}

		rowsAffected, _ := result.RowsAffected()
		updated += rowsAffected
	}

	if err := tx.Commit(); err != nil {
		log.Printf("Transaction commit error: %v", err)
		return
	}

	log.Printf("Ticker patch complete: %d rows updated", updated)
}

func ensureTickerColumn(db database.Database) error {
	if strings.EqualFold(Cfg.Database, "postgres") {
		return db.Insert(`ALTER TABLE ask_checker ADD COLUMN IF NOT EXISTS ticker TEXT`)
	}

	err := db.Insert(`ALTER TABLE ask_checker ADD COLUMN ticker TEXT`)
	if err == nil {
		return nil
	}

	errMsg := strings.ToLower(err.Error())
	if strings.Contains(errMsg, "duplicate column") {
		return nil
	}

	return err
}
