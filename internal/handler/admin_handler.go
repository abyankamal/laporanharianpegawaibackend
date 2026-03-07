package handler

import (
	"fmt"
	"strconv"
	"time"

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

// ---------------------------------------------------------
// PENGUMUMAN MANAGEMENT HANDLERS
// ---------------------------------------------------------

// PengumumanResponseItem struktur custom untuk response sesuai requirement UI
type PengumumanResponseItem struct {
	IDPengumuman   string    `json:"id_pengumuman"` // Format: ANNC-2026-001
	Judul          string    `json:"judul"`
	Pesan          string    `json:"pesan"`
	FormatAudience string    `json:"audience"` // "Semua Pegawai", dll berdasarkan user_id
	Status         string    `json:"status"`   // Hardcoded "Aktif" (karena status kolom di model asli tidak dimodifikasi)
	Tanggal        time.Time `json:"tanggal"`  // mapped dari CreatedAt
}

func (h *AdminHandler) GetPengumuman(c fiber.Ctx) error {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	filter := repository.AdminPengumumanFilter{
		Search: search,
		Page:   page,
		Limit:  limit,
	}

	notifData, err := h.adminService.GetPengumumanAdmin(filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil data pengumuman",
			"error":   err.Error(),
		})
	}

	stats, err := h.adminService.GetPengumumanStatistikAdmin()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengambil statistik pengumuman",
			"error":   err.Error(),
		})
	}

	// Mapping format ANNC dan Audience
	var listPengumuman []PengumumanResponseItem
	for _, n := range notifData.Data {
		// Buat unique ID ANNC-Tahun-ID
		uniqueID := fmt.Sprintf("ANNC-%d-%03d", n.CreatedAt.Year(), n.ID)

		// Evaluasi Audience dari user_id (Bisa 0 karena kita akali untuk Semua Pegawai, dsb)
		audienceStr := "Semua Pegawai"
		if n.UserID != 0 {
			audienceStr = fmt.Sprintf("User Spesifik (%d)", n.UserID)
		}

		listPengumuman = append(listPengumuman, PengumumanResponseItem{
			IDPengumuman:   uniqueID,
			Judul:          n.Judul,
			Pesan:          n.Pesan,
			FormatAudience: audienceStr,
			Status:         "Aktif", // Tuntutan bypass
			Tanggal:        n.CreatedAt,
		})
	}

	// Format Output (Tabel dengan proper pagination & statistik dinamis)
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil mengambil data pengumuman",
		"data": fiber.Map{
			"list": listPengumuman,
			"statistik": fiber.Map{
				"aktif":       stats.TotalPengumuman,
				"terjadwal":   0, // Hardcoded sesuai request drop feature
				"kedaluwarsa": 0,
			},
			"pagination": fiber.Map{
				"total_data":   notifData.TotalData,
				"total_page":   notifData.TotalPage,
				"current_page": notifData.CurrentPage,
			},
		},
	})
}

func (h *AdminHandler) CreatePengumuman(c fiber.Ctx) error {
	var body struct {
		Judul    string `json:"judul"`
		Pesan    string `json:"pesan"`
		Audience string `json:"audience"` // bisa 'Semua Pegawai'
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	// Default user_id ke 0 = "Semua Pegawai"
	userID := 0
	if body.Audience != "Semua Pegawai" && body.Audience != "" {
		// Logika kompleks lain jika audience adalah role spesifik (di luar scope table constraints saat ini tanpa ubah model)
		// Kita biarkan 0 jika tidak ada
	}

	pengumuman := domain.Notification{
		Judul:  body.Judul,
		Pesan:  body.Pesan,
		UserID: userID,
	}

	if err := h.adminService.CreatePengumumanAdmin(&pengumuman); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal membuat pengumuman",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "Berhasil membuat pengumuman",
		"data":    pengumuman,
	})
}

func (h *AdminHandler) UpdatePengumuman(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, _ := strconv.Atoi(idParam)

	var body struct {
		Judul    string `json:"judul"`
		Pesan    string `json:"pesan"`
		Audience string `json:"audience"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"message": "Format request tidak valid",
		})
	}

	userID := 0 // default 'Semua'
	pengumuman := domain.Notification{
		Judul:  body.Judul,
		Pesan:  body.Pesan,
		UserID: userID,
	}

	if err := h.adminService.UpdatePengumumanAdmin(uint(id), &pengumuman); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal mengupdate pengumuman",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil memperbarui pengumuman",
	})
}

func (h *AdminHandler) DeletePengumuman(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, _ := strconv.Atoi(idParam)

	if err := h.adminService.DeletePengumumanAdmin(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"message": "Gagal menghapus pengumuman",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Berhasil menghapus pengumuman",
	})
}
