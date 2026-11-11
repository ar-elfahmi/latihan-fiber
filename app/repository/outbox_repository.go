package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"latihan-fiber/app/model"
)

// DBTX menerima pool PostgreSQL atau transaksi yang sedang berjalan.
type DBTX interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

// PostgresOutboxRepository adalah implementasi OutboxRepository di atas
// PostgreSQL. Field DB bisa pool atau transaksi berjalan.
type PostgresOutboxRepository struct {
	DB DBTX
}

// WithTx mengembalikan repository yang terikat ke transaksi berjalan.
func (r PostgresOutboxRepository) WithTx(tx pgx.Tx) OutboxRepository {
	return PostgresOutboxRepository{DB: tx}
}

func (r PostgresOutboxRepository) Insert(ctx context.Context, eventType string, event model.EnrollmentEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = r.DB.Exec(ctx, "INSERT INTO outbox (event_type, payload) VALUES ($1, $2)", eventType, payload)
	return err
}

func (r PostgresOutboxRepository) FetchUnpublished(ctx context.Context, afterID int64, limit int) ([]model.OutboxEvent, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, event_type, payload, created_at
		FROM outbox WHERE published_at IS NULL AND id > $1
		ORDER BY id LIMIT $2 FOR UPDATE SKIP LOCKED`, afterID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	events := []model.OutboxEvent{}
	for rows.Next() {
		var e model.OutboxEvent
		if err := rows.Scan(&e.ID, &e.EventType, &e.Payload, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (r PostgresOutboxRepository) MarkPublished(ctx context.Context, ids []int64) error {
	_, err := r.DB.Exec(ctx, "UPDATE outbox SET published_at = NOW() WHERE id = ANY($1)", ids)
	return err
}

func (r PostgresOutboxRepository) Stats(ctx context.Context) (int, int, time.Duration, error) {
	var pending, published int
	var seconds float64
	err := r.DB.QueryRow(ctx, `SELECT COUNT(*) FILTER (WHERE published_at IS NULL),
		COUNT(*) FILTER (WHERE published_at IS NOT NULL),
		COALESCE(EXTRACT(EPOCH FROM NOW() - MIN(created_at) FILTER (WHERE published_at IS NULL)), 0)
		FROM outbox`).Scan(&pending, &published, &seconds)
	return pending, published, time.Duration(seconds * float64(time.Second)), err
}
