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
	TipeLaporan    bool // true = pokok, false = tambahan
	JudulKegiatan  string
	DeskripsiHasil string
	WaktuMulai     time.Time
	WaktuSelesai   time.Time
	LokasiLat      string
	LokasiLong     string
	AlamatLokasi   string
	File           *multipart.FileHeader // File bukti (optional)
}

// ReportService adalah interface untuk operasi bisnis Laporan.
type ReportService interface {
	CreateReport(input ReportInput) (*domain.Laporan, error)
	GetAllReports(filter repository.ReportFilter) ([]domain.Laporan, int64, error)
}

// reportService adalah implementasi dari ReportService.
type reportService struct {
	reportRepo repository.ReportRepository
}

// NewReportService membuat instance baru ReportService.
func NewReportService(reportRepo repository.ReportRepository) ReportService {
	return &reportService{reportRepo: reportRepo}
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
		JudulKegiatan:  input.JudulKegiatan,
		DeskripsiHasil: input.DeskripsiHasil,
		WaktuMulai:     input.WaktuMulai,
		WaktuSelesai:   input.WaktuSelesai,
		IsOvertime:     isOvertime,
		LokasiLat:      input.LokasiLat,
		LokasiLong:     input.LokasiLong,
		AlamatLokasi:   input.AlamatLokasi,
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

// GetAllReports mengambil semua laporan dengan filter (untuk monitoring atasan).
func (s *reportService) GetAllReports(filter repository.ReportFilter) ([]domain.Laporan, int64, error) {
	return s.reportRepo.GetAll(filter)
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
