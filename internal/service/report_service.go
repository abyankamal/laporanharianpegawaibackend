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
	File           *multipart.FileHeader // File bukti (optional)
}

// ReportService adalah interface untuk operasi bisnis Laporan.
type ReportService interface {
	CreateReport(input ReportInput) (*domain.Laporan, error)
	GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error)
	GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, *domain.FileLaporan, error)
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

	// 4. Proses upload file jika ada
	var filePath string
	if input.File != nil {
		uploadedPath, err := s.saveFile(input.File)
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan file: %v", err)
		}
		filePath = uploadedPath
	}

	// 5. Buat struct Laporan
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
		CreatedAt:      now,
	}

	// 6. Simpan laporan ke database
	err = s.reportRepo.Create(laporan)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan laporan: %v", err)
	}

	// 7. Simpan data file ke tabel file_laporan jika ada file
	if filePath != "" && input.File != nil {
		fileLaporan := &domain.FileLaporan{
			LaporanID:  &laporan.ID,
			TipeFile:   getFileType(input.File.Filename),
			FilePath:   filePath,
			UploadedAt: now,
		}
		err = s.reportRepo.CreateFileLaporan(fileLaporan)
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan data file: %v", err)
		}
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

// GetReportDetail mengambil detail satu laporan dan file lampirannya.
func (s *reportService) GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, *domain.FileLaporan, error) {
	// 1. Ambil data laporan
	laporan, err := s.reportRepo.GetByID(id)
	if err != nil {
		return nil, nil, errors.New("laporan tidak ditemukan")
	}

	// 2. Terapkan RBAC
	// - Lurah: Boleh melihat semua
	// - Sekertaris: Boleh melihat laporan milik staf
	// - Kasi & Staf: Hanya boleh melihat laporan sendiri
	switch requesterRole {
	case "lurah":
		// Bebas akses
	case "sekertaris":
		if laporan.User != nil && laporan.User.Role != "staf" {
			return nil, nil, errors.New("akses ditolak: hanya dapat melihat laporan staf")
		}
	case "kasi", "staf":
		if laporan.UserID != nil && *laporan.UserID != requesterID {
			return nil, nil, errors.New("akses ditolak: hanya dapat melihat laporan milik sendiri")
		}
	default:
		return nil, nil, errors.New("role tidak dikenali")
	}

	// 3. Ambil data file lampiran (jika ada)
	// Kita bisa query langsung melalui GORM ke tabel file_laporan menggunakan db instance,

	// Type assertion untuk mengakses gorm.DB dari repository (sementara, lebih baik ditambahkan di repo interface)
	// Alternatif yang lebih baik: kita akan tambahkan GetFileByReportID di reportRepository.
	// Di sini kita panggil fungsi tersebut.
	file, _ := s.reportRepo.GetFileByReportID(id)

	return laporan, file, nil
}

// saveFile menyimpan file ke folder uploads/reports
func (s *reportService) saveFile(fileHeader *multipart.FileHeader) (string, error) {
	// Pastikan folder uploads/reports ada
	uploadDir := "./uploads/reports"
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

// getFileType menentukan tipe file berdasarkan ekstensi
func getFileType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
		return "image"
	case ".mp4", ".mov", ".avi":
		return "video"
	case ".pdf":
		return "document"
	default:
		return "other"
	}
}
