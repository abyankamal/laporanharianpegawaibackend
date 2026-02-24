package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

type JabatanRepository interface {
	GetAll() ([]domain.RefJabatan, error)
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
