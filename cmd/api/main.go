package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"laporanharianapi/config"
	"laporanharianapi/internal/handler"
	"laporanharianapi/internal/middleware"
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/scheduler"
	"laporanharianapi/internal/service"
)

func main() {
	// 1. Load environment variables dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	// 2. Koneksi ke Database
	config.ConnectDatabase()
	log.Println("✅ Database terhubung")

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

	// --- Report Module ---
	reportRepo := repository.NewReportRepository(config.DB)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportService)

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

	// =============================================
	// 4. SETUP FIBER APP
	// =============================================
	app := fiber.New(fiber.Config{
		AppName: "Laporan Harian API v1.0",
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	// =============================================
	// 5. GLOBAL MIDDLEWARE
	// =============================================

	// CORS Middleware (agar bisa diakses dari HP/Web)
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Static file serving untuk folder uploads/photos (foto profil)
	app.Get("/uploads/photos/*", func(c fiber.Ctx) error {
		return c.SendFile("./" + c.Path())
	})

	// =============================================
	// 6. SETUP ROUTES
	// =============================================
	api := app.Group("/api")

	// --- Public Routes (Tidak perlu login) ---
	api.Post("/login", authHandler.Login)

	// --- Protected Routes (Wajib login dengan JWT) ---
	protected := api.Group("/", middleware.Protected())

	// ===================================================
	// A. PROFILE & DASHBOARD (Semua role yang sudah login)
	// ===================================================
	protected.Get("/profile", userHandler.GetProfile)
	protected.Put("/profile/change-password", userHandler.ChangePassword)
	protected.Put("/profile/change-photo", userHandler.ChangePhoto)
	protected.Get("/dashboard/summary", dashboardHandler.GetSummary)

	// ===================================================
	// B. USER MANAGEMENT - Hanya Sekertaris
	// ===================================================
	userRoutes := protected.Group("/users", middleware.AllowRoles("sekertaris"))
	userRoutes.Get("/", userHandler.GetAll)
	userRoutes.Get("/:id", userHandler.GetOne)
	userRoutes.Post("/", userHandler.Create)
	userRoutes.Put("/:id", userHandler.Update)
	userRoutes.Delete("/:id", userHandler.Delete)

	// ===================================================
	// C. LAPORAN - Semua role bisa create & view (RBAC di service layer)
	// ===================================================
	reportRoutes := protected.Group("/reports")
	reportRoutes.Post("/", reportHandler.Create) // Semua role
	reportRoutes.Get("/", reportHandler.GetAll)  // RBAC ditangani di service layer

	// ===================================================
	// D. PENILAIAN - Create hanya Lurah & Sekertaris
	//    View penilaian diri sendiri boleh semua role
	// ===================================================
	protected.Get("/reviews", reviewHandler.GetMyReviews) // Semua role (lihat nilai sendiri)

	reviewManage := protected.Group("/reviews", middleware.AllowRoles("lurah", "sekertaris"))
	reviewManage.Post("/", reviewHandler.Create)                             // Hanya Lurah & Sekertaris
	reviewManage.Get("/my-submissions", reviewHandler.GetMySubmittedReviews) // Hanya Lurah & Sekertaris

	// ===================================================
	// E. TUGAS POKOK - Create hanya Lurah & Sekertaris
	// ===================================================
	taskRoutes := protected.Group("/tasks", middleware.AllowRoles("lurah", "sekertaris"))
	taskRoutes.Post("/", taskHandler.Create) // Hanya Lurah & Sekertaris

	// My Tasks - Semua role bisa melihat tugas pokok miliknya (untuk dropdown)
	protected.Get("/my-tasks", taskHandler.GetMyTasks)

	// ===================================================
	// F. NOTIFIKASI - Semua role yang sudah login
	// ===================================================
	notifRoutes := protected.Group("/notifications")
	notifRoutes.Get("/", notifHandler.GetMy)            // Ambil semua notifikasi saya
	notifRoutes.Put("/:id/read", notifHandler.MarkRead) // Tandai notifikasi sebagai dibaca

	// =============================================
	// 7. BACKGROUND JOBS
	// =============================================
	scheduler.StartDailyReminder(config.DB, notifRepo)

	// =============================================
	// 8. START SERVER
	// =============================================
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("================================================")
	log.Printf("🚀 Server berjalan di http://localhost:%s", port)
	log.Println("================================================")
	log.Println("")
	log.Println("📌 Daftar Endpoints:")
	log.Println("   [PUBLIC]")
	log.Println("   POST   /api/login                        - Login user")
	log.Println("")
	log.Println("   [PROTECTED - Semua Role]")
	log.Println("   GET    /api/profile                      - Lihat profil user (dari DB)")
	log.Println("   PUT    /api/profile/change-password       - Ubah password")
	log.Println("   PUT    /api/profile/change-photo          - Ubah foto profil")
	log.Println("   GET    /api/dashboard/summary             - Statistik dashboard")
	log.Println("   POST   /api/reports                      - Buat laporan kinerja")
	log.Println("   GET    /api/reports                      - Lihat laporan (RBAC di service)")
	log.Println("   GET    /api/reviews                      - Lihat penilaian saya")
	log.Println("   GET    /api/notifications                - Lihat notifikasi saya")
	log.Println("   PUT    /api/notifications/:id/read       - Tandai notifikasi dibaca")
	log.Println("")
	log.Println("   [RBAC - Sekertaris Only]")
	log.Println("   GET    /api/users                        - Lihat semua user")
	log.Println("   GET    /api/users/:id                    - Lihat detail user")
	log.Println("   POST   /api/users                        - Buat user baru")
	log.Println("   PUT    /api/users/:id                    - Update user")
	log.Println("   DELETE /api/users/:id                    - Hapus user")
	log.Println("")
	log.Println("   [RBAC - Lurah & Sekertaris]")
	log.Println("   POST   /api/reviews                      - Buat penilaian kinerja")
	log.Println("   GET    /api/reviews/my-submissions        - History penilaian dibuat")
	log.Println("   POST   /api/tasks                        - Buat tugas pokok")
	log.Println("")
	log.Println("   [BACKGROUND JOBS]")
	log.Println("   ⏰ Daily Reminder                         - Jam 15:00 hari kerja")
	log.Println("================================================")

	log.Fatal(app.Listen(":" + port))
}
