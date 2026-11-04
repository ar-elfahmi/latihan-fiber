package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
)

type OutboxRelay struct {
	DB       *pgxpool.Pool
	Queue    *asynq.Client
	Interval time.Duration
	Batch    int
}

func (r OutboxRelay) Run(ctx context.Context) {
	ticker := time.NewTicker(r.Interval)
	defer ticker.Stop()

	for ctx.Err() == nil {
		n, err := r.RelayOnce(ctx)
		if err != nil {
			log.Printf("relay_delayed published=%d error=%v", n, err)
		} else if n > 0 {
			log.Printf("relay_done published=%d", n)
		}
		if err == nil && n == r.Batch {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r OutboxRelay) RelayOnce(ctx context.Context) (int, error) {
	var published int
	var enqueueErr error

	err := pgx.BeginFunc(ctx, r.DB, func(tx pgx.Tx) error {
		outbox := repository.OutboxRepository{DB: tx}
		events, err := outbox.FetchUnpublished(ctx, r.Batch)
		if err != nil {
			return err
		}

		ids := make([]int64, 0, len(events))
		for _, event := range events {
			ids = append(ids, event.ID)
			if enqueueErr = r.publish(ctx, event); enqueueErr != nil {
				break
			}
		}
		if len(ids) > 0 {
			if err := outbox.MarkPublished(ctx, ids); err != nil {
				return err
			}
		}
		published = len(ids)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return published, enqueueErr
}

func (r OutboxRelay) publish(ctx context.Context, event model.OutboxEvent) error {
	job := model.NotificationJob{
		OutboxID:  event.ID,
		EventType: event.EventType,
		Payload:   event.Payload,
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal outbox %d: %w", event.ID, err)
	}
	task := asynq.NewTask(model.TaskSendNotification, payload,
		asynq.TaskID(fmt.Sprintf("outbox-%d", event.ID)),
		asynq.Queue(model.QueueNotifications),
		asynq.MaxRetry(5),
		asynq.Timeout(30*time.Second),
		asynq.Retention(24*time.Hour),
	)
	_, err = r.Queue.EnqueueContext(ctx, task)
	if errors.Is(err, asynq.ErrTaskIDConflict) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("enqueue outbox %d: %w", event.ID, err)
	}
	return nil
}
