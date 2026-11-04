package config

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
)

func GetInt(key string, fallback int) int {
	value, err := strconv.Atoi(Get(key))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func RedisOpt() asynq.RedisClientOpt {
	return asynq.RedisClientOpt{
		Addr:         GetWithDefault("REDIS_ADDR", "localhost:6379"),
		Password:     Get("REDIS_PASSWORD"),
		DB:           GetInt("REDIS_DB", 0),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	}
}

func NewWorkerServer() *asynq.Server {
	return asynq.NewServer(RedisOpt(), asynq.Config{
		Concurrency: GetInt("WORKER_CONCURRENCY", 5),
		Queues:      map[string]int{model.QueueNotifications: 1},
		RetryDelayFunc: func(n int, _ error, _ *asynq.Task) time.Duration {
			return time.Duration((n+1)*2) * time.Second
		},
		ShutdownTimeout: 10 * time.Second,
		ErrorHandler: asynq.ErrorHandlerFunc(func(ctx context.Context, task *asynq.Task, err error) {
			retry, _ := asynq.GetRetryCount(ctx)
			maxRetry, _ := asynq.GetMaxRetry(ctx)
			log.Printf("task_failed type=%s attempt=%d/%d error=%v", task.Type(), retry+1, maxRetry+1, err)
		}),
	})
}
