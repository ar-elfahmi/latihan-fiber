package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
)

type StudentRepository struct {
	// DB adalah connection pool; aman untuk request bersamaan.
	DB *pgxpool.Pool
}

// sortColumns whitelist kolom ORDER BY supaya input mentah tidak pernah
// disisipkan ke SQL.
var sortColumns = map[string]bool{
	"id": true, "nim": true, "name": true, "grade": true, "created_at": true,
}

// List mengambil halaman pertama student dengan urutan id DESC.
// Limit + 1 dipakai untuk deteksi has_more di service.
func (r StudentRepository) List(ctx context.Context, limit int) ([]model.Student, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students
		ORDER BY id DESC
		LIMIT $1
	`, limit+1)
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	defer rows.Close()
	return scanStudents(rows)
}

// FindAfterCursor mengambil daftar student dengan id < cursorID,
// urutan id DESC. Field cursorTime diabaikan (student pakai id saja).
func (r StudentRepository) FindAfterCursor(ctx context.Context, cursorID, limit int) ([]model.Student, error) {
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	rows, err := r.DB.Query(ctx, `
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students
		WHERE id < $1
		ORDER BY id DESC
		LIMIT $2
	`, cursorID, limit+1)
	if err != nil {
		return nil, fmt.Errorf("cursor students: %w", err)
	}
	defer rows.Close()
	return scanStudents(rows)
}

func scanStudents(rows pgx.Rows) ([]model.Student, error) {
	var students []model.Student
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan student: %w", err)
		}
		students = append(students, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi student: %w", err)
	}
	return students, nil
}

// buildWhere menyusun klausa WHERE + daftar argumen dari ListQuery.
// Mengembalikan " WHERE ..." (termasuk spasi) atau string kosong.
func buildWhere(q model.ListQuery) (string, []any) {
	var (
		clauses []string
		args    []any
	)

	if q.Search != "" {
		args = append(args, "%"+strings.ToLower(q.Search)+"%")
		clauses = append(clauses, "(LOWER(name) LIKE $"+fmt.Sprint(len(args))+" OR LOWER(nim) LIKE $"+fmt.Sprint(len(args))+")")
	}
	if q.IsActive != nil {
		args = append(args, *q.IsActive)
		clauses = append(clauses, "is_active = $"+fmt.Sprint(len(args)))
	}

	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

// GetAllWithFilter tetap dipakai path lain yang butuh search/is_active.
// Cursor path pakai List + FindAfterCursor.
func (r StudentRepository) GetAllWithFilter(ctx context.Context, q model.ListQuery, limit int) ([]model.Student, error) {
	where, args := buildWhere(q)
	rows, err := r.DB.Query(ctx, fmt.Sprintf(`
		SELECT id, nim, name, grade, is_active, owner_id, created_at
		FROM students%s
		ORDER BY id DESC
		LIMIT $%d
	`, where, len(args)+1), append(args, limit+1)...)
	if err != nil {
		return nil, fmt.Errorf("filter students: %w", err)
	}
	defer rows.Close()
	return scanStudents(rows)
}

// GetByID mengambil satu mahasiswa berdasarkan id.
// Mengembalikan ErrNotFound bila tidak ada baris.
func (r StudentRepository) GetByID(ctx context.Context, id int) (model.Student, error) {
	var s model.Student
	err := r.DB.QueryRow(ctx,
		`SELECT id, nim, name, grade, is_active, owner_id, created_at FROM students WHERE id = $1`,
		id,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		return model.Student{}, fmt.Errorf("ambil student: %w", err)
	}
	return s, nil
}

// Create menyimpan mahasiswa baru.
// Modul 6: owner_id dikirim dari service (bukan body). Kalau nil ->
// disimpan NULL (record "global" yang hanya boleh diakses via :any).
// Mengembalikan ErrDuplicate bila NIM sudah ada.
func (r StudentRepository) Create(ctx context.Context, s model.Student) (model.Student, error) {
	// Parameter nil untuk owner akan disimpan pgx sebagai NULL di kolom
	// owner_id (record "belum ada pemilik"). Parameter yang dipakai:
	// $1=NIM, $2=name, $3=grade, $4=is_active, $5=*int (owner) atau nil.
	var owner any
	if s.OwnerID != nil {
		owner = *s.OwnerID
	} else {
		owner = nil
	}
	err := r.DB.QueryRow(ctx,
		`INSERT INTO students (nim, name, grade, is_active, owner_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, owner,
	).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		if isDuplicate(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("simpan student: %w", err)
	}
	return s, nil
}

// Replace memperbarui seluruh field mahasiswa. Mengembalikan ErrNotFound / ErrDuplicate.
func (r StudentRepository) Replace(ctx context.Context, s model.Student) (model.Student, error) {
	err := r.DB.QueryRow(ctx,
		`UPDATE students
		 SET nim = $1, name = $2, grade = $3, is_active = $4
		 WHERE id = $5
		 RETURNING id, nim, name, grade, is_active, owner_id, created_at`,
		s.NIM, s.Name, s.Grade, s.IsActive, s.ID,
	).Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Student{}, ErrNotFound
		}
		if isDuplicate(err) {
			return model.Student{}, ErrDuplicate
		}
		return model.Student{}, fmt.Errorf("update student: %w", err)
	}
	return s, nil
}

// Delete menghapus mahasiswa berdasarkan id.
func (r StudentRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.DB.Exec(ctx, `DELETE FROM students WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("hapus student: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
