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

	// --- Report Module ---
	reportRepo := repository.NewReportRepository(config.DB)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportService)

	// --- Review (Penilaian) Module ---
	reviewRepo := repository.NewReviewRepository(config.DB)
	reviewService := service.NewReviewService(reviewRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)

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

	// =============================================
	// 6. SETUP ROUTES
	// =============================================
	api := app.Group("/api")

	// --- Public Routes (Tidak perlu login) ---
	api.Post("/login", authHandler.Login)

	// --- Protected Routes (Wajib login dengan JWT) ---
	protected := api.Group("/", middleware.Protected())

	// Profile (contoh endpoint protected)
	protected.Get("/profile", func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")
		jabatanID := c.Locals("jabatan_id")
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Data profil berhasil diambil",
			"data": fiber.Map{
				"user_id":    userID,
				"role":       role,
				"jabatan_id": jabatanID,
			},
		})
	})

	// Laporan Kinerja
	protected.Get("/reports", reportHandler.GetAll)
	protected.Post("/reports", reportHandler.Create)

	// Penilaian Kinerja
	protected.Post("/reviews", reviewHandler.Create)
	protected.Get("/reviews", reviewHandler.GetMyReviews)
	protected.Get("/reviews/my-submissions", reviewHandler.GetMySubmittedReviews)

	// Profile Settings
	protected.Put("/profile/change-password", userHandler.ChangePassword)

	// =============================================
	// 7. START SERVER
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
	log.Println("   POST   /api/login          - Login user")
	log.Println("")
	log.Println("   [PROTECTED - Butuh JWT]")
	log.Println("   GET    /api/profile        - Lihat profil user")
	log.Println("   GET    /api/reports        - Lihat semua laporan (dengan filter)")
	log.Println("   POST   /api/reports        - Buat laporan kinerja")
	log.Println("   POST   /api/reviews        - Buat penilaian kinerja")
	log.Println("   GET    /api/reviews        - Lihat penilaian saya (staf)")
	log.Println("   GET    /api/reviews/my-submissions - Lihat history penilaian (atasan)")
	log.Println("   PUT    /api/profile/change-password - Ubah password")
	log.Println("================================================")

	log.Fatal(app.Listen(":" + port))
}
