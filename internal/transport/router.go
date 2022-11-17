package transport

import "net/http"

type Router struct {
	Handler BookmarkHandler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.WriteHeader(http.StatusOK)
	case "/all":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		router.Handler.GetAll(w, r)

	case "/get":
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		router.Handler.GetOne(w, r)

	case "/create":
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		router.Handler.Create(w, r)

	case "/update":
		if r.Method != "PUT" && r.Method != "PATCH" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		router.Handler.Update(w, r)

	case "/delete":
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		router.Handler.Delete(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
