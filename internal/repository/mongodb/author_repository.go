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

type authorRepository struct {
	authors  *mongo.Collection
	counters *mongo.Collection // shared with postRepository; each entity uses a different _id key
}

// NewAuthorRepository returns a MongoDB-backed AuthorRepository.
func NewAuthorRepository(db *mongo.Database) repository.AuthorRepository {
	return &authorRepository{
		authors:  db.Collection("authors"),
		counters: db.Collection("counters"), // same collection as post counters, keyed by "authors"
	}
}

// authorDoc is the MongoDB document shape; separate from domain.Author to keep bson tags out of the domain layer.
type authorDoc struct {
	ID        int64     `bson:"_id"`
	Name      string    `bson:"name"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// toDomain converts the DB document to the domain struct returned to callers.
func (d authorDoc) toDomain() *domain.Author {
	return &domain.Author{
		ID:        d.ID,
		Name:      d.Name,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// nextID atomically increments the "authors" counter and returns the new value.
// Uses findAndModify (upsert) so the counter document is created on first call.
func (r *authorRepository) nextID(ctx context.Context) (int64, error) {
	var counter struct {
		Seq int64 `bson:"seq"`
	}
	err := r.counters.FindOneAndUpdate(
		ctx,
		bson.M{"_id": "authors"},                        // separate counter key from "posts"
		bson.M{"$inc": bson.M{"seq": int64(1)}},
		options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After),
	).Decode(&counter)
	if err != nil {
		return 0, fmt.Errorf("nextID: %w", err)
	}
	return counter.Seq, nil
}

func (r *authorRepository) Create(author *domain.Author) (*domain.Author, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout) // opTimeout defined in post_repository.go
	defer cancel()

	id, err := r.nextID(ctx)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	doc := authorDoc{
		ID:        id,
		Name:      author.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := r.authors.InsertOne(ctx, doc); err != nil {
		return nil, err
	}
	return doc.toDomain(), nil
}

func (r *authorRepository) GetAll() ([]*domain.Author, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	cursor, err := r.authors.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"_id": 1})) // sorted by ID for stable order
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	authors := make([]*domain.Author, 0)
	for cursor.Next(ctx) {
		var doc authorDoc
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		authors = append(authors, doc.toDomain())
	}
	return authors, cursor.Err()
}

func (r *authorRepository) GetByID(id int64) (*domain.Author, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	var doc authorDoc
	err := r.authors.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) { // translate mongo error to shared sentinel
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return doc.toDomain(), nil
}

func (r *authorRepository) Update(id int64, req *domain.UpdateAuthorRequest) (*domain.Author, error) {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	fields := bson.M{"updated_at": time.Now()}
	if req.Name != "" { // only include fields that were actually provided
		fields["name"] = req.Name
	}

	var doc authorDoc
	err := r.authors.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": fields},
		options.FindOneAndUpdate().SetReturnDocument(options.After), // return the updated document
	).Decode(&doc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, repository.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return doc.toDomain(), nil
}

func (r *authorRepository) Delete(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), opTimeout)
	defer cancel()

	res, err := r.authors.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 { // document didn't exist
		return repository.ErrNotFound
	}
	return nil
}
