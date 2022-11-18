package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	limitParamName  = "limit"
	offsetParamName = "offset"
	searchParam     = "search"
)

const (
	defaultLimit  int32 = 25
	defaultOffset int32 = 0
)

func GetListParams(url *url.URL) (limit int32, offset int32, err error) {
	limit = defaultLimit
	offset = defaultOffset

	if url.Query().Has(limitParamName) {
		limitParam := url.Query().Get(limitParamName)
		parsedInt, err := strconv.Atoi(limitParam)
		if err != nil {
			return 0, 0, fmt.Errorf("error parsing list limit")
		}
		limit = int32(parsedInt)
	}

	if url.Query().Has(offsetParamName) {
		offsetParam := url.Query().Get(offsetParamName)
		parsedInt, err := strconv.Atoi(offsetParam)
		if err != nil {
			return 0, 0, fmt.Errorf("error parsing list offset")
		}
		offset = int32(parsedInt)
	}

	return limit, offset, nil
}

func GetJson(r *http.Request, target interface{}) error {
	return json.NewDecoder(r.Body).Decode(target)
}
