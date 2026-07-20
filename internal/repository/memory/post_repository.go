package memory

import (
	"sync"
	"time"

	"webapp/internal/domain"
	"webapp/internal/repository"
)

type postRepository struct {
	mu      sync.RWMutex       // guards store and counter for concurrent access
	store   map[int64]*domain.Post
	counter int64 // monotonically incrementing ID; replaces DB auto-increment
}

// NewPostRepository returns a thread-safe in-memory PostRepository.
func NewPostRepository() repository.PostRepository {
	return &postRepository{store: make(map[int64]*domain.Post)}
}

func (r *postRepository) Create(post *domain.Post) (*domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	post.ID = r.counter
	post.CreatedAt = time.Now()
	post.UpdatedAt = time.Now()

	cloned := *post        // store a copy so the caller can't mutate internal state
	r.store[post.ID] = &cloned
	return post, nil
}

func (r *postRepository) GetAll() ([]*domain.Post, error) {
	r.mu.RLock() // read lock: multiple readers allowed concurrently
	defer r.mu.RUnlock()

	posts := make([]*domain.Post, 0, len(r.store))
	for _, p := range r.store {
		cloned := *p // return copies so callers can't mutate store entries
		posts = append(posts, &cloned)
	}
	return posts, nil
}

func (r *postRepository) GetByID(id int64) (*domain.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cloned := *p
	return &cloned, nil
}

func (r *postRepository) Update(id int64, req *domain.UpdatePostRequest) (*domain.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if req.Title != "" { // empty string means "no change" — mirrors COALESCE logic in sqldb
		p.Title = req.Title
	}
	if req.Content != "" {
		p.Content = req.Content
	}
	p.UpdatedAt = time.Now()

	cloned := *p
	return &cloned, nil
}

func (r *postRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.store, id)
	return nil
}
