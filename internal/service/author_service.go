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
	repo repository.AuthorRepository
}

func NewAuthorService(repo repository.AuthorRepository) AuthorService {
	return &authorService{repo: repo}
}

func (s *authorService) Create(req *domain.CreateAuthorRequest) (*domain.Author, error) {
	return s.repo.Create(&domain.Author{Name: req.Name})
}

func (s *authorService) GetAll() ([]*domain.Author, error) {
	return s.repo.GetAll()
}

func (s *authorService) GetByID(id int64) (*domain.Author, error) {
	return s.repo.GetByID(id)
}

func (s *authorService) Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error) {
	return s.repo.Update(id, req)
}

func (s *authorService) Delete(id int64) error {
	return s.repo.Delete(id)
}
