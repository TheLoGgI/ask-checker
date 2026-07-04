package migrate

import (
	"ask-checker/database"
	"encoding/csv"
	"fmt"
	"log"
	"log/slog"
	"os"
)

func MigrateSQLite_csv(db database.Database) {
	var csvFilename = "maj-2026.csv"
	fmt.Println("Tring To read File")
	file, err := os.Open(csvFilename)
	if err != nil {
		slog.Info("Error", "err", err.Error())
		return
	}
	defer file.Close()

	fmt.Println("File has been read")

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		slog.Info("Error", "err", err.Error())
		return
	}

	log.Println("Records has been read!")

	var insertCount int
	for i, record := range records {
		if i == 0 {
			continue // skip header row
		}
		var ISINcode = record[1]
		var name = record[5]
		var lei = record[4]

		err := db.Insert(
			`INSERT OR IGNORE INTO ask_checker (isin, name, lei) VALUES (?, ?, ?)`,
			ISINcode, name, lei,
		)
		if err != nil {
			log.Default()
		} else {
			insertCount++
		}
	}
	fmt.Printf("Migration complete - inserted %d records\n", insertCount)
}

func MigratePG_csv(db database.Database) {
	var csvFilename = "maj-2026.csv"
	file, err := os.Open(csvFilename)
	if err != nil {
		slog.Info("Error", "err", err.Error())
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		slog.Info("Error", "err", err.Error())
		return
	}

	var insertCount int
	insertSQL := `INSERT INTO ask_checker (isin, name, lei) VALUES ($1, $2, $3) ON CONFLICT (isin) DO NOTHING`

	for i, record := range records {
		if i == 0 {
			continue // skip header row
		}
		var ISINcode = record[1]
		var name = record[5]
		var lei = record[4]

		err := db.Insert(
			insertSQL,
			ISINcode, name, lei,
		)
		if err != nil {
			log.Printf("Insert failed for ISIN %s: %v", ISINcode, err)
		} else {
			insertCount++
		}
	}
	fmt.Printf("Migration complete - inserted %d records\n", insertCount)
}
