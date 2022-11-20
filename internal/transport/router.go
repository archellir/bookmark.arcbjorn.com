package transport

import (
	"net/http"
	"regexp"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	handlers "github.com/archellir/bookmark.arcbjorn.com/internal/transport/handlers"
)

type Router struct {
	Bookmarks handlers.BookmarkHandler
	Tags      handlers.TagHandler
	Groups    handlers.GroupHandler
	Users     handlers.UserHandler
}

var (
	isHealthCheck = regexp.MustCompile(`^/$`)
	isBookmark    = regexp.MustCompile(`^/bm`)
	isTag         = regexp.MustCompile(`^/tags`)
	isGroup       = regexp.MustCompile(`^/groups`)
	isUser        = regexp.MustCompile(`^/usr`)
)

func NewRouter(store *orm.Store) *Router {
	router := &Router{
		Bookmarks: *handlers.NewBookmarkHandler(store),
		Tags:      *handlers.NewTagHandler(store),
		Groups:    *handlers.NewGroupHandler(store),
		Users:     *handlers.NewUserHandler(store),
	}

	return router
}

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
	case isUser.MatchString(r.URL.Path):
		router.Users.Handle(w, r)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
