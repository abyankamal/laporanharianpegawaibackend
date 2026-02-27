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
	UserRole  string // Filter laporan berdasarkan role user (untuk RBAC)
	OwnID     int    // Filter tambahan untuk melihat laporan milik sendiri (untuk Sekertaris)
	Limit     int
	Offset    int
}

// ReportRecapResponse adalah struct rekapitulasi agregat laporan.
type ReportRecapResponse struct {
	TotalLaporan   int `json:"total_laporan" gorm:"column:total_laporan"`
	TotalDisetujui int `json:"total_disetujui" gorm:"column:total_disetujui"`
	TotalMenunggu  int `json:"total_menunggu" gorm:"column:total_menunggu"`
	TotalDitolak   int `json:"total_ditolak" gorm:"column:total_ditolak"`
	TotalJamKerja  int `json:"total_jam_kerja" gorm:"column:total_jam_kerja"`
}

// ReportRepository adalah interface untuk operasi database Laporan.
type ReportRepository interface {
	Create(laporan *domain.Laporan) error
	CreateFileLaporan(file *domain.FileLaporan) error
	GetFileByReportID(reportID uint) (*domain.FileLaporan, error)
	CheckIsHoliday(date time.Time) (bool, error)
	GetAll(filter ReportFilter) ([]domain.Laporan, int64, error)
	GetByID(id uint) (*domain.Laporan, error)
	GetReportRecap(userID uint, startDate time.Time, endDate time.Time) (*ReportRecapResponse, error)
	Update(laporan *domain.Laporan) error
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

// Update memperbarui data laporan di database (misal status/komentar).
func (r *reportRepository) Update(laporan *domain.Laporan) error {
	return r.db.Save(laporan).Error
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

	// Filter berdasarkan tanggal mulai (menggunakan waktu_pelaporan)
	if filter.StartDate != "" {
		query = query.Where("laporan.waktu_pelaporan >= ?", filter.StartDate+" 00:00:00")
	}

	// Filter berdasarkan tanggal akhir (menggunakan waktu_pelaporan)
	if filter.EndDate != "" {
		query = query.Where("laporan.waktu_pelaporan <= ?", filter.EndDate+" 23:59:59")
	}

	// Filter berdasarkan user_id
	if filter.UserID > 0 {
		query = query.Where("laporan.user_id = ?", filter.UserID)
	}

	// Flag untuk tracking apakah sudah JOIN users
	needsJoin := false

	// Filter berdasarkan role user (untuk RBAC - misal sekertaris hanya lihat laporan staf)
	if filter.UserRole != "" {
		needsJoin = true
		if filter.OwnID > 0 {
			// Jika ada OwnID (Sekertaris), tampilkan laporan staf OR laporan miliknya sendiri
			query = query.Joins("JOIN users ON users.id = laporan.user_id").
				Where("users.role = ? OR laporan.user_id = ?", filter.UserRole, filter.OwnID)
		} else {
			query = query.Joins("JOIN users ON users.id = laporan.user_id").
				Where("users.role = ?", filter.UserRole)
		}
	}

	// Filter berdasarkan jabatan_id (melalui join tabel users)
	if filter.JabatanID > 0 {
		if !needsJoin {
			query = query.Joins("JOIN users ON users.id = laporan.user_id")
		}
		query = query.Where("users.jabatan_id = ?", filter.JabatanID)
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

// GetByID mengambil satu laporan berdasarkan ID beserta relasinya.
func (r *reportRepository) GetByID(id uint) (*domain.Laporan, error) {
	var laporan domain.Laporan
	err := r.db.
		Preload("User").
		Preload("User.Jabatan").
		Preload("TugasOrganisasi").
		Where("id = ?", id).
		First(&laporan).Error

	if err != nil {
		return nil, err
	}
	return &laporan, nil
}

// GetFileByReportID mengambil file lampiran berdasarkan ID laporan.
func (r *reportRepository) GetFileByReportID(reportID uint) (*domain.FileLaporan, error) {
	var fileLaporan domain.FileLaporan
	err := r.db.Where("laporan_id = ?", reportID).First(&fileLaporan).Error
	if err != nil {
		return nil, err
	}
	return &fileLaporan, nil
}

// GetReportRecap menghitung agregasi status dan jam kerja laporan.
func (r *reportRepository) GetReportRecap(userID uint, startDate time.Time, endDate time.Time) (*ReportRecapResponse, error) {
	var rekap ReportRecapResponse

	err := r.db.Model(&domain.Laporan{}).
		Select("COUNT(id) as total_laporan, "+
			"SUM(CASE WHEN status = 'Disetujui' THEN 1 ELSE 0 END) as total_disetujui, "+
			"SUM(CASE WHEN status = 'Menunggu' THEN 1 ELSE 0 END) as total_menunggu, "+
			"SUM(CASE WHEN status = 'Ditolak' THEN 1 ELSE 0 END) as total_ditolak, "+
			"COALESCE(SUM(jam_kerja), 0) as total_jam_kerja").
		Where("user_id = ? AND waktu_pelaporan BETWEEN ? AND ?", userID, startDate, endDate).
		Scan(&rekap).Error

	if err != nil {
		return nil, err
	}
	return &rekap, nil
}
