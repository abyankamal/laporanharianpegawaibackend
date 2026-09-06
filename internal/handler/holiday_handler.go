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
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil daftar hari libur")
	}

	return SendSuccess(c, fiber.StatusOK, "Daftar hari libur berhasil diambil", holidays)
}

// CreateHolidayRequest adalah struct untuk request create hari libur.
type CreateHolidayRequest struct {
	TanggalMulai   string `json:"tanggal_mulai" validate:"required,min=10,max=10"`
	TanggalSelesai string `json:"tanggal_selesai" validate:"required,min=10,max=10"`
	Keterangan     string `json:"keterangan" validate:"required"`
}

// CreateHoliday menyimpan data hari libur baru.
func (h *HolidayHandler) CreateHoliday(c fiber.Ctx) error {
	var req CreateHolidayRequest
	if !BindAndValidate(c, &req) {
		return nil
	}

	holiday, err := h.service.CreateHoliday(req.TanggalMulai, req.TanggalSelesai, req.Keterangan)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "Hari libur berhasil ditambahkan", holiday)
}

// UpdateHolidayRequest request body
type UpdateHolidayRequest struct {
	TanggalMulai   string `json:"tanggal_mulai" validate:"required,min=10,max=10"`
	TanggalSelesai string `json:"tanggal_selesai" validate:"required,min=10,max=10"`
	Keterangan     string `json:"keterangan" validate:"required"`
}

// UpdateHoliday memperbarui data hari libur.
func (h *HolidayHandler) UpdateHoliday(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID hari libur tidak valid")
	}

	var req UpdateHolidayRequest
	if !BindAndValidate(c, &req) {
		return nil
	}

	holiday, err := h.service.UpdateHoliday(uint(id), req.TanggalMulai, req.TanggalSelesai, req.Keterangan)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Hari libur berhasil diperbarui", holiday)
}

// DeleteHoliday menghapus hari libur tertentu.
func (h *HolidayHandler) DeleteHoliday(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID hari libur tidak valid")
	}

	err = h.service.DeleteHoliday(uint(id))
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal menghapus hari libur")
	}

	return SendSuccess(c, fiber.StatusOK, "Hari libur berhasil dihapus", nil)
}
