package service

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

// ReviewService melayani ulasan kelas (MongoDB, koleksi course_reviews).
// Satu mahasiswa hanya boleh menulis satu ulasan per kelas (unique index
// course_user_unique); hanya peserta kelas yang boleh menulis, daftar
// ulasan cukup dibaca user login mana pun.
type ReviewService struct {
	Reviews repository.ReviewRepository
	Courses repository.CourseRepository
}

// NewReviewService merakit ReviewService dari repository ulasan (MongoDB)
// dan repository kelas (PostgreSQL).
func NewReviewService(reviews repository.ReviewRepository, courses repository.CourseRepository) ReviewService {
	return ReviewService{Reviews: reviews, Courses: courses}
}

// translateReviewError memetakan error repository ulasan menjadi *helper.AppError.
// Tabel modul (S-005 white-box test): ErrInvalidID→400 BAD_REQUEST,
// ErrNotFound→404 NOT_FOUND, ErrDuplicate→409 ALREADY_REVIEWED,
// ErrNotOwner→403 FORBIDDEN, ErrOwnReview→409 OWN_REVIEW,
// ErrAlreadyVoted→409 ALREADY_VOTED, ErrReplyLimit→409 REPLY_LIMIT,
// wrapped ErrDuplicate→409 (via errors.Is), lain→500 INTERNAL.
//
// Catatan kode error: repo memakai helper.CodeValidationFailed
// ("VALIDATION_FAILED", D-002 stabil), bukan "VALIDATION_ERROR" seperti
// tertulis di modul; S-005 menyelaraskan ekspektasi test.
func translateReviewError(err error) error {
	switch {
	case errors.Is(err, repository.ErrInvalidID):
		return helper.BadRequest(err.Error())
	case errors.Is(err, pgx.ErrNoRows),
		errors.Is(err, repository.ErrNotFound):
		return helper.NotFound("kelas")
	case errors.Is(err, repository.ErrDuplicate):
		return &helper.AppError{Status: fiber.StatusConflict, Code: "ALREADY_REVIEWED", Message: "sudah menulis ulasan untuk kelas ini"}
	case errors.Is(err, repository.ErrNotOwner):
		return helper.Forbidden(err.Error())
	case errors.Is(err, repository.ErrOwnReview):
		return &helper.AppError{Status: fiber.StatusConflict, Code: "OWN_REVIEW", Message: err.Error()}
	case errors.Is(err, repository.ErrAlreadyVoted):
		return &helper.AppError{Status: fiber.StatusConflict, Code: "ALREADY_VOTED", Message: err.Error()}
	case errors.Is(err, repository.ErrReplyLimit):
		return &helper.AppError{Status: fiber.StatusConflict, Code: "REPLY_LIMIT", Message: err.Error()}
	default:
		return helper.Internal(err)
	}
}

// Create menulis satu ulasan untuk kelas id.
// Urutan cek (disepakati, lolos 4 kasus modul): parse JSON (400) → validasi
// body (422) → cari kelas (404) → cek peserta (403 NOT_ENROLLED) →
// NormalizeTags → Create (409 ALREADY_REVIEWED bila duplikat).
func (s ReviewService) Create(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	var req model.CreateReviewRequest
	if err := helper.ParseBody(c, &req); err != nil {
		return helper.BadRequest("body harus berupa JSON yang valid")
	}
	if fields := helper.ValidateStruct(req); len(fields) > 0 {
		return helper.Validation(fields)
	}
	if _, err := s.Courses.FindByID(c.UserContext(), id); err != nil {
		return translateReviewError(err)
	}
	enrolled, err := s.Courses.IsEnrolled(c.UserContext(), id, CurrentUserID(c))
	if err != nil {
		return helper.Internal(err)
	}
	if !enrolled {
		return &helper.AppError{Status: fiber.StatusForbidden, Code: "NOT_ENROLLED", Message: "hanya peserta kelas yang boleh menulis ulasan"}
	}
	tags := NormalizeTags(req.Tags)
	if tags == nil {
		tags = []string{}
	}
	saved, err := s.Reviews.Create(c.UserContext(), model.Review{
		CourseID: id,
		UserID:   CurrentUserID(c),
		Rating:   req.Rating,
		Comment:  req.Comment,
		Tags:     tags,
	})
	if err != nil {
		return translateReviewError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(model.WebResponse{
		Success: true,
		Message: "ulasan berhasil ditulis",
		Data:    saved,
	})
}

// ListByCourse mengambil ulasan satu kelas, terbaru dulu.
// Cursor pagination: ?limit (default 10, maks 50), ?before (id ulasan
// terakhir yang sudah dilihat); opsional ?tag untuk saring satu tag.
// Kelas kosong → data: [] (bukan null); kelas tidak ada → 404.
func (s ReviewService) ListByCourse(c *fiber.Ctx) error {
	id, err := courseID(c)
	if err != nil {
		return err
	}
	if _, err := s.Courses.FindByID(c.UserContext(), id); err != nil {
		return translateReviewError(err)
	}
	want := c.QueryInt("limit", 10)
	if want <= 0 || want > 50 {
		want = 10
	}
	reviews, err := s.Reviews.FindByCourse(c.UserContext(), id)
	if err != nil {
		return translateReviewError(err)
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	if tag := strings.ToLower(strings.TrimSpace(c.Query("tag", ""))); tag != "" {
		filtered := []model.Review{}
		for _, r := range reviews {
			for _, t := range r.Tags {
				if t == tag {
					filtered = append(filtered, r)
					break
				}
			}
		}
		reviews = filtered
	}
	start := 0
	if before := c.Query("before", ""); before != "" {
		idx := -1
		for i, r := range reviews {
			if r.ID == before {
				idx = i
				break
			}
		}
		if idx < 0 {
			return translateReviewError(repository.ErrInvalidID)
		}
		start = idx + 1
	}
	page := reviews[start:]
	hasMore := len(page) > want
	out := page
	if hasMore {
		out = page[:want]
	}
	if out == nil {
		out = []model.Review{}
	}
	next := ""
	if hasMore && len(out) > 0 {
		next = out[len(out)-1].ID
	}
	return c.JSON(model.ReviewPage{
		Success: true,
		Message: "daftar ulasan berhasil diambil",
		Data:    out,
		Meta:    model.ReviewPageMeta{Limit: want, HasMore: hasMore, NextBefore: next},
	})
}
