package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
)

type PrestasiRepository struct {
	// DB adalah connection pool; aman untuk request bersamaan.
	DB *pgxpool.Pool
}

var prestasiSortColumns = map[string]bool{
	"id":         true,
	"student_id": true,
	"keterangan": true,
	"juara":      true,
	"created_at": true,
}

func buildPrestasiWhere(q model.ListQuery) (string, []any) {
	var (
		clauses []string
		args    []any
	)
	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		clauses = append(clauses, fmt.Sprintf("keterangan ILIKE $%d", len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " WHERE " + strings.Join(clauses, " AND "), args
}

// GetAll mengambil daftar prestasi (paginasi + filter + urutan).
func (r PrestasiRepository) GetAll(ctx context.Context, q model.ListQuery) ([]model.Prestasi, int, error) {
	where, args := buildPrestasiWhere(q)

	var total int
	if err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM prestasi"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("hitung prestasi: %w", err)
	}

	sort := q.Sort
	if !prestasiSortColumns[sort] {
		sort = "id"
	}
	order := "ASC"
	if strings.ToLower(q.Order) == "desc" {
		order = "DESC"
	}

	args = append(args, q.Limit, q.Offset())
	sqlText := fmt.Sprintf(
		"SELECT id, student_id, keterangan, juara FROM prestasi%s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		where, sort, order, len(args)-1, len(args),
	)

	rows, err := r.DB.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("ambil daftar prestasi: %w", err)
	}
	defer rows.Close()

	out := []model.Prestasi{}
	for rows.Next() {
		var p model.Prestasi
		if err := rows.Scan(&p.ID, &p.StudentID, &p.Keterangan, &p.Juara); err != nil {
			return nil, 0, fmt.Errorf("baca baris prestasi: %w", err)
		}
		out = append(out, p)
	}
	return out, total, rows.Err()
}


// GetByStudentID mengambil semua prestasi milik seorang mahasiswa.
func (r PrestasiRepository) GetByStudentID(ctx context.Context, studentID int) ([]model.Prestasi, error) {
	rows, err := r.DB.Query(ctx,
		`SELECT id, student_id, keterangan, juara FROM prestasi WHERE student_id = $1 ORDER BY id`,
		studentID,
	)
	if err != nil {
		return nil, fmt.Errorf("ambil prestasi mahasiswa: %w", err)
	}
	defer rows.Close()

	out := []model.Prestasi{}
	for rows.Next() {
		var p model.Prestasi
		if err := rows.Scan(&p.ID, &p.StudentID, &p.Keterangan, &p.Juara); err != nil {
			return nil, fmt.Errorf("baca baris prestasi: %w", err)
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Create menyimpan prestasi baru. Mengembalikan ErrStudentNotFound bila
// student_id tidak ada (pelanggaran foreign key).
func (r PrestasiRepository) Create(ctx context.Context, p model.Prestasi) (model.Prestasi, error) {
	err := r.DB.QueryRow(ctx,
		`INSERT INTO prestasi (student_id, keterangan, juara)
		 VALUES ($1, $2, $3)
		 RETURNING id`,
		p.StudentID, p.Keterangan, p.Juara,
	).Scan(&p.ID)

	if err != nil {
		if isForeignKey(err) {
			return model.Prestasi{}, ErrStudentNotFound
		}
		return model.Prestasi{}, fmt.Errorf("simpan prestasi: %w", err)
	}
	return p, nil
}

