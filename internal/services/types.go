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

type tCreateGroupDTO struct {
	Name string `json:"name"`
}

type tUpdateGroupParams struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type tUserDTO struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type tLoginUserResponse struct {
	AccessToken string `json:"access_token"`
	User        string `json:"username"`
}
