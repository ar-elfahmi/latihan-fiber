package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type NotificationWorker struct {
	Users         repository.UserRepository
	Notifications repository.NotificationRepository
	Mailer        helper.Mailer
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

	user, err := w.Users.FindByID(ctx, event.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return fmt.Errorf("%w: user %d tidak ditemukan", asynq.SkipRetry, event.UserID)
	}
	if err != nil {
		return err
	}
	message, err := BuildMessage(job.EventType, event, user)
	if err != nil {
		return fmt.Errorf("%w: %v", asynq.SkipRetry, err)
	}

	alreadySent := false
	err = pgx.BeginFunc(ctx, w.Users.DB, func(tx pgx.Tx) error {
		claimed, err := w.Notifications.Claim(ctx, model.Notification{
			OutboxID: job.OutboxID, UserID: user.ID, Channel: "email",
			Subject: message.Subject, Body: message.Body,
		})
		if err != nil {
			return err
		}
		if !claimed {
			alreadySent = true
			return errors.New("claim notifikasi berhasil dibuat")
		}
		return w.Mailer.Send(ctx, message)
	})
	if errors.Is(err, helper.ErrPermanentDelivery) {
		return fmt.Errorf("%v: %w", asynq.SkipRetry, err)
	}
	if err != nil {
		return err
	}

	retry, _ := asynq.GetRetryCount(ctx)
	log.Printf("notification_processed outbox_id=%d event=%s retry=%d already_sent=%t",
		job.OutboxID, job.EventType, retry, alreadySent)
	return nil
}
