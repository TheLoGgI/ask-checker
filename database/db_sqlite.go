package database

import (
	"context"
	"database/sql"
	"log"
)

type sqlite struct {
	db *sql.DB
}

func connectSqLite() sqlite {
	sqliteDB, err := sql.Open("sqlite", "./ask.db")
	if err != nil {
		log.Fatal(err)
	}

	sqlite := &sqlite{
		db: sqliteDB,
	}

	return *sqlite

}

func (s sqlite) Ping() error {
	return s.db.Ping()
}

func (p sqlite) Connect() (*sql.Conn, error) {
	return p.db.Conn(context.Background())
}

func (s sqlite) TableInit() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS ask_checker (
		isin TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		lei  TEXT NOT NULL
	)`)

	return err
}

func (s sqlite) Insert(query string, args ...any) error {
	result, err := s.db.Exec(query, args...)

	// Debug Log
	lastId, _ := result.LastInsertId()
	log.Println("Inserted Row:", lastId)

	return err
}

func (s sqlite) Query(query string, args ...any) (*sql.Rows, error) {
	rows, err := s.db.Query(query, args...)

	return rows, err
}

func (p sqlite) FindOne(query string, args ...any) *sql.Row {
	return p.db.QueryRow(query, args...)
}
