package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"

	"laporanharianapi/config"
	"laporanharianapi/internal/handler"
	"laporanharianapi/internal/middleware"
	"laporanharianapi/internal/repository"

	// "laporanharianapi/internal/scheduler"
	"laporanharianapi/internal/service"
	"laporanharianapi/pkg/fcm"
)

func main() {
	// 0. Set Global Timezone to WIB (Asia/Jakarta)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		log.Printf("⚠️ Gagal memuat lokasi waktu Asia/Jakarta: %v", err)
	} else {
		time.Local = loc
		log.Println("🌍 Timezone diatur ke Asia/Jakarta (WIB)")
	}

	// 1. Load environment variables dari file .env
	err = godotenv.Load()
	if err != nil {
		log.Println("⚠️  File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	// 2. Koneksi ke Database
	config.ConnectDatabase()
	log.Println("✅ Database terhubung")

	// 2.5 Inisialisasi Firebase Admin SDK
	if err := fcm.InitFirebase(); err != nil {
		log.Printf("⚠️ Gagal inisialisasi Firebase (lihat log FCM): %v", err)
	}

	// =============================================
	// 3. DEPENDENCY INJECTION (Wiring)
	// =============================================

	// --- User & Auth Module ---
	userRepo := repository.NewUserRepository(config.DB)
	authService := service.NewAuthService(userRepo)
	authHandler := handler.NewAuthHandler(authService)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// --- Notification Module ---
	notifRepo := repository.NewNotificationRepository(config.DB)
	notifService := service.NewNotificationService(notifRepo)
	notifHandler := handler.NewNotificationHandler(notifService)

	// --- Supervisor Lurah Module ---
	supervisorRepo := repository.NewSupervisorRepository(config.DB)
	supervisorRepo.SeedDefault()

	// --- Work Hour & Holiday Modules ---
	workHourRepo := repository.NewWorkHourRepository(config.DB)
	workHourService := service.NewWorkHourService(workHourRepo)
	workHourHandler := handler.NewWorkHourHandler(workHourService)

	workHourRepo.SeedDefault()

	holidayRepo := repository.NewHolidayRepository(config.DB)
	holidayService := service.NewHolidayService(holidayRepo)
	holidayHandler := handler.NewHolidayHandler(holidayService)

	// --- Report Module ---
	reportRepo := repository.NewReportRepository(config.DB)
	reportService := service.NewReportService(reportRepo, holidayRepo, workHourRepo, supervisorRepo)
	reportHandler := handler.NewReportHandler(reportService, userService)

	// --- Review (Penilaian) Module ---
	reviewRepo := repository.NewReviewRepository(config.DB)
	reviewService := service.NewReviewService(reviewRepo, userRepo, notifRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)

	// --- Task (Tugas Pokok) Module ---
	taskRepo := repository.NewTaskRepository(config.DB)
	taskService := service.NewTaskService(taskRepo, userRepo, notifRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	// --- Dashboard Module ---
	dashboardRepo := repository.NewDashboardRepository(config.DB)
	dashboardService := service.NewDashboardService(dashboardRepo, userRepo)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)

	// --- Jabatan Module ---
	jabatanRepo := repository.NewJabatanRepository(config.DB)
	jabatanService := service.NewJabatanService(jabatanRepo)
	jabatanHandler := handler.NewJabatanHandler(jabatanService)

	// --- Admin Module ---
	adminRepo := repository.NewAdminRepository(config.DB)
	adminService := service.NewAdminService(adminRepo, userRepo, supervisorRepo)
	adminHandler := handler.NewAdminHandler(adminService)

	// =============================================
	// 4. SETUP FIBER APP
	// =============================================
	app := fiber.New(fiber.Config{
		AppName:         "Laporan Harian API v1.0",
		BodyLimit:       300 * 1024 * 1024, // Increase to 300 MB
		ReadTimeout:     60 * time.Second,  // Give more time for slow mobile uploads
		WriteTimeout:    60 * time.Second,
		IdleTimeout:     120 * time.Second,
		ReadBufferSize:  16 * 1024, // 16KB buff
		WriteBufferSize: 16 * 1024,
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			message := err.Error()

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			if code >= 500 {
				log.Printf("[SERVER ERROR] %d %s: %v", code, c.Path(), err)
				message = "Terjadi kesalahan pada server"
			}

			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": message,
			})
		},
	})

	// =============================================
	// 5. GLOBAL MIDDLEWARE
	// =============================================

	app.Use(recover.New())

	allowedOriginsEnv := os.Getenv("ALLOWED_ORIGINS")
	var allowedOrigins []string
	if allowedOriginsEnv == "" {
		log.Println("⚠️  ALLOWED_ORIGINS tidak diset. Menggunakan '*' (tidak aman untuk production!)")
		allowedOrigins = []string{"*"}
	} else {
		for _, o := range strings.Split(allowedOriginsEnv, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				allowedOrigins = append(allowedOrigins, trimmed)
			}
		}
		log.Printf("✅ CORS diizinkan untuk: %v", allowedOrigins)
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins: allowedOrigins,
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Static file serving
	app.Get("/api/uploads/photos/*", func(c fiber.Ctx) error { return c.SendFile("./uploads/photos/" + c.Params("*")) })
	app.Get("/uploads/photos/*", func(c fiber.Ctx) error { return c.SendFile("./uploads/photos/" + c.Params("*")) })
	app.Get("/api/uploads/reports/*", func(c fiber.Ctx) error { return c.SendFile("./uploads/reports/" + c.Params("*")) })
	app.Get("/uploads/reports/*", func(c fiber.Ctx) error { return c.SendFile("./uploads/reports/" + c.Params("*")) })

	// =============================================
	// 6. SETUP ROUTES
	// =============================================
	h := Handlers{
		Auth:       authHandler,
		User:       userHandler,
		Notif:      notifHandler,
		WorkHour:   workHourHandler,
		Holiday:    holidayHandler,
		Report:     reportHandler,
		Review:     reviewHandler,
		Task:       taskHandler,
		Dashboard:  dashboardHandler,
		Jabatan:    jabatanHandler,
		Admin:      adminHandler,
	}
	
	setupRoutes(app, h)

	// =============================================
	// 7. BACKGROUND JOBS
	// =============================================
	// scheduler.StartDailyReminder(config.DB, notifRepo, workHourRepo)

	// =============================================
	// 8. START SERVER
	// =============================================
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("================================================")
	log.Printf("🚀 Server berjalan di http://localhost:%s", port)
	log.Println("🌍 Timezone : Asia/Jakarta (WIB)")
	log.Println("✅ FCM Push  : Ready (Firebase Admin SDK)")
	log.Println("================================================")

	log.Fatal(app.Listen(":" + port))
}
