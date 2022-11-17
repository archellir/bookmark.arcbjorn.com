package transport

import (
	"net/http"
	"regexp"

	handlers "github.com/archellir/bookmark.arcbjorn.com/internal/transport/handlers"
)

type Router struct {
	Bookmarks handlers.BookmarkHandler
	Tags      handlers.TagHandler
	Groups    handlers.GroupHandler
}

var (
	isHealthCheck = regexp.MustCompile(`^/$`)
	isBookmark    = regexp.MustCompile(`^/bm`)
	isTag         = regexp.MustCompile(`^/tags`)
	isGroup       = regexp.MustCompile(`^/groups`)
)

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case isHealthCheck.MatchString(r.URL.Path):
		w.WriteHeader(http.StatusOK)
	case isBookmark.MatchString(r.URL.Path):
		router.Bookmarks.Handle(w, r)
	case isTag.MatchString(r.URL.Path):
		router.Tags.Handle(w, r)
	case isGroup.MatchString(r.URL.Path):
		router.Groups.Handle(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
