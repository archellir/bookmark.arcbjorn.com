package services

import (
	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

func FormatBookmark(bookmark orm.Bookmark) *tFormattedBookmark {
	return &tFormattedBookmark{
		ID:        bookmark.ID,
		Name:      bookmark.Name,
		Url:       bookmark.Url,
		GroupID:   bookmark.GroupID.Int32,
		CreatedAt: bookmark.CreatedAt,
	}
}

func FormatBookmarks(bookmarks []orm.Bookmark) []*tFormattedBookmark {
	var formattedBookmarks []*tFormattedBookmark

	for idx, bookmark := range bookmarks {
		formattedBookmarks[idx] = FormatBookmark(bookmark)
	}

	return formattedBookmarks
}
