// Package mongodb provides a MongoDB implementation of repository.PostRepository.
// Integer IDs are generated via an atomic counter document so the domain model
// stays consistent with the SQL implementation.
package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"webapp/internal/domain"
	"webapp/internal/repository"
)

const opTimeout = 10 * time.Second

type postRepository struct {
	posts    *mongo.Collection
	counters *mongo.Collection
}

// New returns a MongoDB-backed PostRepository.
func New(db *mongo.Database) repository.PostRepository {
	return &postRepository{
		posts:    db.Collection("posts"),
		counters: db.Collection("counters"),
	}
}

type postDoc struct {
	ID        int64     `bson:"_id"`
	Title     string    `bson:"title"`
	Content   string    `bson:"content"`
	AuthorID  int64     `bson:"author_id"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func (d postDoc) toDomain() *domain.Post {
	return &domain.Post{
		ID:        d.ID,
		Title:     d.Title,
		Content:   d.Content,
		AuthorID:  d.AuthorID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// nextID atomically increments the "posts" counter and returns the new value.
func (r *postRepository) nextID(ctx context.Context) (int64, error) {
	var counter struct {
		Seq int64 `bson:"seq"`
	}
	err := r.counters.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "posts"},
		bson.M{"$inc": bson.M{"seq": int64(1)}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("nextID: %w", err)
	}
	return counter.Seq, nil
}

func (r *postRepository) Create(post *domain.Post) (*domain.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := postDoc{
		ID:        id,
		Title:     post.Title,
		Content:   post.Content,
		AuthorID:  post.AuthorID,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := r.posts.InsertOne(ctx, doc); err != nil {
		return nil, err
	}
	return doc.toDomain(), nil
}

func (r *postRepository) GetAll() ([]*domain.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	cursor, err := r.posts.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	posts := make([]*domain.Post, 0)
	for cursor.Next(ctx) {
		var doc postDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		posts = append(posts, doc.toDomain())
	}
	return posts, cursor.Err()
}

func (r *postRepository) GetByID(id int64) (*domain.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	var doc postDoc
	err := r.posts.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return doc.toDomain(), nil
}

func (r *postRepository) Update(id int64, req *domain.UpdatePostRequest) (*domain.Post, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	fields := bson.M{"updated_at": time.Now()}
	if req.Title != "" {
		fields["title"] = req.Title
	}
	if req.Content != "" {
		fields["content"] = req.Content
	}

	var doc postDoc
	err := r.posts.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": fields},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return doc.toDomain(), nil
}

func (r *postRepository) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	res, err := r.posts.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return repository.ErrNotFound
	}
	return nil
}
