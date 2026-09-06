package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// IzinHandler menangani HTTP request untuk fitur pengajuan izin.
type IzinHandler struct {
	izinService service.IzinService
}

// NewIzinHandler membuat instance baru IzinHandler.
func NewIzinHandler(izinService service.IzinService) *IzinHandler {
	return &IzinHandler{izinService: izinService}
}

// CreateByAdmin menangani pencatatan izin/sakit/cuti pegawai oleh Admin/Lurah dari Web Admin.
// POST /api/web/izin
func (h *IzinHandler) CreateByAdmin(c fiber.Ctx) error {
	adminIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	targetUserID, _ := strconv.Atoi(c.FormValue("user_id"))
	jenisIzin := c.FormValue("jenis_izin")
	tanggalMulai := c.FormValue("tanggal_mulai")
	tanggalSelesai := c.FormValue("tanggal_selesai")
	keterangan := c.FormValue("keterangan")

	fileDokumen, _ := c.FormFile("dokumen")

	input := service.PengajuanIzinInput{
		UserID:         uint(targetUserID),
		JenisIzin:      jenisIzin,
		TanggalMulai:   tanggalMulai,
		TanggalSelesai: tanggalSelesai,
		Keterangan:     keterangan,
		FileDokumen:    fileDokumen,
	}

	izin, err := h.izinService.CreateByAdmin(input, uint(adminIDFloat))
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "Pencatatan izin pegawai berhasil disimpan dan disetujui", izin)
}

// Create menangani request pembuatan pengajuan izin baru.
// POST /api/mobile/izin/
func (h *IzinHandler) Create(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	jenisIzin := c.FormValue("jenis_izin")
	tanggalMulai := c.FormValue("tanggal_mulai")
	tanggalSelesai := c.FormValue("tanggal_selesai")
	keterangan := c.FormValue("keterangan")

	fileDokumen, _ := c.FormFile("dokumen")

	input := service.PengajuanIzinInput{
		UserID:         uint(userIDFloat),
		JenisIzin:      jenisIzin,
		TanggalMulai:   tanggalMulai,
		TanggalSelesai: tanggalSelesai,
		Keterangan:     keterangan,
		FileDokumen:    fileDokumen,
	}

	izin, err := h.izinService.CreatePengajuan(input)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "Pengajuan izin berhasil dibuat", izin)
}

// GetMy menangani request daftar pengajuan izin milik user.
// GET /api/mobile/izin/
func (h *IzinHandler) GetMy(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	list, err := h.izinService.GetMyPengajuan(uint(userIDFloat))
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data pengajuan izin")
	}

	return SendSuccess(c, fiber.StatusOK, "Data pengajuan izin berhasil diambil", list)
}

// GetPending menangani request daftar pengajuan izin yang menunggu approval (Lurah).
// GET /api/mobile/izin/pending
func (h *IzinHandler) GetPending(c fiber.Ctx) error {
	list, err := h.izinService.GetPendingApprovals()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data pengajuan")
	}

	return SendSuccess(c, fiber.StatusOK, "Data pengajuan pending berhasil diambil", list)
}

// GetAll menangani request seluruh daftar pengajuan izin (Web Admin).
// GET /api/web/izin
func (h *IzinHandler) GetAll(c fiber.Ctx) error {
	page, limit, _ := ParsePagination(c, 10)
	list, totalItems, err := h.izinService.GetAllPengajuan(page, limit)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data pengajuan izin", err.Error())
	}

	totalPages := CalculateTotalPages(totalItems, limit)
	return SendPaginated(c, fiber.StatusOK, "Data pengajuan izin berhasil diambil", list, page, limit, totalItems, totalPages)
}

// Approve menangani request approval/rejection pengajuan izin (Lurah).
// PUT /api/mobile/izin/:id/approve
func (h *IzinHandler) Approve(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	// Validasi role — hanya Lurah yang boleh approve
	role, _ := c.Locals("role").(string)
	if strings.ToLower(role) != "lurah" && strings.ToLower(role) != "admin" {
		return SendError(c, fiber.StatusForbidden, "Hanya Lurah yang berhak menyetujui/menolak pengajuan izin")
	}

	izinID, err := strconv.Atoi(c.Params("id"))
	if err != nil || izinID < 1 {
		return SendError(c, fiber.StatusBadRequest, "ID pengajuan tidak valid")
	}

	// Parse body
	type ApproveRequest struct {
		Approved *bool  `json:"approved"`
		Status   string `json:"status"`
		Komentar string `json:"komentar"`
	}

	var req ApproveRequest
	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	isApproved := false
	if req.Approved != nil {
		isApproved = *req.Approved
	} else if req.Status != "" {
		st := strings.ToLower(strings.TrimSpace(req.Status))
		isApproved = (st == "disetujui" || st == "approved")
	}

	err = h.izinService.ApprovePengajuan(uint(izinID), uint(userIDFloat), isApproved, req.Komentar)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	statusMsg := "disetujui"
	if !isApproved {
		statusMsg = "ditolak"
	}

	return SendSuccess(c, fiber.StatusOK, "Pengajuan izin berhasil "+statusMsg, nil)
}
