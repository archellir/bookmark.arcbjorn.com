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
	Http   *http.Server
	config *utils.Config
}

func NewServer(config *utils.Config) (*Server, error) {
	store := orm.InitStore(config.DatabaseDriver, config.DatabaseSource)

	tokenMaker, err := auth.NewPasetoMaker(config.TokenSymmetricKey)
	if err != nil {
		return nil, fmt.Errorf("cannot create token maker: %w", err)
	}

	router := transport.NewRouter(store, config, tokenMaker)

	httpServer := &http.Server{
		Addr:    config.ServerAddress,
		Handler: router,
	}

	server := &Server{
		Http:   httpServer,
		config: config,
	}

	return server, nil
}

func (server *Server) Start() {
	log.Println("Listening and serving HTTP on", server.config.ServerAddress)
	log.Fatal(server.Http.ListenAndServe())
}
