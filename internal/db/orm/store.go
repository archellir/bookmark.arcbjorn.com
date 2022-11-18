package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

const (
	dbDriver = "postgres"
	dbSource = "postgresql://root:root@localhost:5435/arc_bookmark?sslmode=disable"
)

type Store struct {
	Queries *Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
	}
}

func InitStore() *Store {
	db, dbErr := sql.Open(dbDriver, dbSource)

	if dbErr != nil {
		log.Fatal("cannot connect to db:", dbErr)
	}

	store := NewStore(db)

	return store
}
