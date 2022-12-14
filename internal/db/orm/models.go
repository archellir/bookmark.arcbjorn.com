// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.16.0

package db

import (
	"database/sql"
	"time"
)

type Bookmark struct {
	ID int32 `json:"id"`
	// Title of the web page document
	Name      string        `json:"name"`
	Url       string        `json:"url"`
	GroupID   sql.NullInt32 `json:"group_id"`
	CreatedAt time.Time     `json:"created_at"`
}

type BookmarksTag struct {
	BookmarkID int32 `json:"bookmark_id"`
	TagID      int32 `json:"tag_id"`
}

type Group struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Tag struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type User struct {
	ID             int32     `json:"id"`
	Username       string    `json:"username"`
	HashedPassword string    `json:"hashed_password"`
	CreatedAt      time.Time `json:"created_at"`
}
