package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

// WorkHourRepository adalah interface untuk operasi database WorkHour.
type WorkHourRepository interface {
	Get() (*domain.WorkHour, error)
	Update(workHour *domain.WorkHour) error
	SeedDefault() error
}

type workHourRepository struct {
	db *gorm.DB
}

// NewWorkHourRepository membuat instance baru WorkHourRepository.
func NewWorkHourRepository(db *gorm.DB) WorkHourRepository {
	return &workHourRepository{db: db}
}

// Get mengambil data workHour (selalu row dengan ID=1).
func (r *workHourRepository) Get() (*domain.WorkHour, error) {
	var workHour domain.WorkHour
	// Cari berdasarkan ID 1
	err := r.db.Where("id = ?", 1).First(&workHour).Error
	if err != nil {
		return nil, err
	}
	return &workHour, nil
}

// Update memperbarui data workHour.
func (r *workHourRepository) Update(workHour *domain.WorkHour) error {
	// Pastikan ID selalu 1
	workHour.ID = 1
	return r.db.Save(workHour).Error
}

// SeedDefault mengisi data default jika tabel kosong.
func (r *workHourRepository) SeedDefault() error {
	var count int64
	r.db.Model(&domain.WorkHour{}).Count(&count)

	if count == 0 {
		defaultWorkHour := domain.WorkHour{
			ID:             1,
			JamMasuk:       "07:00",
			JamPulang:      "18:00",
			JamMasukJumat:  "07:00",
			JamPulangJumat: "16:00",
		}
		return r.db.Create(&defaultWorkHour).Error
	}
	return nil
}
