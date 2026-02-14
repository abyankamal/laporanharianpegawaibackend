package repository

import (
	"time"

	"gorm.io/gorm"
)

// DashboardStats adalah struct untuk menyimpan statistik dashboard.
type DashboardStats struct {
	JumlahLaporanBulanIni int64 `json:"jumlah_laporan_bulan_ini"`
	JumlahTugasPokok      int64 `json:"jumlah_tugas_pokok"`
	LaporanMasukHariIni   int64 `json:"laporan_masuk_hari_ini"` // Hanya untuk atasan
}

// DashboardRepository adalah interface untuk query statistik dashboard.
type DashboardRepository interface {
	CountLaporanByUserAndMonth(userID uint, year int, month int) (int64, error)
	CountTugasPokokByUser(userID uint) (int64, error)
	CountLaporanHariIni() (int64, error)
	CountLaporanHariIniByRole(role string) (int64, error)
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
