package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Store struct {
	Queries *Queries
}

func NewStore(db *sql.DB) *Store {
	return &Store{
		Queries: New(db),
	}
}

func InitStore(dbDriver string, dbSource string) *Store {
	db, dbErr := sql.Open(dbDriver, dbSource)

	if dbErr != nil {
		log.Fatal("cannot connect to db:", dbErr)
	}

	store := NewStore(db)

	return store
}
