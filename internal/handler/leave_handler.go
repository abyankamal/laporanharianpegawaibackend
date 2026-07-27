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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Pencatatan izin pegawai berhasil disimpan dan disetujui",
		"data":    izin,
	})
}

// Create menangani request pembuatan pengajuan izin baru.
// POST /api/mobile/izin/
func (h *IzinHandler) Create(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Pengajuan izin berhasil dibuat",
		"data":    izin,
	})
}

// GetMy menangani request daftar pengajuan izin milik user.
// GET /api/mobile/izin/
func (h *IzinHandler) GetMy(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
	}

	list, err := h.izinService.GetMyPengajuan(uint(userIDFloat))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil data pengajuan izin",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   list,
	})
}

// GetPending menangani request daftar pengajuan izin yang menunggu approval (Lurah).
// GET /api/mobile/izin/pending
func (h *IzinHandler) GetPending(c fiber.Ctx) error {
	list, err := h.izinService.GetPendingApprovals()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil data pengajuan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   list,
	})
}

// Approve menangani request approval/rejection pengajuan izin (Lurah).
// PUT /api/mobile/izin/:id/approve
func (h *IzinHandler) Approve(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
	}

	// Validasi role — hanya Lurah yang boleh approve
	role, _ := c.Locals("role").(string)
	if strings.ToLower(role) != "lurah" && strings.ToLower(role) != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status": "error", "message": "Hanya Lurah yang berhak menyetujui/menolak pengajuan izin",
		})
	}

	izinID, err := strconv.Atoi(c.Params("id"))
	if err != nil || izinID < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "ID pengajuan tidak valid",
		})
	}

	// Parse body
	type ApproveRequest struct {
		Approved bool   `json:"approved"`
		Komentar string `json:"komentar"`
	}

	var req ApproveRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Format request tidak valid",
		})
	}

	err = h.izinService.ApprovePengajuan(uint(izinID), uint(userIDFloat), req.Approved, req.Komentar)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": err.Error(),
		})
	}

	statusMsg := "disetujui"
	if !req.Approved {
		statusMsg = "ditolak"
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Pengajuan izin berhasil " + statusMsg,
	})
}
