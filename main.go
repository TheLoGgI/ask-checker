package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

var globalDB *sql.DB

func init() {
	db, err := sql.Open("sqlite", "./ask.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS ask_list (
		isin TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		lei  TEXT NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}

	globalDB = db

	fmt.Println("Database ready")
}

// func extractData(db *sql.DB) {
// 	var datasetURL = "https://skat.dk/media/r1dn5su0/maj-2026-abis-liste-til-offentliggoerelse-2021-2026.xlsx"

// }

func migrate_csv() {
	var csvFilename = "maj-2026.csv"
	file, err := os.Open(csvFilename)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var insertCount int
	for i, record := range records {
		if i == 0 {
			continue // skip header row
		}
		var ISINcode = record[1]
		var name = record[5]
		var lei = record[4]

		_, err := globalDB.Exec(
			`INSERT OR IGNORE INTO ask_list (isin, name, lei) VALUES (?, ?, ?)`,
			ISINcode, name, lei,
		)
		if err != nil {
			log.Printf("Row %d: %v", i, err)
		} else {
			insertCount++
		}
	}
	fmt.Printf("Migration complete - inserted %d records\n", insertCount)
}

func main() {

	fmt.Println("ASK Checker Running!")

	migrate_csv()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World!"))
	})

	r.Get("/ask", func(w http.ResponseWriter, r *http.Request) {

		var query = r.URL.Query()
		var isin = query.Get("isin")
		if isin == "" {
			isin = query.Get("isni") // fallback for legacy param name
		}
		log.Printf("Searching for ISIN: '%s'", isin)
		if isin == "" {
			http.Error(w, "Missing isin parameter", http.StatusBadRequest)
			return
		}
		// var lei = query.Get("lei")
		// var tickerCode = query.Get("ticker")

		rows, err := globalDB.Query("SELECT isin, name FROM ask_list WHERE isin = ? LIMIT 1", isin)
		if err != nil {
			log.Printf("Error: %v", err)
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var askStruct ask
		if !rows.Next() {
			log.Printf("No rows found for ISIN: '%s'", isin)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if err := rows.Scan(&askStruct.Isni, &askStruct.Name); err != nil {
			log.Printf("Scan error: %v", err)
			http.Error(w, "Scan error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(askStruct)
	})

	http.ListenAndServe(":3000", r)
}

type ask struct {
	Isni string `json:"isni"`
	Name string `json:"name"`
	Lai  string `json:"lai"`
}
