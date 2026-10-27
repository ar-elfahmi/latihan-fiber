package model

import "time"

type Portfolio struct {
    ID           string        `json:"id"`
    StudentID    int           `json:"student_id"`
    Headline     string        `json:"headline"`
    Bio          string        `json:"bio"`
    Projects     []Project     `json:"projects"`
    Certificates []Certificate `json:"certificates"`
    CreatedAt    time.Time     `json:"created_at"`
    UpdatedAt    time.Time     `json:"updated_at"`
}

type Project struct {
    ID           string   `json:"id"`
    Title        string   `json:"title"`
    Technologies []string `json:"technologies"`
    URL          string   `json:"url"`
    Year         int      `json:"year"`
}

type Certificate struct {
	Name   string `json:"name"`
	Issuer string `json:"issuer"`
	Date   string `json:"date"`
}

// Untuk PUT /students/:id/portfolio
type UpsertPortfolioRequest struct {
    Headline     string             `json:"headline"     validate:"required,max=200"`
    Bio          string             `json:"bio"          validate:"required,max=1000"`
    Certificates []CertificateInput `json:"certificates" validate:"omitempty,max=20,dive"`
}

// Untuk POST /students/:id/portfolio/projects
type AddProjectRequest struct {
    Title        string   `json:"title"        validate:"required,max=200"`
    Technologies []string `json:"technologies" validate:"required,min=1,max=10,dive,min=2,max=50"`
    URL          string   `json:"url"          validate:"required,url"`
    Year         int      `json:"year"         validate:"required,min=2000,max=2100"`
}

// Untuk Certificate — kalau PUT portfolio juga terima certificates
type CertificateInput struct {
    Name   string `json:"name"   validate:"required,max=200"`
    Issuer string `json:"issuer" validate:"required,max=200"`
    Date   string `json:"date"   validate:"required"`
}

type TagCount struct {
    Technology string `json:"technology"`
    Total      int    `json:"total"`
}