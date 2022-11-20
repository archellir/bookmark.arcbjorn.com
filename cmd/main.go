package main

import (
	"log"

	api "github.com/archellir/bookmark.arcbjorn.com/api"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"
)

func main() {
	config, err := utils.LoadConfig(".")
	if err != nil {
		log.Fatal("can not load config: ", err)
	}

	server, err := api.NewServer(config.DatabaseDriver, config.DatabaseSource)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	server.Start(config.ServerAddress)
}
