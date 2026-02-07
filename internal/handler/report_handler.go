package handler

import (
	"time"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// ReportHandler menangani request laporan.
type ReportHandler struct {
	reportService service.ReportService
}

// NewReportHandler membuat instance baru ReportHandler.
func NewReportHandler(reportService service.ReportService) *ReportHandler {
	return &ReportHandler{reportService: reportService}
}

// Create menangani pembuatan laporan baru.
func (h *ReportHandler) Create(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token (via Locals dari middleware)
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := uint(userIDFloat)

	// 2. Parse Form Data
	tipeLaporanStr := c.FormValue("tipe_laporan") // "pokok" atau "tambahan"
	tipeLaporan := tipeLaporanStr == "pokok"      // true jika pokok

	judulKegiatan := c.FormValue("judul_kegiatan")
	deskripsiHasil := c.FormValue("deskripsi_hasil")
	lokasiLat := c.FormValue("lokasi_lat")
	lokasiLong := c.FormValue("lokasi_long")
	alamatLokasi := c.FormValue("alamat_lokasi")

	// Parse waktu
	waktuMulaiStr := c.FormValue("waktu_mulai")
	waktuSelesaiStr := c.FormValue("waktu_selesai")

	waktuMulai, err := time.Parse("2006-01-02 15:04:05", waktuMulaiStr)
	if err != nil {
		// Coba format lain
		waktuMulai, err = time.Parse("2006-01-02T15:04:05", waktuMulaiStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format waktu_mulai tidak valid (gunakan: 2006-01-02 15:04:05)",
			})
		}
	}

	waktuSelesai, err := time.Parse("2006-01-02 15:04:05", waktuSelesaiStr)
	if err != nil {
		// Coba format lain
		waktuSelesai, err = time.Parse("2006-01-02T15:04:05", waktuSelesaiStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format waktu_selesai tidak valid (gunakan: 2006-01-02 15:04:05)",
			})
		}
	}

	// 3. Validasi input sederhana
	if deskripsiHasil == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Deskripsi hasil wajib diisi",
		})
	}

	// 4. Ambil file bukti (optional)
	fileHeader, _ := c.FormFile("file_bukti")

	// 5. Susun input untuk service
	input := service.ReportInput{
		UserID:         userID,
		TipeLaporan:    tipeLaporan,
		JudulKegiatan:  judulKegiatan,
		DeskripsiHasil: deskripsiHasil,
		WaktuMulai:     waktuMulai,
		WaktuSelesai:   waktuSelesai,
		LokasiLat:      lokasiLat,
		LokasiLong:     lokasiLong,
		AlamatLokasi:   alamatLokasi,
		File:           fileHeader,
	}

	// 6. Panggil service
	laporan, err := h.reportService.CreateReport(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 7. Return response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Laporan berhasil dibuat",
		"data": fiber.Map{
			"id":          laporan.ID,
			"is_overtime": laporan.IsOvertime,
			"created_at":  laporan.CreatedAt,
		},
	})
}
