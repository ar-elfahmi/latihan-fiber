package repository

import (
	"context"

	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
)

// PostgresNotificationRepository adalah implementasi NotificationRepository
// di atas PostgreSQL. Field DB bisa pool atau transaksi berjalan.
type PostgresNotificationRepository struct {
	DB DBTX
}

// WithTx mengembalikan repository yang terikat ke transaksi berjalan.
func (r PostgresNotificationRepository) WithTx(tx pgx.Tx) NotificationRepository {
	return PostgresNotificationRepository{DB: tx}
}

func (r PostgresNotificationRepository) Claim(ctx context.Context, n model.Notification) (bool, error) {
	tag, err := r.DB.Exec(ctx, `INSERT INTO notifications (outbox_id, user_id, channel, subject, body)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT (outbox_id, user_id) DO NOTHING`,
		n.OutboxID, n.UserID, n.Channel, n.Subject, n.Body)
	return tag.RowsAffected() == 1, err
}

func (r PostgresNotificationRepository) FindByUser(ctx context.Context, userID, limit int) ([]model.Notification, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, outbox_id, user_id, channel, subject, body, sent_at
		FROM notifications WHERE user_id = $1 ORDER BY outbox_id DESC LIMIT $2`, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		if err := rows.Scan(&n.ID, &n.OutboxID, &n.UserID, &n.Channel, &n.Subject, &n.Body, &n.SentAt); err != nil {
			return nil, err
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func (r PostgresNotificationRepository) Count(ctx context.Context) (int, error) {
	var total int
	err := r.DB.QueryRow(ctx, "SELECT COUNT(*) FROM notifications").Scan(&total)
	return total, err
}
