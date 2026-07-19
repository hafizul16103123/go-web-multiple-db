package repository

import "webapp/internal/domain"

// AuthorRepository is the data-access contract for authors.
// ErrNotFound (defined in post_repository.go) is shared across all repositories.
type AuthorRepository interface {
	Create(author *domain.Author) (*domain.Author, error)
	GetAll() ([]*domain.Author, error)
	GetByID(id int64) (*domain.Author, error)
	Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error)
	Delete(id int64) error
}
