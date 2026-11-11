package model

import "time"

// Review adalah satu ulasan kelas yang ditulis mahasiswa.
// Satu mahasiswa hanya boleh menulis satu ulasan per kelas
// (dijaga unique index course_user_unique di MongoDB).
type Review struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	CourseID  int       `json:"course_id" bson:"course_id"`
	UserID    int       `json:"user_id" bson:"user_id"`
	Rating    int       `json:"rating" bson:"rating"`
	Comment   string    `json:"comment" bson:"comment"`
	Tags      []string  `json:"tags" bson:"tags"`
	Replies   []Reply   `json:"replies" bson:"replies"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

// Reply adalah balasan atas sebuah ulasan.
// Endpoint reply (/reviews/{id}, summary) DITUNDA ke S-007/luar Modul 12;
// struct minimal ini disiapkan supaya bentuk Review stabil sejak S-004.
type Reply struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	UserID    int       `json:"user_id" bson:"user_id"`
	Comment   string    `json:"comment" bson:"comment"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
}

// CreateReviewRequest adalah body POST /courses/:id/reviews.
// Aturan panjang comment (min 10) disepakati dengan solo dev (S-004 Q4);
// modul hanya mensyaratkan "cukup panjang" tanpa angka.
type CreateReviewRequest struct {
	Rating  int      `json:"rating" validate:"required,min=1,max=5"`
	Comment string   `json:"comment" validate:"required,min=10,max=1000"`
	Tags    []string `json:"tags" validate:"omitempty,max=10,dive,min=2,max=50"`
}

// ReviewPage adalah response GET /courses/:id/reviews. Bentuk meta-nya
// berbeda dari WebResponse (cursor, bukan nomor halaman), maka ia memiliki
// struct sendiri agar dapat didokumentasikan dengan jujur (S-007).
type ReviewPage struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Data    []Review       `json:"data"`
	Meta    ReviewPageMeta `json:"meta"`
}

// ReviewPageMeta adalah cursor pagination untuk daftar ulasan.
type ReviewPageMeta struct {
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"has_more"`
	NextBefore string `json:"next_before"`
}

// ReviewResponse adalah response satu ulasan (GET /reviews/:id).
type ReviewResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    Review `json:"data"`
}

// ReviewSummary adalah ringkasan ulasan satu kelas: jumlah ulasan,
// rata-rata rating (2 desimal), dan distribusi rating 1..5.
type ReviewSummary struct {
	CourseID      int            `json:"course_id"`
	TotalReviews  int            `json:"total_reviews"`
	AverageRating float64        `json:"average_rating"`
	Distribution  map[string]int `json:"distribution"`
}

// ReviewSummaryResponse adalah response GET /courses/:id/reviews/summary.
type ReviewSummaryResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    ReviewSummary `json:"data"`
}
