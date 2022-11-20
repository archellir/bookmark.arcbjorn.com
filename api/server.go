package api

import (
	"fmt"
	"net/http"

	auth "github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	transport "github.com/archellir/bookmark.arcbjorn.com/internal/transport"
)

type Server struct {
	Router     *transport.Router
	tokenMaker auth.IMaker
}

func NewServer(dbDriver string, dbSource string) (*Server, error) {
	store := orm.InitStore(dbDriver, dbSource)
	router := transport.NewRouter(store)

	tokenMaker, err := auth.NewPasetoMaker("12345678901234567890123456789012")
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	server := &Server{
		Router:     router,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

func (server *Server) Start(serverAddress string) {
	// addr := fmt.Sprint("localhost:", os.Getenv("SERVER_PORT"))

	http.ListenAndServe(serverAddress, server.Router)
}
