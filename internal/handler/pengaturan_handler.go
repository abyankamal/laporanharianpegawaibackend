package handler

import (
	"laporanharianapi/internal/service"

	"github.com/gofiber/fiber/v3"
)

// PengaturanHandler menangani request ke pengaturan sistem.
type PengaturanHandler struct {
	service service.PengaturanService
}

// NewPengaturanHandler membuat instance baru PengaturanHandler.
func NewPengaturanHandler(service service.PengaturanService) *PengaturanHandler {
	return &PengaturanHandler{service: service}
}

// GetPengaturan mengambil data pengaturan saat ini.
func (h *PengaturanHandler) GetPengaturan(c fiber.Ctx) error {
	pengaturan, err := h.service.GetPengaturan()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil pengaturan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data pengaturan berhasil diambil",
		"data":    pengaturan,
	})
}

// UpdatePengaturanRequest adalah struct untuk request update pengaturan.
type UpdatePengaturanRequest struct {
	JamMasuk  string `json:"jam_masuk"`
	JamPulang string `json:"jam_pulang"`
}

// UpdatePengaturan memperbarui konfigurasi jam masuk dan pulang.
func (h *PengaturanHandler) UpdatePengaturan(c fiber.Ctx) error {
	var req UpdatePengaturanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	pengaturan, err := h.service.UpdatePengaturan(req.JamMasuk, req.JamPulang)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Pengaturan berhasil diperbarui",
		"data":    pengaturan,
	})
}
