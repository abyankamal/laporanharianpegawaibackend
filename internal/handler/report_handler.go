package handler

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/repository"
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

// GetAll menangani request untuk mengambil semua laporan dengan filter (dengan RBAC).
// Query params: start_date, end_date, user_id, jabatan_id, limit, page
func (h *ReportHandler) GetAll(c fiber.Ctx) error {
	// 1. Ambil requester dari JWT Token
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	requesterID := uint(requesterIDFloat)

	requesterRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// 2. Parse query parameters
	startDate := c.Query("start_date") // Format: YYYY-MM-DD
	endDate := c.Query("end_date")     // Format: YYYY-MM-DD

	userID, _ := strconv.Atoi(c.Query("user_id"))
	jabatanID, _ := strconv.Atoi(c.Query("jabatan_id"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))

	// 3. Set default value
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 4. Validasi format tanggal jika diberikan
	if startDate != "" {
		if _, err := time.Parse("2006-01-02", startDate); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format start_date tidak valid (gunakan: YYYY-MM-DD)",
			})
		}
	}
	if endDate != "" {
		if _, err := time.Parse("2006-01-02", endDate); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format end_date tidak valid (gunakan: YYYY-MM-DD)",
			})
		}
	}

	// 5. Buat filter
	filter := repository.ReportFilter{
		StartDate: startDate,
		EndDate:   endDate,
		UserID:    userID,
		JabatanID: jabatanID,
		Limit:     limit,
		Offset:    offset,
	}

	// 6. Panggil service (dengan RBAC)
	reports, total, err := h.reportService.GetAllReports(filter, requesterRole, requesterID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data laporan: " + err.Error(),
		})
	}

	// 6. Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 7. Return response sukses dengan metadata pagination
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data laporan berhasil diambil",
		"data":    reports,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
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
	lokasiLat := c.FormValue("lokasi_lat")           // opsional
	lokasiLong := c.FormValue("lokasi_long")         // opsional
	alamatLokasi := c.FormValue("alamat_lokasi")     // opsional
	tugasPokokIDStr := c.FormValue("tugas_pokok_id") // opsional (wajib jika tipe_laporan = pokok)

	var tugasPokokID *uint
	if tugasPokokIDStr != "" {
		if id, err := strconv.ParseUint(tugasPokokIDStr, 10, 32); err == nil {
			val := uint(id)
			tugasPokokID = &val
		}
	}

	// Parse waktu pelaporan
	waktuPelaporanStr := c.FormValue("waktu_pelaporan")
	waktuPelaporan, err := time.Parse("2006-01-02 15:04:05", waktuPelaporanStr)
	if err != nil {
		// Coba format lain
		waktuPelaporan, err = time.Parse("2006-01-02T15:04:05", waktuPelaporanStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format waktu_pelaporan tidak valid (gunakan: YYYY-MM-DD HH:mm:ss)",
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
		TugasPokokID:   tugasPokokID,
		JudulKegiatan:  judulKegiatan,
		DeskripsiHasil: deskripsiHasil,
		WaktuPelaporan: waktuPelaporan,
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
