package handler

import (
	"laporanharianapi/internal/service"

	"github.com/gofiber/fiber/v3"
)

// WorkHourHandler menangani request ke pengaturan sistem.
type WorkHourHandler struct {
	service service.WorkHourService
}

// NewWorkHourHandler membuat instance baru WorkHourHandler.
func NewWorkHourHandler(service service.WorkHourService) *WorkHourHandler {
	return &WorkHourHandler{service: service}
}

// GetWorkHour mengambil data pengaturan saat ini.
func (h *WorkHourHandler) GetWorkHour(c fiber.Ctx) error {
	workHour, err := h.service.GetWorkHour()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil pengaturan",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data pengaturan berhasil diambil",
		"data":    workHour,
	})
}

// UpdateWorkHourRequest adalah struct untuk request update pengaturan.
type UpdateWorkHourRequest struct {
	JamMasuk  string `json:"jam_masuk"`
	JamPulang string `json:"jam_pulang"`
}

// UpdateWorkHour memperbarui konfigurasi jam masuk dan pulang.
func (h *WorkHourHandler) UpdateWorkHour(c fiber.Ctx) error {
	var req UpdateWorkHourRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	workHour, err := h.service.UpdateWorkHour(req.JamMasuk, req.JamPulang)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Pengaturan berhasil diperbarui",
		"data":    workHour,
	})
}
