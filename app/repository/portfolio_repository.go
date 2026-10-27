package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"latihan-fiber/app/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrProjectLimit = errors.New("jumlah proyek sudah mencapai batas")

type portfolioDocument struct {
	ID           bson.ObjectID         `bson:"_id,omitempty"`
	StudentID    int                   `bson:"student_id"`
	Headline     string                `bson:"headline"`
	Bio          string                `bson:"bio"`
	Projects     []projectDocument     `bson:"projects"`
	Certificates []certificateDocument `bson:"certificates"`
	CreatedAt    time.Time             `bson:"created_at"`
	UpdatedAt    time.Time             `bson:"updated_at"`
}

type projectDocument struct {
	ID           bson.ObjectID `bson:"_id,omitempty"`
	Title        string        `bson:"title"`
	Technologies []string      `bson:"technologies"`
	URL          string        `bson:"url"`
	Year         int           `bson:"year"`
}

type certificateDocument struct {
	Name   string `bson:"name"`
	Issuer string `bson:"issuer"`
	Date   string `bson:"date"`
}

type PortfolioRepository interface {
	EnsureIndexes(ctx context.Context) error
	Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error)
	FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error)
	AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error)
	DeleteProject(ctx context.Context, studentID int, projectID string) error
	StatsTechnologies(ctx context.Context) ([]model.TagCount, error)
}

func (d portfolioDocument) toModel() model.Portfolio {
	projects := make([]model.Project, 0, len(d.Projects))
	for _, p := range d.Projects {
		projects = append(projects, model.Project{
			ID:           p.ID.Hex(), // ObjectID → string
			Title:        p.Title,
			Technologies: p.Technologies,
			URL:          p.URL,
			Year:         p.Year,
		})
	}

	certs := make([]model.Certificate, 0, len(d.Certificates))
	for _, c := range d.Certificates {
		certs = append(certs, model.Certificate{
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   c.Date,
		})
	}

	return model.Portfolio{
		ID:           d.ID.Hex(), // ObjectID → string
		StudentID:    d.StudentID,
		Headline:     d.Headline,
		Bio:          d.Bio,
		Projects:     projects,
		Certificates: certs,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

type portfolioMongoRepository struct {
	coll *mongo.Collection
}

func NewPortfolioRepository(db *mongo.Database) PortfolioRepository {
	return &portfolioMongoRepository{
		coll: db.Collection("student_portfolios"),
	}
}

func now() time.Time {
	return time.Now().UTC().Truncate(time.Millisecond)
}

func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
	var doc portfolioDocument
	err := r.coll.FindOne(ctx, bson.M{"student_id": studentID}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Portfolio{}, ErrNotFound
		}
		return model.Portfolio{}, fmt.Errorf("mengambil portofolio: %w", err)
	}
	return doc.toModel(), nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error) {
	doc := projectDocument{
		ID:           bson.NewObjectID(),
		Title:        p.Title,
		Technologies: p.Technologies,
		URL:          p.URL,
		Year:         p.Year,
	}

	// "projects.19 tidak ada" = array belum berisi 20 elemen (indeks 0–19)
	// Pemeriksaan batas dan penambahan terjadi dalam SATU operasi atomik
	limitField := "projects.19"
	var updated portfolioDocument
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{
			"student_id": studentID,
			limitField:   bson.M{"$exists": false},
		},
		bson.M{"$push": bson.M{"projects": doc}},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	).Decode(&updated)

	if err == nil {
		return updated.toModel(), nil
	}
	if !errors.Is(err, mongo.ErrNoDocuments) {
		return model.Portfolio{}, fmt.Errorf("menambah proyek: %w", err)
	}

	// Cari tahu kenapa tidak ada dokumen yang cocok
	count, err := r.coll.CountDocuments(ctx, bson.M{"student_id": studentID})
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("memeriksa portofolio: %w", err)
	}
	if count == 0 {
		return model.Portfolio{}, ErrNotFound
	}
	return model.Portfolio{}, ErrProjectLimit
}

func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
	oid, err := bson.ObjectIDFromHex(projectID)
	if err != nil {
		return ErrNotFound
	}

	res, err := r.coll.UpdateOne(
		ctx,
		bson.M{"student_id": studentID},
		bson.M{"$pull": bson.M{"projects": bson.M{"_id": oid}}},
	)
	if err != nil {
		return fmt.Errorf("menghapus proyek: %w", err)
	}
	// ModifiedCount 0 = dokumen tidak ada ATAU proyek yang dimaksud tidak ada
	// di dalam array. Dua-duanya "tidak ditemukan" -> 404, bukan 204 senyap.
	if res.ModifiedCount == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *portfolioMongoRepository) StatsTechnologies(ctx context.Context) ([]model.TagCount, error) {
	pipeline := mongo.Pipeline{
		// 1. Pecah array projects: 1 dokumen → N dokumen (per proyek)
		{{Key: "$unwind", Value: "$projects"}},
		// 2. Pecah array technologies: 1 dokumen → M dokumen (per teknologi)
		{{Key: "$unwind", Value: "$projects.technologies"}},
		// 3. Hitung kemunculan tiap teknologi
		{{Key: "$group", Value: bson.M{
			"_id":   "$projects.technologies",
			"total": bson.M{"$sum": 1},
		}}},
		// 4. Urutkan terbanyak dulu
		{{Key: "$sort", Value: bson.D{{Key: "total", Value: -1}}}},
		// 5. Ambil 10 teratas
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := r.coll.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("agregasi teknologi: %w", err)
	}
	defer cursor.Close(ctx)

	var out []struct {
		Tech  string `bson:"_id"`
		Total int    `bson:"total"`
	}
	if err := cursor.All(ctx, &out); err != nil {
		return nil, fmt.Errorf("membaca hasil agregasi: %w", err)
	}

	result := make([]model.TagCount, 0, len(out))
	for _, o := range out {
		result = append(result, model.TagCount{
			Technology: o.Tech,
			Total:      o.Total,
		})
	}
	return result, nil
}

func (r *portfolioMongoRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.coll.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			// Satu mahasiswa hanya boleh punya SATU portofolio
			Keys:    bson.D{{Key: "student_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("student_unique"),
		},
	})
	if err != nil {
		return fmt.Errorf("membuat index student_portfolios: %w", err)
	}
	return nil
}

func (r *portfolioMongoRepository) Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error) {
	t := now()

	// Data yang akan disimpan/diupdate
	update := bson.M{
		"$set": bson.M{
			"headline":     p.Headline,
			"bio":          p.Bio,
			"certificates": certsToDoc(p.Certificates),
			"updated_at":   t,
		},
		"$setOnInsert": bson.M{
			"student_id": p.StudentID,
			"projects":   []projectDocument{},
			"created_at": t,
		},
	}

	opts := options.FindOneAndUpdate().
		SetUpsert(true).
		SetReturnDocument(options.After)

	var doc portfolioDocument
	err := r.coll.FindOneAndUpdate(
		ctx,
		bson.M{"student_id": p.StudentID},
		update,
		opts,
	).Decode(&doc)
	if err != nil {
		return model.Portfolio{}, fmt.Errorf("upsert portofolio: %w", err)
	}
	return doc.toModel(), nil
}

// certsToDoc mengubah []model.Certificate menjadi []certificateDocument
func certsToDoc(certs []model.Certificate) []certificateDocument {
	result := make([]certificateDocument, 0, len(certs))
	for _, c := range certs {
		result = append(result, certificateDocument{
			Name:   c.Name,
			Issuer: c.Issuer,
			Date:   c.Date,
		})
	}
	return result
}
