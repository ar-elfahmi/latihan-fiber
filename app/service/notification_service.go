package service

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/helper"
)

type NotificationService struct {
	Notifications repository.NotificationRepository
	Outbox        repository.OutboxRepository
	Inspector     *asynq.Inspector
}

// Mine mengembalikan notifikasi milik user yang sedang login.
//
// @Summary		Daftar notifikasi saya
// @Description	Mengembalikan notifikasi milik user yang sedang login (20 terbaru).
// @Tags			Notifications
// @Produce		json
// @Success		200	{array}		model.Notification
// @Failure		401	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/notifications/mine [get]
func (s NotificationService) Mine(c *fiber.Ctx) error {
	userID := CurrentUserID(c)
	if userID == 0 {
		return helper.Unauthorized("belum terautentikasi")
	}
	list, err := s.Notifications.FindByUser(c.UserContext(), userID, 20)
	if err != nil {
		return helper.Internal(err)
	}
	return c.JSON(list)
}

// Monitor mengembalikan statistik antrean (outbox + Redis).
//
// @Summary		Statistik antrean
// @Description	Mengembalikan statistik outbox (pending/published), jumlah notifikasi terkirim, dan info antrean Redis.
// @Tags			Notifications
// @Produce		json
// @Success		200	{object}	model.QueueStats
// @Failure		403	{object}	model.ErrorResponse
// @Security		BearerAuth
// @Router			/admin/queue [get]
func (s NotificationService) Monitor(c *fiber.Ctx) error {
	ctx := c.UserContext()
	pending, published, oldest, err := s.Outbox.Stats(ctx)
	if err != nil {
		return helper.Internal(err)
	}
	sent, err := s.Notifications.Count(ctx)
	if err != nil {
		return helper.Internal(err)
	}

	stats := model.QueueStats{
		OutboxPending:     pending,
		OutboxPublished:   published,
		OldestPendingAge:  oldest.Round(time.Millisecond).String(),
		NotificationsSent: sent,
	}
	info, err := s.Inspector.GetQueueInfo(model.QueueNotifications)
	if err != nil {
		stats.QueueError = "queue tidak dapat dibaca: Redis tidak dapat dihubungi atau queue belum pernah dipakai"
	} else {
		stats.Queue = &model.QueueInfo{
			Name:      info.Queue,
			Pending:   info.Pending,
			Active:    info.Active,
			Scheduled: info.Scheduled,
			Retry:     info.Retry,
			Archived:  info.Archived,
			Completed: info.Completed,
			Processed: info.Processed,
			Failed:    info.Failed,
		}
	}
	return c.JSON(stats)
}
