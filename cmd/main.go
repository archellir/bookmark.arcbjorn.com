package main

import (
	"log"

	api "github.com/archellir/bookmark.arcbjorn.com/api"
)

func main() {
	server, err := api.NewServer()
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	server.Start()
}
