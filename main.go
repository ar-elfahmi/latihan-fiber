package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hibiken/asynq"

	"latihan-fiber/app/repository"
	"latihan-fiber/app/service"
	"latihan-fiber/config"
	"latihan-fiber/database"
	"latihan-fiber/helper"
	"latihan-fiber/middleware"
	"latihan-fiber/route"
)

func main() {
	// 0. Baca file .env ke environment variable, lalu susun satu
	//    konfigurasi aplikasi (config.NewApp) yang dipakai semua lapisan.
	config.LoadEnv()
	appCfg := config.NewApp()

	role := flag.String("role", config.GetWithDefault("APP_ROLE", "api"), "api | worker | all")
	flag.Parse()
	if *role != "api" && *role != "worker" && *role != "all" {
		log.Fatalf("role tidak dikenal: %q", *role)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 1. Buka koneksi database. Fatal bila gagal.
	db, err := database.ConnectWithDSN(ctx, appCfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if *role == "worker" {
		shutdown, err := startWorker(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		<-ctx.Done()
		shutdown()
		return
	}

	mongoClient, err := database.NewMongoClient(ctx, appCfg.MongoURI)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = mongoClient.Disconnect(ctx)
	}()

	mongoDB := mongoClient.Database(appCfg.MongoDatabase)

	portfolioRepo := repository.NewPortfolioRepository(mongoDB)
	if err := portfolioRepo.EnsureIndexes(context.Background()); err != nil {
		log.Fatal(err)
	}
	reviewRepo := repository.NewReviewRepository(mongoDB)
	if err := reviewRepo.EnsureIndexes(context.Background()); err != nil {
		log.Fatal(err)
	}

	studentRepo := repository.PostgresStudentRepository{DB: db}
	prestasiRepo := repository.PostgresPrestasiRepository{DB: db}
	userRepo := repository.PostgresUserRepository{DB: db}
	rbacRepo := repository.PostgresRBACRepository{DB: db}

	// Aturan domain KRS (batas SKS) berasal dari app/service dan diinjeksi ke
	// repository supaya app/repository tidak mengimpor app/service (siklus
	// import) tapi tetap memakai implementasi aturan yang sama.
	courseRepo, err := repository.NewPostgresCourseRepository(db, repository.CreditRules{
		WithinCreditLimit: service.WithinCreditLimit,
		CreditsAfterSwap:  service.CreditsAfterSwap,
	})
	if err != nil {
		log.Fatal(err)
	}

	secret := appCfg.JWTSecret
	inspector := asynq.NewInspector(appCfg.Redis)
	defer func() {
		if err := inspector.Close(); err != nil {
			log.Printf("menutup inspector antrean: %v", err)
		}
	}()

	svc := service.Services{
		Student:  service.StudentService{Repo: studentRepo},
		Prestasi: service.PrestasiService{Repo: prestasiRepo, Student: studentRepo},
		Auth: service.AuthService{
			Repo:    userRepo,
			RBAC:    rbacRepo,
			Tokens:  helper.NewJWTManager(secret),
			Limiter: helper.NewLoginLimiter(helper.DefaultLoginLimit, helper.DefaultLoginWindow),
		},
		Course:   service.CourseService{Repo: courseRepo},
		User:     service.UserService{Repo: userRepo},
		Portfolio: service.PortfolioService{
			Repo:        portfolioRepo,
			StudentRepo: studentRepo,
		},
		Review: service.NewReviewService(reviewRepo, courseRepo),
		Notification: service.NotificationService{
			Notifications: repository.PostgresNotificationRepository{DB: db},
			Outbox:        repository.PostgresOutboxRepository{DB: db},
			Inspector:     inspector,
		},
	}

	// 3. Fiber + middleware + route.
	app := fiber.New(fiber.Config{ErrorHandler: config.ErrorHandler})
	app.Use(middleware.Logger)
	// Protect butuh rbacRepo untuk fetch permission live (Modul 6).
	route.Setup(app, route.Dependencies{
		Services:    svc,
		RBAC:        rbacRepo,
		JWTSecret:   secret,
		Pool:        db,
		MongoClient: mongoClient,
	})
	// 4. Penampung 404 (URL tidak dikenal). Dikembalikan sebagai
	// *AppError agar ErrorHandler membentuk JSON seragam.
	app.Use(func(c *fiber.Ctx) error {
		return helper.NotFound("endpoint")
	})

	port := appCfg.Port
	if *role == "all" {
		shutdown, err := startWorker(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		defer shutdown()
	}

	apiDone := make(chan error, 1)
	log.Printf("server berjalan di :%s", port)
	go func() {
		apiDone <- app.Listen(":" + port)
	}()

	select {
	case err := <-apiDone:
		if err != nil {
			log.Printf("server HTTP berhenti: %v", err)
		}
	case <-ctx.Done():
		if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
			log.Printf("shutdown HTTP: %v", err)
		}
	}
}
