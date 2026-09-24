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

// GetAll mengambil daftar mahasiswa dengan paginasi, pencarian, dan urutan.
func (r StudentRepository) GetAll(ctx context.Context, q model.ListQuery) ([]model.Student, int, error) {
	where, args := buildWhere(q)

	// Hitung total baris (untuk metadata paginasi).
	var total int
	if err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM students"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung student: %w", err)
	}

	// Pilih kolom ORDER BY hanya dari whitelist.
	sort := q.Sort
	if !sortColumns[sort] {
		sort = "id"
	}
	order := "ASC"
	if strings.ToLower(q.Order) == "desc" {
		order = "DESC"
	}

	// Tambahkan LIMIT/OFFSET di akhir query.
	args = append(args, q.Limit, q.Offset())
	sqlText := fmt.Sprintf(
		"SELECT id, nim, name, grade, is_active, owner_id, created_at FROM students%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sort, order, len(args)-1, len(args),
	)

	rows, err := r.DB.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ambil daftar student: %w", err)
	}
	defer rows.Close()

	students := []model.Student{}
	for rows.Next() {
		var s model.Student
		if err := rows.Scan(&s.ID, &s.NIM, &s.Name, &s.Grade, &s.IsActive, &s.OwnerID, &s.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("baca baris student: %w", err)
		}
		students = append(students, s)
	}
	return students, total, rows.Err()
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
