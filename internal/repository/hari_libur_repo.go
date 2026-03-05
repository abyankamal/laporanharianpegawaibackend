package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// HariLiburRepository adalah interface untuk operasi database HariLibur.
type HariLiburRepository interface {
	GetAll() ([]domain.HariLibur, error)
	Create(hariLibur *domain.HariLibur) error
	Delete(id uint) error
	CheckIsHoliday(date time.Time) (bool, error)
}

type hariLiburRepository struct {
	db *gorm.DB
}

// NewHariLiburRepository membuat instance baru HariLiburRepository.
func NewHariLiburRepository(db *gorm.DB) HariLiburRepository {
	return &hariLiburRepository{db: db}
}

// GetAll mengambil semua data hari libur, diurutkan dari yang terbaru.
func (r *hariLiburRepository) GetAll() ([]domain.HariLibur, error) {
	var list []domain.HariLibur
	err := r.db.Order("tanggal desc").Find(&list).Error
	return list, err
}

// Create menyimpan data hari libur baru.
func (r *hariLiburRepository) Create(hariLibur *domain.HariLibur) error {
	return r.db.Create(hariLibur).Error
}

// Delete menghapus hari libur berdasarkan ID.
func (r *hariLiburRepository) Delete(id uint) error {
	return r.db.Delete(&domain.HariLibur{}, id).Error
}

// CheckIsHoliday mengecek apakah tanggal tertentu adalah hari libur.
func (r *hariLiburRepository) CheckIsHoliday(date time.Time) (bool, error) {
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
