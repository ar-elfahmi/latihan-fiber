package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"latihan-fiber/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// reviewDocument adalah bentuk penyimpanan ulasan di MongoDB.
type reviewDocument struct {
	ID        bson.ObjectID   `bson:"_id,omitempty"`
	CourseID  int             `bson:"course_id"`
	UserID    int             `bson:"user_id"`
	Rating    int             `bson:"rating"`
	Comment   string          `bson:"comment"`
	Tags      []string        `bson:"tags"`
	Replies   []replyDocument `bson:"replies"`
	CreatedAt time.Time       `bson:"created_at"`
	UpdatedAt time.Time       `bson:"updated_at"`
}

type replyDocument struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	UserID    int           `bson:"user_id"`
	Comment   string        `bson:"comment"`
	CreatedAt time.Time     `bson:"created_at"`
}

func (d reviewDocument) toModel() model.Review {
	tags := d.Tags
	if tags == nil {
		tags = []string{}
	}
	replies := make([]model.Reply, 0, len(d.Replies))
	for _, r := range d.Replies {
		replies = append(replies, model.Reply{
			ID:        r.ID.Hex(),
			UserID:    r.UserID,
			Comment:   r.Comment,
			CreatedAt: r.CreatedAt,
		})
	}
	return model.Review{
		ID:        d.ID.Hex(),
		CourseID:  d.CourseID,
		UserID:    d.UserID,
		Rating:    d.Rating,
		Comment:   d.Comment,
		Tags:      tags,
		Replies:   replies,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

// reviewMongoRepository menyimpan ulasan kelas di koleksi course_reviews.
type reviewMongoRepository struct {
	coll *mongo.Collection
}

// NewReviewRepository membuat repository ulasan kelas di atas database MongoDB.
func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewMongoRepository{coll: db.Collection("course_reviews")}
}

// Compile-time assertion: implementasi MongoDB wajib memenuhi kontrak.
var _ ReviewRepository = (*reviewMongoRepository)(nil)

// EnsureIndexes membuat unique index course_user_unique (course_id, user_id)
// supaya satu mahasiswa hanya bisa menulis satu ulasan per kelas.
func (r *reviewMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "user_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("course_user_unique"),
		},
		{
			Keys:    bson.D{{Key: "course_id", Value: 1}, {Key: "_id", Value: -1}},
			Options: options.Index().SetName("course_created_idx"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index course_reviews: %w", err)
	}
	return nil
}

// isMongoDuplicate mendeteksi pelanggaran unique index (kode 11000).
func isMongoDuplicate(err error) bool {
	var writeErr mongo.WriteException
	if errors.As(err, &writeErr) {
		for _, e := range writeErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
		if strings.Contains(writeErr.Error(), "E11000") {
			return true
		}
	}
	var bulkErr mongo.BulkWriteException
	if errors.As(err, &bulkErr) {
		for _, e := range bulkErr.WriteErrors {
			if e.Code == 11000 {
				return true
			}
		}
	}
	return strings.Contains(err.Error(), "E11000")
}

// Create menulis satu ulasan. Duplikat (course_id, user_id) → ErrDuplicate.
func (r *reviewMongoRepository) Create(ctx context.Context, rev model.Review) (model.Review, error) {
	t := time.Now().UTC().Truncate(time.Millisecond)
	doc := reviewDocument{
		CourseID:  rev.CourseID,
		UserID:    rev.UserID,
		Rating:    rev.Rating,
		Comment:   rev.Comment,
		Tags:      rev.Tags,
		Replies:   []replyDocument{},
		CreatedAt: t,
		UpdatedAt: t,
	}
	if doc.Tags == nil {
		doc.Tags = []string{}
	}
	res, err := r.coll.InsertOne(ctx, doc)
	if err != nil {
		if isMongoDuplicate(err) {
			return model.Review{}, ErrDuplicate
		}
		return model.Review{}, fmt.Errorf("menulis ulasan: %w", err)
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		doc.ID = oid
	}
	return doc.toModel(), nil
}

// FindByCourse mengambil ulasan satu kelas, terbaru dulu (ObjectID menurun).
func (r *reviewMongoRepository) FindByCourse(ctx context.Context, courseID int) ([]model.Review, error) {
	cursor, err := r.coll.Find(ctx, bson.M{"course_id": courseID},
		options.Find().SetSort(bson.D{{Key: "_id", Value: -1}}))
	if err != nil {
		return nil, fmt.Errorf("mengambil ulasan: %w", err)
	}
	defer cursor.Close(ctx)

	reviews := []model.Review{}
	for cursor.Next(ctx) {
		var doc reviewDocument
		if err := cursor.Decode(&doc); err != nil {
			return nil, fmt.Errorf("membaca ulasan: %w", err)
		}
		reviews = append(reviews, doc.toModel())
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("iterasi ulasan: %w", err)
	}
	return reviews, nil
}
