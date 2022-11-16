package api

import (
	"log"
	"net/http"

	auth "github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	transport "github.com/archellir/bookmark.arcbjorn.com/internal/transport"
)

type Server struct {
	tokenMaker auth.IMaker
}

func (s *Server) Start() {
	tokenMaker, err := auth.NewPasetoMaker("12345678901234567890123456789012")
	if err != nil {
		log.Fatalf("cannot create token maker: %w", err)
	}

	bookmarkHandler := &transport.BookmarkHandler{}
	router := &transport.Router{
		Handler: *bookmarkHandler,
	}

	// addr := fmt.Sprint("localhost:", os.Getenv("SERVER_PORT"))

	http.ListenAndServe("localhost:8080", router)
}
