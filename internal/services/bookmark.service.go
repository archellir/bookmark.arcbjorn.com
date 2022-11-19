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
		ReturnResponseWithError(w, response, ErrorTitleBookmark, err)
		return
	}

	args := &orm.ListBookmarksParams{
		Limit:  limit,
		Offset: offset,
	}

	bookmarks, err := service.Store.Queries.ListBookmarks(context.Background(), *args)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarksNotFound, err)
		return
	}

	if len(bookmarks) == 0 {
		bookmarks = []orm.Bookmark{}
	}

	response.Data = FormatBookmarks(bookmarks)
	ReturnJson(w, response)
}

func (service *BookmarkService) GetOne(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	id, err := GetIdFromUrlQuery(r.URL)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmark, err)
		return
	}

	var bookmark orm.Bookmark

	bookmark, err = service.Store.Queries.GetBookmarkById(context.Background(), int32(id))
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotFound, err)
		return
	}

	response.Data = FormatBookmark(bookmark)
	ReturnJson(w, response)
}

func (service *BookmarkService) Create(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var createBookmarkDTO orm.CreateBookmarkParams
	err = GetJson(r, &createBookmarkDTO)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkCreateDtoNotParsed, err)
		return
	}

	bookmark, err := service.Store.Queries.CreateBookmark(context.Background(), createBookmarkDTO)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotCreated, err)
		return
	}

	response.Data = FormatBookmark(bookmark)
	ReturnJson(w, response)
}

func (service *BookmarkService) SearchByNameAndUrl(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	searchString := r.URL.Query().Get(searchParam)
	bookmarks := []orm.Bookmark{}

	if searchString != "" {
		bookmarks, err = service.Store.Queries.SearchBookmarkByNameAndUrl(context.Background(), "%"+searchString+"%")
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleBookmarksNotFound, err)
			return
		}
	}

	response.Data = FormatBookmarks(bookmarks)
	ReturnJson(w, response)
}

func (service *BookmarkService) Update(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var updateBookmarkDTO tUpdateBookmarkParams
	err = GetJson(r, &updateBookmarkDTO)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkUpdateDtoNotParsed, err)
		return
	}

	if updateBookmarkDTO.ID == 0 {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNoId, err)
		return
	}

	var bookmark orm.Bookmark

	_, err = service.Store.Queries.GetBookmarkById(context.Background(), updateBookmarkDTO.ID)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotFound, err)
		return
	}

	if updateBookmarkDTO.Name != "" {
		nameDto := &orm.UpdateBookmarkNameParams{
			ID:   updateBookmarkDTO.ID,
			Name: updateBookmarkDTO.Name,
		}

		bookmark, err = service.Store.Queries.UpdateBookmarkName(context.Background(), *nameDto)
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleBookmarkNameNotUpdated, err)
			return
		}
	}

	if updateBookmarkDTO.Url != "" {
		nameDto := &orm.UpdateBookmarkUrlParams{
			ID:  updateBookmarkDTO.ID,
			Url: updateBookmarkDTO.Url,
		}

		bookmark, err = service.Store.Queries.UpdateBookmarkUrl(context.Background(), *nameDto)
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleBookmarkUrlNotUpdated, err)
			return
		}
	}

	if updateBookmarkDTO.GroupID != 0 {
		_, err = service.Store.Queries.GetGroupById(context.Background(), updateBookmarkDTO.GroupID)
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleGroupNotFound, err)
			return
		}

		groupDto := &orm.UpdateBookmarkGroupIdParams{
			ID:      updateBookmarkDTO.ID,
			GroupID: *Int32ToSqlNullInt32(updateBookmarkDTO.GroupID),
		}

		bookmark, err = service.Store.Queries.UpdateBookmarkGroupId(context.Background(), *groupDto)
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleBookmarkGroupIdNotUpdated, err)
			return
		}
	}

	response.Data = FormatBookmark(bookmark)
	ReturnJson(w, response)
}

func (service *BookmarkService) Delete(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	id, err := GetIdFromUrlQuery(r.URL)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmark, err)
		return
	}

	err = service.Store.Queries.DeleteBookmark(context.Background(), int32(id))
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotDeleted, err)
		return
	}

	response.Data = true
	ReturnJson(w, response)
}
