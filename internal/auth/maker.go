package auth

import "time"

// Maker is an interface for managing tokens
type IMaker interface {
	// for specific username & duration
	CreateToken(username string, duration time.Duration) (string, error)

	VerifyToken(token string) (*Token, error)
}
