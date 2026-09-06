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
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data jabatan: "+err.Error())
	}

	var response []JabatanModelResponse
	for _, j := range jabatans {
		response = append(response, JabatanModelResponse{
			ID:   j.ID,
			Nama: j.NamaJabatan,
		})
	}

	return SendSuccess(c, fiber.StatusOK, "Data jabatan berhasil diambil", response)
}

func (h *JabatanHandler) GetOne(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	jabatan, err := h.jabatanService.GetJabatanByID(uint(id))
	if err != nil {
		return SendError(c, fiber.StatusNotFound, "Jabatan tidak ditemukan")
	}

	return SendSuccess(c, fiber.StatusOK, "Data jabatan berhasil diambil", JabatanModelResponse{
		ID:   jabatan.ID,
		Nama: jabatan.NamaJabatan,
	})
}

func (h *JabatanHandler) Create(c fiber.Ctx) error {
	var req CreateJabatanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	if req.Nama == "" {
		return SendError(c, fiber.StatusBadRequest, "Nama jabatan wajib diisi")
	}

	jabatan, err := h.jabatanService.CreateJabatan(req.Nama)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal membuat jabatan: "+err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "Jabatan berhasil dibuat", JabatanModelResponse{
		ID:   jabatan.ID,
		Nama: jabatan.NamaJabatan,
	})
}

func (h *JabatanHandler) Update(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	var req CreateJabatanRequest
	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	jabatan, err := h.jabatanService.UpdateJabatan(uint(id), req.Nama)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal memperbarui jabatan: "+err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Jabatan berhasil diperbarui", JabatanModelResponse{
		ID:   jabatan.ID,
		Nama: jabatan.NamaJabatan,
	})
}

func (h *JabatanHandler) Delete(c fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	if err := h.jabatanService.DeleteJabatan(uint(id)); err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal menghapus jabatan: "+err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Jabatan berhasil dihapus", nil)
}
