package internal

import "net/http"

type Router struct {
	Handler BookmarkHandler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		w.WriteHeader(http.StatusOK)
	case "/all":
		router.Handler.GetAll(w, r)
	case "/get":
		router.Handler.GetOne(w, r)
	case "/create":
		router.Handler.Create(w, r)
	case "/update":
		router.Handler.Update(w, r)
	case "/delete":
		router.Handler.Delete(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
