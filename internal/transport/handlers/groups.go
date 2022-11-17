package transport

import (
	"net/http"
)

type GroupHandler struct{}

func (handler *GroupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/groups/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.GetAll(w, r)

	case "/groups/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.GetOne(w, r)

	case "/groups/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Create(w, r)

	case "/groups/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Update(w, r)

	case "/groups/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		handler.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (handler *GroupHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("all groups"))
}

func (handler *GroupHandler) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one groups"))
}

func (handler *GroupHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created groups"))
}

func (handler *GroupHandler) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated groups"))
}

func (handler *GroupHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
