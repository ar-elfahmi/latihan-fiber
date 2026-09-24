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
type CreateStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// ReplaceStudentRequest adalah body PUT /students/:id (ganti seluruh field).
type ReplaceStudentRequest struct {
	NIM      string  `json:"nim"`
	Name     string  `json:"name"`
	Grade    float64 `json:"grade"`
	IsActive bool    `json:"is_active"`
}

// PatchStudentRequest adalah body PATCH /students/:id (ubah sebagian).
// Field bertipe pointer: nil artinya tidak dikirim / tidak diubah.
type PatchStudentRequest struct {
	NIM      *string  `json:"nim,omitempty"`
	Name     *string  `json:"name,omitempty"`
	Grade    *float64 `json:"grade,omitempty"`
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

// Offset dipakai di query SQL LIMIT/OFFSET.
func (q ListQuery) Offset() int {
	return (q.Page - 1) * q.Limit
}
