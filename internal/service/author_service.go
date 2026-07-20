package service

import (
	"webapp/internal/domain"
	"webapp/internal/repository"
)

// AuthorService defines the business-logic contract for authors.
type AuthorService interface {
	Create(req *domain.CreateAuthorRequest) (*domain.Author, error)
	GetAll() ([]*domain.Author, error)
	GetByID(id int64) (*domain.Author, error)
	Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error)
	Delete(id int64) error
}

type authorService struct {
	repo repository.AuthorRepository // the concrete DB implementation injected at startup
}

// NewAuthorService wires the service to its repository dependency.
func NewAuthorService(repo repository.AuthorRepository) AuthorService {
	return &authorService{repo: repo}
}

func (s *authorService) Create(req *domain.CreateAuthorRequest) (*domain.Author, error) {
	// map the request DTO to the domain struct before passing to the repo
	return s.repo.Create(&domain.Author{Name: req.Name})
}

func (s *authorService) GetAll() ([]*domain.Author, error) {
	return s.repo.GetAll()
}

func (s *authorService) GetByID(id int64) (*domain.Author, error) {
	return s.repo.GetByID(id) // returns repository.ErrNotFound if id doesn't exist
}

func (s *authorService) Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error) {
	return s.repo.Update(id, req) // empty fields in req are treated as "no change" by each repo
}

func (s *authorService) Delete(id int64) error {
	return s.repo.Delete(id)
}
