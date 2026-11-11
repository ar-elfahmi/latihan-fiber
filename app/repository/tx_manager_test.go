package repository

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// fakeTx adalah tiruan pgx.Tx yang mencatat commit/rollback.
type fakeTx struct {
	committed   bool
	rolledBack  bool
	commitErr   error
	rollbackErr error
}

func (f *fakeTx) Begin(context.Context) (pgx.Tx, error) { return f, nil }

func (f *fakeTx) Commit(context.Context) error {
	f.committed = true
	return f.commitErr
}

func (f *fakeTx) Rollback(context.Context) error {
	f.rolledBack = true
	return f.rollbackErr
}

func (f *fakeTx) CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error) {
	return 0, nil
}

func (f *fakeTx) SendBatch(context.Context, *pgx.Batch) pgx.BatchResults { return nil }
func (f *fakeTx) LargeObjects() pgx.LargeObjects                         { return pgx.LargeObjects{} }

func (f *fakeTx) Prepare(context.Context, string, string) (*pgconn.StatementDescription, error) {
	return nil, nil
}

func (f *fakeTx) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return pgconn.CommandTag{}, nil
}

func (f *fakeTx) Query(context.Context, string, ...any) (pgx.Rows, error) { return nil, nil }
func (f *fakeTx) QueryRow(context.Context, string, ...any) pgx.Row        { return nil }
func (f *fakeTx) Conn() *pgx.Conn                                         { return nil }

// newTestTxManager membuat TxManager dengan beginTx tiruan supaya
// perilaku commit/rollback bisa diuji tanpa database.
func newTestTxManager() (*TxManager, *fakeTx) {
	fake := &fakeTx{}
	m := &TxManager{begin: func(context.Context) (pgx.Tx, error) { return fake, nil }}
	return m, fake
}

// TestNewTxManagerNilPool memastikan manager tanpa pool mengembalikan
// error terkontrol, bukan panic.
func TestNewTxManagerNilPool(t *testing.T) {
	t.Parallel()

	m := NewTxManager(nil)
	if m == nil {
		t.Fatal("NewTxManager(nil) harus mengembalikan manager, bukan nil")
	}

	err := m.WithTx(context.Background(), func(pgx.Tx) error {
		t.Error("fungsi callback tidak boleh dijalankan tanpa pool")
		return nil
	})
	if !errors.Is(err, ErrNoPool) {
		t.Fatalf("err = %v, mau ErrNoPool", err)
	}
}

// TestWithTxCommit memastikan transaksi di-commit saat callback sukses.
func TestWithTxCommit(t *testing.T) {
	t.Parallel()

	m, fake := newTestTxManager()
	called := false

	err := m.WithTx(context.Background(), func(pgx.Tx) error {
		called = true
		return nil
	})
	if err != nil {
		t.Fatalf("err = %v, mau nil", err)
	}
	if !called {
		t.Fatal("callback harus dijalankan")
	}
	if !fake.committed {
		t.Fatal("transaksi harus di-commit")
	}
	if fake.rolledBack {
		t.Fatal("transaksi sukses tidak boleh di-rollback")
	}
}

// TestWithTxRollbackOnError memastikan error dari callback memicu rollback
// dan error diteruskan ke pemanggil.
func TestWithTxRollbackOnError(t *testing.T) {
	t.Parallel()

	m, fake := newTestTxManager()
	sentinel := errors.New("gagal di tengah transaksi")

	err := m.WithTx(context.Background(), func(pgx.Tx) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, mau %v", err, sentinel)
	}
	if !fake.rolledBack {
		t.Fatal("transaksi harus di-rollback")
	}
	if fake.committed {
		t.Fatal("transaksi gagal tidak boleh di-commit")
	}
}

// TestWithTxRollbackOnPanic memastikan panic di dalam callback diubah
// menjadi error dan transaksi di-rollback (tidak panic keluar).
func TestWithTxRollbackOnPanic(t *testing.T) {
	t.Parallel()

	m, fake := newTestTxManager()

	err := m.WithTx(context.Background(), func(pgx.Tx) error {
		panic("bom")
	})
	if err == nil {
		t.Fatal("panic harus diubah menjadi error")
	}
	if !strings.Contains(err.Error(), "bom") {
		t.Fatalf("err = %v, mau memuat nilai panic", err)
	}
	if !fake.rolledBack {
		t.Fatal("transaksi harus di-rollback setelah panic")
	}
	if fake.committed {
		t.Fatal("transaksi yang panic tidak boleh di-commit")
	}
}

// TestWithTxBeginError memastikan error saat membuka transaksi
// diteruskan tanpa menjalankan callback.
func TestWithTxBeginError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("koneksi database gagal")
	m := &TxManager{begin: func(context.Context) (pgx.Tx, error) { return nil, sentinel }}

	err := m.WithTx(context.Background(), func(pgx.Tx) error {
		t.Error("callback tidak boleh dijalankan bila transaksi gagal dibuka")
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("err = %v, mau %v", err, sentinel)
	}
}

// TestWithTxCommitError memastikan error commit memicu rollback.
func TestWithTxCommitError(t *testing.T) {
	t.Parallel()

	m, fake := newTestTxManager()
	fake.commitErr = errors.New("commit ditolak")

	err := m.WithTx(context.Background(), func(pgx.Tx) error { return nil })
	if err == nil {
		t.Fatal("error commit harus dikembalikan")
	}
	if !fake.rolledBack {
		t.Fatal("transaksi harus di-rollback bila commit gagal")
	}
}
