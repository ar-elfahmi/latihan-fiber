package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RBACRepository membaca peta peran -> permission dari database.
// Setiap query berparameter; nama tabel/kolom hardcoded (bukan input user).
type RBACRepository struct {
	DB *pgxpool.Pool
}

// GetPermissionsByUserID mengembalikan himpunan permission (nama) yang dimiliki user.
// Hasil dipakai middleware RequirePermission untuk O(1) lookup.
// Kalau user tak punya role, hasilnya kosong (bukan error) -> fail closed di middleware.
func (r RBACRepository) GetPermissionsByUserID(ctx context.Context, userID int) (map[string]bool, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT DISTINCT rp.permission_name
		FROM role_permissions rp
		JOIN user_roles ur ON ur.role_name = rp.role_name
		WHERE ur.user_id = $1
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ambil permission user: %w", err)
	}
	defer rows.Close()

	perms := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		perms[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterasi permission: %w", err)
	}
	return perms, nil
}

// GetRolesByUserID mengembalikan daftar nama role untuk user.
// Dipakai diagnostik / laporan; tidak wajib untuk hot-path.
func (r RBACRepository) GetRolesByUserID(ctx context.Context, userID int) ([]string, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT role_name FROM user_roles WHERE user_id = $1 ORDER BY role_name
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("ambil role user: %w", err)
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		roles = append(roles, name)
	}
	return roles, rows.Err()
}
