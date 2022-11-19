package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

const (
	idParam         = "id"
	searchParam     = "search"
	limitParamName  = "limit"
	offsetParamName = "offset"
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

func ReturnJson(data interface{}, w http.ResponseWriter) {
	json, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("can not generate json" + err.Error()))
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
	if !url.Query().Has(idParam) {
		return 0, fmt.Errorf("ID is not provided")
	}

	idStr := url.Query().Get(idParam)

	idInt64, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("ID is not valid: " + err.Error())
	}

	return int32(idInt64), nil
}
