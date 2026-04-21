package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func OpenDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Open DB error:", err)
	}
	return db
}
