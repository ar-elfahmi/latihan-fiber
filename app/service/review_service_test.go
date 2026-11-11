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
)

// fakeReviewRepo adalah fake repository.ReviewRepository.
//
// PENTING: fake selalu dibuat PER TEST lewat newFakeReviewRepo(). Tidak ada
// variabel global yang menyimpan state fake, karena fake global bocor antar
// test (temuan B1 #4).
type fakeReviewRepo struct {
	repository.ReviewRepository // method lain tidak dipakai

	mu      sync.Mutex
	reviews []model.Review
	seq     int
}

// newFakeReviewRepo membuat fake baru yang kosong untuk setiap test.
func newFakeReviewRepo() *fakeReviewRepo {
	return &fakeReviewRepo{}
}

func (f *fakeReviewRepo) Create(_ context.Context, r model.Review) (model.Review, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, existing := range f.reviews {
		if existing.CourseID == r.CourseID && existing.UserID == r.UserID {
			return model.Review{}, repository.ErrDuplicate
		}
	}
	f.seq++
	r.ID = "ulasan-" + string(rune('0'+f.seq))
	if r.Tags == nil {
		r.Tags = []string{}
	}
	r.Replies = []model.Reply{}
	f.reviews = append(f.reviews, r)
	return r, nil
}

func (f *fakeReviewRepo) FindByCourse(_ context.Context, courseID int) ([]model.Review, error) {
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

// fakeReviewCourses adalah fake repository.CourseRepository yang dipakai
// ReviewService.
type fakeReviewCourses struct {
	repository.CourseRepository // method lain tidak dipakai

	courses  map[int]model.Course
	enrolled map[[2]int]bool
}

func newFakeReviewCourses() *fakeReviewCourses {
	return &fakeReviewCourses{
		courses: map[int]model.Course{
			1: {ID: 1, Code: "IF101", Name: "Pemrograman Backend Lanjut", Credits: 3},
			2: {ID: 2, Code: "IF102", Name: "Basis Data Lanjut", Credits: 3},
		},
		enrolled: map[[2]int]bool{{1, 7}: true, {1, 8}: true},
	}
}

func (f *fakeReviewCourses) FindByID(_ context.Context, id int) (model.Course, error) {
	c, ok := f.courses[id]
	if !ok {
		return model.Course{}, repository.ErrNotFound
	}
	return c, nil
}

func (f *fakeReviewCourses) IsEnrolled(_ context.Context, courseID, userID int) (bool, error) {
	return f.enrolled[[2]int{courseID, userID}], nil
}

var _ repository.ReviewRepository = (*fakeReviewRepo)(nil)
var _ repository.CourseRepository = (*fakeReviewCourses)(nil)

// serveReview memanggil handler Create/ListByCourse lewat Fiber supaya
// route param dan ErrorHandler produksi ikut teruji.
func serveReview(t *testing.T, svc ReviewService, userID int, method, path, body string) (int, map[string]any) {
	t.Helper()

	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Post("/courses/:id/reviews", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return svc.Create(c)
	})
	app.Get("/courses/:id/reviews", func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		return svc.ListByCourse(c)
	})

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test: %v", err)
	}
	defer resp.Body.Close()

	parsed := map[string]any{}
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		t.Fatalf("response bukan JSON: %v", err)
	}
	return resp.StatusCode, parsed
}

// TestReviewServiceCreateNormalizesTags memastikan tag ulasan disimpan dalam
// bentuk ternormalisasi.
func TestReviewServiceCreateNormalizesTags(t *testing.T) {
	t.Parallel()

	repo := newFakeReviewRepo()
	svc := NewReviewService(repo, newFakeReviewCourses())

	status, body := serveReview(t, svc, 7, "POST", "/courses/1/reviews",
		`{"rating":5,"comment":"Materinya runtut dan contohnya nyata.","tags":[" Seru ","seru","Go"]}`)
	if status != 201 {
		t.Fatalf("status = %d, mau 201 (body: %v)", status, body)
	}
	data := body["data"].(map[string]any)
	tags := data["tags"].([]any)
	if len(tags) != 2 || tags[0] != "seru" || tags[1] != "go" {
		t.Errorf("tags = %v, mau [seru go] setelah dinormalisasi", tags)
	}
}

// TestReviewServiceCreateDuplicate memastikan ulasan kedua untuk kelas yang
// sama ditolak dengan 409 ALREADY_REVIEWED.
func TestReviewServiceCreateDuplicate(t *testing.T) {
	t.Parallel()

	repo := newFakeReviewRepo()
	svc := NewReviewService(repo, newFakeReviewCourses())
	body := `{"rating":4,"comment":"Komentar yang cukup panjang untuk lolos."}`

	if status, _ := serveReview(t, svc, 7, "POST", "/courses/1/reviews", body); status != 201 {
		t.Fatalf("ulasan pertama: status = %d, mau 201", status)
	}
	status, parsed := serveReview(t, svc, 7, "POST", "/courses/1/reviews", body)
	if status != 409 || parsed["code"] != "ALREADY_REVIEWED" {
		t.Fatalf("ulasan kedua: dapat %d %v, mau 409 ALREADY_REVIEWED", status, parsed["code"])
	}
}

// TestReviewServiceCreateRejectsNonParticipant memastikan hanya peserta kelas
// yang boleh menulis ulasan (403 NOT_ENROLLED).
func TestReviewServiceCreateRejectsNonParticipant(t *testing.T) {
	t.Parallel()

	svc := NewReviewService(newFakeReviewRepo(), newFakeReviewCourses())

	status, parsed := serveReview(t, svc, 7, "POST", "/courses/2/reviews",
		`{"rating":4,"comment":"Komentar yang cukup panjang untuk lolos."}`)
	if status != 403 || parsed["code"] != "NOT_ENROLLED" {
		t.Fatalf("dapat %d %v, mau 403 NOT_ENROLLED", status, parsed["code"])
	}
}

// TestReviewServiceListFiltersByTag memastikan filter ?tag bekerja dan tag
// dibandingkan setelah dinormalisasi.
func TestReviewServiceListFiltersByTag(t *testing.T) {
	t.Parallel()

	repo := newFakeReviewRepo()
	svc := NewReviewService(repo, newFakeReviewCourses())

	if status, _ := serveReview(t, svc, 7, "POST", "/courses/1/reviews",
		`{"rating":5,"comment":"Komentar pertama yang cukup panjang.","tags":["Seru"]}`); status != 201 {
		t.Fatalf("ulasan pertama: status = %d, mau 201", status)
	}
	if status, _ := serveReview(t, svc, 8, "POST", "/courses/1/reviews",
		`{"rating":3,"comment":"Komentar kedua yang cukup panjang.","tags":["sql"]}`); status != 201 {
		t.Fatalf("ulasan kedua: status = %d, mau 201", status)
	}

	status, body := serveReview(t, svc, 7, "GET", "/courses/1/reviews?tag=seru", "")
	if status != 200 {
		t.Fatalf("status = %d, mau 200 (body: %v)", status, body)
	}
	data := body["data"].([]any)
	if len(data) != 1 {
		t.Fatalf("ulasan dengan tag \"seru\" = %d, mau 1", len(data))
	}
}

// TestReviewServiceListUsesBeforeCursor memastikan cursor ?before memangkas
// ulasan sebelum id tersebut.
func TestReviewServiceListUsesBeforeCursor(t *testing.T) {
	t.Parallel()

	repo := newFakeReviewRepo()
	svc := NewReviewService(repo, newFakeReviewCourses())

	if status, _ := serveReview(t, svc, 7, "POST", "/courses/1/reviews",
		`{"rating":5,"comment":"Komentar pertama yang cukup panjang."}`); status != 201 {
		t.Fatalf("ulasan pertama: status = %d, mau 201", status)
	}
	if status, _ := serveReview(t, svc, 8, "POST", "/courses/1/reviews",
		`{"rating":4,"comment":"Komentar kedua yang cukup panjang."}`); status != 201 {
		t.Fatalf("ulasan kedua: status = %d, mau 201", status)
	}

	list, err := repo.FindByCourse(context.Background(), 1)
	if err != nil {
		t.Fatalf("FindByCourse: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("jumlah ulasan = %d, mau 2", len(list))
	}

	status, body := serveReview(t, svc, 7, "GET", "/courses/1/reviews?before="+list[0].ID, "")
	if status != 200 {
		t.Fatalf("status = %d, mau 200 (body: %v)", status, body)
	}
	if got := len(body["data"].([]any)); got != 1 {
		t.Errorf("ulasan setelah cursor = %d, mau 1", got)
	}
}

// TestReviewServiceFakeIsolatedPerTest memastikan fake repository dibuat per
// test: fake yang baru selalu kosong, tidak mewarisi data test lain
// (temuan B1 #4 — fake global bocor antar test).
func TestReviewServiceFakeIsolatedPerTest(t *testing.T) {
	t.Parallel()

	repo := newFakeReviewRepo()

	got, err := repo.FindByCourse(context.Background(), 1)
	if err != nil {
		t.Fatalf("FindByCourse: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("fake baru berisi %d ulasan, mau 0 (state bocor dari test lain)", len(got))
	}
}
