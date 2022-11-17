package transport

import (
	"net/http"
	"regexp"

	handlers "github.com/archellir/bookmark.arcbjorn.com/internal/transport/handlers"
)

type Router struct {
	Bookmarks handlers.BookmarkHandler
}

var (
	isBookmark    = regexp.MustCompile(`^/bm`)
	isHealthCheck = regexp.MustCompile(`^/$`)
)

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case isHealthCheck.MatchString(r.URL.Path):
		w.WriteHeader(http.StatusOK)
	case isBookmark.MatchString(r.URL.Path):
		router.Bookmarks.Handle(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
