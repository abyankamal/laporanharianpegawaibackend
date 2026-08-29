package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/helmet"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/joho/godotenv"

	"laporanharianapi/config"
	"laporanharianapi/internal/handler"
	"laporanharianapi/internal/repository"

	"laporanharianapi/internal/scheduler"
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

	// Validasi JWT_SECRET wajib ada
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("❌ FATAL: JWT_SECRET tidak dikonfigurasi! Server tidak bisa berjalan tanpa JWT_SECRET.")
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

	// --- Supervisor Lurah Module ---
	supervisorRepo := repository.NewSupervisorRepository(config.DB)
	supervisorRepo.SeedDefault()

	userService := service.NewUserService(userRepo, supervisorRepo)
	userHandler := handler.NewUserHandler(userService)

	// --- Notification Module ---
	notifRepo := repository.NewNotificationRepository(config.DB)
	notifService := service.NewNotificationService(notifRepo)
	notifHandler := handler.NewNotificationHandler(notifService)

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
	reportService := service.NewReportService(reportRepo, holidayRepo, workHourRepo, supervisorRepo, userRepo, notifRepo)
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

	// --- Absensi Module ---
	absensiRepo := repository.NewAbsensiRepository(config.DB)
	absensiService := service.NewAbsensiService(absensiRepo, holidayRepo, workHourRepo, userRepo)
	absensiHandler := handler.NewAbsensiHandler(absensiService, userService)

	// --- Izin Module ---
	izinRepo := repository.NewIzinRepository(config.DB)
	izinService := service.NewIzinService(izinRepo, absensiRepo)
	izinHandler := handler.NewIzinHandler(izinService)

	// =============================================
	// 4. SETUP FIBER APP
	// =============================================
	app := fiber.New(fiber.Config{
		AppName:         "Laporan Harian API v1.0",
		BodyLimit:       50 * 1024 * 1024, // 50 MB
		ReadTimeout:     60 * time.Second, // Give more time for mobile uploads
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
	app.Use(helmet.New())

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

	// Static file serving dengan proteksi Path Traversal
	serveUploadFile := func(baseDir string) fiber.Handler {
		return func(c fiber.Ctx) error {
			rawParam := c.Params("*")
			// Bersihkan parameter path untuk mencegah traversal seperti ../..
			cleanParam := filepath.Clean("/" + rawParam)
			targetPath := filepath.Join(baseDir, cleanParam)

			// Pastikan targetPath berada di dalam baseDir
			absBase, err1 := filepath.Abs(baseDir)
			absTarget, err2 := filepath.Abs(targetPath)
			if err1 != nil || err2 != nil || !strings.HasPrefix(absTarget, absBase) {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
					"status":  "error",
					"message": "Akses file ditolak",
				})
			}

			// Cek apakah file fisik ada di disk dan bukan direktori
			info, err := os.Stat(absTarget)
			if err != nil || info.IsDir() {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"status":  "error",
					"message": "File tidak ditemukan",
				})
			}

			return c.SendFile(absTarget)
		}
	}

	app.Get("/api/uploads/photos/*", serveUploadFile("./uploads/photos"))
	app.Get("/uploads/photos/*", serveUploadFile("./uploads/photos"))
	app.Get("/api/uploads/reports/*", serveUploadFile("./uploads/reports"))
	app.Get("/uploads/reports/*", serveUploadFile("./uploads/reports"))
	app.Get("/api/uploads/attendance/*", serveUploadFile("./uploads/attendance"))
	app.Get("/uploads/attendance/*", serveUploadFile("./uploads/attendance"))
	app.Get("/api/uploads/leave/*", serveUploadFile("./uploads/leave"))
	app.Get("/uploads/leave/*", serveUploadFile("./uploads/leave"))

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
		Absensi:    absensiHandler,
		Izin:       izinHandler,
	}
	
	setupRoutes(app, h)

	// =============================================
	// 7. BACKGROUND JOBS
	// =============================================
	scheduler.StartDailyReminder(config.DB, notifRepo, workHourRepo, holidayRepo)

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
