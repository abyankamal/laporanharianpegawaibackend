package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// HolidayRepository adalah interface untuk operasi database Holiday.
type HolidayRepository interface {
	GetAll() ([]domain.Holiday, error)
	Create(holiday *domain.Holiday) error
	Delete(id uint) error
	CheckIsHoliday(date time.Time) (bool, error)
}

type holidayRepository struct {
	db *gorm.DB
}

// NewHolidayRepository membuat instance baru HolidayRepository.
func NewHolidayRepository(db *gorm.DB) HolidayRepository {
	return &holidayRepository{db: db}
}

// GetAll mengambil semua data hari libur, diurutkan dari yang terbaru.
func (r *holidayRepository) GetAll() ([]domain.Holiday, error) {
	var list []domain.Holiday
	err := r.db.Order("tanggal desc").Find(&list).Error
	return list, err
}

// Create menyimpan data hari libur baru.
func (r *holidayRepository) Create(holiday *domain.Holiday) error {
	return r.db.Create(holiday).Error
}

// Delete menghapus hari libur berdasarkan ID.
func (r *holidayRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Holiday{}, id).Error
}

// CheckIsHoliday mengecek apakah tanggal tertentu adalah hari libur.
func (r *holidayRepository) CheckIsHoliday(date time.Time) (bool, error) {
	var count int64
	// Format tanggal untuk query (hanya tanggal, tanpa jam)
	dateOnly := date.Format("2006-01-02")

	err := r.db.Table("holiday").
		Where("tanggal = ?", dateOnly).
		Count(&count).Error

	if err != nil {
		// Jika tabel tidak ada atau error, asumsikan bukan hari libur
		return false, nil
	}

	return count > 0, nil
}
