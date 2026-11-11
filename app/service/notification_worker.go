package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type NotificationWorker struct {
	Users         repository.UserRepository
	Notifications repository.NotificationRepository
	// Pool dipakai untuk membuka transaksi claim+mail. Repo di atas hanya
	// interface, jadi pool tidak bisa diambil dari sana.
	Pool   *pgxpool.Pool
	Mailer helper.Mailer
}

func (w NotificationWorker) HandleSend(ctx context.Context, task *asynq.Task) error {
	var job model.NotificationJob
	if err := json.Unmarshal(task.Payload(), &job); err != nil {
		return fmt.Errorf("%w: payload task rusak: %v", asynq.SkipRetry, err)
	}
	var event model.EnrollmentEvent
	if err := json.Unmarshal(job.Payload, &event); err != nil {
		return fmt.Errorf("%w: payload event %d rusak: %v", asynq.SkipRetry, job.OutboxID, err)
	}
	if job.EventType == model.EventCourseQuotaChanged {
		var permanentErr error
		for _, userID := range event.RecipientIDs {
			if err := w.send(ctx, job, event, userID); err != nil {
				if !errors.Is(err, asynq.SkipRetry) {
					return err
				}
				permanentErr = errors.Join(permanentErr, err)
			}
		}
		return permanentErr
	}
	return w.send(ctx, job, event, event.UserID)
}

func (w NotificationWorker) send(ctx context.Context, job model.NotificationJob, event model.EnrollmentEvent, userID int) error {
	user, err := w.Users.FindByID(ctx, userID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, userID)
	}
	if err != nil {
		return err
	}
	message, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = pgx.BeginFunc(ctx, w.Pool, func(tx pgx.Tx) error {
		claimed, err := w.Notifications.WithTx(tx).Claim(ctx, model.Notification{
			OutboxID: job.OutboxID, UserID: user.ID, Channel: "email",
			Subject: message.Subject, Body: message.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return nil
		}
		return w.Mailer.Send(ctx, message)
	})
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	log.Printf("notification_processed outbox_id=%d event=%s retry=%d already_sent=%t",
		job.OutboxID, job.EventType, retry, alreadySent)
	return nil
}
