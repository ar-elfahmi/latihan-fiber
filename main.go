package main

import (
	"context"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
	"latihan-fiber/route"
)

func main() {
	// 0. Baca file .env ke environment variable.
	config.LoadEnv()

	ctx := context.Background()

	// 1. Buka koneksi database. Fatal bila gagal.
	db, err := database.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mongoClient, err := database.NewMongoClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    _ = mongoClient.Disconnect(ctx)
	}()

	mongoDB := mongoClient.Database(config.GetWithDefault("MONGO_DATABASE", "praktikum_backend"))
	_ = mongoDB

	studentRepo := repository.StudentRepository{DB: db}
	prestasiRepo := repository.PrestasiRepository{DB: db}
	userRepo := repository.UserRepository{DB: db}
	rbacRepo := repository.RBACRepository{DB: db}

	secret := []byte(config.GetWithDefault("JWT_SECRET", "rahasia-latihan"))

	svc := service.Services{
		Student:  service.StudentService{Repo: studentRepo},
		Prestasi: service.PrestasiService{Repo: prestasiRepo, Student: studentRepo},
		Auth:     service.AuthService{Repo: userRepo, RBAC: rbacRepo, Secret: secret},
		User:     service.UserService{Repo: userRepo},
	}

	// 3. Fiber + middleware + route.
	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Use(middleware.Logger)
	// Protect butuh rbacRepo untuk fetch permission live (Modul 6).
	route.Setup(app, svc, rbacRepo, secret, db, mongoClient)
	// 4. Penampung 404 (URL tidak dikenal). Dikembalikan sebagai
	// *AppError agar ErrorHandler membentuk JSON seragam.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint")
	})

	port := config.GetWithDefault("APP_PORT", "3000")
	log.Printf("server berjalan di :%s", port)
	log.Fatal(app.Listen(":" + port))
}
