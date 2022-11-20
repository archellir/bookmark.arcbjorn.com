package transport

import (
	"net/http"
	"strings"

	"github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	handlers "github.com/archellir/bookmark.arcbjorn.com/internal/transport/handlers"
)

type Router struct {
	Bookmarks handlers.BookmarkHandler
	Tags      handlers.TagHandler
	Groups    handlers.GroupHandler
	Users     handlers.UserHandler
}

const (
	apiRoutePrefix    = "/api"
	healthCheckPrefix = "/api/healthcheck"
	bookmarkPrefix    = "/api/bm"
	tagPrefix         = "/api/tags"
	groupPrefix       = "/api/groups"
	userPrefix        = "/api/usr"
)

func NewRouter(store *orm.Store, config *utils.Config, tokenMaker auth.IMaker) *Router {
	router := &Router{
		Bookmarks: *handlers.NewBookmarkHandler(store),
		Tags:      *handlers.NewTagHandler(store),
		Groups:    *handlers.NewGroupHandler(store),
		Users:     *handlers.NewUserHandler(store, config, tokenMaker),
	}

	return router
}

func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == healthCheckPrefix:
		w.WriteHeader(http.StatusOK)

	case strings.HasPrefix(r.URL.Path, bookmarkPrefix):
		router.Bookmarks.Handle(w, r)
	case strings.HasPrefix(r.URL.Path, tagPrefix):
		router.Tags.Handle(w, r)
	case strings.HasPrefix(r.URL.Path, groupPrefix):
		router.Groups.Handle(w, r)
	case strings.HasPrefix(r.URL.Path, userPrefix):
		router.Users.Handle(w, r)

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
