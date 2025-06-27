// db/db.go
package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Connect(dbPath ...string) {
	var err error
	path := "./db/loan_service.db" // default
	if len(dbPath) > 0 {
		path = dbPath[0]
	}

	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal("DB open error:", err)
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
}
