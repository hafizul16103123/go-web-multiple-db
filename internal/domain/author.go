package domain

import "time"

// Author is a person who writes posts.
type Author struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateAuthorRequest struct {
	Name string `json:"name"`
}

type UpdateAuthorRequest struct {
	Name string `json:"name"`
}
