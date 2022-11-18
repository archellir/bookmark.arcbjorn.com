package transport

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type TagHandler struct {
	Store *orm.Store
}

func (handler *TagHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/tags/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.GetAll(w, r)

	case "/tags/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.GetOne(w, r)

	case "/tags/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Create(w, r)

	case "/tags/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Update(w, r)

	case "/tags/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (handler *TagHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all tags"))
}

func (handler *TagHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one tag"))
}

func (handler *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created tag"))
}

func (handler *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated tag"))
}

func (handler *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
