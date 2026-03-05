package handler

import (
	"archive/zip"
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v3"
	"github.com/xuri/excelize/v2"

	"laporanharianapi/internal/domain"
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
		// Set end of day for endDate
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)
	} else {
		// Default to current month if not provided
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, -1)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)
	}

	// 3. Panggil service
	rekap, err := h.reportService.GetReportRecap(userID, startDate, endDate)
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

// ExportReportPDFHandler mengekspor laporan harian menjadi file PDF berformat F4.
// - Staf/Kasi   : hanya laporan diri sendiri
// - Lurah/Sekertaris : gabungan semua pegawai yang relevan, tapi bisa difilter per user_id
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
		// Hanya diri sendiri, tidak bisa memilih user lain
		user, err := h.userService.GetUserByID(requesterID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
		}
		targetUsers = []domain.User{*user}

	case "sekertaris", "sekretaris":
		if targetUserID > 0 {
			// Filter ke 1 user tertentu (diri sendiri atau staf)
			user, err := h.userService.GetUserByID(uint(targetUserID))
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "User tidak ditemukan"})
			}
			targetUsers = []domain.User{*user}
		} else {
			// Gabungan semua staf dan sekertaris
			users, err := h.userService.GetUsersByRoles([]string{"staf", "Staf", "kasi", "Kasi", "sekertaris", "Sekertaris"})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
			}
			targetUsers = users
		}

	default: // Lurah
		if targetUserID > 0 {
			// Filter ke 1 user tertentu
			user, err := h.userService.GetUserByID(uint(targetUserID))
			if err != nil {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "User tidak ditemukan"})
			}
			targetUsers = []domain.User{*user}
		} else {
			// Semua user
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

	// 3. Setup PDF (F4 = 215.9mm x 330.2mm)
	pdf := fpdf.New("P", "mm", "A4", "") // akan di-override ke F4
	pdf.AddPageFormat("P", fpdf.SizeType{Wd: 215.9, Ht: 330.2})

	// Hapus halaman A4 default yang ditambahkan New()
	// Kita manggil AddPage di loop user, jadi kita tidak perlu halaman default

	// Lebar halaman efektif setelah margin (L:15, R:15)
	marginL, marginR := 15.0, 15.0
	pdf.SetMargins(marginL, 20, marginR)
	pageW := 215.9 - marginL - marginR // 185.9mm

	// Lebar kolom: No | Tanggal | Jenis | Judul | Deskripsi | Foto
	colW := []float64{8, 22, 28, 35, 47, 45.9}
	colHeaders := []string{"No", "Tanggal", "Jenis\nLaporan", "Judul\nLaporan", "Deskripsi", "Foto"}

	// Warna header tabel
	headerBgR, headerBgG, headerBgB := 31, 73, 125 // Biru tua
	headerFgR, headerFgG, headerFgB := 255, 255, 255

	// Fungsi draw header tabel
	drawTableHeader := func() {
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetFillColor(headerBgR, headerBgG, headerBgB)
		pdf.SetTextColor(headerFgR, headerFgG, headerFgB)
		pdf.SetDrawColor(200, 200, 200)
		for i, w := range colW {
			pdf.CellFormat(w, 10, colHeaders[i], "1", 0, "C", true, 0, "")
		}
		pdf.Ln(-1)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFillColor(255, 255, 255)
	}

	// Fungsi helper untuk menghitung tinggi teks multi-baris (agar bisa menentukan tinggi baris)
	calcTextRows := func(text string, colWidth float64, fontSize float64) int {
		if text == "" {
			return 1
		}
		// estimasi jumlah karakter per baris (kasar: lebar/fontSize * 1.8)
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
		lineCount++ // Baris terakhir
		return lineCount
	}

	rowFillAlt := false // Warna baris bergantian

	// Fungsi helper menambah satu baris laporan ke PDF
	addReportRow := func(no int, laporan domain.Laporan) {
		lineH := 4.5 // tinggi per baris teks

		// Hitung tinggi minimum dari kolom teks
		jenis := "Tambahan"
		if laporan.TipeLaporan {
			jenis = "Pokok"
			if laporan.TugasOrganisasi != nil {
				jenis = "Pokok\n" + laporan.TugasOrganisasi.JudulTugas
			}
		}

		judul := laporan.JudulKegiatan
		desc := laporan.DeskripsiHasil
		tanggal := laporan.WaktuPelaporan.Format("02/01/2006\n15:04")

		// Hitung jumlah baris per kolom
		rowsJenis := calcTextRows(jenis, colW[2], 7)
		rowsJudul := calcTextRows(judul, colW[3], 7)
		rowsDesc := calcTextRows(desc, colW[4], 7)
		rowsTanggal := 2

		// Tinggi minimum dari konten teks
		maxTextRows := rowsJenis
		if rowsJudul > maxTextRows {
			maxTextRows = rowsJudul
		}
		if rowsDesc > maxTextRows {
			maxTextRows = rowsDesc
		}
		if rowsTanggal > maxTextRows {
			maxTextRows = rowsTanggal
		}
		rowH := float64(maxTextRows)*lineH + 4

		// Cek apakah ada foto dan hitung dimensinya
		photoH := 0.0
		photoW := 0.0
		photoPath := ""
		photoType := ""

		if laporan.FotoURL != nil && *laporan.FotoURL != "" {
			localPath := filepath.Join(".", *laporan.FotoURL)
			if _, statErr := os.Stat(localPath); statErr == nil {
				// Cek dimensi gambar
				f, openErr := os.Open(localPath)
				if openErr == nil {
					imgCfg, _, decErr := image.DecodeConfig(f)
					f.Close()
					if decErr == nil {
						// Max lebar foto di tabel = colW[5] - 2mm padding
						maxW := colW[5] - 2
						maxH := 50.0 // Max tinggi foto dalam sel

						imgRatio := float64(imgCfg.Height) / float64(imgCfg.Width)
						scaledW := maxW
						scaledH := scaledW * imgRatio

						// Perperkecil jika tinggi melebihi maxH
						if scaledH > maxH {
							scaledH = maxH
							scaledW = scaledH / imgRatio
						}

						photoW = scaledW
						photoH = scaledH
						photoPath = localPath
						ext := strings.ToLower(filepath.Ext(localPath))
						if ext == ".jpg" || ext == ".jpeg" {
							photoType = "JPG"
						} else if ext == ".png" {
							photoType = "PNG"
						}
					}
				}
			}
		}

		// Tinggi baris final: maks antara teks dan foto
		if photoH+4 > rowH {
			rowH = photoH + 4
		}
		if rowH < 12 {
			rowH = 12
		}

		// Cek apakah cukup ruang di halaman
		_, pageH := pdf.GetPageSize()
		bottomMargin := 20.0
		if pdf.GetY()+rowH > pageH-bottomMargin {
			pdf.AddPage()
			drawTableHeader()
		}

		// Posisi awal baris
		startX := pdf.GetX()
		startY := pdf.GetY()

		// Warna baris bergantian
		if rowFillAlt {
			pdf.SetFillColor(240, 245, 255)
		} else {
			pdf.SetFillColor(255, 255, 255)
		}
		rowFillAlt = !rowFillAlt

		pdf.SetFont("Helvetica", "", 7.5)
		pdf.SetDrawColor(200, 200, 200)

		// Sel: No
		pdf.SetXY(startX, startY)
		pdf.CellFormat(colW[0], rowH, strconv.Itoa(no), "1", 0, "C", true, 0, "")

		// Sel: Tanggal
		pdf.SetXY(startX+colW[0], startY)
		pdf.MultiCell(colW[1], lineH, tanggal, "1", "C", true)
		pdf.SetXY(startX+colW[0]+colW[1], startY)

		// Sel: Jenis Laporan
		pdf.SetXY(startX+colW[0]+colW[1], startY)
		pdf.MultiCell(colW[2], lineH, jenis, "1", "C", true)

		// Sel: Judul
		pdf.SetXY(startX+colW[0]+colW[1]+colW[2], startY)
		pdf.MultiCell(colW[3], lineH, judul, "1", "L", true)

		// Sel: Deskripsi
		pdf.SetXY(startX+colW[0]+colW[1]+colW[2]+colW[3], startY)
		pdf.MultiCell(colW[4], lineH, desc, "1", "L", true)

		// Sel: Foto (kotak kosong dulu, lalu gambar di-overlay)
		fotoX := startX + colW[0] + colW[1] + colW[2] + colW[3] + colW[4]
		pdf.SetXY(fotoX, startY)
		pdf.CellFormat(colW[5], rowH, "", "1", 0, "C", true, 0, "")

		// Overlay gambar di dalam sel foto
		if photoPath != "" && photoType != "" {
			// Hitung posisi agar gambar terpusat di dalam sel
			imgX := fotoX + (colW[5]-photoW)/2
			imgY := startY + (rowH-photoH)/2
			opt := fpdf.ImageOptions{ImageType: photoType, ReadDpi: false}
			pdf.ImageOptions(photoPath, imgX, imgY, photoW, photoH, false, opt, 0, "")
		} else if laporan.FotoURL != nil && *laporan.FotoURL != "" {
			// File tidak ditemukan, tulis keterangan teks
			pdf.SetXY(fotoX, startY+(rowH/2)-2)
			pdf.SetFont("Helvetica", "I", 6)
			pdf.CellFormat(colW[5], 4, "File tdk ditemukan", "", 0, "C", false, 0, "")
		} else {
			// Tidak ada foto
			pdf.SetXY(fotoX, startY+(rowH/2)-2)
			pdf.SetFont("Helvetica", "I", 6)
			pdf.CellFormat(colW[5], 4, "- Tanpa Foto -", "", 0, "C", false, 0, "")
		}

		// Reset posisi ke bawah baris
		pdf.SetXY(startX, startY+rowH)
	}

	// 4. Generate PDF per user
	for uIdx, user := range targetUsers {
		filter := repository.ReportFilter{
			UserID:    int(user.ID),
			Limit:     100000,
			Offset:    0,
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   endDate.Format("2006-01-02"),
		}

		reports, _, err := h.reportService.GetAllReports(filter, "lurah", requesterID) // role lurah agar dapat semua
		if err != nil || len(reports) == 0 {
			continue
		}

		pdf.AddPage()

		// --- Kop halaman ---
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(pageW, 8, "Laporan Harian Pegawai", "", 1, "C", false, 0, "")
		pdf.SetFont("Helvetica", "", 9)

		nameStr := user.Nama
		if user.Jabatan != nil {
			nameStr += " — " + user.Jabatan.NamaJabatan
		}
		pdf.CellFormat(pageW, 6, nameStr, "", 1, "C", false, 0, "")
		pdf.CellFormat(pageW, 6,
			fmt.Sprintf("Periode: %s s/d %s", startDate.Format("02 Jan 2006"), endDate.Format("02 Jan 2006")),
			"", 1, "C", false, 0, "")
		pdf.Ln(2)

		drawTableHeader()

		rowFillAlt = false
		rowCounter := 1
		for _, laporan := range reports {
			addReportRow(rowCounter, laporan)
			rowCounter++
		}

		// Halaman baru jika masih ada user lagi
		_ = uIdx
	}

	// Cek jika tidak ada data
	if pdf.PageCount() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Tidak ada laporan dalam periode tersebut"})
	}

	// 5. Tulis ke buffer dan kirim
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal generate PDF: " + err.Error()})
	}

	filename := fmt.Sprintf("laporan_harian_%s_sd_%s.pdf", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/pdf")
	return c.Send(buf.Bytes())
}
