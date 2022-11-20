package transport

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	services "github.com/archellir/bookmark.arcbjorn.com/internal/services"
)

type UserHandler struct {
	Service *services.UserService
}

func NewUserHandler(store *orm.Store) *UserHandler {
	tagService := &services.UserService{
		Store: store,
	}
	tagHandler := &UserHandler{
		Service: tagService,
	}

	return tagHandler
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

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
