package handler

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"
)


// ReportHandler menangani request laporan.
type ReportHandler struct {
	reportService service.ReportService
	userService   service.UserService
}

// NewReportHandler membuat instance baru ReportHandler.
func NewReportHandler(reportService service.ReportService, userService service.UserService) *ReportHandler {
	return &ReportHandler{reportService: reportService, userService: userService}
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

	// 3. Set default value & cap limit to prevent unbounded fetches
	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 4. Validasi format tanggal jika diberikan
	if startDate != "" {
		if _, err := time.ParseInLocation("2006-01-02", startDate, time.Local); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format start_date tidak valid (gunakan: YYYY-MM-DD)",
			})
		}
	}
	if endDate != "" {
		if _, err := time.ParseInLocation("2006-01-02", endDate, time.Local); err != nil {
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
		"success": true,
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
	isOfflineSyncStr := c.FormValue("is_offline_sync")         // penanda sinkronisasi dari mode offline
	isOfflineSync := isOfflineSyncStr == "true" || isOfflineSyncStr == "1"

	var tugasOrganisasiID *uint
	if tugasOrganisasiIDStr != "" {
		if id, err := strconv.ParseUint(tugasOrganisasiIDStr, 10, 32); err == nil {
			val := uint(id)
			tugasOrganisasiID = &val
		}
	}

	// Parse waktu pelaporan dengan berbagai format (ISO8601, MySQL, etc)
	waktuPelaporanStr := c.FormValue("waktu_pelaporan")
	var waktuPelaporan time.Time
	var parseErr error

	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		time.RFC3339,
		"2006-01-02",
	}

	for _, fmtStr := range formats {
		waktuPelaporan, parseErr = time.ParseInLocation(fmtStr, waktuPelaporanStr, time.Local)
		if parseErr == nil {
			break
		}
	}

	if parseErr != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": fmt.Sprintf("Format waktu_pelaporan tidak valid: %s. Gunakan format YYYY-MM-DD HH:mm:ss", waktuPelaporanStr),
		})
	}

	// 3. Validasi input sederhana
	if deskripsiHasil == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Deskripsi hasil wajib diisi",
		})
	}

	// 4. Ambil file foto (opsional)
	fileFoto, fotoErr := c.FormFile("foto")
	if fotoErr != nil {
		// Jika errornya bukan karena file tidak ada, return error
		if fotoErr.Error() != "there is no uploaded file associated with the given key" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Gagal membaca file foto: %v", fotoErr),
			})
		}
		// Jika tidak ada file, biarkan nil
		fileFoto = nil
	}

	// 5. Ambil file dokumen (opsional)
	fileDokumen, dokErr := c.FormFile("dokumen")
	if dokErr != nil {
		if dokErr.Error() != "there is no uploaded file associated with the given key" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Gagal membaca file dokumen: %v", dokErr),
			})
		}
		fileDokumen = nil
	}

	// 6. Susun input untuk service
	role, _ := c.Locals("role").(string)
	input := service.ReportInput{
		UserID:            userID,
		UserRole:          role,
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
		IsOfflineSync:     isOfflineSync,
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
		"success": true,
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
		return ErrorResponse(c, err)
	}

	// 4. Susun response
	responseMap := fiber.Map{
		"id":                laporan.ID,
		"status":            laporan.Status,
		"jenis_tugas":       "Tugas Individu",
		"judul_laporan":     laporan.JudulKegiatan,
		"waktu_pelaksanaan": laporan.WaktuPelaporan,
		"lokasi":            laporan.AlamatLokasi,
		"lokasi_lat":        laporan.LokasiLat,
		"lokasi_long":       laporan.LokasiLong,
		"deskripsi_hasil":   laporan.DeskripsiHasil,
		"komentar_atasan":   laporan.KomentarAtasan,
		"foto_url":          laporan.FotoURL,
		"dokumen_url":       laporan.DokumenURL,
		"owner_role":        "", // Akan diisi di bawah
	}

	if laporan.User != nil {
		responseMap["owner_role"] = laporan.User.Role
		
		userMap := fiber.Map{
			"nama": laporan.User.Nama,
			"nip":  laporan.User.NIP,
		}
		if laporan.User.Jabatan != nil {
			userMap["jabatan"] = fiber.Map{
				"nama_jabatan": laporan.User.Jabatan.NamaJabatan,
			}
		}
		responseMap["user"] = userMap

		if laporan.User.Supervisor != nil {
			responseMap["supervisor_nip"] = laporan.User.Supervisor.NIP
			responseMap["supervisor_nama"] = laporan.User.Supervisor.Nama
		}
	}

	if laporan.TipeLaporan {
		responseMap["jenis_tugas"] = "Tugas Organisasi"
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
		"success": true,
		"status":  "success",
		"message": "Detail laporan berhasil diambil",
		"data":    responseMap,
	})
}

// GetReportRecapHandler mengambil rekapitulasi agregasi laporan.
func (h *ReportHandler) GetReportRecapHandler(c fiber.Ctx) error {
	// 1. Ambil requester data dari JWT Token
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User tidak terautentikasi"})
	}
	requesterID := uint(requesterIDFloat)
	requesterRole, _ := c.Locals("role").(string)
	roleBase := strings.ToLower(requesterRole)

	roleFilter := c.Query("role")

	// 2. Tentukan target userID (dari query param atau default ke diri sendiri)
	targetUserIDStr := c.Query("user_id")
	targetUserID := requesterID // Default ke diri sendiri

	if roleFilter == "" && targetUserIDStr != "" {
		parsedID, _ := strconv.Atoi(targetUserIDStr)
		// Hanya Lurah/Sekertaris yang boleh ganti targetUserID
		if roleBase == "lurah" || roleBase == "sekertaris" || roleBase == "sekretaris" {
			targetUserID = uint(parsedID)
		}
	} else if roleFilter != "" {
		targetUserID = 0 // Jika ada role filter, kita fallback ke aggregated mode
	}

	// 3. Parse query parameters (tanggal)
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	var startDate, endDate time.Time
	var err error

	if startDateStr != "" && endDateStr != "" {
		startDate, err = time.ParseInLocation("2006-01-02", startDateStr, time.Local)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Format start_date tidak valid"})
		}
		endDate, err = time.ParseInLocation("2006-01-02", endDateStr, time.Local)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Format end_date tidak valid"})
		}
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)
	} else {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, -1)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)
	}

	// 4. Panggil service
	var rekap *repository.ReportRecapResponse
	
	if roleFilter != "" {
		filter := repository.ReportFilter{
			StartDate: startDateStr,
			EndDate:   endDateStr,
			UserRole:  roleFilter,
			UserID:    int(targetUserID), // Harus 0
		}
		rekap, err = h.reportService.GetReportRecapAggregated(filter, requesterRole, requesterID)
	} else {
		rekap, err = h.reportService.GetReportRecap(targetUserID, startDate, endDate)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil rekap laporan: " + err.Error()})
	}

	// 5. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
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
		return ErrorResponse(c, err)
	}

	// 4. Sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Evaluasi laporan berhasil disimpan",
	})
}

// Update menangani pembaruan data laporan (Judul & Detail).
func (h *ReportHandler) Update(c fiber.Ctx) error {
	// 1. Ambil ID dari URL parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID laporan tidak valid",
		})
	}

	// 2. Ambil info requester dari JWT
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	requesterID := uint(requesterIDFloat)
	requesterRole, _ := c.Locals("role").(string)

	// 3. Parse Request Body (Mendukung JSON dan Form-Data)
	type UpdateRequest struct {
		JudulKegiatan  string `json:"judul_kegiatan" form:"judul_kegiatan"`
		DeskripsiHasil string `json:"deskripsi_hasil" form:"deskripsi_hasil"`
	}
	var req UpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		req.JudulKegiatan = c.FormValue("judul_kegiatan")
		req.DeskripsiHasil = c.FormValue("deskripsi_hasil")
	} else if c.FormValue("judul_kegiatan") != "" || c.FormValue("deskripsi_hasil") != "" {
		req.JudulKegiatan = c.FormValue("judul_kegiatan")
		req.DeskripsiHasil = c.FormValue("deskripsi_hasil")
	}

	// Coba ambil file foto (opsional, hanya valid untuk status ditolak)
	fileFoto, _ := c.FormFile("foto")

	// 4. Panggil Service
	err = h.reportService.UpdateReport(uint(id), req.JudulKegiatan, req.DeskripsiHasil, fileFoto, requesterID, requesterRole)
	if err != nil {
		return ErrorResponse(c, err)
	}

	// 5. Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Laporan berhasil diperbarui",
	})
}

// Delete menangani penghapusan laporan (Hanya Lurah).
func (h *ReportHandler) Delete(c fiber.Ctx) error {
	// 1. Ambil ID dari URL parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID laporan tidak valid",
		})
	}

	// 2. Ambil info requester (Hanya role Lurah yang diizinkan)
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	requesterID := uint(requesterIDFloat)
	requesterRole, _ := c.Locals("role").(string)

	// 3. Panggil service Delete
	err = h.reportService.DeleteReport(uint(id), requesterID, requesterRole)
	if err != nil {
		return ErrorResponse(c, err)
	}

	// 4. Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Laporan berhasil dihapus",
	})
}
