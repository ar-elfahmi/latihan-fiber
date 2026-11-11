package repository

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoPool dikembalikan saat TxManager dipakai tanpa connection pool.
var ErrNoPool = errors.New("transaction manager tanpa connection pool")

// beginTx membuka transaksi baru. Berisi *pgxpool.Pool.Begin pada
// implementasi nyata; unit test menyuntikkan versi tiruan agar
// perilaku rollback/commit bisa diuji tanpa database.
type beginTx func(ctx context.Context) (pgx.Tx, error)

// TxManager membungkus connection pool PostgreSQL untuk menjalankan
// operasi lintas tabel dalam satu transaksi.
//
// Transaction manager TIDAK dipakai service pada slice S-002; kontrak ini
// disiapkan agar slice berikutnya (S-003+) bisa memakai satu transaksi
// tanpa mengubah signature repository.
type TxManager struct {
	begin beginTx
}

// NewTxManager membuat TxManager di atas pool PostgreSQL.
// Pool boleh nil; pemanggilan WithTx pada manager tanpa pool akan
// mengembalikan error terkontrol (bukan panic).
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	if pool == nil {
		return &TxManager{}
	}
	return &TxManager{begin: pool.Begin}
}

// WithTx menjalankan fn di dalam satu transaksi.
//
// Perilaku:
//   - fn mengembalikan error -> transaksi di-rollback, error dikembalikan.
//   - fn panic -> transaksi di-rollback, panic diubah menjadi error.
//   - fn sukses -> transaksi di-commit.
//
// Koneksi selalu dikembalikan ke pool: commit, rollback, maupun panic.
func (m *TxManager) WithTx(ctx context.Context, fn func(tx pgx.Tx) error) (err error) {
	if m == nil || m.begin == nil {
		return ErrNoPool
	}

	tx, err := m.begin(ctx)
	if err != nil {
		return fmt.Errorf("mulai transaksi: %w", err)
	}

	// Rollback selalu dijadwalkan lebih dulu: kalau fn commit berhasil,
	// transaksi sudah ditutup dan rollback berikutnya jadi no-op.
	committed := false
	defer func() {
		if committed {
			return
		}
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			log.Printf("rollback transaksi gagal: %v", rbErr)
		}
	}()

	defer func() {
		if rec := recover(); rec != nil {
			err = fmt.Errorf("transaksi dibatalkan (panic): %v", rec)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaksi: %w", err)
	}
	committed = true
	return nil
}
