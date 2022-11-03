package internal

import (
	"net/http"

	internal "github.com/archellir/bookmark.arcbjorn.com/internal/transport"
)

type Server struct{}

func (s *Server) Start() {
	bookmarkHandler := &internal.BookmarkHandler{}
	router := &internal.Router{
		Handler: *bookmarkHandler,
	}

	// addr := fmt.Sprint("localhost:", os.Getenv("SERVER_PORT"))

	http.ListenAndServe("localhost:8080", router)
}
