package services

import (
	"context"
	"fmt"
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type BookmarkService struct {
	Store *orm.Store
}

func (service *BookmarkService) List(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := GetListParams(r.URL)
	if err != nil {
		return err
	}

	args := &orm.ListBookmarksParams{
		Limit:  limit,
		Offset: offset,
	}

	bookmarks, err := service.Store.Queries.ListBookmarks(context.Background(), *args)
	if err != nil {
		return fmt.Errorf("can not retrieve bookmarks: %w", err)
	}

	if len(bookmarks) == 0 {
		bookmarks = []orm.Bookmark{}
	}

	formattedBookmarks := FormatBookmarks(bookmarks)

	err = ReturnJson(formattedBookmarks, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (service *BookmarkService) GetOne(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("one bookmark"))
}

func (service *BookmarkService) Create(w http.ResponseWriter, r *http.Request) error {
	var createBookmarkDTO orm.CreateBookmarkParams
	err := GetJson(r, &createBookmarkDTO)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("can not parse createBookmarkDTO: %w", err)
	}

	result, err := service.Store.Queries.CreateBookmark(context.Background(), createBookmarkDTO)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("can not create bookmark: %w", err)
	}

	err = ReturnJson(result, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (service *BookmarkService) SearchByNameAndUrl(w http.ResponseWriter, r *http.Request) error {
	searchString := r.URL.Query().Get(searchParam)
	bookmarks := []orm.Bookmark{}
	var err error

	if searchString != "" {
		bookmarks, err = service.Store.Queries.SearchBookmarkByNameAndUrl(context.Background(), "%"+searchString+"%")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("can not create bookmark: %w", err)
		}
	}

	formattedBookmarks := FormatBookmarks(bookmarks)

	err = ReturnJson(formattedBookmarks, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (service *BookmarkService) Update(w http.ResponseWriter, r *http.Request) error {
	var updateBookmarkDTO tUpdateBookmarkParams
	err := GetJson(r, &updateBookmarkDTO)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("can not parse updateBookmarkDTO: %w", err)
	}

	if updateBookmarkDTO.ID == 0 {
		w.WriteHeader(http.StatusNotFound)
		return nil
	}

	if updateBookmarkDTO.Name != "" {
		nameDto := &orm.UpdateBookmarkNameParams{
			ID:   updateBookmarkDTO.ID,
			Name: updateBookmarkDTO.Name,
		}

		_, err := service.Store.Queries.UpdateBookmarkName(context.Background(), *nameDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("can not update bookmark name: %w", err)
		}
	}

	if updateBookmarkDTO.Url != "" {
		nameDto := &orm.UpdateBookmarkUrlParams{
			ID:  updateBookmarkDTO.ID,
			Url: updateBookmarkDTO.Url,
		}

		_, err := service.Store.Queries.UpdateBookmarkUrl(context.Background(), *nameDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("can not update bookmark url: %w", err)
		}
	}

	if updateBookmarkDTO.GroupID != 0 {
		urlDto := &orm.UpdateBookmarkUrlParams{
			ID:  updateBookmarkDTO.ID,
			Url: updateBookmarkDTO.Url,
		}

		_, err := service.Store.Queries.UpdateBookmarkUrl(context.Background(), *urlDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("can not update bookmark url: %w", err)
		}
	}

	if updateBookmarkDTO.GroupID != 0 {
		groupDto := &orm.UpdateBookmarkGroupIdParams{
			ID:      updateBookmarkDTO.ID,
			GroupID: *Int32ToSqlNullInt32(updateBookmarkDTO.GroupID),
		}

		_, err := service.Store.Queries.UpdateBookmarkGroupId(context.Background(), *groupDto)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return fmt.Errorf("can not update bookmark url: %w", err)
		}
	}

	ReturnJson(updateBookmarkDTO, w)

	return nil
}

func (service *BookmarkService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
