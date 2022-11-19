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
	formattedBookmarks := make([]*tFormattedBookmark, 0)

	for _, bookmark := range bookmarks {
		formattedBookmarks = append(formattedBookmarks, FormatBookmark(bookmark))
	}

	return formattedBookmarks
}
