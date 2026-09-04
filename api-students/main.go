package main

import (
	"context"
	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"log"
)

func main() {
	config.LoadEnv()
	pool, e := database.NewPool(context.Background())
	if e != nil {
		log.Fatalf("database: %v", e)
	}
	defer pool.Close()
	logger, closeLog, e := config.NewLogger("logs/app.log")
	if e != nil {
		log.Fatal(e)
	}
	defer closeLog()
	repo := repository.NewStudentRepository(pool)
	handler := service.NewStudentHandler(repo)
	app := config.NewApp(logger, handler, pool)
	log.Fatal(app.Listen(":" + config.GetEnv("APP_PORT", "3000")))
}
