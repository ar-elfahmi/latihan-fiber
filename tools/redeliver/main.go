package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
	"latihan-fiber/config"
	"latihan-fiber/database"
)

func main() {
	outboxID := flag.Int64("outbox", 0, "ID event outbox")
	copies := flag.Int("copies", 5, "jumlah salinan task")
	flag.Parse()
	if *outboxID < 1 || *copies < 1 {
		log.Fatal("outbox dan copies harus lebih dari nol")
	}
	config.LoadEnv()
	if config.Get("DB_NAME") != "praktikum_modul11_uji" {
		log.Fatal("redelivery hanya boleh dijalankan di praktikum_modul11_uji")
	}
	if config.RedisOpt().DB != 11 {
		log.Fatal("redelivery hanya boleh memakai Redis DB 11")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	job := model.NotificationJob{OutboxID: *outboxID}
	err = db.QueryRow(ctx, "SELECT event_type, payload FROM outbox WHERE id = $1", *outboxID).
		Scan(&job.EventType, &job.Payload)
	if err != nil {
		log.Fatal(err)
	}
	payload, err := json.Marshal(job)
	if err != nil {
		log.Fatal(err)
	}
	client := asynq.NewClient(config.RedisOpt())
	defer client.Close()
	for i := 0; i < *copies; i++ {
		task := asynq.NewTask(model.TaskSendNotification, payload)
		info, err := client.EnqueueContext(ctx, task, asynq.Queue(model.QueueNotifications),
			asynq.MaxRetry(5), asynq.Timeout(30*time.Second), asynq.Retention(24*time.Hour))
		if err != nil {
			log.Fatalf("terkirim %d dari %d salinan: %v", i, *copies, err)
		}
		fmt.Printf("outbox_id=%d task_id=%s\n", *outboxID, info.ID)
	}
}
