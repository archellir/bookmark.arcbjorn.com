package services

import "time"

type tResponse struct {
	Data  interface{} `json:"data"`
	Error interface{} `json:"error"`
}

type tUpdateBookmarkParams struct {
	ID      int32  `json:"id"`
	Name    string `json:"name"`
	Url     string `json:"url"`
	GroupID int32  `json:"group_id"`
}

type tFormattedBookmark struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Url       string    `json:"url"`
	GroupID   int32     `json:"group_id"`
	CreatedAt time.Time `json:"created_at"`
}
