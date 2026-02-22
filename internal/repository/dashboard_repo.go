package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// DashboardRepository adalah interface untuk query statistik dashboard.
type DashboardRepository interface {
	CountLaporanByUserAndMonth(userID uint, year int, month int) (int64, error)
	CountTugasPokokByUser(userID uint) (int64, error)
	CountLaporanHariIni() (int64, error)
	CountLaporanHariIniByRole(role string) (int64, error)
	CountTugasPendingHariIni(userID uint) (int64, error)
	GetRecentLaporan(userID uint, limit int) ([]domain.Laporan, error)
}

// dashboardRepository adalah implementasi dari DashboardRepository.
type dashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository membuat instance baru DashboardRepository.
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &dashboardRepository{db: db}
}

// CountLaporanByUserAndMonth menghitung jumlah laporan user pada bulan tertentu.
func (r *dashboardRepository) CountLaporanByUserAndMonth(userID uint, year int, month int) (int64, error) {
	var count int64
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, 0) // Awal bulan berikutnya

	err := r.db.Table("laporan").
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, startDate, endDate).
		Count(&count).Error

	return count, err
}

// CountTugasPokokByUser menghitung jumlah tugas pokok milik user.
func (r *dashboardRepository) CountTugasPokokByUser(userID uint) (int64, error) {
	var count int64
	err := r.db.Table("tugas_pokok").
		Where("user_id = ?", userID).
		Count(&count).Error
	return count, err
}

// CountLaporanHariIni menghitung total semua laporan yang masuk hari ini.
func (r *dashboardRepository) CountLaporanHariIni() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")

	err := r.db.Table("laporan").
		Where("created_at >= ? AND created_at < ?", today+" 00:00:00", today+" 23:59:59").
		Count(&count).Error
	return count, err
}

// CountLaporanHariIniByRole menghitung laporan hari ini dari user dengan role tertentu.
func (r *dashboardRepository) CountLaporanHariIniByRole(role string) (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")

	err := r.db.Table("laporan").
		Joins("JOIN users ON users.id = laporan.user_id").
		Where("users.role = ? AND laporan.created_at >= ? AND laporan.created_at < ?", role, today+" 00:00:00", today+" 23:59:59").
		Count(&count).Error
	return count, err
}

// CountTugasPendingHariIni menghitung jumlah tugas pokok user
// yang belum memiliki laporan (tipe_laporan = true) pada hari ini.
// Menggunakan LEFT JOIN + IS NULL untuk menemukan tugas tanpa laporan.
func (r *dashboardRepository) CountTugasPendingHariIni(userID uint) (int64, error) {
	var count int64
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	todayEnd := todayStart.AddDate(0, 0, 1)

	err := r.db.Table("tugas_pokok").
		Joins("LEFT JOIN laporan ON laporan.tugas_pokok_id = tugas_pokok.id AND laporan.tipe_laporan = ? AND laporan.created_at >= ? AND laporan.created_at < ?", true, todayStart, todayEnd).
		Where("tugas_pokok.user_id = ? AND laporan.id IS NULL", userID).
		Count(&count).Error

	return count, err
}

// GetRecentLaporan mengambil laporan terbaru milik user berdasarkan created_at DESC.
func (r *dashboardRepository) GetRecentLaporan(userID uint, limit int) ([]domain.Laporan, error) {
	var reports []domain.Laporan
	if limit <= 0 {
		limit = 5
	}
	err := r.db.Model(&domain.Laporan{}).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&reports).Error
	return reports, err
}
