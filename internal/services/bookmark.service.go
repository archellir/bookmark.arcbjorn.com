package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type BookmarkService struct {
	Store *orm.Store
}

func (service *BookmarkService) List(w http.ResponseWriter, r *http.Request) error {
	limit := defaultLimit
	offset := defaultOffset

	if r.URL.Query().Has(limitParamName) {
		limitParam := r.URL.Query().Get(limitParamName)
		parsedInt, err := strconv.Atoi(limitParam)
		if err != nil {
			return fmt.Errorf("error parsing list limit")
		}
		limit = parsedInt
	}

	if r.URL.Query().Has(offsetParamName) {
		offsetParam := r.URL.Query().Get(offsetParamName)
		parsedInt, err := strconv.Atoi(offsetParam)
		if err != nil {
			return fmt.Errorf("error parsing list offset")
		}
		offset = parsedInt
	}

	args := &orm.ListBookmarksParams{
		Limit:  int32(limit),
		Offset: int32(offset),
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

func (service *BookmarkService) Create(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("created bookmark"))
}

func (service *BookmarkService) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated bookmark"))
}

func (service *BookmarkService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
