package internal

import "github.com/google/uuid"

type BookMark struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}
