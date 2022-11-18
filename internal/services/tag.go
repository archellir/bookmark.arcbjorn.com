package services

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type TagService struct {
	Store *orm.Store
}

func (service *TagService) List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all tags"))
}

func (service *TagService) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one tag"))
}

func (service *TagService) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created tag"))
}

func (service *TagService) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated tag"))
}

func (service *TagService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
