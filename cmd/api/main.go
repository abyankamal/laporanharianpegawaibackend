package main

import (
	"log"
	"os"
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

	// Recover Middleware (menangkap panic dan mengubahnya menjadi HTTP 500)
	app.Use(recover.New())

	// CORS Middleware (agar bisa diakses dari Web Admin/Mobile)
	app.Use(cors.New(cors.Config{
		// Mengizinkan semua origin untuk tahap development.
		// Untuk production, ganti "*" menjadi domain spesifik, misal:
		// AllowOrigins: []string{"https://admin.siopik.com", "http://localhost:5173"},
		AllowOrigins: []string{"*"},
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
	protected.Put("/users/fcm-token", userHandler.UpdateFCMToken)
	protected.Get("/dashboard/summary", dashboardHandler.GetSummary)
	protected.Get("/jabatan", jabatanHandler.GetAll)

	// ===================================================
	// B. USER MANAGEMENT
	// ===================================================
	// Read-only access untuk Lurah (butuh untuk dropdown bawahan) & Sekertaris
	userRoutesRead := protected.Group("/users", middleware.AllowRoles("Sekertaris", "lurah", "sekertaris"))
	userRoutesRead.Get("/", userHandler.GetAll)
	userRoutesRead.Get("/supervisors", userHandler.GetSupervisors)
	userRoutesRead.Get("/:id", userHandler.GetOne)

	// Write access khusus Sekertaris
	userRoutesWrite := protected.Group("/users", middleware.AllowRoles("sekertaris", "Sekertaris"))
	userRoutesWrite.Post("/", userHandler.Create)
	userRoutesWrite.Put("/:id", userHandler.Update)
	userRoutesWrite.Delete("/:id", userHandler.Delete)

	// ===================================================
	// APP SETTINGS - Hanya Sekertaris dan Lurah (Sekarang menggunakan AdminOnly)
	// ===================================================
	adminRoutes := protected.Group("/admin", middleware.AdminOnly())
	adminRoutes.Get("/jam-kerja", workHourHandler.GetWorkHour)
	adminRoutes.Put("/jam-kerja", workHourHandler.UpdateWorkHour)
	adminRoutes.Get("/hari-libur", holidayHandler.GetHolidays)
	adminRoutes.Post("/hari-libur", holidayHandler.CreateHoliday)
	adminRoutes.Delete("/hari-libur/:id", holidayHandler.DeleteHoliday)

	// ===================================================
	// C. LAPORAN - Semua role bisa create & view (RBAC di service layer)
	// ===================================================
	reportRoutes := protected.Group("/reports")
	reportRoutes.Post("/", reportHandler.Create) // Semua role
	reportRoutes.Get("/", reportHandler.GetAll)  // RBAC ditangani di service layer
	reportRoutes.Get("/recap", reportHandler.GetReportRecapHandler)
	reportRoutes.Get("/recap/export/excel", reportHandler.ExportReportRecapExcelHandler)
	reportRoutes.Get("/recap/export/attachments", reportHandler.ExportReportAttachmentsHandler)
	reportRoutes.Get("/recap/export/pdf", reportHandler.ExportReportPDFHandler) // Ekspor PDF laporan harian
	reportRoutes.Put("/evaluate", reportHandler.EvaluateReportHandler)          // Evaluasi laporan (Lurah/Sekertaris)
	reportRoutes.Get("/:id", reportHandler.GetOne)                              // Mengambil detail laporan

	// ===================================================
	// D. PENILAIAN - Create hanya Lurah & Sekertaris
	//    View penilaian diri sendiri boleh semua role
	// ===================================================
	protected.Get("/reviews", reviewHandler.GetMyReviews) // Semua role (lihat nilai sendiri)

	reviewManage := protected.Group("/reviews", middleware.AllowRoles("lurah", "sekertaris"))
	reviewManage.Post("/", reviewHandler.Create)                             // Hanya Lurah & Sekertaris
	reviewManage.Get("/my-submissions", reviewHandler.GetMySubmittedReviews) // Hanya Lurah & Sekertaris

	// ===================================================
	// E. TUGAS ORGANISASI - Create, Update, Delete hanya Lurah
	// ===================================================
	taskRoutes := protected.Group("/tasks", middleware.AllowRoles("lurah"))
	taskRoutes.Post("/", taskHandler.Create)      // Hanya Lurah
	taskRoutes.Get("/", taskHandler.GetAll)       // Hanya Lurah
	taskRoutes.Put("/:id", taskHandler.Update)    // Hanya Lurah
	taskRoutes.Delete("/:id", taskHandler.Delete) // Hanya Lurah

	// My Tasks - Semua role bisa melihat tugas organisasi miliknya (untuk dropdown)
	protected.Get("/my-tasks", taskHandler.GetMyTasks)

	// Detail Tugas - Bisa dilihat oleh Lurah maupun assignee
	protected.Get("/tasks/:id", taskHandler.GetByID)

	// ===================================================
	// F. NOTIFIKASI - Semua role yang sudah login
	// ===================================================
	notifRoutes := protected.Group("/notifications")
	notifRoutes.Get("/", notifHandler.GetMy)            // Ambil semua notifikasi saya
	notifRoutes.Get("/:id", notifHandler.GetByID)       // Ambil detail notifikasi
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
	log.Println("   GET    /api/jabatan                      - Lihat daftar jabatan")
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
	log.Println("   [RBAC - Lurah Only]")
	log.Println("   POST   /api/tasks                        - Buat tugas organisasi")
	log.Println("   GET    /api/tasks                        - Lihat seluruh tugas organisasi")
	log.Println("   PUT    /api/tasks/:id                    - Update tugas organisasi")
	log.Println("   DELETE /api/tasks/:id                    - Hapus tugas organisasi")
	log.Println("   [RBAC - Semua Role]")
	log.Println("   GET    /api/my-tasks                     - Lihat tugas organisasi saya")
	log.Println("")
	log.Println("   [RBAC - Sekertaris & Lurah]")
	log.Println("   GET    /api/admin/jam-kerja              - Lihat pengaturan jam kerja")
	log.Println("   PUT    /api/admin/jam-kerja              - Update pengaturan jam kerja")
	log.Println("   GET    /api/admin/hari-libur             - Lihat daftar hari libur")
	log.Println("   POST   /api/admin/hari-libur             - Tambah hari libur")
	log.Println("   DELETE /api/admin/hari-libur/:id          - Hapus hari libur")
	log.Println("   [BACKGROUND JOBS]")
	log.Println("   ⏰ Daily Reminder                         - Jam 15:00 hari kerja")
	log.Println("================================================")

	log.Fatal(app.Listen(":" + port))
}
