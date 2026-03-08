package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// HolidayRepository adalah interface untuk operasi database Holiday.
type HolidayRepository interface {
	GetAll() ([]domain.Holiday, error)
	GetByID(id uint) (*domain.Holiday, error)
	Create(holiday *domain.Holiday) error
	Update(holiday *domain.Holiday) error
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
	err := r.db.Order("tanggal_mulai asc").Find(&list).Error
	return list, err
}

// Create menyimpan data hari libur baru.
func (r *holidayRepository) Create(holiday *domain.Holiday) error {
	return r.db.Create(holiday).Error
}

// GetByID mengambil satu data hari libur berdasarkan ID.
func (r *holidayRepository) GetByID(id uint) (*domain.Holiday, error) {
	var holiday domain.Holiday
	err := r.db.First(&holiday, id).Error
	if err != nil {
		return nil, err
	}
	return &holiday, nil
}

// Update memperbarui data hari libur yang sudah ada.
func (r *holidayRepository) Update(holiday *domain.Holiday) error {
	return r.db.Save(holiday).Error
}

// Delete menghapus hari libur berdasarkan ID.
func (r *holidayRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Holiday{}, id).Error
}

// CheckIsHoliday mengecek apakah tanggal tertentu adalah hari libur (dalam rentang tanggal_mulai dan tanggal_selesai).
func (r *holidayRepository) CheckIsHoliday(date time.Time) (bool, error) {
	var count int64
	// Format tanggal untuk query (hanya tanggal, tanpa jam)
	dateOnly := date.Format("2006-01-02")

	err := r.db.Table("holiday").
		Where("? BETWEEN tanggal_mulai AND tanggal_selesai", dateOnly).
		Count(&count).Error

	if err != nil {
		// Jika tabel tidak ada atau error, asumsikan bukan hari libur
		return false, nil
	}

	return count > 0, nil
}
