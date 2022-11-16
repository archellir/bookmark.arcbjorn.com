package transport

import "net/http"

type BookmarkHandler struct{}

func (handler *BookmarkHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("all bookmarks"))
}

func (handler *BookmarkHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("one bookmark"))
}

func (handler *BookmarkHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("created bookmark"))
}

func (handler *BookmarkHandler) Update(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" && r.Method != "PATCH" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("updated bookmark"))
}

func (handler *BookmarkHandler) Delete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte("true"))
}
