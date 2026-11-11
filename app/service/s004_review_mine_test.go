package service

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/config"
	"latihan-fiber/helper"
)

// s004FakeReviews adalah fake repository.ReviewRepository khusus slice S-004.
// Fake baru dibuat di tiap test (tanpa state global) — sesuai docs/TESTING.md §4.
type s004FakeReviews struct {
	repository.ReviewRepository // method lain: panic bila terpanggil

	mu      sync.Mutex
	reviews []model.Review
}

func newS004FakeReviews() *s004FakeReviews { return &s004FakeReviews{} }

func (f *s004FakeReviews) Create(_ context.Context, r model.Review) (model.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	// Meniru unique index course_user_unique di MongoDB.
	for _, existing := range f.reviews {
		if existing.CourseID == r.CourseID && existing.UserID == r.UserID {
			return model.Review{}, repository.ErrDuplicate
		}
	}
	r.ID = string(rune('0' + len(f.reviews) + 1))
	tags := r.Tags
	if tags == nil {
		tags = []string{}
	}
	r.Tags = tags
	r.Replies = []model.Reply{}
	f.reviews = append(f.reviews, r)
	return r, nil
}

func (f *s004FakeReviews) FindByCourse(_ context.Context, courseID int) ([]model.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	result := []model.Review{}
	for i := len(f.reviews) - 1; i >= 0; i-- {
		if f.reviews[i].CourseID == courseID {
			result = append(result, f.reviews[i])
		}
	}
	return result, nil
}

// s004FakeCourses hanya mengisi method yang dipakai ReviewService/CourseService.
type s004FakeCourses struct {
	repository.CourseRepository // method lain: panic bila terpanggil

	courses  map[int]model.Course
	enrolled map[[2]int]bool
	listErr  error
}

func (f *s004FakeCourses) FindByID(_ context.Context, id int) (model.Course, error) {
	c, ok := f.courses[id]
	if !ok {
		return model.Course{}, repository.ErrNotFound
	}
	return c, nil
}

func (f *s004FakeCourses) IsEnrolled(_ context.Context, courseID, userID int) (bool, error) {
	return f.enrolled[[2]int{courseID, userID}], nil
}

func (f *s004FakeCourses) ListByUser(_ context.Context, userID int) ([]model.Course, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	out := []model.Course{}
	for _, c := range f.courses {
		if f.enrolled[[2]int{c.ID, userID}] {
			out = append(out, c)
		}
	}
	return out, nil
}

var _ repository.ReviewRepository = (*s004FakeReviews)(nil)
var _ repository.CourseRepository = (*s004FakeCourses)(nil)

func s004ReviewEnv(courses *s004FakeCourses, reviews *s004FakeReviews) ReviewService {
	return NewReviewService(reviews, courses)
}

// s004ServeCreate memanggil handler dengan route param Fiber yang asli
// (/courses/:id/reviews) supaya courseID terbaca seperti di production.
func s004ServeCreate(t *testing.T, svc ReviewService, courseID, body string) (int, map[string]any) {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Post("/courses/:id/reviews", func(c *fiber.Ctx) error {
		c.Locals("user_id", 7)
		return svc.Create(c)
	})
	req := httptest.NewRequest("POST", "/courses/"+courseID+"/reviews", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("POST /courses/%s/reviews: response bukan JSON: %v", courseID, err)
	}
	return resp.StatusCode, parsed
}

// s004ServeMine memanggil CourseService.Mine (tanpa path param).
func s004ServeMine(t *testing.T, svc CourseService) (int, map[string]any) {
	t.Helper()
	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Get("/mine", func(c *fiber.Ctx) error {
		c.Locals("user_id", 7)
		return svc.Mine(c)
	})
	req := httptest.NewRequest("GET", "/mine", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()
	var parsed map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("GET /mine: response bukan JSON: %v", err)
	}
	return resp.StatusCode, parsed
}

func s004Courses() *s004FakeCourses {
	return &s004FakeCourses{
		courses: map[int]model.Course{
			1: {ID: 1, Code: "IF101", Name: "Pemrograman Backend Lanjut", Credits: 3},
			2: {ID: 2, Code: "IF102", Name: "Basis Data Lanjut", Credits: 3},
		},
		enrolled: map[[2]int]bool{{1, 7}: true},
	}
}

// TestS004MineEmpty memastikan GET /courses/mine milik user tanpa KRS
// mengembalikan courses: [] (bukan null) + total 0.
func TestS004MineEmpty(t *testing.T) {
	t.Parallel()
	courses := s004Courses()
	courses.enrolled = map[[2]int]bool{} // user 7 tidak ambil kelas apa pun
	svc := CourseService{Repo: courses}

	status, body := s004ServeMine(t, svc)
	if status != 200 {
		t.Fatalf("status = %d, mau 200 (body: %v)", status, body)
	}
	data := body["data"].(map[string]any)
	if got := len(data["courses"].([]any)); got != 0 {
		t.Errorf("courses = %d, mau 0", got)
	}
	if got := data["total_credits"].(float64); got != 0 {
		t.Errorf("total_credits = %v, mau 0", got)
	}
	if got := data["max_credits"].(float64); got != float64(MaxCreditsPerSemester) {
		t.Errorf("max_credits = %v, mau %d", got, MaxCreditsPerSemester)
	}
}

// TestS004MineFilled memastikan GET /courses/mine hanya mengembalikan
// kelas milik user login beserta total SKS.
func TestS004MineFilled(t *testing.T) {
	t.Parallel()
	svc := CourseService{Repo: s004Courses()}

	status, body := s004ServeMine(t, svc)
	if status != 200 {
		t.Fatalf("status = %d, mau 200 (body: %v)", status, body)
	}
	data := body["data"].(map[string]any)
	list := data["courses"].([]any)
	if len(list) != 1 {
		t.Fatalf("courses = %d, mau 1", len(list))
	}
	if got := list[0].(map[string]any)["code"]; got != "IF101" {
		t.Errorf("code = %v, mau IF101", got)
	}
	if got := data["total_credits"].(float64); got != 3 {
		t.Errorf("total_credits = %v, mau 3", got)
	}
}

// TestS004CreateReviewSuccess memastikan tulis ulasan sukses 201,
// tags ternormalisasi ([Seru seru] → [seru]).
func TestS004CreateReviewSuccess(t *testing.T) {
	t.Parallel()
	reviews := newS004FakeReviews()
	svc := s004ReviewEnv(s004Courses(), reviews)

	status, body := s004ServeCreate(t, svc, "1",
		`{"rating":5,"comment":"Materinya runtut dan contohnya nyata.","tags":["Seru","seru"]}`)
	if status != 201 {
		t.Fatalf("status = %d, mau 201 (body: %v)", status, body)
	}
	data := body["data"].(map[string]any)
	tags := data["tags"].([]any)
	if len(tags) != 1 || tags[0] != "seru" {
		t.Errorf("tags = %v, mau [seru] setelah dinormalisasi", tags)
	}
}

// TestS004CreateReviewDuplicate memastikan ulasan kedua untuk pasangan
// (kelas, mahasiswa) yang sama → 409 ALREADY_REVIEWED.
func TestS004CreateReviewDuplicate(t *testing.T) {
	t.Parallel()
	reviews := newS004FakeReviews()
	svc := s004ReviewEnv(s004Courses(), reviews)
	body := `{"rating":5,"comment":"Materinya runtut dan contohnya nyata."}`

	if status, _ := s004ServeCreate(t, svc, "1", body); status != 201 {
		t.Fatalf("ulasan pertama: status = %d, mau 201", status)
	}
	status, resp := s004ServeCreate(t, svc, "1", body)
	if status != 409 || resp["code"] != "ALREADY_REVIEWED" {
		t.Errorf("ulasan kedua: dapat %d %v, mau 409 ALREADY_REVIEWED", status, resp["code"])
	}
}

// TestS004CreateReviewGuards memastikan urutan cek Create lolos 4 kasus modul:
// 400 body rusak, 422 rating-9, 404 kelas tidak ada, 403 bukan peserta.
func TestS004CreateReviewGuards(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name       string
		courseID   string
		body       string
		wantStatus int
		wantCode   string
	}{
		{"body bukan JSON", "1", `{rating:`, 400, helper.CodeBadRequest},
		{"rating di luar 1..5", "1", `{"rating":9,"comment":"Komentar yang cukup panjang."}`, 422, helper.CodeValidationFailed},
		{"kelas tidak ada", "99", `{"rating":5,"comment":"Materinya runtut dan contohnya nyata."}`, 404, helper.CodeNotFound},
		{"bukan peserta kelas", "2", `{"rating":5,"comment":"Materinya runtut dan contohnya nyata."}`, 403, "NOT_ENROLLED"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			svc := s004ReviewEnv(s004Courses(), newS004FakeReviews())
			status, parsed := s004ServeCreate(t, svc, tc.courseID, tc.body)
			if status != tc.wantStatus || parsed["code"] != tc.wantCode {
				t.Errorf("dapat %d %v, mau %d %s", status, parsed["code"], tc.wantStatus, tc.wantCode)
			}
		})
	}
}
