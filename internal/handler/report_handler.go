package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"

	embedImages "laporanharianapi/images"
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"
)

var (
	cachedLogos     map[string]*bytes.Buffer
	cachedLogosOnce sync.Once
)

func loadLogosToCache() {
	cachedLogos = make(map[string]*bytes.Buffer)
	allImages := []string{
		"images/logo.png",
		"images/logo_berakhlak.png",
		"images/Logo_EVP.png",
		"images/splash_illustration.png",
	}

	for _, f := range allImages {
		name := filepath.Base(f)
		file, errFS := embedImages.FS.Open(name)
		if errFS != nil {
			continue
		}
		img, errOpen := imaging.Decode(file, imaging.AutoOrientation(true))
		file.Close()
		if errOpen == nil {
			whiteBg := image.NewRGBA(img.Bounds())
			draw.Draw(whiteBg, whiteBg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
			draw.Draw(whiteBg, whiteBg.Bounds(), img, img.Bounds().Min, draw.Over)

			var buf bytes.Buffer
			if jpeg.Encode(&buf, whiteBg, &jpeg.Options{Quality: 90}) == nil {
				cachedLogos[name] = &buf
			}
		}
	}
}

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
		"jenis_tugas":       "Tugas Individu",
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
		status := fiber.StatusInternalServerError
		if err.Error() == "laporan tidak ditemukan" {
			status = fiber.StatusNotFound
		} else if err.Error() == "Anda tidak memiliki hak untuk mengevaluasi laporan pegawai ini" || err.Error() == "akses ditolak" {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 4. Sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Evaluasi laporan berhasil disimpan",
	})
}

// ExportReportRecapExcelHandler mengambil rekapitulasi agregasi laporan dan mengekspornya ke file Excel.
func (h *ReportHandler) ExportReportRecapExcelHandler(c fiber.Ctx) error {
	// 1. Ambil requester
	requesterRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Role tidak ditemukan"})
	}
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User tidak terautentikasi"})
	}
	requesterID := uint(requesterIDFloat)

	var targetUsers []domain.User
	roleBase := strings.ToLower(requesterRole)
	switch roleBase {
	case "staf", "kasi":
		// Only self
		user, err := h.userService.GetUserByID(requesterID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
		}
		targetUsers = []domain.User{*user}
	case "sekertaris", "sekretaris":
		// Sendiri dan staf
		users, err := h.userService.GetUsersByRoles([]string{"staf", "Staf", "sekertaris", "Sekertaris"})
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
		}
		targetUsers = users
	default:
		// Lurah (All users)
		users, err := h.userService.GetAllUsers()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
		}
		targetUsers = users
	}

	// 2. Parse query parameters
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

	// 3. Setup excelize
	f := excelize.NewFile()
	defer f.Close()
	sheetName := "Rekap Laporan"
	f.SetSheetName("Sheet1", sheetName)

	headers := []string{"No", "NIP", "Nama Pegawai", "Jabatan", "Total Laporan", "Sudah Direview", "Menunggu Review", "Total Jam Kerja"}
	for i, head := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, head)
	}

	// Style header
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
	})
	f.SetRowStyle(sheetName, 1, 1, style)

	row := 2
	for index, user := range targetUsers {
		rekap, err := h.reportService.GetReportRecap(user.ID, startDate, endDate)
		if err != nil {
			continue // skip error
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), index+1)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), user.NIP)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user.Nama)
		if user.Jabatan != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), user.Jabatan.NamaJabatan)
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), "-")
		}

		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), rekap.TotalLaporan)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), rekap.TotalSudahDireview)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), rekap.TotalMenunggu)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), rekap.TotalJamKerja)
		row++
	}

	// 4. Send to client as download
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal generate excel"})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=rekap_laporan_%s_to_%s.xlsx", startDate.Format("20060102"), endDate.Format("20060102")))
	return c.Type("xlsx").Send(buffer.Bytes())
}

// ExportReportAttachmentsHandler mendownload semua lampiran laporan dalam bentuk ZIP
func (h *ReportHandler) ExportReportAttachmentsHandler(c fiber.Ctx) error {
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User tidak terautentikasi"})
	}
	requesterID := uint(requesterIDFloat)
	requesterRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Role tidak ditemukan"})
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	userID := 0
	if u := c.Query("user_id"); u != "" {
		userID, _ = strconv.Atoi(u)
	}

	roleBase := strings.ToLower(requesterRole)
	switch roleBase {
	case "staf", "kasi", "pegawai":
		userID = int(requesterID)
	case "sekertaris", "sekretaris":
		// Validasi apakah target user bawahan atau staff jika userID beda dengan requester
		if userID != 0 && userID != int(requesterID) {
			targetUser, err := h.userService.GetUserByID(uint(userID))
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "User yang dipilih tidak ditemukan"})
			}
			targetRole := strings.ToLower(targetUser.Role)
			if targetRole != "staf" {
				return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": "Akses ditolak: Hanya dapat mengunduh laporan staf atau diri sendiri"})
			}
		}
	} // Lurah tidak butuh validasi target user

	if startDateStr == "" || endDateStr == "" {
		// default to beginning and end of current month if not provided
		now := time.Now()
		startDateStr = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local).Format("2006-01-02")
		endDateStr = time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, time.Local).Format("2006-01-02")
	}

	filter := repository.ReportFilter{
		StartDate: startDateStr,
		EndDate:   endDateStr,
		UserID:    int(userID),
		JabatanID: 0,
		Limit:     100000, // get all without pagination
		Offset:    0,
	}

	reports, _, err := h.reportService.GetAllReports(filter, requesterRole, requesterID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data laporan: " + err.Error()})
	}

	// Buat buffer untuk zip
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	for _, laporan := range reports {
		dateFolder := laporan.WaktuPelaporan.Format("20060102")
		employeeName := "Unknown"
		if laporan.User != nil {
			employeeName = laporan.User.Nama
		}

		folderName := fmt.Sprintf("%s_%s", dateFolder, employeeName)

		// Fungsi helper untuk include file ke zip
		addFileToZip := func(fileUrl string, folder string) {
			if fileUrl == "" {
				return
			}

			// Di database, path sudah disimpan relatif (misal: uploads/reports/images/uuid.jpg)
			// Kita tinggal gunakan path tersebut.
			localPath := filepath.Join(".", fileUrl)

			if _, err := os.Stat(localPath); err == nil {
				f, err := os.Open(localPath)
				if err != nil {
					return
				}
				defer f.Close()

				// Nama file di dalam zip
				filename := filepath.Base(fileUrl)
				zipEntryPath := filepath.Join(folder, filename)
				w, err := zipWriter.Create(zipEntryPath)
				if err != nil {
					return
				}
				io.Copy(w, f)
			}
		}

		if laporan.FotoURL != nil {
			addFileToZip(*laporan.FotoURL, folderName)
		}
		if laporan.DokumenURL != nil {
			addFileToZip(*laporan.DokumenURL, folderName)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal membuat zip"})
	}

	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=lampiran_laporan_%s_to_%s.zip", startDateStr, endDateStr))
	return c.Type("zip").Send(buf.Bytes())
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

	// 2. Ambil info requester untuk RBAC
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	requesterID := uint(requesterIDFloat)
	requesterRole, _ := c.Locals("role").(string)

	// 3. Parse Request Body
	// Mendukung JSON body
	type UpdateRequest struct {
		JudulKegiatan  string `json:"judul_kegiatan"`
		DeskripsiHasil string `json:"deskripsi_hasil"`
	}
	var req UpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		// Jika gagal parse JSON, coba ambil dari form value (fallback)
		req.JudulKegiatan = c.FormValue("judul_kegiatan")
		req.DeskripsiHasil = c.FormValue("deskripsi_hasil")
	}

	// 4. Panggil Service
	err = h.reportService.UpdateReport(uint(id), req.JudulKegiatan, req.DeskripsiHasil, requesterID, requesterRole)
	if err != nil {
		status := fiber.StatusInternalServerError
		if err.Error() == "laporan tidak ditemukan" {
			status = fiber.StatusNotFound
		} else if strings.Contains(err.Error(), "akses ditolak") {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
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
		status := fiber.StatusInternalServerError
		if err.Error() == "laporan tidak ditemukan" {
			status = fiber.StatusNotFound
		} else if strings.Contains(err.Error(), "akses ditolak") {
			status = fiber.StatusForbidden
		}
		return c.Status(status).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 4. Success
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Laporan berhasil dihapus",
	})
}

// ExportReportPDFHandler mengekspor laporan harian menjadi file PDF berformat F4.
func (h *ReportHandler) ExportReportPDFHandler(c fiber.Ctx) error {
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User tidak terautentikasi"})
	}
	requesterID := uint(requesterIDFloat)

	requesterRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Role tidak ditemukan"})
	}
	roleBase := strings.ToLower(requesterRole)

	// 1. Tentukan target users berdasarkan RBAC
	targetUserIDStr := c.Query("user_id") // Lurah/Sekertaris bisa filter per user
	targetUserID, _ := strconv.Atoi(targetUserIDStr)

	var targetUsers []domain.User

	switch roleBase {
	case "staf", "kasi":
		user, err := h.userService.GetUserByID(requesterID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
		}
		targetUsers = []domain.User{*user}

	case "sekertaris", "sekretaris":
		if targetUserID > 0 {
			user, err := h.userService.GetUserByID(uint(targetUserID))
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "User tidak ditemukan"})
			}
			targetUsers = []domain.User{*user}
		} else {
			users, err := h.userService.GetUsersByRoles([]string{"staf", "Staf", "kasi", "Kasi", "sekertaris", "Sekertaris"})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
			}
			targetUsers = users
		}

	default: // Lurah
		if targetUserID > 0 {
			user, err := h.userService.GetUserByID(uint(targetUserID))
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "User tidak ditemukan"})
			}
			targetUsers = []domain.User{*user}
		} else {
			users, err := h.userService.GetAllUsers()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
			}
			targetUsers = users
		}
	}

	// 2. Parse tanggal
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

	// 3. Fetch laporan in a single bulk query (Solusi N+1)
	filter := repository.ReportFilter{
		Limit:     1000000, 
		Offset:    0,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		SortOrder: "asc", 
	}
	if targetUserID > 0 {
		filter.UserID = targetUserID
	}

	// Menggunakan requesterRole asli dari token untuk RBAC
	allReports, _, err := h.reportService.GetAllReports(filter, requesterRole, requesterID)
	if err != nil {
		allReports = []domain.Laporan{} // Fallback to empty slice
	}

	// Group reports by user ID
	reportsByUser := make(map[uint][]domain.Laporan)
	for _, rp := range allReports {
		if rp.UserID != nil {
			reportsByUser[*rp.UserID] = append(reportsByUser[*rp.UserID], rp)
		}
	}

	// 4. Parallel Image Processing using Goroutine Pool
	exePath, exeErr := os.Executable()
	var baseDir string
	if exeErr == nil {
		baseDir = filepath.Dir(exePath)
	} else {
		baseDir, _ = os.Getwd()
	}
	if strings.Contains(baseDir, "go-build") || strings.Contains(baseDir, "/tmp") {
		baseDir, _ = os.Getwd()
	}

	colW := []float64{8, 22, 25, 30, 50, 50.9}

	type ProcessedImage struct {
		ReportID uint
		Img      image.Image
		Width    float64
		Height   float64
		Path     string
	}

	processedImages := make(map[uint]ProcessedImage)
	var mu sync.Mutex
	var wg sync.WaitGroup

	type Job struct {
		Report domain.Laporan
	}
	jobs := make(chan Job, len(allReports))

	numWorkers := 4
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if job.Report.FotoURL == nil || *job.Report.FotoURL == "" {
					continue
				}

				p := strings.ReplaceAll(*job.Report.FotoURL, "\\", "/")
				p = strings.TrimPrefix(p, "/")
				localPath := filepath.Join(baseDir, filepath.FromSlash(p))
				if _, statErr := os.Stat(localPath); statErr != nil {
					cwd, _ := os.Getwd()
					localPath = filepath.Join(cwd, filepath.FromSlash(p))
				}

				if _, statErr := os.Stat(localPath); statErr == nil {
					img, openErr := imaging.Open(localPath, imaging.AutoOrientation(true))
					if openErr == nil {
						maxW := colW[5] - 2
						maxH := 70.0

						imgW := float64(img.Bounds().Dx())
						imgH := float64(img.Bounds().Dy())
						imgRatio := imgH / imgW

						scaledW := maxW
						scaledH := scaledW * imgRatio
						if scaledH > maxH {
							scaledH = maxH
							scaledW = scaledH / imgRatio
						}

						targetPxW := int(scaledW * 3.78)
						targetPxH := int(scaledH * 3.78)
						if targetPxW > img.Bounds().Dx() {
							targetPxW = img.Bounds().Dx()
							targetPxH = img.Bounds().Dy()
						}

						// Use Box filter instead of Lanczos for massive speedup
						resized := imaging.Resize(img, targetPxW, targetPxH, imaging.Box)

						mu.Lock()
						processedImages[job.Report.ID] = ProcessedImage{
							ReportID: job.Report.ID,
							Img:      resized,
							Width:    scaledW,
							Height:   scaledH,
							Path:     localPath,
						}
						mu.Unlock()
					}
				}
			}
		}()
	}

	for _, rp := range allReports {
		jobs <- Job{Report: rp}
	}
	close(jobs)
	wg.Wait() // Tunggu semua foto diproses

	// 5. Setup PDF (F4 = 215.9mm x 330.2mm)
	pdf := fpdf.New("P", "mm", "A4", "") 
	f4Size := fpdf.SizeType{Wd: 215.9, Ht: 330.2}

	marginL, marginR := 15.0, 15.0
	pdf.SetMargins(marginL, 20, marginR)
	pageW := 215.9 - marginL - marginR 

	// Init cached logos
	cachedLogosOnce.Do(loadLogosToCache)

	for name, buf := range cachedLogos {
		opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
		pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(buf.Bytes()))
	}

	type logoTarget struct {
		file   string
		height float64
	}

	logoFiles := []logoTarget{
		{"logo.png", 12.0},             
		{"logo_berakhlak.png", 8.0},    
		{"Logo_EVP.png", 8.0},          
	}

	splashName := "splash_illustration.png"
	logoGap := 2.0
	pdf.SetHeaderFuncMode(func() {
		// Watermark
		if infoSplash := pdf.GetImageInfo(splashName); infoSplash != nil {
			splashW := 90.0 
			splashH := splashW * float64(infoSplash.Height()) / float64(infoSplash.Width())
			
			splashX := (215.9 - splashW) / 2
			splashY := (330.2 - splashH) / 2

			pdf.SetAlpha(0.15, "Normal") 
			opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
			pdf.ImageOptions(splashName, splashX, splashY, splashW, splashH, false, opt, 0, "")
			pdf.SetAlpha(1.0, "Normal") 
		}

		// Logo Header
		logoX := marginL
		baseY := 3.0
		baseH := 12.0 

		for _, logo := range logoFiles {
			if info := pdf.GetImageInfo(logo.file); info != nil {
				opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
				offsetY := (baseH - logo.height) / 2
				currentY := baseY + offsetY

				pdf.ImageOptions(logo.file, logoX, currentY, 0, logo.height, false, opt, 0, "")
				logoW := logo.height * float64(info.Width()) / float64(info.Height())
				logoX += logoW + logoGap
			}
		}
	}, true)

	getIndonesianMonth := func(m time.Month) string {
		months := map[time.Month]string{
			time.January:   "Januari", time.February:  "Februari", time.March:     "Maret",
			time.April:     "April", time.May:       "Mei", time.June:      "Juni",
			time.July:      "Juli", time.August:    "Agustus", time.September: "September",
			time.October:   "Oktober", time.November:  "November", time.December:  "Desember",
		}
		return months[m]
	}

	formatDateIndo := func(t time.Time) string {
		return fmt.Sprintf("%02d %s %d", t.Day(), getIndonesianMonth(t.Month()), t.Year())
	}

	colHeaders := []string{"No", "Waktu\nPelaksanaan", "Jenis\nLaporan", "Judul\nLaporan", "Deskripsi", "Foto"}

	headerBgR, headerBgG, headerBgB := 255, 255, 255
	headerFgR, headerFgG, headerFgB := 0, 0, 0

	drawTableHeader := func() {
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(headerBgR, headerBgG, headerBgB)
		pdf.SetTextColor(headerFgR, headerFgG, headerFgB)
		pdf.SetDrawColor(200, 200, 200)

		headerH := 10.0
		startY := pdf.GetY()
		startX := marginL

		for i, w := range colW {
			pdf.SetXY(startX, startY)
			pdf.CellFormat(w, headerH, "", "1", 0, "C", false, 0, "")
			lines := strings.Split(colHeaders[i], "\n")
			lineH := 4.0
			totalTextH := float64(len(lines)) * lineH
			offsetY := (headerH - totalTextH) / 2

			pdf.SetXY(startX, startY+offsetY)
			pdf.MultiCell(w, lineH, colHeaders[i], "0", "C", false)
			startX += w
		}
		pdf.SetXY(marginL, startY+headerH)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFillColor(255, 255, 255)
	}

	calcTextRows := func(text string, colWidth float64, fontSize float64) int {
		if text == "" {
			return 1
		}
		charsPerLine := int((colWidth / fontSize) * 1.9)
		if charsPerLine < 1 {
			charsPerLine = 1
		}
		lineCount := 0
		words := strings.Fields(text)
		currentLen := 0
		for _, w := range words {
			if currentLen+len(w)+1 > charsPerLine {
				lineCount++
				currentLen = len(w)
			} else {
				currentLen += len(w) + 1
			}
		}
		lineCount++ 
		return lineCount
	}

	rowFillAlt := false 

	addReportRow := func(no int, laporan domain.Laporan) {
		lineH := 4.5 

		jenis := "Individu"
		if laporan.TipeLaporan {
			jenis = "Organisasi"
			if laporan.TugasOrganisasi != nil {
				jenis = "Organisasi\n" + laporan.TugasOrganisasi.JudulTugas
			}
		}

		judul := laporan.JudulKegiatan
		desc := laporan.DeskripsiHasil
		tanggal := fmt.Sprintf("%02d/%02d/%d\n%02d:%02d", laporan.WaktuPelaporan.Day(), int(laporan.WaktuPelaporan.Month()), laporan.WaktuPelaporan.Year(), laporan.WaktuPelaporan.Hour(), laporan.WaktuPelaporan.Minute())

		rowsJenis := calcTextRows(jenis, colW[2], 7)
		rowsJudul := calcTextRows(judul, colW[3], 7)
		rowsDesc := calcTextRows(desc, colW[4], 7)
		rowsTanggal := 2

		maxTextRows := rowsJenis
		if rowsJudul > maxTextRows { maxTextRows = rowsJudul }
		if rowsDesc > maxTextRows { maxTextRows = rowsDesc }
		if rowsTanggal > maxTextRows { maxTextRows = rowsTanggal }
		rowH := float64(maxTextRows)*lineH + 4

		photoH := 0.0
		photoW := 0.0
		photoPath := ""
		var photoImg image.Image 

		if pimg, ok := processedImages[laporan.ID]; ok {
			photoW = pimg.Width
			photoH = pimg.Height
			photoPath = pimg.Path
			photoImg = pimg.Img
		}

		if photoH+4 > rowH {
			rowH = photoH + 4
		}
		if rowH < 12 {
			rowH = 12
		}

		_, pageH := pdf.GetPageSize()
		bottomMargin := 20.0
		if pdf.GetY()+rowH > pageH-bottomMargin {
			pdf.AddPageFormat("P", f4Size)
			drawTableHeader()
		}

		startX := pdf.GetX()
		startY := pdf.GetY()

		pdf.SetFillColor(255, 255, 255)
		rowFillAlt = !rowFillAlt

		pdf.SetFont("Arial", "", 7.5)
		pdf.SetDrawColor(200, 200, 200)

		drawCell := func(x, y, w, h float64, txt string, align string) {
			pdf.SetXY(x, y)
			pdf.CellFormat(w, h, "", "1", 0, "", false, 0, "")
			if txt != "" {
				rows := calcTextRows(txt, w, 7)
				realTextH := float64(rows) * lineH
				offsetY := (h - realTextH) / 2
				if offsetY < 0 { offsetY = 0 }
				pdf.SetXY(x, y+offsetY)
				pdf.MultiCell(w, lineH, txt, "0", align, false)
			}
		}

		drawCell(startX, startY, colW[0], rowH, strconv.Itoa(no), "C")
		drawCell(startX+colW[0], startY, colW[1], rowH, tanggal, "C")
		drawCell(startX+colW[0]+colW[1], startY, colW[2], rowH, jenis, "C")
		drawCell(startX+colW[0]+colW[1]+colW[2], startY, colW[3], rowH, judul, "L")
		drawCell(startX+colW[0]+colW[1]+colW[2]+colW[3], startY, colW[4], rowH, desc, "L")

		fotoX := startX + colW[0] + colW[1] + colW[2] + colW[3] + colW[4]
		pdf.SetXY(fotoX, startY)
		pdf.CellFormat(colW[5], rowH, "", "1", 0, "C", false, 0, "")

		if photoPath != "" {
			imgX := fotoX + (colW[5]-photoW)/2
			imgY := startY + (rowH-photoH)/2

			var buf bytes.Buffer
			if jpeg.Encode(&buf, photoImg, &jpeg.Options{Quality: 70}) == nil {
				imgName := fmt.Sprintf("opt_%d", laporan.ID)
				opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
				pdf.RegisterImageOptionsReader(imgName, opt, &buf)
				pdf.ImageOptions(imgName, imgX, imgY, photoW, photoH, false, opt, 0, "")
			}
		} else if laporan.FotoURL != nil && *laporan.FotoURL != "" {
			pdf.SetXY(fotoX, startY+(rowH/2)-2)
			pdf.SetFont("Arial", "I", 6)
			pdf.CellFormat(colW[5], 4, "File tdk ditemukan", "", 0, "C", false, 0, "")
		} else {
			pdf.SetXY(fotoX, startY+(rowH/2)-2)
			pdf.SetFont("Arial", "I", 6)
			pdf.CellFormat(colW[5], 4, "- Tanpa Foto -", "", 0, "C", false, 0, "")
		}

		pdf.SetXY(startX, startY+rowH)
	}

	// 6. Generate PDF per user using bulk-fetched reports
	for _, user := range targetUsers {
		reports := reportsByUser[user.ID]
		if len(reports) == 0 {
			continue
		}

		pdf.AddPageFormat("P", f4Size)

		// Kop halaman
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(pageW, 8, "Laporan Harian Pegawai", "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(pageW, 6, user.Nama, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 9)
		if user.Jabatan != nil {
			pdf.CellFormat(pageW, 6, user.Jabatan.NamaJabatan, "", 1, "C", false, 0, "")
		}
		pdf.CellFormat(pageW, 6,
			fmt.Sprintf("Periode: %s s/d %s", formatDateIndo(startDate), formatDateIndo(endDate)),
			"", 1, "C", false, 0, "")
		pdf.Ln(2)

		drawTableHeader()

		rowFillAlt = false
		for i, laporan := range reports {
			addReportRow(i+1, laporan)
		}

		pdf.Ln(10)

		_, pageH := pdf.GetPageSize()
		bottomMargin := 20.0
		if pdf.GetY()+40 > pageH-bottomMargin {
			pdf.AddPageFormat("P", f4Size)
		}

		sigY := pdf.GetY()
		pdf.SetFont("Arial", "", 10)

		leftX := marginL
		pdf.SetXY(leftX, sigY)
		pdf.CellFormat(60, 5, "Mengetahui,", "", 2, "C", false, 0, "")
		pdf.CellFormat(60, 5, "Pejabat Penilai Kinerja,", "", 2, "C", false, 0, "")

		pdf.Ln(20)

		pdf.SetX(leftX)
		supervisorName := ""
		supervisorNIP := ""
		if strings.ToLower(user.Role) == "lurah" {
			supervisorName = "Rena Sudrajat, S.Sos., M.Si"
			supervisorNIP = "197208241992031003"
		} else if user.Supervisor != nil {
			supervisorName = user.Supervisor.Nama
			supervisorNIP = user.Supervisor.NIP
		}

		if supervisorName != "" {
			pdf.SetFont("Arial", "BU", 10)
			pdf.CellFormat(60, 5, supervisorName, "", 2, "C", false, 0, "")
			pdf.SetFont("Arial", "", 10)
			pdf.CellFormat(60, 5, "NIP. "+supervisorNIP, "", 1, "C", false, 0, "")
		} else {
			pdf.SetFont("Arial", "", 10)
			pdf.CellFormat(60, 5, "( ........................................ )", "", 2, "C", false, 0, "")
			pdf.CellFormat(60, 5, "NIP. ........................................ ", "", 1, "C", false, 0, "")
		}

		pdf.SetFont("Arial", "", 10)
		rightX := pageW - 60 + marginL 
		pdf.SetXY(rightX, sigY)
		pdf.CellFormat(60, 5, "", "", 2, "C", false, 0, "") 
		pdf.CellFormat(60, 5, "Yang Dinilai,", "", 2, "C", false, 0, "")

		pdf.Ln(20)

		pdf.SetX(rightX)
		pdf.SetFont("Arial", "BU", 10)
		pdf.CellFormat(60, 5, user.Nama, "", 2, "C", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(60, 5, "NIP. "+user.NIP, "", 1, "C", false, 0, "")
	}

	if pdf.PageCount() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Tidak ada laporan dalam periode tersebut"})
	}

	// 7. Tulis ke file/buffer
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal generate PDF: " + err.Error()})
	}

	filename := fmt.Sprintf("laporan_harian_%s_sd_%s.pdf", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/pdf")
	return c.Send(buf.Bytes())
}
