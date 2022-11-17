package transport

import (
	"net/http"
)

type BookmarkHandler struct{}

func (handler *BookmarkHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/bm/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.GetAll(w, r)

	case "/bm/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.GetOne(w, r)

	case "/bm/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Create(w, r)

	case "/bm/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Update(w, r)

	case "/bm/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (handler *BookmarkHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all bookmarks"))
}

func (handler *BookmarkHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one bookmark"))
}

func (handler *BookmarkHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created bookmark"))
}

func (handler *BookmarkHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated bookmark"))
}

func (handler *BookmarkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
