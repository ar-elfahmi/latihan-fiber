package model

import "time"

// Student adalah baris pada tabel students.
type Student struct {
	ID        int       `json:"id"`
	NIM       string    `json:"nim"`
	Name      string    `json:"name"`
	Grade     float64   `json:"grade"`
	IsActive  bool      `json:"is_active"`
	OwnerID   *int      `json:"owner_id,omitempty"` // pemilik data (yang mendaftarkan); NULL = belum ada pemilik
	CreatedAt time.Time `json:"created_at"`
}

// CreateStudentRequest adalah body POST /students.
// Validasi sepenuhnya lewat tag validate; tidak ada pemeriksaan manual
// tambahan di service.
type CreateStudentRequest struct {
	NIM      string  `json:"nim"      validate:"required,nimformat,max=20"`
	Name     string  `json:"name"     validate:"required,min=1,max=255"`
	Grade    float64 `json:"grade"    validate:"min=0,max=4"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest adalah body PUT /students/:id (ganti seluruh field).
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"      validate:"required,nimformat,max=20"`
	Name     string  `json:"name"     validate:"required,min=1,max=255"`
	Grade    float64 `json:"grade"    validate:"min=0,max=4"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest adalah body PATCH /students/:id (ubah sebagian).
// Field bertipe pointer dan tag omitnil: field yang tidak dikirim
// (nil) dilewati; field yang dikirim divalidasi sesuai tag.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"      validate:"omitnil,required,nimformat,max=20"`
	Name     *string  `json:"name,omitempty"     validate:"omitnil,required,min=1,max=255"`
	Grade    *float64 `json:"grade,omitempty"    validate:"omitnil,min=0,max=4"`
	IsActive *bool    `json:"is_active,omitempty"`
}

// ListQuery membawa parameter query string untuk GET /students.
type ListQuery struct {
	Page     int    // ?page=
	Limit    int    // ?limit=
	Search   string // ?search=
	Sort     string // ?sort=
	Order    string // ?order=
	IsActive *bool  // ?is_active=
}

// Offset dipakai di query SQL LIMIT/OFFSET (legacy; D.3 pakai cursor).
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
