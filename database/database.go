package database

import (
	"database/sql"
	"fmt"
)

type Database interface {
	Ping() error
	TableInit() error
	Insert(query string, args ...any) error
	Query(query string, args ...any) (*sql.Rows, error)
	Connect() (*sql.Conn, error)
	FindOne(query string, args ...any) *sql.Row
}

func InitDb(databaseName string) (Database, error) {

	switch databaseName {
	case "postgres":
		return connectPostgres(), nil
	case "sqlite":
		return connectSqLite(), nil
	default:
		return nil, fmt.Errorf("unsupported database: %s", databaseName)

	}

}
