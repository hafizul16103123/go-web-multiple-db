package domain

import "time"

// Post belongs to an Author via AuthorID.
type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  int64     `json:"author_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreatePostRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID int64  `json:"author_id"`
}

// UpdatePostRequest only allows updating content fields, not ownership.
type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
