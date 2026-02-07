package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// ReportRepository adalah interface untuk operasi database Laporan.
type ReportRepository interface {
	Create(laporan *domain.Laporan) error
	CreateFileLaporan(file *domain.FileLaporan) error
	CheckIsHoliday(date time.Time) (bool, error)
}

// reportRepository adalah implementasi dari ReportRepository.
type reportRepository struct {
	db *gorm.DB
}

// NewReportRepository membuat instance baru ReportRepository.
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &reportRepository{db: db}
}

// Create menyimpan data laporan ke database.
func (r *reportRepository) Create(laporan *domain.Laporan) error {
	return r.db.Create(laporan).Error
}

// CreateFileLaporan menyimpan data file laporan ke database.
func (r *reportRepository) CreateFileLaporan(file *domain.FileLaporan) error {
	return r.db.Create(file).Error
}

// CheckIsHoliday mengecek apakah tanggal tertentu adalah hari libur.
// CATATAN: Fungsi ini memerlukan tabel hari_libur aktif di database.
func (r *reportRepository) CheckIsHoliday(date time.Time) (bool, error) {
	var count int64
	// Format tanggal untuk query (hanya tanggal, tanpa jam)
	dateOnly := date.Format("2006-01-02")

	err := r.db.Table("hari_libur").
		Where("tanggal = ?", dateOnly).
		Count(&count).Error

	if err != nil {
		// Jika tabel tidak ada atau error, asumsikan bukan hari libur
		return false, nil
	}

	return count > 0, nil
}
