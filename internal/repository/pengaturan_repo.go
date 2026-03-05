package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

// PengaturanRepository adalah interface untuk operasi database Pengaturan.
type PengaturanRepository interface {
	Get() (*domain.Pengaturan, error)
	Update(pengaturan *domain.Pengaturan) error
	SeedDefault() error
}

type pengaturanRepository struct {
	db *gorm.DB
}

// NewPengaturanRepository membuat instance baru PengaturanRepository.
func NewPengaturanRepository(db *gorm.DB) PengaturanRepository {
	return &pengaturanRepository{db: db}
}

// Get mengambil data pengaturan (selalu row dengan ID=1).
func (r *pengaturanRepository) Get() (*domain.Pengaturan, error) {
	var pengaturan domain.Pengaturan
	// Cari berdasarkan ID 1
	err := r.db.Where("id = ?", 1).First(&pengaturan).Error
	if err != nil {
		return nil, err
	}
	return &pengaturan, nil
}

// Update memperbarui data pengaturan.
func (r *pengaturanRepository) Update(pengaturan *domain.Pengaturan) error {
	// Pastikan ID selalu 1
	pengaturan.ID = 1
	return r.db.Save(pengaturan).Error
}

// SeedDefault mengisi data default jika tabel kosong.
func (r *pengaturanRepository) SeedDefault() error {
	var count int64
	r.db.Model(&domain.Pengaturan{}).Count(&count)

	if count == 0 {
		defaultPengaturan := domain.Pengaturan{
			ID:        1,
			JamMasuk:  "07:00",
			JamPulang: "18:00",
		}
		return r.db.Create(&defaultPengaturan).Error
	}
	return nil
}
