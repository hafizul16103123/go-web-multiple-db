package domain

import "time"

// Post belongs to an Author via AuthorID.
type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	AuthorID  int64     `json:"author_id"` // foreign key — must reference an existing Author
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreatePostRequest struct {
	Title    string `json:"title"`
	Content  string `json:"content"`
	AuthorID int64  `json:"author_id"` // required; service validates the author exists
}

// UpdatePostRequest only allows updating content fields, not ownership.
// AuthorID is intentionally absent — reassigning a post to a different author is not supported.
type UpdatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}
