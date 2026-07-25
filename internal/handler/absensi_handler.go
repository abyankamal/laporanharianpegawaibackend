package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
	}

	faceVerified := c.FormValue("face_verified") == "true"
	lokasiLat := c.FormValue("lokasi_lat")
	lokasiLong := c.FormValue("lokasi_long")

	// Ambil file selfie
	fileSelfie, _ := c.FormFile("selfie")

	if fileSelfie == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Selfie wajib dikirim untuk verifikasi absensi",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Absensi masuk berhasil",
		"data":    absensi,
	})
}

// CheckOut menangani request absensi pulang.
// POST /api/mobile/absensi/check-out
func (h *AbsensiHandler) CheckOut(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
	}

	faceVerified := c.FormValue("face_verified") == "true"
	lokasiLat := c.FormValue("lokasi_lat")
	lokasiLong := c.FormValue("lokasi_long")

	fileSelfie, _ := c.FormFile("selfie")
	if fileSelfie == nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Selfie wajib dikirim untuk verifikasi absensi",
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Absensi pulang berhasil",
		"data":    absensi,
	})
}

// GetTodayStatus menangani request status absensi hari ini.
// GET /api/mobile/absensi/today
func (h *AbsensiHandler) GetTodayStatus(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
	}

	absensi, isWorkday, err := h.absensiService.GetTodayStatus(uint(userIDFloat))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil status absensi",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":     "success",
		"is_workday": isWorkday,
		"data":       absensi,
	})
}

// GetMonthlyRecap menangani request rekap absensi bulanan per user.
// GET /api/mobile/absensi/recap?bulan=1&tahun=2026
func (h *AbsensiHandler) GetMonthlyRecap(c fiber.Ctx) error {
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status": "error", "message": "User tidak terautentikasi",
		})
	}

	bulan, _ := strconv.Atoi(c.Query("bulan"))
	tahun, _ := strconv.Atoi(c.Query("tahun"))

	if bulan < 1 || bulan > 12 || tahun < 2020 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Parameter bulan (1-12) dan tahun wajib diisi",
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil rekap absensi",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"recap":   recap,
		"details": details,
	})
}

// GetAllRecap menangani request rekap absensi semua pegawai (Lurah/Sekretaris).
// GET /api/mobile/absensi/recap/all?bulan=1&tahun=2026
func (h *AbsensiHandler) GetAllRecap(c fiber.Ctx) error {
	bulan, _ := strconv.Atoi(c.Query("bulan"))
	tahun, _ := strconv.Atoi(c.Query("tahun"))

	if bulan < 1 || bulan > 12 || tahun < 2020 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Parameter bulan (1-12) dan tahun wajib diisi",
		})
	}

	// Ambil semua user (exclude admin)
	allUsers, err := h.userService.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil data pegawai",
		})
	}

	// Filter hanya pegawai non-admin
	var users []interface{}
	for _, u := range allUsers {
		if strings.ToLower(u.Role) != "admin" {
			users = append(users, u)
		}
	}

	recaps, err := h.absensiService.GetAllMonthlyRecap(bulan, tahun, allUsers)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil rekap absensi",
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data":   recaps,
	})
}
