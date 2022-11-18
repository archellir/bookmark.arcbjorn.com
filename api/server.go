package api

import (
	"fmt"
	"net/http"

	auth "github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	transport "github.com/archellir/bookmark.arcbjorn.com/internal/transport"
	handlers "github.com/archellir/bookmark.arcbjorn.com/internal/transport/handlers"
)

type Server struct {
	Router     *transport.Router
	Store      *orm.Store
	tokenMaker auth.IMaker
}

func NewServer() (*Server, error) {
	store := orm.InitStore()

	tokenMaker, err := auth.NewPasetoMaker("12345678901234567890123456789012")
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	bookmarkHandler := &handlers.BookmarkHandler{}
	tagsHandler := &handlers.TagHandler{}
	groupsHandler := &handlers.GroupHandler{}

	router := &transport.Router{
		Bookmarks: *bookmarkHandler,
		Tags:      *tagsHandler,
		Groups:    *groupsHandler,
	}

	server := &Server{
		Router:     router,
		Store:      store,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

func (server *Server) Start() {
	// addr := fmt.Sprint("localhost:", os.Getenv("SERVER_PORT"))

	http.ListenAndServe("localhost:8080", server.Router)
}
