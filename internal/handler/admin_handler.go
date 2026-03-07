package handler

import (
	"strconv"

	"laporanharianapi/internal/domain"
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
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	filter := repository.AdminReportFilter{
		StartDate:    c.Query("start_date"),
		EndDate:      c.Query("end_date"),
		StatusWaktu:  c.Query("status_waktu"),
		StatusReview: c.Query("status_review"),
		Search:       c.Query("search"),
		Page:         page,
		Limit:        limit,
	}

	// 2. Panggil service untuk mengambil data sesuai filter
	rekapData, err := h.adminService.GetRekapLaporanAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil rekap laporan",
			"error":   err.Error(),
		})
	}

	// 3. Kembalikan response JSON pagination seragam
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Data rekap laporan berhasil diambil",
		"data":    rekapData.Data,
		"pagination": fiber.Map{
			"current_page": rekapData.CurrentPage,
			"limit":        limit,
			"total_data":   rekapData.TotalData,
			"total_pages":  rekapData.TotalPage,
		},
	})
}

// GetLaporanExport menghandle request export data Excel/PDF
// Endpoint ini menggunakan filter yang sama tanpa limit dan offset
func (h *AdminHandler) GetLaporanExport(c fiber.Ctx) error {
	filter := repository.AdminReportFilter{
		StartDate:    c.Query("start_date"),
		EndDate:      c.Query("end_date"),
		StatusWaktu:  c.Query("status_waktu"),
		StatusReview: c.Query("status_review"),
		Search:       c.Query("search"),
		// Tanpa Page / Limit -> Menarik semua data yang match dengan filter
	}

	reports, err := h.adminService.GetLaporanExportAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data export",
			"error":   err.Error(),
		})
	}

	// TODO: Di sinilah Anda bisa menambahkan logika library Golang (misal excelize atau gofpdf)
	// untuk mengolah array `reports` ke dalam bentuk file binary (.xlsx atau .pdf)
	// Untuk saat ini, fungsi akan me-return JSON murni yang berisi seluruh data terfilter.

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil menarik seluruh data (siap diexport)",
		"total":   len(reports),
		"data":    reports,
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

// ---------------------------------------------------------
// PEGAWAI MANAGEMENT HANDLERS
// ---------------------------------------------------------

// GetPegawai menghandle request GET /api/admin/pegawai
func (h *AdminHandler) GetPegawai(c fiber.Ctx) error {
	// 1. Tangkap Query Parameters
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	filter := repository.AdminPegawaiFilter{
		Search: search,
		Page:   page,
		Limit:  limit,
	}

	// 2. Ambil data List Pegawai
	pegawaiData, err := h.adminService.GetPegawaiAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data pegawai",
			"error":   err.Error(),
		})
	}

	// 3. Ambil data Statistik
	stats, err := h.adminService.GetPegawaiStatistikAdmin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil statistik pegawai",
			"error":   err.Error(),
		})
	}

	// 4. Gabungkan Response (Sesuai request)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil mengambil daftar pegawai",
		"data": fiber.Map{
			"list":      pegawaiData.Data,
			"statistik": stats,
			"pagination": fiber.Map{
				"total_data":   pegawaiData.TotalData,
				"total_page":   pegawaiData.TotalPage,
				"current_page": pegawaiData.CurrentPage,
			},
		},
	})
}

// CreatePegawai menghandle request POST /api/admin/pegawai
func (h *AdminHandler) CreatePegawai(c fiber.Ctx) error {
	var user domain.User
	if err := c.Bind().JSON(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
			"error":   err.Error(),
		})
	}

	// Minimal data NIP, Nama, Password, Role wajib ada
	if user.NIP == "" || user.Nama == "" || user.Password == "" || user.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "NIP, Nama, Password, dan Role wajib diisi",
		})
	}

	if err := h.adminService.CreatePegawaiAdmin(&user); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat data pegawai",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil menambahkan pegawai",
		"data":    user,
	})
}

// UpdatePegawai menghandle request PUT /api/admin/pegawai/:id
func (h *AdminHandler) UpdatePegawai(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID pegawai tidak valid",
		})
	}

	var req domain.User
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
			"error":   err.Error(),
		})
	}

	if err := h.adminService.UpdatePegawaiAdmin(uint(id), &req); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal memperbarui data pegawai",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil memperbarui data pegawai",
	})
}

// DeletePegawai menghandle request DELETE /api/admin/pegawai/:id
func (h *AdminHandler) DeletePegawai(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "ID pegawai tidak valid",
		})
	}

	if err := h.adminService.DeletePegawaiAdmin(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus data pegawai",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil menghapus data pegawai",
	})
}
