// Package sqldb provides a PostgreSQL implementation of repository.PostRepository
// using only the standard library's database/sql package.
package sqldb

import (
	"database/sql"
	"errors"

	_ "github.com/lib/pq" // registers the "postgres" driver with database/sql

	"webapp/internal/domain"
	"webapp/internal/repository"
)

type postRepository struct {
	db *sql.DB
}

// New runs the posts schema migration and returns a PostgreSQL-backed PostRepository.
// Call NewAuthorRepository first so the authors table already exists.
func New(db *sql.DB) (repository.PostRepository, error) {
	r := &postRepository{db: db}
	if err := r.migrate(); err != nil {
		return nil, err
	}
	return r, nil
}

func (r *postRepository) migrate() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS posts (
			id         BIGSERIAL    PRIMARY KEY,
			title      TEXT         NOT NULL,
			content    TEXT         NOT NULL DEFAULT '',
			author_id  BIGINT       NOT NULL,
			created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func (r *postRepository) Create(post *domain.Post) (*domain.Post, error) {
	const q = `
		INSERT INTO posts (title, content, author_id, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, title, content, author_id, created_at, updated_at
	`
	row := r.db.QueryRow(q, post.Title, post.Content, post.AuthorID)
	return scanPost(row)
}

func (r *postRepository) GetAll() ([]*domain.Post, error) {
	const q = `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM posts
		ORDER BY id
	`
	rows, err := r.db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*domain.Post, 0)
	for rows.Next() {
		p := &domain.Post{}
		if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *postRepository) GetByID(id int64) (*domain.Post, error) {
	const q = `
		SELECT id, title, content, author_id, created_at, updated_at
		FROM posts WHERE id = $1
	`
	p, err := scanPost(r.db.QueryRow(q, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return p, err
}

func (r *postRepository) Update(id int64, req *domain.UpdatePostRequest) (*domain.Post, error) {
	const q = `
		UPDATE posts
		SET    title      = COALESCE(NULLIF($1, ''), title),
		       content    = COALESCE(NULLIF($2, ''), content),
		       updated_at = NOW()
		WHERE  id = $3
		RETURNING id, title, content, author_id, created_at, updated_at
	`
	p, err := scanPost(r.db.QueryRow(q, req.Title, req.Content, id))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNotFound
	}
	return p, err
}

func (r *postRepository) Delete(id int64) error {
	res, err := r.db.Exec(`DELETE FROM posts WHERE id = $1`, id)
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

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanPost(row scanner) (*domain.Post, error) {
	p := &domain.Post{}
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.AuthorID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}
