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

	case "/api/usr":

		switch r.Method {

		case http.MethodPost:
			handler.Service.Create(w, r)
			return

		case http.MethodPut:
			handler.Service.UpdatePassword(w, r)
			return

		case http.MethodDelete:
			handler.Service.Delete(w, r)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	case "/usr/login":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		handler.Service.LoginUser(w, r)
		return

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
