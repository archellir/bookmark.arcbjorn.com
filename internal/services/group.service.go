package services

import (
	"context"
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type GroupService struct {
	Store *orm.Store
}

func (service *GroupService) List(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	limit, offset, err := GetListParams(r.URL)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleGroup, err)
		return
	}

	args := &orm.ListGroupsParams{
		Limit:  limit,
		Offset: offset,
	}

	groups, err := service.Store.Queries.ListGroups(context.Background(), *args)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleGroupsNotFound, err)
		return
	}

	if len(groups) == 0 {
		groups = []orm.Group{}
	}

	response.Data = groups
	ReturnJson(w, response)
}

func (service *GroupService) GetOne(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	id, err := GetIdFromUrlQuery(r.URL)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleGroup, err)
		return
	}

	var group orm.Group

	group, err = service.Store.Queries.GetGroupById(context.Background(), int32(id))
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleGroupNotFound, err)
		return
	}

	response.Data = group
	ReturnJson(w, response)
}

func (service *GroupService) Create(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var createGroupDTO tCreateGroupDTO
	err = GetJson(r, &createGroupDTO)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleGroupCreateDtoNotParsed, err)
		return
	}

	if createGroupDTO.Name == "" {
		ReturnResponseWithError(w, response, ErrorTitleGroupNoName, err)
		return
	}

	group, err := service.Store.Queries.CreateGroup(context.Background(), createGroupDTO.Name)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleGroupNotCreated, err)
		return
	}

	response.Data = group
	ReturnJson(w, response)
}

func (service *GroupService) Update(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("updated group"))
}

func (service *GroupService) Delete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("true"))
}
