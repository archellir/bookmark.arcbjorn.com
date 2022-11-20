package main

import (
	"log"

	api "github.com/archellir/bookmark.arcbjorn.com/api"
)

const (
	databaseDriver = "postgres"
	databaseSource = "postgresql://root:root@localhost:5435/arc_bookmark?sslmode=disable"
	serverAddress  = ":8080"
)

func main() {
	server, err := api.NewServer(databaseDriver, databaseSource)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	server.Start(serverAddress)
}
