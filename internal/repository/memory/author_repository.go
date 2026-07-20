package memory

import (
	"sync"
	"time"

	"webapp/internal/domain"
	"webapp/internal/repository"
)

type authorRepository struct {
	mu      sync.RWMutex         // guards store and counter for concurrent access
	store   map[int64]*domain.Author
	counter int64 // monotonically incrementing ID; replaces DB auto-increment
}

// NewAuthorRepository returns a thread-safe in-memory AuthorRepository.
func NewAuthorRepository() repository.AuthorRepository {
	return &authorRepository{store: make(map[int64]*domain.Author)}
}

func (r *authorRepository) Create(author *domain.Author) (*domain.Author, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.counter++
	author.ID = r.counter
	author.CreatedAt = time.Now()
	author.UpdatedAt = time.Now()

	cloned := *author        // store a copy so the caller can't mutate internal state
	r.store[author.ID] = &cloned
	return author, nil
}

func (r *authorRepository) GetAll() ([]*domain.Author, error) {
	r.mu.RLock() // read lock: multiple readers allowed concurrently
	defer r.mu.RUnlock()

	authors := make([]*domain.Author, 0, len(r.store))
	for _, a := range r.store {
		cloned := *a // return copies so callers can't mutate store entries
		authors = append(authors, &cloned)
	}
	return authors, nil
}

func (r *authorRepository) GetByID(id int64) (*domain.Author, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	a, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	cloned := *a
	return &cloned, nil
}

func (r *authorRepository) Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.store[id]
	if !ok {
		return nil, repository.ErrNotFound
	}
	if req.Name != "" { // empty string means "no change" — mirrors COALESCE logic in sqldb
		a.Name = req.Name
	}
	a.UpdatedAt = time.Now()

	cloned := *a
	return &cloned, nil
}

func (r *authorRepository) Delete(id int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.store[id]; !ok {
		return repository.ErrNotFound
	}
	delete(r.store, id)
	return nil
}
