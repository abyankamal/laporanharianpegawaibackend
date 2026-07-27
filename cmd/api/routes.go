package main

import (
	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/handler"
	"laporanharianapi/internal/middleware"
)

// Handlers struct menyimpan semua handler yang dibutuhkan
type Handlers struct {
	Auth       *handler.AuthHandler
	User       *handler.UserHandler
	Notif      *handler.NotificationHandler
	WorkHour   *handler.WorkHourHandler
	Holiday    *handler.HolidayHandler
	Report     *handler.ReportHandler
	Review     *handler.ReviewHandler
	Task       *handler.TaskHandler
	Dashboard  *handler.DashboardHandler
	Jabatan    *handler.JabatanHandler
	Admin      *handler.AdminHandler
	Absensi    *handler.AbsensiHandler
	Izin       *handler.IzinHandler
}

// setupRoutes mengatur semua routing aplikasi
func setupRoutes(app *fiber.App, h Handlers) {
	api := app.Group("/api")
	
	setupMobileRoutes(api, h)
	setupWebRoutes(api, h)
}

func setupMobileRoutes(api fiber.Router, h Handlers) {
	mobile := api.Group("/mobile")

	// Public Routes
	mobile.Post("/login", h.Auth.Login)
	mobile.Post("/refresh", h.Auth.RefreshToken)

	// Protected Routes
	mProtected := mobile.Group("", middleware.Protected())

	// Profile & Dashboard
	mProtected.Get("/profile", h.User.GetProfile)
	mProtected.Put("/profile/change-password", h.User.ChangePassword)
	mProtected.Put("/profile/change-photo", h.User.ChangePhoto)
	mProtected.Put("/users/fcm-token", h.User.UpdateFCMToken)
	mProtected.Get("/dashboard/summary", h.Dashboard.GetSummary)

	// Directory
	mProtected.Get("/rekan-kerja", h.Admin.GetPegawai)

	// Laporan
	mReport := mProtected.Group("/reports")
	mReport.Post("/", h.Report.Create)
	mReport.Get("/", h.Report.GetAll)
	mReport.Get("/recap", h.Report.GetReportRecapHandler)
	mReport.Get("/recap-pegawai", h.Admin.GetRekapLaporan, middleware.AllowRoles("lurah", "sekertaris", "admin"))
	mReport.Get("/export", h.Admin.GetLaporanExport)
	mReport.Get("/export/excel", h.Report.ExportReportRecapExcelHandler)
	mReport.Get("/export/pdf", h.Report.ExportReportPDFHandler)
	mReport.Get("/export/attachments", h.Report.ExportReportAttachmentsHandler)
	mReport.Put("/evaluate", h.Report.EvaluateReportHandler, middleware.AllowRoles("lurah", "sekertaris"))
	mReport.Put("/:id", h.Report.Update)
	mReport.Delete("/:id", h.Report.Delete)
	mReport.Get("/:id", h.Report.GetOne)

	// Tugas & Notifikasi
	mProtected.Get("/my-tasks", h.Task.GetMyTasks)
	mProtected.Get("/notifications", h.Notif.GetMy)
	mProtected.Get("/notifications/:id", h.Notif.GetByID)
	mProtected.Put("/notifications/:id/read", h.Notif.MarkRead)

	// Manajemen Tugas (Lurah)
	mTasks := mProtected.Group("/tasks")
	mTasks.Get("/:id", h.Task.GetByID) // Bisa diakses Lurah & Assignee
	mTasks.Post("/", h.Task.Create, middleware.AllowRoles("lurah"))
	mTasks.Get("/", h.Task.GetAll, middleware.AllowRoles("lurah"))
	mTasks.Put("/:id", h.Task.Update, middleware.AllowRoles("lurah"))
	mTasks.Delete("/:id", h.Task.Delete, middleware.AllowRoles("lurah"))

	// Penilaian
	mProtected.Get("/reviews", h.Review.GetMyReviews)

	// Manajemen Penilaian
	mReviewManage := mProtected.Group("/reviews", middleware.AllowRoles("lurah", "sekertaris"))
	mReviewManage.Post("/", h.Review.Create)
	mReviewManage.Get("/submissions", h.Review.GetMySubmittedReviews)

	// Absensi
	mAbsensi := mProtected.Group("/absensi")
	mAbsensi.Post("/check-in", h.Absensi.CheckIn)
	mAbsensi.Post("/check-out", h.Absensi.CheckOut)
	mAbsensi.Get("/today", h.Absensi.GetTodayStatus)
	mAbsensi.Get("/recap", h.Absensi.GetMonthlyRecap)
	mAbsensi.Get("/recap/all", h.Absensi.GetAllRecap, middleware.AllowRoles("lurah", "sekertaris"))
	mAbsensi.Get("/export/pdf", h.Absensi.ExportPDF)
}

func setupWebRoutes(api fiber.Router, h Handlers) {
	web := api.Group("/web")

	// Public Routes
	web.Post("/login", h.Auth.Login)
	web.Post("/refresh", h.Auth.RefreshToken)

	// Protected Routes
	wProtected := web.Group("", middleware.Protected())

	// Dashboard & Profile
	wProtected.Get("/profile", h.User.GetProfile)
	wProtected.Get("/dashboard/summary", h.Admin.GetDashboardSummary)

	// Rekap Laporan & Export
	wReports := wProtected.Group("/reports")
	wReports.Get("/", h.Report.GetAll)
	wReports.Post("/", h.Report.Create)
	wReports.Get("/recap", h.Admin.GetRekapLaporan)
	wReports.Get("/export", h.Admin.GetLaporanExport)
	wReports.Get("/export/excel", h.Report.ExportReportRecapExcelHandler)
	wReports.Get("/export/pdf", h.Report.ExportReportPDFHandler)
	wReports.Get("/export/attachments", h.Report.ExportReportAttachmentsHandler)
	wReports.Put("/evaluate", h.Report.EvaluateReportHandler)
	wReports.Get("/:id", h.Report.GetOne)

	// Admin Specific
	adminOnly := wProtected.Group("/admin", middleware.AdminOnly())

	// App Settings
	adminOnly.Get("/jam-kerja", h.WorkHour.GetWorkHour)
	adminOnly.Put("/jam-kerja", h.WorkHour.UpdateWorkHour)
	adminOnly.Get("/hari-libur", h.Holiday.GetHolidays)
	adminOnly.Post("/hari-libur", h.Holiday.CreateHoliday)
	adminOnly.Put("/hari-libur/:id", h.Holiday.UpdateHoliday)
	adminOnly.Delete("/hari-libur/:id", h.Holiday.DeleteHoliday)
	
	// Supervisor Lurah Settings
	adminOnly.Get("/supervisor-lurah", h.Admin.GetSupervisorLurah)
	adminOnly.Put("/supervisor-lurah", h.Admin.UpdateSupervisorLurah)

	// User Management
	userManage := wProtected.Group("/users", middleware.AllowRoles("lurah", "sekertaris"))
	userManage.Get("/", h.User.GetAll)
	userManage.Get("/supervisors", h.User.GetSupervisors)
	userManage.Get("/:id", h.User.GetOne)
	userManage.Post("/", h.User.Create)
	userManage.Put("/:id", h.User.Update)
	userManage.Delete("/:id", h.User.Delete)

	// Manajemen Pegawai
	pegawaiManage := adminOnly.Group("/pegawai")
	pegawaiManage.Get("/", h.Admin.GetPegawai)
	pegawaiManage.Post("/", h.Admin.CreatePegawai)
	pegawaiManage.Put("/:id", h.Admin.UpdatePegawai)
	pegawaiManage.Delete("/:id", h.Admin.DeletePegawai)

	// Alias admin/reports
	wAdminReports := adminOnly.Group("/reports")
	wAdminReports.Get("/", h.Report.GetAll)
	wAdminReports.Post("/", h.Report.Create)
	wAdminReports.Get("/recap", h.Admin.GetRekapLaporan)
	wAdminReports.Get("/export", h.Admin.GetLaporanExport)
	wAdminReports.Get("/export/excel", h.Report.ExportReportRecapExcelHandler)
	wAdminReports.Get("/export/pdf", h.Report.ExportReportPDFHandler)
	wAdminReports.Get("/export/attachments", h.Report.ExportReportAttachmentsHandler)
	wAdminReports.Put("/evaluate", h.Report.EvaluateReportHandler)
	wAdminReports.Get("/:id", h.Report.GetOne)

	// Pusat Pengumuman
	pengumuman := adminOnly.Group("/pengumuman")
	pengumuman.Get("/", h.Admin.GetPengumuman)
	pengumuman.Post("/", h.Admin.CreatePengumuman)
	pengumuman.Put("/:id", h.Admin.UpdatePengumuman)
	pengumuman.Delete("/:id", h.Admin.DeletePengumuman)

	// Manajemen Tugas (Lurah)
	wTasks := wProtected.Group("/tasks", middleware.AllowRoles("lurah"))
	wTasks.Post("/", h.Task.Create)
	wTasks.Get("/", h.Task.GetAll)
	wTasks.Put("/:id", h.Task.Update)
	wTasks.Delete("/:id", h.Task.Delete)

	// Manajemen Penilaian
	wReviews := wProtected.Group("/reviews", middleware.AllowRoles("lurah", "sekertaris"))
	wReviews.Post("/", h.Review.Create)
	wReviews.Get("/submissions", h.Review.GetMySubmittedReviews)

	// Manajemen Jabatan
	adminOnly.Get("/jabatan", h.Jabatan.GetAll)
	adminOnly.Get("/jabatan/:id", h.Jabatan.GetOne)
	adminOnly.Post("/jabatan", h.Jabatan.Create)
	adminOnly.Put("/jabatan/:id", h.Jabatan.Update)
	adminOnly.Delete("/jabatan/:id", h.Jabatan.Delete)

	// Absensi (Web Admin)
	wAbsensi := wProtected.Group("/absensi")
	wAbsensi.Get("/recap", h.Absensi.GetAllRecap)
	wAbsensi.Get("/export/pdf", h.Absensi.ExportPDF)

	// Geofencing Settings (Admin Only)
	adminOnly.Get("/geofencing", h.WorkHour.GetWorkHour)
	adminOnly.Put("/geofencing", h.WorkHour.UpdateWorkHour)

	// Pencatatan & Kelola Izin/Sakit/Cuti (Web Admin - Admin, Lurah, Sekertaris)
	wIzin := wProtected.Group("/izin", middleware.AllowRoles("admin", "lurah", "sekertaris"))
	wIzin.Post("/", h.Izin.CreateByAdmin)
	wIzin.Get("/pending", h.Izin.GetPending)
	wIzin.Put("/:id/approve", h.Izin.Approve)
}
