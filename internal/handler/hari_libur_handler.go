package handler

import (
	"strconv"

	"laporanharianapi/internal/service"

	"github.com/gofiber/fiber/v3"
)

// HariLiburHandler menangani API kelola jadwal hari libur.
type HariLiburHandler struct {
	service service.HariLiburService
}

// NewHariLiburHandler membuat instance baru HariLiburHandler.
func NewHariLiburHandler(service service.HariLiburService) *HariLiburHandler {
	return &HariLiburHandler{service: service}
}

// GetHariLibur mengambil semua tanggal merah/hari libur.
func (h *HariLiburHandler) GetHariLibur(c fiber.Ctx) error {
	hariLibur, err := h.service.GetHariLibur()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil daftar hari libur",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Daftar hari libur berhasil diambil",
		"data":    hariLibur,
	})
}

// CreateHariLiburRequest adalah struct untuk request create hari libur.
type CreateHariLiburRequest struct {
	Tanggal    string `json:"tanggal"`
	Keterangan string `json:"keterangan"`
}

// CreateHariLibur menyimpan data hari libur baru.
func (h *HariLiburHandler) CreateHariLibur(c fiber.Ctx) error {
	var req CreateHariLiburRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	hariLibur, err := h.service.CreateHariLibur(req.Tanggal, req.Keterangan)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Hari libur berhasil ditambahkan",
		"data":    hariLibur,
	})
}

// DeleteHariLibur menghapus hari libur tertentu.
func (h *HariLiburHandler) DeleteHariLibur(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID hari libur tidak valid",
		})
	}

	err = h.service.DeleteHariLibur(uint(id))
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
