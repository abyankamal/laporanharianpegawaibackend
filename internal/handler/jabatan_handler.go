package handler

import (
	"laporanharianapi/internal/service"

	"github.com/gofiber/fiber/v3"
)

type JabatanHandler struct {
	jabatanService service.JabatanService
}

func NewJabatanHandler(jabatanService service.JabatanService) *JabatanHandler {
	return &JabatanHandler{jabatanService: jabatanService}
}

type JabatanModelResponse struct {
	ID         uint   `json:"id"`
	Nama       string `json:"nama"`
	Keterangan string `json:"keterangan,omitempty"`
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
		"status":  "success",
		"message": "Data jabatan berhasil diambil",
		"data":    response,
	})
}
