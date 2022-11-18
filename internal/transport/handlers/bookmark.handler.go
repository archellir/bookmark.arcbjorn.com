package transport

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	services "github.com/archellir/bookmark.arcbjorn.com/internal/services"
)

type BookmarkHandler struct {
	Service *services.BookmarkService
}

func NewBookmarkHandler(store *orm.Store) *BookmarkHandler {
	bookmarkService := &services.BookmarkService{
		Store: store,
	}
	bookmarkHandler := &BookmarkHandler{
		Service: bookmarkService,
	}

	return bookmarkHandler
}

func (handler *BookmarkHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/bm/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.List(w, r)

	case "/bm/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.GetOne(w, r)

	case "/bm/search":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.SearchByNameAndUrl(w, r)

	case "/bm/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Create(w, r)

	case "/bm/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Update(w, r)

	case "/bm/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Service.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
