package main

import (
	internal "github.com/archellir/bookmark.arcbjorn.com/api"
)

func main() {
	router := &internal.Server{}
	router.Start()
}
