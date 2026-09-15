package db

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/lib/pq"
)

func OpenDB(dsn string) *sql.DB {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Open DB error:", err)
	}
	for attempt := 1; attempt <= 10; attempt++ {
		err = db.Ping()
		if err == nil {
			return db
		}
		log.Printf("Postgres ping attempt %d/10 failed: %v", attempt, err)
		time.Sleep(time.Duration(attempt) * 2 * time.Second)
	}
	log.Fatal("Ping DB error:", err)
	return db
}
