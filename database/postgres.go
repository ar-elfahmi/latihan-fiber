package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/config"
)

// Connect membuka connection pool ke PostgreSQL memakai DSN dari
// environment variable (config.DatabaseDSN).
//
// Pakai pgxpool (bukan pgx.Conn) karena pool AMAN untuk request bersamaan —
// Fiber jalankan tiap request di goroutine terpisah. Jika pakai pgx.Conn
// tunggal, request yang bertabrakan akan error 500.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	return ConnectWithDSN(ctx, config.DatabaseDSN())
}

// ConnectWithDSN membuka connection pool ke PostgreSQL dengan DSN yang
// sudah jadi (dibaca dari config.NewApp). Dipakai main.go supaya DSN
// hanya disusun di satu tempat.
func ConnectWithDSN(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, dsn)
}
