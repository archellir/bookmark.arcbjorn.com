package models

import "errors"

var (
	// Common errors
	ErrNotFound    = errors.New("not found")
	ErrInvalidID   = errors.New("invalid ID")
	ErrDuplicate   = errors.New("duplicate entry")
	ErrInvalidData = errors.New("invalid data")

	// Bookmark errors
	ErrBookmarkNotFound = errors.New("bookmark not found")
	ErrBookmarkExists   = errors.New("bookmark already exists")

	// Tag errors
	ErrTagNotFound = errors.New("tag not found")
	ErrTagExists   = errors.New("tag already exists")

	// Folder errors
	ErrFolderNotFound   = errors.New("folder not found")
	ErrFolderExists     = errors.New("folder already exists")
	ErrCyclicFolder     = errors.New("cyclic folder relationship not allowed")
	ErrFolderHasContent = errors.New("folder contains bookmarks or subfolders")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrTokenExpired       = errors.New("token expired")
	ErrTokenInvalid       = errors.New("token invalid")
)