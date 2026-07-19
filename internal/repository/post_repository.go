package repository

import (
	"errors"

	"webapp/internal/domain"
)

// ErrNotFound is returned by any repository implementation when a record does not exist.
var ErrNotFound = errors.New("post not found")

// PostRepository is the data-access contract.
// Swap implementations (SQL, MongoDB, in-memory) without touching any other layer.
type PostRepository interface {
	Create(post *domain.Post) (*domain.Post, error)
	GetAll() ([]*domain.Post, error)
	GetByID(id int64) (*domain.Post, error)
	Update(id int64, req *domain.UpdatePostRequest) (*domain.Post, error)
	Delete(id int64) error
}
