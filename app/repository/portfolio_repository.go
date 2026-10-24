package repository

import (
    "context"
    "time"

    "go.mongodb.org/mongo-driver/v2/bson"
    "go.mongodb.org/mongo-driver/v2/mongo"
    "latihan-fiber/app/model"
)

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
	Technologies []string        `bson:"technologies"`
	URL          string        `bson:"url"`
	Year         int        `bson:"year"`
}

type certificateDocument struct {
	Name string `bson:"name"`
	Issuer          string `bson:"issuer"`
	Date         string `bson:"date"`
}

type PortfolioRepository interface {
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
            ID:           p.ID.Hex(),  // ObjectID → string
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
        ID:           d.ID.Hex(),  // ObjectID → string
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

func (r *portfolioMongoRepository) Upsert(ctx context.Context, p model.Portfolio) (model.Portfolio, error) {
    return model.Portfolio{}, nil
}

func (r *portfolioMongoRepository) FindByStudentID(ctx context.Context, studentID int) (model.Portfolio, error) {
    return model.Portfolio{}, nil
}

func (r *portfolioMongoRepository) AddProject(ctx context.Context, studentID int, p model.Project) (model.Portfolio, error) {
    return model.Portfolio{}, nil
}

func (r *portfolioMongoRepository) DeleteProject(ctx context.Context, studentID int, projectID string) error {
    return nil
}

func (r *portfolioMongoRepository) StatsTechnologies(ctx context.Context) ([]model.TagCount, error) {
    return nil, nil
}