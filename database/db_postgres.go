package database

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type postgres struct {
	db *sql.DB
}

func connectPostgres() postgres {
	var databaseURL = os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("Missing Database URL ENV")
	}

	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		log.Fatalf("Invalid DATABASE_URL: %v", err)
	}

	query := parsedURL.Query()
	if query.Get("statement_cache_capacity") == "" {
		query.Set("statement_cache_capacity", "0")
	}

	mode := strings.ToLower(query.Get("default_query_exec_mode"))
	if mode == "" || mode == "cache_statement" {
		// Required when statement cache is disabled.
		query.Set("default_query_exec_mode", "exec")
	}
	parsedURL.RawQuery = query.Encode()

	postgresDb, err := sql.Open("pgx", parsedURL.String())
	if err != nil {
		log.Fatalf("Failed to open postgres driver: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := postgresDb.Ping(); err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	var currentDB, currentUser string
	if err := postgresDb.QueryRowContext(ctx, "SELECT current_database(), current_user").Scan(&currentDB, &currentUser); err != nil {
		slog.Warn("Connected to postgres but identity query failed", "error", err)
	} else {
		slog.Info("Postgres connection verified", "database", currentDB, "user", currentUser)
	}

	pg := &postgres{
		db: postgresDb,
	}

	return *pg
}

func (p postgres) Connect() (*sql.Conn, error) {
	return p.db.Conn(context.Background())
}

func (p postgres) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return p.db.PingContext(ctx)
}

func (p postgres) TableInit() error {
	result, err := p.db.Exec(`CREATE TABLE IF NOT EXISTS ask_checker (
		id BIGSERIAL PRIMARY KEY,
		isin TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		lei  TEXT NOT NULL
	)`)
	if err != nil {
		return err
	}

	// Backfill constraint for older tables created without UNIQUE on isin.
	if _, err = p.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_ask_checker_isin_unique ON ask_checker (isin)`); err != nil {
		return err
	}

	slog.Debug("TableInit result", "result", result)

	return err
}

func (p postgres) Insert(query string, args ...any) error {
	result, err := p.db.Exec(query, args...)
	if err != nil {
		slog.Error("Insert failed", "error", err, "query", query)
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	slog.Debug("Insert result", "rows_affected", rowsAffected)

	return nil
}

func (p postgres) Query(query string, args ...any) (*sql.Rows, error) {
	return p.db.Query(query, args...)
}

func (p postgres) FindOne(query string, args ...any) *sql.Row {
	return p.db.QueryRow(query, args...)
}
