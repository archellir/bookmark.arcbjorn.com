package services

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type BookmarkService struct {
	Store *orm.Store
}

func (service *BookmarkService) List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all bookmarks"))
}

func (service *BookmarkService) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one bookmark"))
}

func (service *BookmarkService) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created bookmark"))
}

func (service *BookmarkService) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated bookmark"))
}

func (service *BookmarkService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
