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
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil pengaturan")
	}

	return SendSuccess(c, fiber.StatusOK, "Data pengaturan berhasil diambil", workHour)
}

// UpdateWorkHourRequest adalah struct untuk request update pengaturan jam kerja.
type UpdateWorkHourRequest struct {
	JamMasuk       string `json:"jam_masuk" validate:"required,min=5,max=5"`
	JamPulang      string `json:"jam_pulang" validate:"required,min=5,max=5"`
	JamMasukJumat  string `json:"jam_masuk_jumat" validate:"required,min=5,max=5"`
	JamPulangJumat string `json:"jam_pulang_jumat" validate:"required,min=5,max=5"`
}

// UpdateWorkHour memperbarui konfigurasi jam masuk dan pulang.
func (h *WorkHourHandler) UpdateWorkHour(c fiber.Ctx) error {
	var req UpdateWorkHourRequest
	if !BindAndValidate(c, &req) {
		return nil
	}

	workHour, err := h.service.UpdateWorkHour(req.JamMasuk, req.JamPulang, req.JamMasukJumat, req.JamPulangJumat)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Pengaturan jam kerja berhasil diperbarui", workHour)
}

// UpdateGeofencingRequest adalah struct untuk request update koordinat geofencing.
type UpdateGeofencingRequest struct {
	KantorLat         *string `json:"kantor_lat" validate:"required"`
	KantorLong        *string `json:"kantor_long" validate:"required"`
	RadiusMeter       int     `json:"radius_meter" validate:"min=1"`
	GeofencingEnabled bool    `json:"geofencing_enabled"`
}

// UpdateGeofencing memperbarui konfigurasi koordinat dan radius geofencing kantor.
func (h *WorkHourHandler) UpdateGeofencing(c fiber.Ctx) error {
	var req UpdateGeofencingRequest
	if !BindAndValidate(c, &req) {
		return nil
	}

	workHour, err := h.service.UpdateGeofencing(req.KantorLat, req.KantorLong, req.RadiusMeter, req.GeofencingEnabled)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Pengaturan geofencing berhasil diperbarui", workHour)
}
