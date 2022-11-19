package services

import (
	"database/sql"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

func Int32ToSqlNullInt32(n int32) *sql.NullInt32 {
	valid := true
	if n == 0 {
		valid = false
	}

	return &sql.NullInt32{
		Int32: n,
		Valid: valid,
	}
}

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
