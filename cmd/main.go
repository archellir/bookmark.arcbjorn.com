package main

import (
	"log"
	"os"

	"github.com/archellir/bookmark.arcbjorn.com/api"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"
)

func main() {
	// detect production environment
	var productionFlag string
	if len(os.Args) > 1 {
		productionFlag = os.Args[1]
	}

	config, err := utils.LoadConfig(".", productionFlag)
	if err != nil {
		log.Fatal("can not load config: ", err)
	}

	server, err := api.NewServer(config)
	if err != nil {
		log.Fatal("cannot create server", err)
	}

	server.Start()
}
