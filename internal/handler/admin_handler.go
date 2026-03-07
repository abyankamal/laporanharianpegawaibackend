package handler

import (
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"

	"github.com/gofiber/fiber/v3"
)

type AdminHandler struct {
	adminService service.AdminService
}

func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

// GetRekapLaporan menghandle request GET /api/admin/rekap-laporan
func (h *AdminHandler) GetRekapLaporan(c fiber.Ctx) error {
	// 1. Ekstrak parameter query string dari URL
	// Contoh: /api/admin/rekap-laporan?start_date=2023-10-01&end_date=2023-10-31&search=Budi
	filter := repository.AdminReportFilter{
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
		StatusWaktu: c.Query("status_waktu"),
		Search:      c.Query("search"),
	}

	// 2. Panggil service untuk mengambil data sesuai filter
	laporanList, err := h.adminService.GetRekapLaporanAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil rekap laporan",
			"error":   err.Error(),
		})
	}

	// 3. Kembalikan response JSON yang rapi
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Berhasil mengambil rekap laporan admin",
		"data":    laporanList,
	})
}

// GetDashboardSummary menghandle request GET /api/admin/dashboard/summary
func (h *AdminHandler) GetDashboardSummary(c fiber.Ctx) error {
	summary, err := h.adminService.GetDashboardSummaryAdmin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil statistik dashboard",
			"error":   err.Error(),
		})
	}

	// Memastikan kembalian JSON persis seperti yang direquest
	// { "success": true, "data": { "statistik": ..., "laporan_terbaru": ..., "notifikasi": ..., "agenda": ... } }
	return c.JSON(fiber.Map{
		"success": true,
		"data":    summary, // summary secara otomatis berbentuk object JSON karena struct response layer repository
	})
}
