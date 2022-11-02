package internal

import (
	"net/http"

	"github.com/archellir/bookmark.arcbjorn.com/internal"
)

type Server struct{}

func (s *Server) Start() {
	bookmarkHandler := &internal.BookmarkHandler{}
	router := &internal.Router{
		Handler: *bookmarkHandler,
	}

	http.ListenAndServe("localhost:8080", router)
}
