package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// ReportInput adalah struct untuk input pembuatan laporan.
type ReportInput struct {
	UserID         uint
	TipeLaporan    bool  // true = pokok, false = tambahan
	TugasPokokID   *uint // ID tugas pokok (wajib jika TipeLaporan = true)
	JudulKegiatan  string
	DeskripsiHasil string
	WaktuPelaporan time.Time
	LokasiLat      string                // opsional, bisa kosong
	LokasiLong     string                // opsional, bisa kosong
	AlamatLokasi   string                // opsional, bisa kosong
	FileFoto       *multipart.FileHeader // File foto lampiran (opsional)
	FileDokumen    *multipart.FileHeader // File dokumen lampiran (opsional)
}

// EvaluateReportRequest adalah struct untuk input evaluasi laporan.
type EvaluateReportRequest struct {
	ReportID uint   `json:"report_id" validate:"required"`
	Status   string `json:"status" validate:"required"` // 'Disetujui' atau 'Ditolak'
	Komentar string `json:"komentar"`
}

// ReportService adalah interface untuk operasi bisnis Laporan.
type ReportService interface {
	CreateReport(input ReportInput) (*domain.Laporan, error)
	GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error)
	GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error)
	GetReportRecap(userID uint, period string, targetDate time.Time) (*repository.ReportRecapResponse, error)
	EvaluateReport(assessorID uint, assessorRole string, req EvaluateReportRequest) error
}

// reportService adalah implementasi dari ReportService.
type reportService struct {
	reportRepo repository.ReportRepository
}

// NewReportService membuat instance baru ReportService.
func NewReportService(reportRepo repository.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
}

// toStringPtr mengkonversi string ke pointer. Mengembalikan nil jika string kosong.
func toStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CreateReport membuat laporan baru dengan validasi bisnis.
func (s *reportService) CreateReport(input ReportInput) (*domain.Laporan, error) {
	now := time.Now()

	// 1. Validasi: Cek apakah hari ini adalah hari libur
	isHoliday, err := s.reportRepo.CheckIsHoliday(now)
	if err != nil {
		return nil, errors.New("gagal mengecek status hari libur")
	}
	if isHoliday {
		return nil, errors.New("laporan tidak dapat dibuat pada hari libur")
	}

	// 2. Validasi input berdasarkan tipe laporan
	// true = pokok (pelaporan tugas pokok)
	// false = tambahan (kegiatan tambahan, wajib ada judul)
	if !input.TipeLaporan && input.JudulKegiatan == "" {
		return nil, errors.New("judul kegiatan wajib diisi untuk laporan tambahan")
	}

	// 3. Cek jam kerja (07:00 - 16:00)
	// Jika di luar jam kerja, tandai sebagai overtime
	currentHour := now.Hour()
	isOvertime := currentHour < 7 || currentHour >= 16

	// 4. Proses upload file foto jika ada
	var fotoURL *string
	if input.FileFoto != nil {
		uploadedPath, err := s.saveFile(input.FileFoto, "images")
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan file foto: %v", err)
		}
		fotoURL = &uploadedPath
	}

	// 5. Proses upload file dokumen jika ada
	var dokumenURL *string
	if input.FileDokumen != nil {
		uploadedPath, err := s.saveFile(input.FileDokumen, "documents")
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan file dokumen: %v", err)
		}
		dokumenURL = &uploadedPath
	}

	// 6. Buat struct Laporan
	userID := input.UserID
	laporan := &domain.Laporan{
		UserID:         &userID,
		TipeLaporan:    input.TipeLaporan,
		TugasPokokID:   input.TugasPokokID,
		JudulKegiatan:  input.JudulKegiatan,
		DeskripsiHasil: input.DeskripsiHasil,
		WaktuPelaporan: input.WaktuPelaporan,
		IsOvertime:     isOvertime,
		LokasiLat:      toStringPtr(input.LokasiLat),
		LokasiLong:     toStringPtr(input.LokasiLong),
		AlamatLokasi:   toStringPtr(input.AlamatLokasi),
		FotoURL:        fotoURL,
		DokumenURL:     dokumenURL,
		CreatedAt:      now,
	}

	// 7. Simpan laporan ke database
	err = s.reportRepo.Create(laporan)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan laporan: %v", err)
	}

	return laporan, nil
}

// GetAllReports mengambil laporan dengan filter berdasarkan role requester (RBAC).
// - Lurah: Boleh melihat SEMUA laporan.
// - Sekertaris: HANYA boleh melihat laporan milik Staf.
// - Kasi & Staf: HANYA boleh melihat laporan DIRI SENDIRI.
func (s *reportService) GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error) {
	switch requesterRole {
	case "lurah":
		// Lurah boleh melihat semua laporan — tidak ada filter tambahan
	case "sekertaris":
		// Sekertaris hanya boleh melihat laporan milik staf
		filter.UserRole = "staf"
	case "kasi", "staf":
		// Kasi & Staf hanya boleh melihat laporan diri sendiri
		filter.UserID = int(requesterID)
	default:
		return nil, 0, errors.New("role tidak dikenali")
	}

	return s.reportRepo.GetAll(filter)
}

// GetReportDetail mengambil detail satu laporan.
func (s *reportService) GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error) {
	// 1. Ambil data laporan
	laporan, err := s.reportRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}

	// 2. Terapkan RBAC
	switch requesterRole {
	case "lurah":
		// Bebas akses
	case "sekertaris":
		if laporan.User != nil && laporan.User.Role != "staf" {
			return nil, errors.New("akses ditolak: hanya dapat melihat laporan staf")
		}
	case "kasi", "staf":
		if laporan.UserID != nil && *laporan.UserID != requesterID {
			return nil, errors.New("akses ditolak: hanya dapat melihat laporan milik sendiri")
		}
	default:
		return nil, errors.New("role tidak dikenali")
	}

	return laporan, nil
}

// saveFile menyimpan file ke subfolder uploads/reports/<subDir>
func (s *reportService) saveFile(fileHeader *multipart.FileHeader, subDir string) (string, error) {
	// Pastikan folder uploads/reports/<subDir> ada
	uploadDir := filepath.Join("./uploads/reports", subDir)
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Generate nama file unik dengan UUID
	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	// Buka source file
	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Buat destination file
	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy isi file
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return destPath, nil
}

// GetReportRecap menghitung agregasi status dan total jam kerja laporan untuk rentang waktu tertentu.
func (s *reportService) GetReportRecap(userID uint, period string, targetDate time.Time) (*repository.ReportRecapResponse, error) {
	var startDate, endDate time.Time

	switch period {
	case "harian":
		// Awal hari sampai akhir hari
		startDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
		endDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, targetDate.Location())
	case "mingguan":
		// Cari hari Senin di pekan tersebut
		offsetToMonday := int(time.Monday - targetDate.Weekday())
		if offsetToMonday > 0 {
			offsetToMonday = -6 // Jika target hari Minggu (0), mundur 6 hari
		}
		startDate = time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day()+offsetToMonday, 0, 0, 0, 0, targetDate.Location())
		endDate = startDate.AddDate(0, 0, 6) // Tambah 6 hari ke depan (Minggu)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, targetDate.Location())
	case "bulanan":
		// Tanggal 1 sampai akhir bulan
		startDate = time.Date(targetDate.Year(), targetDate.Month(), 1, 0, 0, 0, 0, targetDate.Location())
		endDate = startDate.AddDate(0, 1, -1) // Mundur 1 hari dari bulan depannya
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, targetDate.Location())
	default:
		return nil, errors.New("period tidak valid (harian, mingguan, bulanan)")
	}

	return s.reportRepo.GetReportRecap(userID, startDate, endDate)
}

// EvaluateReport mengevaluasi laporan (Approve/Reject) berdasarkan RBAC.
func (s *reportService) EvaluateReport(assessorID uint, assessorRole string, req EvaluateReportRequest) error {
	if req.Status != "Disetujui" && req.Status != "Ditolak" {
		return errors.New("status evaluasi tidak valid (harus 'Disetujui' atau 'Ditolak')")
	}

	// 1. Ambil data laporan beserta relasi User pengirimnya
	laporan, err := s.reportRepo.GetByID(req.ReportID)
	if err != nil {
		return errors.New("laporan tidak ditemukan")
	}

	targetUser := laporan.User
	if targetUser == nil {
		return errors.New("data user pemilik laporan tidak valid")
	}

	// 2. Terapkan RBAC Hierarki Penilaian
	switch assessorRole {
	case "lurah":
		// Bebas menilai laporan siapapun
	case "sekertaris":
		// Sekertaris hanya boleh menilai staf atau user yang SupervisorID-nya adalah dirinya
		isStaf := targetUser.Role == "staf"
		isDirectSubordinate := targetUser.SupervisorID != nil && *targetUser.SupervisorID == assessorID

		if !isStaf && !isDirectSubordinate {
			return errors.New("Anda tidak memiliki hak untuk mengevaluasi laporan pegawai ini")
		}
	case "kasi", "staf":
		// Kasi / Staf tidak punya hak approve laporan general
		return errors.New("akses ditolak")
	default:
		return errors.New("role tidak dikenali")
	}

	// 3. Update field
	laporan.Status = req.Status
	if req.Komentar != "" {
		laporan.KomentarAtasan = &req.Komentar
	} else {
		laporan.KomentarAtasan = nil
	}

	// 4. Save ke database
	err = s.reportRepo.Update(laporan)
	if err != nil {
		return fmt.Errorf("gagal mengevaluasi laporan: %v", err)
	}

	return nil
}
