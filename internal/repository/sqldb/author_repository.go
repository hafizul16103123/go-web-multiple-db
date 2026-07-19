package sqldb

import (
	"database/sql"
	"errors"

	"webapp/internal/domain"
	"webapp/internal/repository"
)

type authorRepository struct {
	db *sql.DB
}

// NewAuthorRepository creates the authors table if needed and returns the repository.
// Call this before NewPostRepository so the authors table exists first.
func NewAuthorRepository(db *sql.DB) (repository.AuthorRepository, error) {
	r := &authorRepository{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *authorRepository) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS authors (
			id         BIGSERIAL    PRIMARY KEY,
			name       TEXT         NOT NULL,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (r *authorRepository) Create(author *domain.Author) (*domain.Author, error) {
	const q = `
		INSERT INTO authors (name, created_at, updated_at)
		VALUES ($1, NOW(), NOW())
		RETURNING id, name, created_at, updated_at
	`
	row := r.db.QueryRow(q, author.Name)
	return scanAuthor(row)
}

func (r *authorRepository) GetAll() ([]*domain.Author, error) {
	const q = `SELECT id, name, created_at, updated_at FROM authors ORDER BY id`
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	authors := make([]*domain.Author, 0)
	for rows.Next() {
		a := &domain.Author{}
		if err := rows.Scan(&a.ID, &a.Name, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		authors = append(authors, a)
	}
	return authors, rows.Err()
}

func (r *authorRepository) GetByID(id int64) (*domain.Author, error) {
	const q = `SELECT id, name, created_at, updated_at FROM authors WHERE id = $1`
	a, err := scanAuthor(r.db.QueryRow(q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return a, err
}

func (r *authorRepository) Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error) {
	const q = `
		UPDATE authors
		SET    name       = COALESCE(NULLIF($1, ''), name),
		       updated_at = NOW()
		WHERE  id = $2
		RETURNING id, name, created_at, updated_at
	`
	a, err := scanAuthor(r.db.QueryRow(q, req.Name, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return a, err
}

func (r *authorRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM authors WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrNotFound
	}
	return nil
}

func scanAuthor(row scanner) (*domain.Author, error) {
	a := &domain.Author{}
	err := row.Scan(&a.ID, &a.Name, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}
