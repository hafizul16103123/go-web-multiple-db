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
	Name string `json:"name"` // required; handler rejects empty string
}

type UpdateAuthorRequest struct {
	Name string `json:"name"` // empty string means "no change" (COALESCE in SQL, nil-check in memory/mongo)
}
