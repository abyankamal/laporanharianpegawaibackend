package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

type JabatanRepository interface {
	GetAll() ([]domain.RefJabatan, error)
	GetByID(id uint) (*domain.RefJabatan, error)
	Create(jabatan *domain.RefJabatan) error
	Update(jabatan *domain.RefJabatan) error
	Delete(id uint) error
}

type jabatanRepository struct {
	db *gorm.DB
}

func NewJabatanRepository(db *gorm.DB) JabatanRepository {
	return &jabatanRepository{db: db}
}

func (r *jabatanRepository) GetAll() ([]domain.RefJabatan, error) {
	var jabatans []domain.RefJabatan
	err := r.db.Find(&jabatans).Error
	return jabatans, err
}

func (r *jabatanRepository) GetByID(id uint) (*domain.RefJabatan, error) {
	var jabatan domain.RefJabatan
	err := r.db.First(&jabatan, id).Error
	return &jabatan, err
}

func (r *jabatanRepository) Create(jabatan *domain.RefJabatan) error {
	return r.db.Create(jabatan).Error
}

func (r *jabatanRepository) Update(jabatan *domain.RefJabatan) error {
	return r.db.Save(jabatan).Error
}

func (r *jabatanRepository) Delete(id uint) error {
	return r.db.Delete(&domain.RefJabatan{}, id).Error
}
