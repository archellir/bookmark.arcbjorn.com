package api

import (
	"fmt"
	"net/http"

	auth "github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	transport "github.com/archellir/bookmark.arcbjorn.com/internal/transport"
	handlers "github.com/archellir/bookmark.arcbjorn.com/internal/transport/handlers"
)

type Server struct {
	Router     *transport.Router
	tokenMaker auth.IMaker
}

func NewServer() (*Server, error) {
	tokenMaker, err := auth.NewPasetoMaker("12345678901234567890123456789012")
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	bookmarkHandler := &handlers.BookmarkHandler{}
	tagsHandler := &handlers.TagHandler{}
	router := &transport.Router{
		Bookmarks: *bookmarkHandler,
		Tags:      *tagsHandler,
	}

	server := &Server{
		Router:     router,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

func (server *Server) Start() {
	// addr := fmt.Sprint("localhost:", os.Getenv("SERVER_PORT"))

	http.ListenAndServe("localhost:8080", server.Router)
}
