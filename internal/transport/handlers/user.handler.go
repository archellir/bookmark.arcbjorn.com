package transport

import (
	"net/http"

	"github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	"github.com/archellir/bookmark.arcbjorn.com/internal/services"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(store *orm.Store, config *utils.Config, tokenMaker auth.IMaker) *UserHandler {
	userService := services.NewUserService(store, config, tokenMaker)
	userHandler := &UserHandler{
		Service: userService,
	}

	return userHandler
}

func (handler *UserHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/usr":

		switch r.Method {

		case "POST":
			handler.Service.Create(w, r)
			return

		case "PUT":
			handler.Service.UpdatePassword(w, r)
			return

		case "DELETE":
			handler.Service.Delete(w, r)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	case "/usr/login":
		handler.Service.LoginUser(w, r)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
