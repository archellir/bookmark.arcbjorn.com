package transport

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	services "github.com/archellir/bookmark.arcbjorn.com/internal/services"
)

type GroupHandler struct {
	Service *services.GroupService
}

func NewGroupHandler(store *orm.Store) *GroupHandler {
	groupService := &services.GroupService{
		Store: store,
	}
	groupHandler := &GroupHandler{
		Service: groupService,
	}

	return groupHandler
}

func (handler *GroupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/groups/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.List(w, r)

	case "/groups/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.GetOne(w, r)

	case "/groups/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Create(w, r)

	case "/groups/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Update(w, r)

	case "/groups/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
