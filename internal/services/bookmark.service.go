package services

import (
	"context"
	"encoding/json"
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

	result, err := service.Store.Queries.ListBookmarks(context.Background(), *args)
	if err != nil {
		return fmt.Errorf("can not retrieve bookmarks: %w", err)
	}

	if len(result) == 0 {
		result = []orm.Bookmark{}
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("can not generate json: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

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

	err = service.returnJson(result, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (service *BookmarkService) SearchByNameAndUrl(w http.ResponseWriter, r *http.Request) error {
	searchString := r.URL.Query().Get(searchParam)
	result, err := service.Store.Queries.SearchBookmarkByNameAndUrl(context.Background(), "%"+searchString+"%")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return fmt.Errorf("can not create bookmark: %w", err)
	}

	err = service.returnJson(result, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}

func (service *BookmarkService) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated bookmark"))
}

func (service *BookmarkService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}

func (service *BookmarkService) returnJson(data interface{}, w http.ResponseWriter) error {
	json, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("can not generate json: %w", err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)

	return nil
}
