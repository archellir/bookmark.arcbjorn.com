package services

import (
	"context"
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type BookmarkService struct {
	Store *orm.Store
}

func (service *BookmarkService) List(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	limit, offset, err := GetListParams(r.URL)
	if err != nil {
		response.Error = err.Error()
	}

	args := &orm.ListBookmarksParams{
		Limit:  limit,
		Offset: offset,
	}

	bookmarks, err := service.Store.Queries.ListBookmarks(context.Background(), *args)
	if err != nil {
		response.Error = "can not retrieve bookmarks: " + err.Error()
	}

	if len(bookmarks) == 0 {
		bookmarks = []orm.Bookmark{}
	}

	if response.Error == nil {
		response.Data = FormatBookmarks(bookmarks)
	}

	ReturnJson(response, w)
}

func (service *BookmarkService) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one bookmark"))
}

func (service *BookmarkService) Create(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var createBookmarkDTO orm.CreateBookmarkParams
	err = GetJson(r, &createBookmarkDTO)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Error = "can not parse createBookmarkDTO: " + err.Error()
	}

	bookmark, err := service.Store.Queries.CreateBookmark(context.Background(), createBookmarkDTO)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Error = "can not create bookmark: " + err.Error()
	}

	if response.Error == nil {
		response.Data = FormatBookmark(bookmark)
	}

	ReturnJson(response, w)
}

func (service *BookmarkService) SearchByNameAndUrl(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	searchString := r.URL.Query().Get(searchParam)
	bookmarks := []orm.Bookmark{}

	if searchString != "" {
		bookmarks, err = service.Store.Queries.SearchBookmarkByNameAndUrl(context.Background(), "%"+searchString+"%")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Error = "can not find bookmarks: " + err.Error()
		}
	}

	if response.Error == nil {
		response.Data = FormatBookmarks(bookmarks)
	}

	ReturnJson(response, w)
}

func (service *BookmarkService) Update(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var updateBookmarkDTO tUpdateBookmarkParams
	err = GetJson(r, &updateBookmarkDTO)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		response.Error = "can not parse updateBookmarkDTO: " + err.Error()
	}

	if updateBookmarkDTO.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		response.Error = "can get bookmark ID: " + err.Error()
	}

	var bookmark orm.Bookmark

	_, err = service.Store.Queries.GetBookmarkById(context.Background(), updateBookmarkDTO.ID)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		response.Error = "can not find bookmark to update: " + err.Error()
	}

	if updateBookmarkDTO.Name != "" {
		nameDto := &orm.UpdateBookmarkNameParams{
			ID:   updateBookmarkDTO.ID,
			Name: updateBookmarkDTO.Name,
		}

		bookmark, err = service.Store.Queries.UpdateBookmarkName(context.Background(), *nameDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Error = "can not update bookmark name: " + err.Error()
		}
	}

	if updateBookmarkDTO.Url != "" {
		nameDto := &orm.UpdateBookmarkUrlParams{
			ID:  updateBookmarkDTO.ID,
			Url: updateBookmarkDTO.Url,
		}

		bookmark, err = service.Store.Queries.UpdateBookmarkUrl(context.Background(), *nameDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Error = "can not update bookmark url: " + err.Error()
		}
	}

	if updateBookmarkDTO.GroupID != 0 {
		_, err = service.Store.Queries.GetGroupById(context.Background(), updateBookmarkDTO.GroupID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Error = "can not find group: " + err.Error()
		}

		groupDto := &orm.UpdateBookmarkGroupIdParams{
			ID:      updateBookmarkDTO.ID,
			GroupID: *Int32ToSqlNullInt32(updateBookmarkDTO.GroupID),
		}

		bookmark, err = service.Store.Queries.UpdateBookmarkGroupId(context.Background(), *groupDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			response.Error = "can not update bookmark group: " + err.Error()
		}
	}

	if response.Error == nil {
		response.Data = FormatBookmark(bookmark)
	}

	ReturnJson(response, w)
}

func (service *BookmarkService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
