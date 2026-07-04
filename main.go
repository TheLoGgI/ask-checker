package main

import (
	"ask-checker/database"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "modernc.org/sqlite"
)

// func extractData(db *sql.DB) {
// 	var datasetURL = "https://skat.dk/media/r1dn5su0/maj-2026-abis-liste-til-offentliggoerelse-2021-2026.xlsx"

// }

func main() {
	databaseType := Cfg.Database
	db, err := database.InitDb(databaseType)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	if Cfg.Port == "" {
		log.Fatalf("Port Requried to be configured in ENV")
	}

	// if err = db.TableInit(); err != nil {
	// 	log.Fatalf("Database could not create table: %v", err)
	// }

	fmt.Println("Database ready")
	fmt.Println("ASK Checker Running!")

	// if databaseType == "postgres" {
	// 	migrate.MigratePG_csv(db)
	// } else {
	// 	migrate.MigrateSQLite_csv(db)
	// }

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	// Liveness endpoint for load balancers and container orchestration probes.
	r.Get("/livez", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Readiness endpoint ensures dependencies are available before serving traffic.
	r.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			http.Error(w, "not ready", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Get("/search", func(w http.ResponseWriter, r *http.Request) {

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
		var resultRow ask
		selectSQL := "SELECT isin, name FROM ask_checker WHERE isin = ?"
		if databaseType == "postgres" {
			selectSQL = "SELECT isin, name FROM ask_checker WHERE isin = $1"
		}

		err := db.FindOne(selectSQL, isin).Scan(&resultRow.Isni, &resultRow.Name)
		if err != nil {
			log.Printf("Error: %v", err)
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Not Found", http.StatusNotFound)
			} else {
				http.Error(w, "Database error", http.StatusInternalServerError)
			}
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resultRow)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = Cfg.Port
	}

	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("HTTP server failed: %v", err)
	}
}

type ask struct {
	Isni string `json:"isni"`
	Name string `json:"name"`
	Lai  string `json:"lai"`
}
