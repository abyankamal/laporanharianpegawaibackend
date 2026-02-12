package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// ReportFilter adalah struct parameter input untuk filtering laporan.
type ReportFilter struct {
	StartDate string // Format YYYY-MM-DD
	EndDate   string // Format YYYY-MM-DD
	UserID    int
	JabatanID int
	Limit     int
	Offset    int
}

// ReportRepository adalah interface untuk operasi database Laporan.
type ReportRepository interface {
	Create(laporan *domain.Laporan) error
	CreateFileLaporan(file *domain.FileLaporan) error
	CheckIsHoliday(date time.Time) (bool, error)
	GetAll(filter ReportFilter) ([]domain.Laporan, int64, error)
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

// GetAll mengambil semua laporan berdasarkan filter yang diberikan.
// Mengembalikan slice laporan, total count, dan error.
func (r *reportRepository) GetAll(filter ReportFilter) ([]domain.Laporan, int64, error) {
	var reports []domain.Laporan
	var total int64

	query := r.db.Model(&domain.Laporan{})

	// Filter berdasarkan tanggal mulai
	if filter.StartDate != "" {
		query = query.Where("laporan.created_at >= ?", filter.StartDate+" 00:00:00")
	}

	// Filter berdasarkan tanggal akhir
	if filter.EndDate != "" {
		query = query.Where("laporan.created_at <= ?", filter.EndDate+" 23:59:59")
	}

	// Filter berdasarkan user_id
	if filter.UserID > 0 {
		query = query.Where("laporan.user_id = ?", filter.UserID)
	}

	// Filter berdasarkan jabatan_id (melalui join tabel users)
	if filter.JabatanID > 0 {
		query = query.Joins("JOIN users ON users.id = laporan.user_id").
			Where("users.jabatan_id = ?", filter.JabatanID)
	}

	// Hitung total data sebelum pagination (untuk metadata response)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Default limit jika tidak diset
	limit := filter.Limit
	if limit <= 0 {
		limit = 10
	}

	// Terapkan pagination dan sorting
	err := query.
		Preload("User").
		Preload("User.Jabatan").
		Order("laporan.created_at DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&reports).Error

	if err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}
