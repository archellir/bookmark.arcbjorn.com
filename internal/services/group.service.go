package services

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type GroupService struct {
	Store *orm.Store
}

func (service *GroupService) List(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all groups"))
}

func (service *GroupService) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one group"))
}

func (service *GroupService) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created group"))
}

func (service *GroupService) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated group"))
}

func (service *GroupService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
