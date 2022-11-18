package transport

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	services "github.com/archellir/bookmark.arcbjorn.com/internal/services"
)

type TagHandler struct {
	Service *services.TagService
}

func NewTagHandler(store *orm.Store) *TagHandler {
	tagService := &services.TagService{
		Store: store,
	}
	tagHandler := &TagHandler{
		Service: tagService,
	}

	return tagHandler
}

func (handler *TagHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/tags/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.List(w, r)

	case "/tags/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.GetOne(w, r)

	case "/tags/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Create(w, r)

	case "/tags/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Update(w, r)

	case "/tags/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
