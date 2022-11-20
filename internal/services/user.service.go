package services

import (
	"context"
	"net/http"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"
)

type UserService struct {
	Store *orm.Store
}

func (service *UserService) Create(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var userDto tUserDTO
	err = GetJson(r, &userDto)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserDtoNotParsed, err)
		return
	}

	if userDto.Username == "" {
		ReturnResponseWithError(w, response, ErrorTitleUserNoUsername, err)
		return
	}

	if userDto.Password == "" {
		ReturnResponseWithError(w, response, ErrorTitleUserNoPassword, err)
		return
	}

	hashedPassword, err := utils.HashPassword(userDto.Password)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUser, err)
		return
	}

	createUserParams := &orm.CreateUserParams{
		Username:       userDto.Username,
		HashedPassword: hashedPassword,
	}

	user, err := service.Store.Queries.CreateUser(context.Background(), *createUserParams)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserNotCreated, err)
		return
	}

	response.Data = user
	ReturnJson(w, response)
}

func (service *UserService) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var userDto tUserDTO
	err = GetJson(r, &userDto)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserDtoNotParsed, err)
		return
	}

	if userDto.Username == "" {
		ReturnResponseWithError(w, response, ErrorTitleUserNoUsername, err)
		return
	}

	if userDto.Password == "" {
		ReturnResponseWithError(w, response, ErrorTitleUserNoPassword, err)
		return
	}

	hashedPassword, err := utils.HashPassword(userDto.Password)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUser, err)
		return
	}

	args := &orm.UpdateUserPasswordParams{
		Username:       userDto.Username,
		HashedPassword: hashedPassword,
	}

	user, err := service.Store.Queries.UpdateUserPassword(context.Background(), *args)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserPasswordNotUpdated, err)
		return
	}

	response.Data = user
	ReturnJson(w, response)
}

func (service *UserService) Delete(w http.ResponseWriter, r *http.Request) {
	response := CreateResponse(nil, nil)
	var err error

	var userDto tUserDTO
	err = GetJson(r, &userDto)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserDtoNotParsed, err)
		return
	}

	if userDto.Username == "" {
		ReturnResponseWithError(w, response, ErrorTitleUserNoUsername, err)
		return
	}

	err = service.Store.Queries.DeleteUser(context.Background(), userDto.Username)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserNotDeleted, err)
		return
	}

	response.Data = true
	ReturnJson(w, response)
}
