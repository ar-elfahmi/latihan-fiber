package model

import "time"

// MaxCreditsPerSemester adalah batas SKS yang boleh diambil satu mahasiswa
// dalam satu semester. Ini satu-satunya tempat angka batas SKS ditulis di
// kode produksi; app/service mengeksposnya ulang sebagai
// service.MaxCreditsPerSemester.
const MaxCreditsPerSemester = 24

type Course struct {
	ID            int       `json:"id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Credits       int       `json:"credits"`
	Quota         int       `json:"quota"`
	EnrolledCount int       `json:"enrolled_count"`
	Version       int       `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
}

type SwapRequest struct {
	FromCourseID int `json:"from_course_id" validate:"required,min=1,max=2147483647"`
	ToCourseID   int `json:"to_course_id" validate:"required,min=1,max=2147483647"`
}

type UpdateQuotaRequest struct {
	Quota   int `json:"quota" validate:"required,min=1,max=500"`
	Version int `json:"version" validate:"required,min=1,max=2147483647"`
}
