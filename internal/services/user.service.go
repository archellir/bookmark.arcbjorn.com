package services

import (
	"context"
	"net/http"

	"github.com/archellir/bookmark.arcbjorn.com/internal/auth"
	"github.com/archellir/bookmark.arcbjorn.com/internal/utils"

	orm "github.com/archellir/bookmark.arcbjorn.com/internal/db/orm"
)

type UserService struct {
	store      *orm.Store
	config     *utils.Config
	tokenMaker auth.IMaker
}

func NewUserService(store *orm.Store, config *utils.Config, tokenMaker auth.IMaker) *UserService {
	return &UserService{
		store:      store,
		config:     config,
		tokenMaker: tokenMaker,
	}
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

	user, err := service.store.Queries.CreateUser(context.Background(), *createUserParams)
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

	user, err := service.store.Queries.UpdateUserPassword(context.Background(), *args)
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

	_, err = service.store.Queries.GetUserByUsername(context.Background(), userDto.Username)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserNotFound, err)
		return
	}

	err = service.store.Queries.DeleteUser(context.Background(), userDto.Username)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserNotDeleted, err)
		return
	}

	response.Data = true
	ReturnJson(w, response)
}

func (service *UserService) LoginUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := service.store.Queries.GetUserByUsername(context.Background(), userDto.Username)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserNotFound, err)
		return
	}

	err = service.store.Queries.DeleteUser(context.Background(), userDto.Username)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserNotDeleted, err)
		return
	}

	err = utils.CheckPassword(userDto.Password, user.HashedPassword)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserWrongPassword, err)
		return
	}

	accessToken, err := service.tokenMaker.CreateToken(
		user.Username,
		service.config.AccessTokenDuration,
	)
	if err != nil {
		ReturnResponseWithError(w, response, ErrorTitleUserAccessTokenNotMade, err)
		return
	}

	loginData := tLoginUserResponse{
		AccessToken: accessToken,
		User:        user.Username,
	}

	response.Data = loginData
	ReturnJson(w, response)
}
