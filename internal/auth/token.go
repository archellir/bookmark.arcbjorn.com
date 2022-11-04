package auth

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("token has expired ")

type Token struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewToken(username string, duration time.Duration) (*Token, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	token := &Token{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return token, nil
}

func (token *Token) Valid() error {
	if time.Now().After(token.ExpiredAt) {
		return ErrInvalidToken
	}
	return nil
}
