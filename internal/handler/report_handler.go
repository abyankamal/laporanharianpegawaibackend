package handler

import (
	"math"
	"path/filepath"
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
// Mendukung upload DUA file sekaligus: foto dan dokumen (keduanya opsional).
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
	lokasiLat := c.FormValue("lokasi_lat")                     // opsional
	lokasiLong := c.FormValue("lokasi_long")                   // opsional
	alamatLokasi := c.FormValue("alamat_lokasi")               // opsional
	tugasOrganisasiIDStr := c.FormValue("tugas_organisasi_id") // opsional (hanya jika link ke tugas organisasi)

	var tugasOrganisasiID *uint
	if tugasOrganisasiIDStr != "" {
		if id, err := strconv.ParseUint(tugasOrganisasiIDStr, 10, 32); err == nil {
			val := uint(id)
			tugasOrganisasiID = &val
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

	// 4. Ambil file foto (opsional)
	fileFoto, _ := c.FormFile("foto")

	// 5. Ambil file dokumen (opsional)
	fileDokumen, _ := c.FormFile("dokumen")

	// 6. Susun input untuk service
	input := service.ReportInput{
		UserID:            userID,
		TipeLaporan:       tipeLaporan,
		TugasOrganisasiID: tugasOrganisasiID,
		JudulKegiatan:     judulKegiatan,
		DeskripsiHasil:    deskripsiHasil,
		WaktuPelaporan:    waktuPelaporan,
		LokasiLat:         lokasiLat,
		LokasiLong:        lokasiLong,
		AlamatLokasi:      alamatLokasi,
		FileFoto:          fileFoto,
		FileDokumen:       fileDokumen,
	}

	// 7. Panggil service
	laporan, err := h.reportService.CreateReport(input)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 8. Return response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Laporan berhasil dibuat",
		"data": fiber.Map{
			"id":          laporan.ID,
			"is_overtime": laporan.IsOvertime,
			"foto_url":    laporan.FotoURL,
			"dokumen_url": laporan.DokumenURL,
			"created_at":  laporan.CreatedAt,
		},
	})
}

// GetOne menangani request untuk mengambil detail satu laporan.
func (h *ReportHandler) GetOne(c fiber.Ctx) error {
	// 1. Ambil ID dari URL parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID laporan tidak valid",
		})
	}

	// 2. Ambil requester dari JWT Token untuk RBAC
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

	// 3. Panggil service GetReportDetail (sekarang return 2 value, bukan 3)
	laporan, err := h.reportService.GetReportDetail(uint(id), requesterRole, requesterID)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "laporan tidak ditemukan" {
			status = fiber.StatusNotFound
		} else if err.Error() == "akses ditolak: hanya dapat melihat laporan staf" || err.Error() == "akses ditolak: hanya dapat melihat laporan milik sendiri" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 4. Susun response
	responseMap := fiber.Map{
		"id":                laporan.ID,
		"status":            laporan.Status,
		"jenis_tugas":       "Tugas Tambahan",
		"judul_laporan":     laporan.JudulKegiatan,
		"waktu_pelaksanaan": laporan.WaktuPelaporan,
		"lokasi":            laporan.AlamatLokasi,
		"deskripsi_hasil":   laporan.DeskripsiHasil,
		"komentar_atasan":   laporan.KomentarAtasan,
		"foto_url":          laporan.FotoURL,
		"dokumen_url":       laporan.DokumenURL,
		"owner_role":        "", // Akan diisi di bawah
	}

	if laporan.User != nil {
		responseMap["owner_role"] = laporan.User.Role
	}

	if laporan.TipeLaporan {
		responseMap["jenis_tugas"] = "Tugas Pokok"
		if laporan.TugasOrganisasi != nil {
			responseMap["jenis_tugas"] = laporan.TugasOrganisasi.JudulTugas
		}
	}

	// Backward compatibility: tampilkan lampiran_url dari foto jika ada
	if laporan.FotoURL != nil {
		responseMap["lampiran_url"] = *laporan.FotoURL
		responseMap["lampiran_nama"] = filepath.Base(*laporan.FotoURL)
		responseMap["is_image"] = true
	} else if laporan.DokumenURL != nil {
		responseMap["lampiran_url"] = *laporan.DokumenURL
		responseMap["lampiran_nama"] = filepath.Base(*laporan.DokumenURL)
		responseMap["is_image"] = false
	} else {
		responseMap["lampiran_url"] = nil
		responseMap["lampiran_nama"] = nil
		responseMap["is_image"] = false
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Detail laporan berhasil diambil",
		"data":    responseMap,
	})
}

// GetReportRecapHandler mengambil rekapitulasi agregasi laporan.
func (h *ReportHandler) GetReportRecapHandler(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := uint(userIDFloat)

	// 2. Parse query parameters
	period := c.Query("period", "bulanan") // harian, mingguan, bulanan
	dateStr := c.Query("date")

	var targetDate time.Time
	var err error
	if dateStr != "" {
		targetDate, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format date tidak valid (gunakan: YYYY-MM-DD)",
			})
		}
	} else {
		targetDate = time.Now()
	}

	// 3. Panggil service
	rekap, err := h.reportService.GetReportRecap(userID, period, targetDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil rekap laporan: " + err.Error(),
		})
	}

	// 4. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Rekap laporan berhasil diambil",
		"data":    rekap,
	})
}

// EvaluateReportHandler memproses aksi atasan menyetujui/menolak laporan.
func (h *ReportHandler) EvaluateReportHandler(c fiber.Ctx) error {
	// 1. Ambil penilai_id dari JWT Token
	assessorIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	assessorID := uint(assessorIDFloat)

	assessorRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// 2. Parse request body
	var req service.EvaluateReportRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	// 3. Panggil service logic
	err := h.reportService.EvaluateReport(assessorID, assessorRole, req)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "laporan tidak ditemukan" {
			status = fiber.StatusNotFound
		} else if err.Error() == "Anda tidak memiliki hak untuk mengevaluasi laporan pegawai ini" || err.Error() == "akses ditolak" {
			status = fiber.StatusForbidden
		} else if err.Error() == "status evaluasi tidak valid (harus 'Disetujui' atau 'Ditolak')" {
			status = fiber.StatusBadRequest
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 4. Sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Evaluasi laporan berhasil disimpan",
	})
}
