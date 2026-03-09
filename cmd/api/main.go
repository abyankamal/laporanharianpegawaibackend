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
	reportService := service.NewReportService(reportRepo, holidayRepo, workHourRepo)
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
	adminService := service.NewAdminService(adminRepo, userRepo)
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
			message := err.Error() // Temporarily show full error for debugging

			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}

			// Log detail error di server
			if code >= 500 {
				log.Printf("[SERVER ERROR] %d %s: %v", code, c.Path(), err)
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

	// Recover Middleware (menangkap panic dan mengubahnya menjadi HTTP 500)
	app.Use(recover.New())

	// CORS Middleware — origins dibaca dari ALLOWED_ORIGINS di .env
	// Contoh production: ALLOWED_ORIGINS=https://siopik.com,https://admin.siopik.com
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

	// Static file serving untuk folder uploads/photos (foto profil)
	// Ditambahkan alias "/api" agar sesuai dengan baseImageUrl frontend yang menggunakan /api
	app.Get("/api/uploads/photos/*", func(c fiber.Ctx) error {
		return c.SendFile("./uploads/photos/" + c.Params("*"))
	})

	app.Get("/uploads/photos/*", func(c fiber.Ctx) error {
		return c.SendFile("./uploads/photos/" + c.Params("*"))
	})

	// Static file serving untuk folder uploads/reports (bukti laporan)
	app.Get("/api/uploads/reports/*", func(c fiber.Ctx) error {
		return c.SendFile("./uploads/reports/" + c.Params("*"))
	})

	app.Get("/uploads/reports/*", func(c fiber.Ctx) error {
		return c.SendFile("./uploads/reports/" + c.Params("*"))
	})

	// =============================================
	// 6. SETUP ROUTES
	// =============================================
	api := app.Group("/api")

	// ===================================================
	// MOBILE ROUTES (FLUTTER) - /api/mobile/v1
	// ===================================================
	mobile := api.Group("/mobile")

	// 1. Authenticated Routes for Mobile
	mobile.Post("/login", authHandler.Login) // Public Login for Mobile

	mProtected := mobile.Group("", middleware.Protected())

	// Profile & Dashboard
	mProtected.Get("/profile", userHandler.GetProfile)
	mProtected.Put("/profile/change-password", userHandler.ChangePassword)
	mProtected.Put("/profile/change-photo", userHandler.ChangePhoto)
	mProtected.Put("/users/fcm-token", userHandler.UpdateFCMToken)
	mProtected.Get("/dashboard/summary", dashboardHandler.GetSummary)

	// Directory
	mProtected.Get("/rekan-kerja", adminHandler.GetPegawai)

	// Laporan (Mobile)
	mReport := mProtected.Group("/reports")
	mReport.Post("/", reportHandler.Create)
	mReport.Get("/", reportHandler.GetAll)
	mReport.Get("/recap", reportHandler.GetReportRecapHandler)
	mReport.Get("/recap-pegawai", adminHandler.GetRekapLaporan, middleware.AllowRoles("lurah", "sekertaris", "admin"))
	mReport.Get("/export", adminHandler.GetLaporanExport)
	mReport.Get("/export/excel", reportHandler.ExportReportRecapExcelHandler)
	mReport.Get("/export/pdf", reportHandler.ExportReportPDFHandler)
	mReport.Get("/export/attachments", reportHandler.ExportReportAttachmentsHandler)
	mReport.Put("/evaluate", reportHandler.EvaluateReportHandler, middleware.AllowRoles("lurah", "sekertaris"))
	mReport.Get("/:id", reportHandler.GetOne)

	// Tugas & Notifikasi (Mobile)
	mProtected.Get("/my-tasks", taskHandler.GetMyTasks)
	mProtected.Get("/notifications", notifHandler.GetMy)
	mProtected.Get("/notifications/:id", notifHandler.GetByID)
	mProtected.Put("/notifications/:id/read", notifHandler.MarkRead)

	// Manajemen Tugas (Khusus Lurah di Mobile)
	mTasks := mProtected.Group("/tasks")
	mTasks.Get("/:id", taskHandler.GetByID) // Bisa diakses Lurah & Assignee
	mTasks.Post("/", taskHandler.Create, middleware.AllowRoles("lurah"))
	mTasks.Get("/", taskHandler.GetAll, middleware.AllowRoles("lurah"))
	mTasks.Put("/:id", taskHandler.Update, middleware.AllowRoles("lurah"))
	mTasks.Delete("/:id", taskHandler.Delete, middleware.AllowRoles("lurah"))

	// Penilaian (Mobile)
	mProtected.Get("/reviews", reviewHandler.GetMyReviews) // Semua role bisa lihat nilai diri sendiri

	// Manajemen Penilaian (Khusus Lurah & Sekertaris di Mobile)
	mReviewManage := mProtected.Group("/reviews", middleware.AllowRoles("lurah", "sekertaris"))
	mReviewManage.Post("/", reviewHandler.Create)
	mReviewManage.Get("/submissions", reviewHandler.GetMySubmittedReviews)

	// ===================================================
	// WEB ROUTES (ADMIN PANEL) - /api/web/v1
	// ===================================================
	web := api.Group("/web")

	// 1. Authenticated Routes for Web
	web.Post("/login", authHandler.Login) // Public Login for Web

	wProtected := web.Group("", middleware.Protected())

	// Dashboard & Profile
	wProtected.Get("/profile", userHandler.GetProfile)
	wProtected.Get("/dashboard/summary", adminHandler.GetDashboardSummary)

	// Rekap Laporan & Export (Kembalikan ke wProtected agar tidak merusak frontend lama)
	wReports := wProtected.Group("/reports")
	wReports.Get("/", reportHandler.GetAll)
	wReports.Get("/recap", adminHandler.GetRekapLaporan)
	wReports.Get("/export", adminHandler.GetLaporanExport)
	wReports.Get("/export/excel", reportHandler.ExportReportRecapExcelHandler)
	wReports.Get("/export/pdf", reportHandler.ExportReportPDFHandler)
	wReports.Get("/export/attachments", reportHandler.ExportReportAttachmentsHandler)
	wReports.Put("/evaluate", reportHandler.EvaluateReportHandler)
	wReports.Get("/:id", reportHandler.GetOne)

	// Admin Specific (Only Lurah/Sekertaris/Admin)
	adminOnly := wProtected.Group("/admin", middleware.AdminOnly())

	// App Settings
	adminOnly.Get("/jam-kerja", workHourHandler.GetWorkHour)
	adminOnly.Put("/jam-kerja", workHourHandler.UpdateWorkHour)
	adminOnly.Get("/hari-libur", holidayHandler.GetHolidays)
	adminOnly.Post("/hari-libur", holidayHandler.CreateHoliday)
	adminOnly.Put("/hari-libur/:id", holidayHandler.UpdateHoliday)
	adminOnly.Delete("/hari-libur/:id", holidayHandler.DeleteHoliday)

	// User Management
	userManage := wProtected.Group("/users", middleware.AllowRoles("lurah", "sekertaris"))
	userManage.Get("/", userHandler.GetAll)
	userManage.Get("/supervisors", userHandler.GetSupervisors)
	userManage.Get("/:id", userHandler.GetOne)
	userManage.Post("/", userHandler.Create)
	userManage.Put("/:id", userHandler.Update)
	userManage.Delete("/:id", userHandler.Delete)

	// Manajemen Pegawai
	pegawaiManage := adminOnly.Group("/pegawai")
	pegawaiManage.Get("/", adminHandler.GetPegawai)
	pegawaiManage.Post("/", adminHandler.CreatePegawai)
	pegawaiManage.Put("/:id", adminHandler.UpdatePegawai)
	pegawaiManage.Delete("/:id", adminHandler.DeletePegawai)

	// Alias untuk standarisasi (agar /api/web/admin/reports juga bekerja)
	wAdminReports := adminOnly.Group("/reports")
	wAdminReports.Get("/", reportHandler.GetAll)
	wAdminReports.Get("/recap", adminHandler.GetRekapLaporan)
	wAdminReports.Get("/export", adminHandler.GetLaporanExport)
	wAdminReports.Get("/export/excel", reportHandler.ExportReportRecapExcelHandler)
	wAdminReports.Get("/export/pdf", reportHandler.ExportReportPDFHandler)
	wAdminReports.Get("/export/attachments", reportHandler.ExportReportAttachmentsHandler)
	wAdminReports.Put("/evaluate", reportHandler.EvaluateReportHandler)
	wAdminReports.Get("/:id", reportHandler.GetOne)

	// Pusat Pengumuman
	pengumuman := adminOnly.Group("/pengumuman")
	pengumuman.Get("/", adminHandler.GetPengumuman)
	pengumuman.Post("/", adminHandler.CreatePengumuman)
	pengumuman.Put("/:id", adminHandler.UpdatePengumuman)
	pengumuman.Delete("/:id", adminHandler.DeletePengumuman)

	// Manajemen Tugas (Lurah)
	wTasks := wProtected.Group("/tasks", middleware.AllowRoles("lurah"))
	wTasks.Post("/", taskHandler.Create)
	wTasks.Get("/", taskHandler.GetAll)
	wTasks.Put("/:id", taskHandler.Update)
	wTasks.Delete("/:id", taskHandler.Delete)

	// Manajemen Penilaian
	wReviews := wProtected.Group("/reviews", middleware.AllowRoles("lurah", "sekertaris"))
	wReviews.Post("/", reviewHandler.Create)
	wReviews.Get("/submissions", reviewHandler.GetMySubmittedReviews)

	// Manajemen Jabatan
	adminOnly.Get("/jabatan", jabatanHandler.GetAll)
	adminOnly.Get("/jabatan/:id", jabatanHandler.GetOne)
	adminOnly.Post("/jabatan", jabatanHandler.Create)
	adminOnly.Put("/jabatan/:id", jabatanHandler.Update)
	adminOnly.Delete("/jabatan/:id", jabatanHandler.Delete)

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
	log.Println("📌 Daftar Endpoints Utama:")
	log.Println("   [MOBILE - /api/mobile/v1]")
	log.Println("   POST   /login                    - Login Mobile")
	log.Println("   GET    /profile                  - Profile Saya")
	log.Println("   POST   /reports                  - Buat Laporan")
	log.Println("   GET    /my-tasks                 - Tugas Saya")
	log.Println("")
	log.Println("   [WEB - /api/web/v1]")
	log.Println("   POST   /login                    - Login Web")
	log.Println("   GET    /dashboard/summary        - Dashboard Admin")
	log.Println("   GET    /users                    - Kelola User")
	log.Println("   GET    /reports/recap            - Rekap Laporan")
	log.Println("================================================")
	log.Println("   [BACKGROUND JOBS]")
	log.Println("   ⏰ Daily Reminder                         - Jam 15:00 hari kerja")
	log.Println("================================================")

	log.Fatal(app.Listen(":" + port))
}
