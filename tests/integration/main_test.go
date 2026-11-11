//go:build integration

// Package integration_test berisi integration test: test yang berbicara
// dengan PostgreSQL sungguhan (database uji berakhiran _test).
//
// File ini menyiapkan segalanya sekali saja lewat TestMain:
//   - membaca TEST_DATABASE_URL dari environment (WAJIB diisi),
//   - menolak database yang namanya tidak berakhiran _test,
//   - membuka pool dengan MaxConns cukup besar supaya uji konkurensi
//     benar-benar konkuren,
//   - menjalankan migration 001..010 (idempoten).
package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
)

// pool dipakai bersama oleh seluruh test di package ini.
var pool *pgxpool.Pool

// seqCounter membuat kode/nama seed selalu unik antar test.
var seqCounter atomic.Int64

func TestMain(m *testing.M) {
	// DSN test dibaca dari environment. Kosong = ERROR, bukan "dilewati":
	// test yang dilewati diam-diam dianggap lulus (temuan B1 #5).
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		fmt.Println("TEST_DATABASE_URL tidak diisi: integration test tidak bisa dijalankan")
		os.Exit(1)
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		fmt.Println("TEST_DATABASE_URL tidak valid:", err)
		os.Exit(1)
	}

	// PENGAMAN. Integration test menulis dan menghapus data. Test hanya
	// boleh berjalan pada database yang namanya berakhiran _test.
	if !strings.HasSuffix(cfg.ConnConfig.Database, "_test") {
		fmt.Printf("menolak berjalan: database %q tidak berakhiran _test\n", cfg.ConnConfig.Database)
		os.Exit(1)
	}

	// Cukup koneksi agar test konkurensi benar-benar berjalan serentak.
	// MaxConns = 1 membuat semua request saling menunggu (temuan B1 #6).
	cfg.MaxConns = 10

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err = pgxpool.NewWithConfig(ctx, cfg)
	if err == nil {
		err = pool.Ping(ctx)
	}
	if err != nil {
		fmt.Println("gagal terhubung ke database test:", err)
		os.Exit(1)
	}

	if err := applyMigrations(ctx); err != nil {
		fmt.Println("gagal menjalankan migration:", err)
		os.Exit(1)
	}

	code := m.Run()
	pool.Close()
	os.Exit(code)
}

// applyMigrations menjalankan seluruh file di folder migrations secara
// berurutan. Migration yang sudah pernah diterapkan boleh gagal diulang
// (mis. 009 menjatuhkan constraint yang sudah tidak ada); skema akhir
// diverifikasi terpisah oleh verifySchema.
func applyMigrations(ctx context.Context) error {
	dir := filepath.Join("..", "..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("membaca folder migration: %w", err)
	}

	names := []string{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		script, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("membaca %s: %w", name, err)
		}
		if err := execScript(ctx, string(script)); err != nil {
			fmt.Printf("migration %s dilewati (sudah diterapkan?): %v\n", name, err)
		}
	}
	return verifySchema(ctx)
}

// execScript menjalankan satu file SQL (bisa berisi banyak statement).
func execScript(ctx context.Context, script string) error {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	results, err := conn.Conn().PgConn().Exec(ctx, script).ReadAll()
	if err != nil {
		// Transaction yang gagal dibatalkan supaya koneksi kembali bersih.
		_, _ = conn.Conn().PgConn().Exec(ctx, "ROLLBACK").ReadAll()
		return err
	}
	for _, r := range results {
		if r.Err != nil {
			_, _ = conn.Conn().PgConn().Exec(ctx, "ROLLBACK").ReadAll()
			return r.Err
		}
	}
	return nil
}

// verifySchema memastikan objek yang dipakai test benar-benar ada.
func verifySchema(ctx context.Context) error {
	required := []string{"courses", "enrollments", "outbox", "notifications", "users"}
	for _, table := range required {
		var exists bool
		err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1)`, table).Scan(&exists)
		if err != nil {
			return fmt.Errorf("cek tabel %s: %w", table, err)
		}
		if !exists {
			return fmt.Errorf("tabel %s tidak ada di database test", table)
		}
	}

	var emailExists bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM information_schema.columns
			WHERE table_name = 'users' AND column_name = 'email')`).Scan(&emailExists); err != nil {
		return fmt.Errorf("cek kolom users.email: %w", err)
	}
	if !emailExists {
		return fmt.Errorf("kolom users.email tidak ada di database test")
	}

	var uniqueOutbox bool
	if err := pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM pg_constraint c
			JOIN pg_attribute a ON a.attrelid = c.conrelid AND a.attnum = ANY(c.conkey)
			WHERE c.conrelid = 'notifications'::regclass
			  AND c.contype = 'u'
			  AND a.attname = 'outbox_id'
			  AND array_length(c.conkey, 1) = 2)`).Scan(&uniqueOutbox); err != nil {
		return fmt.Errorf("cek constraint notifications: %w", err)
	}
	if !uniqueOutbox {
		return fmt.Errorf("constraint unik notifications(outbox_id, user_id) tidak ada di database test")
	}
	return nil
}

// seedCourse membuat satu kelas uji. Pembersihan memakai t.Cleanup, bukan
// defer di dalam helper: defer akan menghapus data sebelum test memakai
// nya (temuan B1 #7).
func seedCourse(t *testing.T, credits, quota int) model.Course {
	t.Helper()

	ctx := context.Background()
	code := fmt.Sprintf("TST-%d-%d", time.Now().UnixNano(), seqCounter.Add(1))

	var course model.Course
	err := pool.QueryRow(ctx, `INSERT INTO courses (code, name, credits, quota)
		VALUES ($1, $2, $3, $4)
		RETURNING id, code, name, credits, quota, enrolled_count, version, created_at`,
		code, "Kelas Uji "+code, credits, quota).
		Scan(&course.ID, &course.Code, &course.Name, &course.Credits, &course.Quota,
			&course.EnrolledCount, &course.Version, &course.CreatedAt)
	if err != nil {
		t.Fatalf("seed course: %v", err)
	}

	t.Cleanup(func() { cleanupCourse(t, course.ID) })
	return course
}

// cleanupCourse menghapus kelas uji (enrollment ikut terhapus via cascade).
func cleanupCourse(t *testing.T, courseID int) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), "DELETE FROM courses WHERE id = $1", courseID); err != nil {
		t.Errorf("membersihkan course %d: %v", courseID, err)
	}
}

// seedStudent membuat satu akun uji.
func seedStudent(t *testing.T) model.User {
	t.Helper()

	ctx := context.Background()
	username := fmt.Sprintf("uji_%d_%d", time.Now().UnixNano(), seqCounter.Add(1))

	var user model.User
	err := pool.QueryRow(ctx, `INSERT INTO users (username, password_hash, email)
		VALUES ($1, $2, $3) RETURNING id, username, email`,
		username, "hash-uji", username+"@uji.local").
		Scan(&user.ID, &user.Username, &user.Email)
	if err != nil {
		t.Fatalf("seed student: %v", err)
	}

	t.Cleanup(func() { cleanupStudent(t, user.ID) })
	return user
}

// cleanupStudent menghapus seluruh jejak data uji milik satu user.
func cleanupStudent(t *testing.T, userID int) {
	t.Helper()
	ctx := context.Background()

	if _, err := pool.Exec(ctx, "DELETE FROM enrollments WHERE user_id = $1", userID); err != nil {
		t.Errorf("membersihkan enrollment user %d: %v", userID, err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM notifications WHERE user_id = $1", userID); err != nil {
		t.Errorf("membersihkan notifications user %d: %v", userID, err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM outbox WHERE payload->>'user_id' = $1",
		strconv.Itoa(userID)); err != nil {
		t.Errorf("membersihkan outbox user %d: %v", userID, err)
	}
	if _, err := pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID); err != nil {
		t.Errorf("membersihkan user %d: %v", userID, err)
	}
}

// newAPI membuat aplikasi Fiber dengan route KRS yang memakai repository
// PostgreSQL sungguhan. Identitas user diambil dari header X-Test-User
// supaya test tidak perlu login (yang diuji adalah KRS + database).
func newAPI(t *testing.T) *fiber.App {
	t.Helper()

	repo, err := repository.NewPostgresCourseRepository(pool, repository.CreditRules{
		WithinCreditLimit: service.WithinCreditLimit,
		CreditsAfterSwap:  service.CreditsAfterSwap,
	})
	if err != nil {
		t.Fatalf("membuat repository kelas: %v", err)
	}

	course := service.CourseService{Repo: repo}

	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Use(func(c *fiber.Ctx) error {
		id, err := strconv.Atoi(c.Get("X-Test-User"))
		if err != nil {
			return fmt.Errorf("header X-Test-User wajib berisi id user")
		}
		c.Locals("user_id", id)
		return c.Next()
	})
	app.Post("/courses/:id/enroll", course.Enroll)
	app.Delete("/courses/:id/enroll", course.Drop)
	app.Post("/enrollments/swap", course.Swap)
	return app
}

// request mengirim request HTTP ke app dan mengembalikan status + body JSON.
func request(t *testing.T, app *fiber.App, method, path string, userID int, body string) (int, map[string]any) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", strconv.Itoa(userID))

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Errorf("request %s %s: %v", method, path, err)
		return 0, map[string]any{}
	}
	defer resp.Body.Close()

	parsed := map[string]any{}
	if resp.StatusCode != fiber.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
			t.Errorf("response %s %s bukan JSON: %v", method, path, err)
		}
	}
	return resp.StatusCode, parsed
}

// countRows menghitung baris hasil query (parameterized).
func countRows(t *testing.T, sql string, args ...any) int {
	t.Helper()

	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatalf("query %q: %v", sql, err)
	}
	return n
}
