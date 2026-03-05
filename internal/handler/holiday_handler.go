package handler

import (
	"strconv"

	"laporanharianapi/internal/service"

	"github.com/gofiber/fiber/v3"
)

// HolidayHandler menangani API kelola jadwal hari libur.
type HolidayHandler struct {
	service service.HolidayService
}

// NewHolidayHandler membuat instance baru HolidayHandler.
func NewHolidayHandler(service service.HolidayService) *HolidayHandler {
	return &HolidayHandler{service: service}
}

// GetHolidays mengambil semua tanggal merah/hari libur.
func (h *HolidayHandler) GetHolidays(c fiber.Ctx) error {
	holidays, err := h.service.GetHolidays()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil daftar hari libur",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Daftar hari libur berhasil diambil",
		"data":    holidays,
	})
}

// CreateHolidayRequest adalah struct untuk request create hari libur.
type CreateHolidayRequest struct {
	Tanggal    string `json:"tanggal"`
	Keterangan string `json:"keterangan"`
}

// CreateHoliday menyimpan data hari libur baru.
func (h *HolidayHandler) CreateHoliday(c fiber.Ctx) error {
	var req CreateHolidayRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	holiday, err := h.service.CreateHoliday(req.Tanggal, req.Keterangan)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Hari libur berhasil ditambahkan",
		"data":    holiday,
	})
}

// DeleteHoliday menghapus hari libur tertentu.
func (h *HolidayHandler) DeleteHoliday(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID hari libur tidak valid",
		})
	}

	err = h.service.DeleteHoliday(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menghapus hari libur",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Hari libur berhasil dihapus",
	})
}
