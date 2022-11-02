package internal

import "net/http"

type Router struct {
	Handler BookmarkHandler
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/all":
		router.Handler.GetAll(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
