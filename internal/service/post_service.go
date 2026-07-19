package service

import (
	"fmt"

	"webapp/internal/domain"
	"webapp/internal/repository"
)

// PostService defines the business-logic contract for posts.
type PostService interface {
	Create(req *domain.CreatePostRequest) (*domain.Post, error)
	GetAll() ([]*domain.Post, error)
	GetByID(id int64) (*domain.Post, error)
	Update(id int64, req *domain.UpdatePostRequest) (*domain.Post, error)
	Delete(id int64) error
}

type postService struct {
	repo       repository.PostRepository
	authorRepo repository.AuthorRepository
}

// NewPostService wires the service to its dependencies.
// authorRepo is used to validate that the referenced author exists on create.
func NewPostService(repo repository.PostRepository, authorRepo repository.AuthorRepository) PostService {
	return &postService{repo: repo, authorRepo: authorRepo}
}

func (s *postService) Create(req *domain.CreatePostRequest) (*domain.Post, error) {
	// Validate the author exists before storing the post.
	if _, err := s.authorRepo.GetByID(req.AuthorID); err != nil {
		return nil, fmt.Errorf("author %d: %w", req.AuthorID, err)
	}

	post := &domain.Post{
		Title:    req.Title,
		Content:  req.Content,
		AuthorID: req.AuthorID,
	}
	return s.repo.Create(post)
}

func (s *postService) GetAll() ([]*domain.Post, error) {
	return s.repo.GetAll()
}

func (s *postService) GetByID(id int64) (*domain.Post, error) {
	return s.repo.GetByID(id)
}

func (s *postService) Update(id int64, req *domain.UpdatePostRequest) (*domain.Post, error) {
	return s.repo.Update(id, req)
}

func (s *postService) Delete(id int64) error {
	return s.repo.Delete(id)
}
