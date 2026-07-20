package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

// SupervisorRepository interface untuk akses data supervisor lurah.
type SupervisorRepository interface {
	GetSupervisor() (*domain.LurahSupervisor, error)
	UpdateSupervisor(nama, nip string) error
	SeedDefault() error
}

type supervisorRepository struct {
	db *gorm.DB
}

// NewSupervisorRepository membuat instance baru SupervisorRepository.
func NewSupervisorRepository(db *gorm.DB) SupervisorRepository {
	return &supervisorRepository{db: db}
}

// GetSupervisor mengambil data atasan lurah (hanya ada 1 record).
func (r *supervisorRepository) GetSupervisor() (*domain.LurahSupervisor, error) {
	var supervisor domain.LurahSupervisor
	err := r.db.First(&supervisor).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Jika belum ada, panggil seed default
			r.SeedDefault()
			return r.GetSupervisor()
		}
		return nil, err
	}
	return &supervisor, nil
}

// UpdateSupervisor memperbarui data atasan lurah.
func (r *supervisorRepository) UpdateSupervisor(nama, nip string) error {
	var supervisor domain.LurahSupervisor
	err := r.db.First(&supervisor).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create baru
			return r.db.Create(&domain.LurahSupervisor{
				Nama: nama,
				NIP:  nip,
			}).Error
		}
		return err
	}

	supervisor.Nama = nama
	supervisor.NIP = nip
	return r.db.Save(&supervisor).Error
}

// SeedDefault mengisi data default jika tabel kosong.
func (r *supervisorRepository) SeedDefault() error {
	var count int64
	r.db.Model(&domain.LurahSupervisor{}).Count(&count)
	if count == 0 {
		defaultSupervisor := domain.LurahSupervisor{
			Nama: "Rena Sudrajat, S.Sos., M.Si",
			NIP:  "197208241992031003",
		}
		return r.db.Create(&defaultSupervisor).Error
	}
	return nil
}
