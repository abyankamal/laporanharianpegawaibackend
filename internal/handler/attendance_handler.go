package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/service"
)

// AbsensiHandler menangani HTTP request untuk fitur absensi.
type AbsensiHandler struct {
	absensiService service.AbsensiService
	userService    service.UserService
}

// NewAbsensiHandler membuat instance baru AbsensiHandler.
func NewAbsensiHandler(absensiService service.AbsensiService, userService service.UserService) *AbsensiHandler {
	return &AbsensiHandler{
		absensiService: absensiService,
		userService:    userService,
	}
}

// CheckIn menangani request absensi masuk.
// POST /api/mobile/absensi/check-in
func (h *AbsensiHandler) CheckIn(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	faceVerified := c.FormValue("face_verified") == "true"
	lokasiLat := c.FormValue("lokasi_lat")
	lokasiLong := c.FormValue("lokasi_long")

	// Ambil file selfie
	fileSelfie, _ := c.FormFile("selfie")

	if fileSelfie == nil {
		return SendError(c, fiber.StatusBadRequest, "Selfie wajib dikirim untuk verifikasi absensi")
	}

	input := service.AbsensiCheckInInput{
		UserID:       uint(userIDFloat),
		LokasiLat:    lokasiLat,
		LokasiLong:   lokasiLong,
		FaceVerified: faceVerified,
		FileSelfie:   fileSelfie,
	}

	absensi, err := h.absensiService.CheckIn(input)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Absensi masuk berhasil", absensi)
}

// CheckOut menangani request absensi pulang.
// POST /api/mobile/absensi/check-out
func (h *AbsensiHandler) CheckOut(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	faceVerified := c.FormValue("face_verified") == "true"
	lokasiLat := c.FormValue("lokasi_lat")
	lokasiLong := c.FormValue("lokasi_long")

	fileSelfie, _ := c.FormFile("selfie")
	if fileSelfie == nil {
		return SendError(c, fiber.StatusBadRequest, "Selfie wajib dikirim untuk verifikasi absensi")
	}

	input := service.AbsensiCheckOutInput{
		UserID:       uint(userIDFloat),
		LokasiLat:    lokasiLat,
		LokasiLong:   lokasiLong,
		FaceVerified: faceVerified,
		FileSelfie:   fileSelfie,
	}

	absensi, err := h.absensiService.CheckOut(input)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusOK, "Absensi pulang berhasil", absensi)
}

// GetTodayStatus menangani request status absensi hari ini.
// GET /api/mobile/absensi/today
func (h *AbsensiHandler) GetTodayStatus(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	absensi, isWorkday, err := h.absensiService.GetTodayStatus(uint(userIDFloat))
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil status absensi")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success":    true,
		"status":     "success",
		"message":    "Status absensi berhasil diambil",
		"is_workday": isWorkday,
		"data": fiber.Map{
			"absensi":    absensi,
			"is_workday": isWorkday,
		},
	})
}

// GetMonthlyRecap menangani request rekap absensi bulanan per user.
// GET /api/mobile/absensi/recap?bulan=1&tahun=2026
func (h *AbsensiHandler) GetMonthlyRecap(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}

	bulan, _ := strconv.Atoi(c.Query("bulan"))
	tahun, _ := strconv.Atoi(c.Query("tahun"))

	if bulan < 1 || bulan > 12 || tahun < 2020 {
		return SendError(c, fiber.StatusBadRequest, "Parameter bulan (1-12) dan tahun wajib diisi")
	}

	// Jika ada query user_id dan requester adalah lurah/sekretaris, ambil data user lain
	targetUserID := uint(userIDFloat)
	if qUserID := c.Query("user_id"); qUserID != "" {
		requesterRole, _ := c.Locals("role").(string)
		roleLower := strings.ToLower(requesterRole)
		if roleLower == "lurah" || roleLower == "sekertaris" || roleLower == "sekretaris" || roleLower == "admin" {
			if id, err := strconv.Atoi(qUserID); err == nil && id > 0 {
				targetUserID = uint(id)
			}
		}
	}

	details, recap, err := h.absensiService.GetMonthlyRecap(targetUserID, bulan, tahun)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil rekap absensi")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Rekap absensi berhasil diambil",
		"recap":   recap,
		"details": details,
		"data": fiber.Map{
			"recap":   recap,
			"details": details,
		},
	})
}

// GetAllRecap menangani request rekap absensi semua pegawai (Lurah/Sekretaris).
// GET /api/mobile/absensi/recap/all?bulan=1&tahun=2026
func (h *AbsensiHandler) GetAllRecap(c fiber.Ctx) error {
	bulan, _ := strconv.Atoi(c.Query("bulan"))
	tahun, _ := strconv.Atoi(c.Query("tahun"))

	if bulan < 1 || bulan > 12 || tahun < 2020 {
		return SendError(c, fiber.StatusBadRequest, "Parameter bulan (1-12) dan tahun wajib diisi")
	}

	// Ambil semua user (exclude admin)
	allUsers, err := h.userService.GetAllUsers()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data pegawai")
	}

	// Filter hanya pegawai non-admin
	var users []domain.User
	for _, u := range allUsers {
		if strings.ToLower(u.Role) != "admin" {
			users = append(users, u)
		}
	}

	recaps, err := h.absensiService.GetAllMonthlyRecap(bulan, tahun, users)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil rekap absensi")
	}

	return SendSuccess(c, fiber.StatusOK, "Rekap seluruh absensi berhasil diambil", recaps)
}
