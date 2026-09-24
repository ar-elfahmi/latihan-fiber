package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
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

	studentRepo := repository.StudentRepository{DB: db}
	prestasiRepo := repository.PrestasiRepository{DB: db}
	userRepo := repository.UserRepository{DB: db}
	rbacRepo := repository.RBACRepository{DB: db}

	secret := []byte(config.GetWithDefault("JWT_SECRET", "rahasia-latihan"))

	svc := service.Services{
		Student:  service.StudentService{Repo: studentRepo},
		Prestasi: service.PrestasiService{Repo: prestasiRepo, Student: studentRepo},
		Auth:     service.AuthService{Repo: userRepo, RBAC: rbacRepo, Secret: secret},
	}

	// 3. Fiber + middleware + route.
	app := fiber.New()
	app.Use(middleware.Logger)
	// Protect butuh rbacRepo untuk fetch permission live (Modul 6).
	route.Setup(app, svc, rbacRepo, secret)
	// 4. Penampung 404 (URL tidak dikenal).
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "endpoint tidak ditemukan"})
	})

	port := config.GetWithDefault("APP_PORT", "3000")
	log.Printf("server berjalan di :%s", port)
	log.Fatal(app.Listen(":" + port))
}
