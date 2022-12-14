package services

import (
	"context"
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type BookmarkService struct {
	Store       *orm.Store
	LinkService *LinkService
}

func (service *BookmarkService) List(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var bookmarks []orm.Bookmark
	var err error

	limit, offset, searchString, err := GetListParams(r.URL)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmark, err)
		return
	}

	if searchString != "" {
		args := &orm.SearchBookmarkByNameAndUrlParams{
			Limit:        limit,
			Offset:       offset,
			SearchString: "%" + searchString + "%",
		}

		bookmarks, err = service.Store.Queries.SearchBookmarkByNameAndUrl(context.Background(), *args)
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleBookmarksNotFound, err)
			return
		}
	} else {
		args := &orm.ListBookmarksParams{
			Limit:  limit,
			Offset: offset,
		}
		bookmarks, err = service.Store.Queries.ListBookmarks(context.Background(), *args)
		if err != nil {
			ReturnResponseWithError(w, response, ErrorTitleBookmarksNotFound, err)
			return
		}
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
	var isValid bool

	var createBookmarkDTO orm.CreateBookmarkParams
	err = GetJson(r, &createBookmarkDTO)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkCreateDtoNotParsed, err)
		return
	}

	if createBookmarkDTO.Url == "" {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNoUrl, err)
		return
	}

	if createBookmarkDTO.Name == "" {
		isValid, title, err := service.LinkService.ProcessLink(createBookmarkDTO.Url)
		if !isValid {
			ReturnResponseWithError(w, response, ErrorTitleBookmark, err)
			return
		}

		createBookmarkDTO.Name = title
	} else {
		isValid, err = service.LinkService.ValidateLink(createBookmarkDTO.Url)
		if !isValid {
			ReturnResponseWithError(w, response, ErrorTitleBookmark, err)
			return
		}
	}

	bookmark, err := service.Store.Queries.CreateBookmark(context.Background(), createBookmarkDTO)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotCreated, err)
		return
	}

	response.Data = FormatBookmark(bookmark)
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

	idInt := int32(id)

	_, err = service.Store.Queries.GetBookmarkById(context.Background(), idInt)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotFound, err)
		return
	}

	err = service.Store.Queries.DeleteBookmark(context.Background(), idInt)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleBookmarkNotDeleted, err)
		return
	}

	response.Data = true
	ReturnJson(w, response)
}
