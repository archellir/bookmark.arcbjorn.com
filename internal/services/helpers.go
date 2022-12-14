package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	IdParam         = "id"
	searchParam     = "search"
	limitParamName  = "limit"
	offsetParamName = "offset"
)

const (
	defaultLimit  int32 = 25
	defaultOffset int32 = 0
)

const (
	ErrorTitleGroup                   string = "group: "
	ErrorTitleGroupNotFound           string = "can not find group: "
	ErrorTitleGroupsNotFound          string = "can not find groups: "
	ErrorTitleGroupNotCreated         string = "can not create group: "
	ErrorTitleGroupNoName             string = "can not get group name: "
	ErrorTitleGroupNoId               string = "can not get group ID: "
	ErrorTitleGroupCreateDtoNotParsed string = "can not parse createGroupDTO: "
	ErrorTitleGroupNameNotUpdated     string = "can not update group name: "
	ErrorTitleGroupUpdateDtoNotParsed string = "can not parse updateGroupDTO: "
	ErrorTitleGroupNotDeleted         string = "can not delete group: "
)

const (
	ErrorTitleUser                   string = "user: "
	ErrorTitleUserNotFound           string = "can not find user: "
	ErrorTitleUserNotCreated         string = "can not create user: "
	ErrorTitleUserNoUsername         string = "can not get user username: "
	ErrorTitleUserNoPassword         string = "can not get user password: "
	ErrorTitleUserNoId               string = "can not get user ID: "
	ErrorTitleUserDtoNotParsed       string = "can not parse userDTO: "
	ErrorTitleUserPasswordNotUpdated string = "can not update user password: "
	ErrorTitleUserNotDeleted         string = "can not delete user: "
	ErrorTitleUserWrongPassword      string = "wrong password: "
	ErrorTitleUserAccessTokenNotMade string = "can not generate access token: "
)

const (
	ErrorTitleBookmark                   string = "bookmark: "
	ErrorTitleBookmarkNoId               string = "can not get bookmark ID: "
	ErrorTitleBookmarkCreateDtoNotParsed string = "can not parse createBookmarkDTO: "
	ErrorTitleBookmarkNotCreated         string = "can not create bookmark: "
	ErrorTitleBookmarkNoUrl              string = "can not get bookmark url: "
	ErrorTitleBookmarkNotFound           string = "can not find bookmark: "
	ErrorTitleBookmarksNotFound          string = "can not find bookmarks: "
	ErrorTitleBookmarkNotDeleted         string = "can not delete bookmark: "
	ErrorTitleBookmarkUpdateDtoNotParsed string = "can not parse updateBookmarkDTO: "
	ErrorTitleBookmarkNameNotUpdated     string = "can not update bookmark name: "
	ErrorTitleBookmarkUrlNotUpdated      string = "can not update bookmark url: "
	ErrorTitleBookmarkGroupIdNotUpdated  string = "can not update bookmark group: "
	ErrorTitleUrlNotStaticallyValid      string = "url is statically not valid"
	ErrorTitleUrlNotValid                string = "can not validate url: "
)

func GetListParams(url *url.URL) (limit int32, offset int32, searchString string, err error) {
	limit = defaultLimit
	offset = defaultOffset
	searchString = ""

	if url.Query().Has(limitParamName) {
		limitParam := url.Query().Get(limitParamName)
		parsedInt, err := strconv.Atoi(limitParam)
		if err != nil {
			return 0, 0, "", fmt.Errorf("error parsing list limit")
		}
		limit = int32(parsedInt)
	}

	if url.Query().Has(offsetParamName) {
		offsetParam := url.Query().Get(offsetParamName)
		parsedInt, err := strconv.Atoi(offsetParam)
		if err != nil {
			return 0, 0, "", fmt.Errorf("error parsing list offset")
		}
		offset = int32(parsedInt)
	}

	if url.Query().Has(searchParam) {
		searchString = url.Query().Get(searchParam)
	}

	return limit, offset, searchString, nil
}

func GetJson(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}

func ReturnJson(w http.ResponseWriter, data interface{}) {
	json, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can not generate json" + err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func CreateResponse(data interface{}, err interface{}) *tResponse {
	return &tResponse{
		Data:  data,
		Error: err,
	}
}

func GetIdFromUrlQuery(url *url.URL) (id int32, err error) {
	if !url.Query().Has(IdParam) {
		return 0, fmt.Errorf("ID is not provided")
	}

	idStr := url.Query().Get(IdParam)

	idInt64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("ID is not valid: " + err.Error())
	}

	return int32(idInt64), nil
}

func ReturnResponseWithError(w http.ResponseWriter, response *tResponse, errorTitle string, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	response.Error = errorTitle + err.Error()

	ReturnJson(w, response)
}
