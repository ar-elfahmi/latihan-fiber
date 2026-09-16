package database

import (
	"context"
	"fmt"

	"latihan-fiber/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect membuka connection pool ke PostgreSQL dari variabel environment.
// Pakai pgxpool (bukan pgx.Conn) karena pool AMAN untuk request bersamaan —
// Fiber jalankan tiap request di goroutine terpisah. Jika pakai pgx.Conn
// tunggal, request yang bertabrakan akan error 500.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.Get("DB_USER"),
		config.Get("DB_PASSWORD"),
		config.Get("DB_HOST"),
		config.Get("DB_PORT"),
		config.Get("DB_NAME"),
		config.Get("DB_SSLMODE"),
	)
	return pgxpool.New(ctx, dsn)
}
