package transport

import (
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	services "github.com/archellir/bookmark.arcbjorn.com/internal/services"
)

type GroupHandler struct {
	Service *services.GroupService
}

func NewGroupHandler(store *orm.Store) *GroupHandler {
	groupService := &services.GroupService{
		Store: store,
	}
	groupHandler := &GroupHandler{
		Service: groupService,
	}

	return groupHandler
}

func (handler *GroupHandler) Handle(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {

	case "/api/groups":

		switch r.Method {

		case http.MethodGet:
			if r.URL.Query().Has(services.IdParam) {
				handler.Service.GetOne(w, r)
			} else {
				handler.Service.List(w, r)
			}
			return

		case http.MethodPost:
			handler.Service.Create(w, r)
			return

		case http.MethodPut:
			handler.Service.Update(w, r)
			return

		case http.MethodDelete:
			handler.Service.Delete(w, r)
			return

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}
