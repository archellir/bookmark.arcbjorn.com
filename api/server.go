package api

import (
	"fmt"
	"log"
	"net/http"

	"github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	"github.com/archellir/bookmark.arcbjorn.com/internal/transport"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type Server struct {
	config     *utils.Config
	Router     *transport.Router
	tokenMaker auth.IMaker
}

func NewServer(config *utils.Config) (*Server, error) {
	store := orm.InitStore(config.DatabaseDriver, config.DatabaseSource)

	tokenMaker, err := auth.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	router := transport.NewRouter(store, config, tokenMaker)

	server := &Server{
		config:     config,
		Router:     router,
		tokenMaker: tokenMaker,
	}

	return server, nil
}

func (server *Server) Start() {
	log.Println("Listening and serving HTTP on", server.config.ServerAddress)
	log.Fatal(http.ListenAndServe(server.config.ServerAddress, server.Router))
}
