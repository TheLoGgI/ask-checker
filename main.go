package main

import (
	"ask-checker/database"
	structs "ask-checker/stucts"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
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

	// migrate_ticker(db)

	fmt.Println("Database ready")
	fmt.Println("ASK Checker Running!")

	// if databaseType == "postgres" {
	// 	migrate.MigratePG_csv(db)
	// } else {
	// 	migrate.MigrateSQLite_csv(db)
	// }

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://www.nordnet.dk", "http://nordet.dk", "https://nordnet.dk"},
		AllowedMethods: []string{"GET", "OPTIONS"},
	}))

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
		var ticker = query.Get("ticker")

		log.Printf("Searching for ISIN: '%s' Ticker: '%s'", isin, ticker)
		if isin == "" && ticker == "" {
			http.Error(w, "Missing isin or ticker parameter", http.StatusBadRequest)
			return
		}

		var resultRow structs.Ask
		selectSQL := "SELECT isin, name, ticker FROM ask_checker WHERE isin = ?"
		args := []any{isin}

		if isin == "" {
			selectSQL = "SELECT isin, name, ticker FROM ask_checker WHERE ticker = ?"
			args = []any{ticker}
		} else if ticker != "" {
			selectSQL = "SELECT isin, name, ticker FROM ask_checker WHERE isin = ? OR ticker = ?"
			args = []any{isin, ticker}
		}

		if databaseType == "postgres" {
			selectSQL = "SELECT isin, name, ticker FROM ask_checker WHERE isin = $1"
			if isin == "" {
				selectSQL = "SELECT isin, name, ticker FROM ask_checker WHERE ticker = $1"
				args = []any{ticker}
			} else if ticker != "" {
				selectSQL = "SELECT isin, name, ticker FROM ask_checker WHERE isin = $1 OR ticker = $2"
				args = []any{isin, ticker}
			}
		}

		err := db.FindOne(selectSQL, args...).Scan(&resultRow.Isin, &resultRow.Name, &resultRow.Ticker)
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
