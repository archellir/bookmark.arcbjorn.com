package internal

import "net/http"

type BookmarkHandler struct{}

func (handler *BookmarkHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all bookmarks"))
}
