package main

import (
	api "github.com/archellir/bookmark.arcbjorn.com/api"
)

func main() {
	server := api.NewServer()
	server.Start()
}
