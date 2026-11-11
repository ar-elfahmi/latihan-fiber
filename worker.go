package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5/pgxpool"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/helper"
)

func startWorker(ctx context.Context, pool *pgxpool.Pool) (func(), error) {
	afterID, err := strconv.ParseInt(config.GetWithDefault("RELAY_AFTER_ID", "0"), 10, 64)
	if err != nil || afterID < 0 {
		return nil, fmt.Errorf("RELAY_AFTER_ID harus berupa bilangan bulat nonnegatif")
	}
	interval := config.GetInt("RELAY_INTERVAL_MS", 1000)
	batch := config.GetInt("RELAY_BATCH", 100)
	if interval == 0 || batch == 0 {
		return nil, fmt.Errorf("RELAY_INTERVAL_MS dan RELAY_BATCH harus lebih dari nol")
	}
	mailer, err := helper.NewLogMailer(
		"logs/mail.log",
		time.Duration(config.GetInt("MAILER_DELAY_MS", 1500))*time.Millisecond,
		config.GetInt("MAILER_FAIL_EVERY", 0),
	)
	if err != nil {
		return nil, err
	}

	worker := service.NotificationWorker{
		Users:         repository.PostgresUserRepository{DB: pool},
		Pool:          pool,
		Notifications: repository.NotificationRepository{DB: pool},
		Mailer:        mailer,
	}
	mux := asynq.NewServeMux()
	mux.HandleFunc(model.TaskSendNotification, worker.HandleSend)

	server := config.NewWorkerServer()
	if err := server.Start(mux); err != nil {
		return nil, fmt.Errorf("menyalakan worker: %w", err)
	}
	client := asynq.NewClient(config.RedisOpt())
	relay := service.OutboxRelay{
		DB:       pool,
		Queue:    client,
		Interval: time.Duration(interval) * time.Millisecond,
		Batch:    batch,
		AfterID:  afterID,
	}
	relayCtx, cancelRelay := context.WithCancel(ctx)
	relayDone := make(chan struct{})
	go func() {
		defer close(relayDone)
		relay.Run(relayCtx)
	}()
	log.Println("worker berjalan")

	return func() {
		cancelRelay()
		<-relayDone
		server.Shutdown()
		if err := client.Close(); err != nil {
			log.Printf("menutup client antrean: %v", err)
		}
	}, nil
}
