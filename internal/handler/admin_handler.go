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
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}

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
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil rekap laporan", err.Error())
	}

	// 3. Kembalikan response JSON pagination seragam
	return SendPaginated(c, fiber.StatusOK, "Data rekap laporan berhasil diambil", rekapData.Data, page, limit, rekapData.TotalData, rekapData.TotalPage)
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
	}

	reports, err := h.adminService.GetLaporanExportAdmin(filter)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data export", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil menarik seluruh data (siap diexport)", reports)
}

// GetDashboardSummary menghandle request GET /api/admin/dashboard/summary
func (h *AdminHandler) GetDashboardSummary(c fiber.Ctx) error {
	summary, err := h.adminService.GetDashboardSummaryAdmin()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil statistik dashboard", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Data dashboard admin berhasil diambil", summary)
}

// ---------------------------------------------------------
// PEGAWAI MANAGEMENT HANDLERS
// ---------------------------------------------------------

// GetPegawai menghandle request GET /api/admin/pegawai
func (h *AdminHandler) GetPegawai(c fiber.Ctx) error {
	// 1. Tangkap Query Parameters
	search := c.Query("search")
	role := c.Query("role")
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}

	filter := repository.AdminPegawaiFilter{
		Search: search,
		Role:   role,
		Page:   page,
		Limit:  limit,
	}

	// 2. Ambil data List Pegawai
	pegawaiData, err := h.adminService.GetPegawaiAdmin(filter)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data pegawai", err.Error())
	}

	// 3. Ambil data Statistik
	stats, err := h.adminService.GetPegawaiStatistikAdmin()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil statistik pegawai", err.Error())
	}

	// 4. Gabungkan Response (Sesuai request)
	return c.JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Berhasil mengambil daftar rekan kerja",
		"data": fiber.Map{
			"list":      pegawaiData.Data,
			"statistik": stats,
			"pagination": fiber.Map{
				"total_data":   pegawaiData.TotalData,
				"total_page":   pegawaiData.TotalPage,
				"total_pages":  pegawaiData.TotalPage,
				"total_items":  pegawaiData.TotalData,
				"current_page": pegawaiData.CurrentPage,
				"page":         pegawaiData.CurrentPage,
				"limit":        limit,
			},
		},
		"pagination": &Pagination{
			Page:        page,
			Limit:       limit,
			TotalItems:  pegawaiData.TotalData,
			TotalPages:  pegawaiData.TotalPage,
			CurrentPage: page,
			TotalData:   pegawaiData.TotalData,
		},
		"meta": &Pagination{
			Page:        page,
			Limit:       limit,
			TotalItems:  pegawaiData.TotalData,
			TotalPages:  pegawaiData.TotalPage,
			CurrentPage: page,
			TotalData:   pegawaiData.TotalData,
		},
	})
}

// CreatePegawai menghandle request POST /api/admin/pegawai
func (h *AdminHandler) CreatePegawai(c fiber.Ctx) error {
	var user domain.User
	if err := c.Bind().JSON(&user); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	// Minimal data NIP, Nama, Password, Role wajib ada
	if user.NIP == "" || user.Nama == "" || user.Password == "" || user.Role == "" {
		return SendError(c, fiber.StatusBadRequest, "NIP, Nama, Password, dan Role wajib diisi")
	}

	if err := h.adminService.CreatePegawaiAdmin(&user); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal membuat data pegawai", err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "Berhasil menambahkan pegawai", user)
}

// UpdatePegawai menghandle request PUT /api/admin/pegawai/:id
func (h *AdminHandler) UpdatePegawai(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID pegawai tidak valid")
	}

	var req domain.User
	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid", err.Error())
	}

	if err := h.adminService.UpdatePegawaiAdmin(uint(id), &req); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui data pegawai", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil memperbarui data pegawai", nil)
}

// DeletePegawai menghandle request DELETE /api/admin/pegawai/:id
func (h *AdminHandler) DeletePegawai(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID pegawai tidak valid")
	}

	if err := h.adminService.DeletePegawaiAdmin(uint(id)); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal menghapus data pegawai", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil menghapus data pegawai", nil)
}

// ResetPasswordPegawai menghandle request PUT /api/web/admin/pegawai/:id/reset-password
func (h *AdminHandler) ResetPasswordPegawai(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID pegawai tidak valid")
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}
	_ = c.Bind().JSON(&req)

	if err := h.adminService.ResetPasswordPegawaiAdmin(uint(id), req.NewPassword); err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Password pegawai berhasil direset", nil)
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
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}

	filter := repository.AdminPengumumanFilter{
		Search: search,
		Page:   page,
		Limit:  limit,
	}

	notifData, err := h.adminService.GetPengumumanAdmin(filter)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data pengumuman", err.Error())
	}

	stats, err := h.adminService.GetPengumumanStatistikAdmin()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil statistik pengumuman", err.Error())
	}

	// Mapping format ANNC dan Audience
	var listPengumuman []PengumumanResponseItem
	for _, n := range notifData.Data {
		uniqueID := fmt.Sprintf("ANNC-%d-%03d", n.CreatedAt.Year(), n.ID)

		audienceStr := "Semua Pegawai"
		if n.UserID != 0 {
			audienceStr = fmt.Sprintf("User Spesifik (%d)", n.UserID)
		}

		listPengumuman = append(listPengumuman, PengumumanResponseItem{
			IDPengumuman:   uniqueID,
			Judul:          n.Judul,
			Pesan:          n.Pesan,
			FormatAudience: audienceStr,
			Status:         "Aktif",
			Tanggal:        n.CreatedAt,
		})
	}

	// Format Output (Tabel dengan proper pagination & statistik dinamis)
	return c.JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Berhasil mengambil data pengumuman",
		"data": fiber.Map{
			"list": listPengumuman,
			"statistik": fiber.Map{
				"aktif":       stats.TotalPengumuman,
				"terjadwal":   0,
				"kedaluwarsa": 0,
			},
			"pagination": fiber.Map{
				"total_data":   notifData.TotalData,
				"total_page":   notifData.TotalPage,
				"total_pages":  notifData.TotalPage,
				"total_items":  notifData.TotalData,
				"current_page": notifData.CurrentPage,
				"page":         notifData.CurrentPage,
				"limit":        limit,
			},
		},
		"pagination": &Pagination{
			Page:        page,
			Limit:       limit,
			TotalItems:  notifData.TotalData,
			TotalPages:  notifData.TotalPage,
			CurrentPage: page,
			TotalData:   notifData.TotalData,
		},
		"meta": &Pagination{
			Page:        page,
			Limit:       limit,
			TotalItems:  notifData.TotalData,
			TotalPages:  notifData.TotalPage,
			CurrentPage: page,
			TotalData:   notifData.TotalData,
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
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	userID := 0
	if body.Audience != "Semua Pegawai" && body.Audience != "" {
	}

	pengumuman := domain.Notification{
		Judul:  body.Judul,
		Pesan:  body.Pesan,
		UserID: userID,
	}

	if err := h.adminService.CreatePengumumanAdmin(&pengumuman); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal membuat pengumuman", err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "Berhasil membuat pengumuman", pengumuman)
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
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	userID := 0
	pengumuman := domain.Notification{
		Judul:  body.Judul,
		Pesan:  body.Pesan,
		UserID: userID,
	}

	if err := h.adminService.UpdatePengumumanAdmin(uint(id), &pengumuman); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengupdate pengumuman", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil memperbarui pengumuman", nil)
}

func (h *AdminHandler) DeletePengumuman(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, _ := strconv.Atoi(idParam)

	if err := h.adminService.DeletePengumumanAdmin(uint(id)); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal menghapus pengumuman", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil menghapus pengumuman", nil)
}

// ---------------------------------------------------------
// SUPERVISOR LURAH HANDLERS
// ---------------------------------------------------------

func (h *AdminHandler) GetSupervisorLurah(c fiber.Ctx) error {
	supervisor, err := h.adminService.GetSupervisorLurahAdmin()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data supervisor lurah", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil mengambil data supervisor lurah", supervisor)
}

func (h *AdminHandler) UpdateSupervisorLurah(c fiber.Ctx) error {
	var body struct {
		Nama string `json:"nama"`
		NIP  string `json:"nip"`
	}

	if err := c.Bind().JSON(&body); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if err := h.adminService.UpdateSupervisorLurahAdmin(body.Nama, body.NIP); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui supervisor lurah", err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Berhasil memperbarui supervisor lurah", nil)
}
