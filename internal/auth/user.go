package internal

import (
	"database/sql"
	"net/http"
)

type userResponse struct {
	Username string `json:"username"`
}

type loginUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginUserResponse struct {
	AccessToken string       `json:"access_token"`
	User        userResponse `json:"user"`
}

// PSEUDO CODE
func loginUser() {
	var req loginUserRequest

	user.err := server.store.GerUser()

	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, error)
			return
		}

		ctx.JSON(http.StatusInternalServerError, error)
		return
	}

	err = util.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, error)
	}

	accessToken, err := server.tokenMaker.CreateToken(user.Username, server.config.AccessTokenDuration)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, error)
		return
	}

	rsp := loginUserResponse{
		AccessToken: accessToken,
		User: newUserResponse(user)
	}

	ctx.JSON(http.StatusOK, rsp)
}