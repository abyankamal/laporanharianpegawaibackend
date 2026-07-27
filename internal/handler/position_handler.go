package handler

import (
	"laporanharianapi/internal/service"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

type JabatanHandler struct {
	jabatanService service.JabatanService
}

func NewJabatanHandler(jabatanService service.JabatanService) *JabatanHandler {
	return &JabatanHandler{jabatanService: jabatanService}
}

type JabatanModelResponse struct {
	ID   uint   `json:"id"`
	Nama string `json:"nama"`
}

type CreateJabatanRequest struct {
	Nama string `json:"nama"`
}

func (h *JabatanHandler) GetAll(c fiber.Ctx) error {
	jabatans, err := h.jabatanService.GetAllJabatan()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data jabatan: " + err.Error(),
		})
	}

	var response []JabatanModelResponse
	for _, j := range jabatans {
		response = append(response, JabatanModelResponse{
			ID:   j.ID,
			Nama: j.NamaJabatan,
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Data jabatan berhasil diambil",
		"data":    response,
	})
}

func (h *JabatanHandler) GetOne(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tidak valid",
		})
	}

	jabatan, err := h.jabatanService.GetJabatanByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Jabatan tidak ditemukan",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Data jabatan berhasil diambil",
		"data": JabatanModelResponse{
			ID:   jabatan.ID,
			Nama: jabatan.NamaJabatan,
		},
	})
}

func (h *JabatanHandler) Create(c fiber.Ctx) error {
	var req CreateJabatanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	if req.Nama == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Nama jabatan wajib diisi",
		})
	}

	jabatan, err := h.jabatanService.CreateJabatan(req.Nama)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal membuat jabatan: " + err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Jabatan berhasil dibuat",
		"data": JabatanModelResponse{
			ID:   jabatan.ID,
			Nama: jabatan.NamaJabatan,
		},
	})
}

func (h *JabatanHandler) Update(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tidak valid",
		})
	}

	var req CreateJabatanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	jabatan, err := h.jabatanService.UpdateJabatan(uint(id), req.Nama)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal memperbarui jabatan: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Jabatan berhasil diperbarui",
		"data": JabatanModelResponse{
			ID:   jabatan.ID,
			Nama: jabatan.NamaJabatan,
		},
	})
}

func (h *JabatanHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tidak valid",
		})
	}

	if err := h.jabatanService.DeleteJabatan(uint(id)); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal menghapus jabatan: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Jabatan berhasil dihapus",
	})
}
