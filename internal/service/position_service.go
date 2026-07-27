package service

import (
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

type JabatanService interface {
	GetAllJabatan() ([]domain.RefJabatan, error)
	GetJabatanByID(id uint) (*domain.RefJabatan, error)
	CreateJabatan(nama string) (*domain.RefJabatan, error)
	UpdateJabatan(id uint, nama string) (*domain.RefJabatan, error)
	DeleteJabatan(id uint) error
}

type jabatanService struct {
	jabatanRepo repository.JabatanRepository
}

func NewJabatanService(jabatanRepo repository.JabatanRepository) JabatanService {
	return &jabatanService{jabatanRepo: jabatanRepo}
}

func (s *jabatanService) GetAllJabatan() ([]domain.RefJabatan, error) {
	return s.jabatanRepo.GetAll()
}

func (s *jabatanService) GetJabatanByID(id uint) (*domain.RefJabatan, error) {
	return s.jabatanRepo.GetByID(id)
}

func (s *jabatanService) CreateJabatan(nama string) (*domain.RefJabatan, error) {
	jabatan := &domain.RefJabatan{
		NamaJabatan: nama,
	}
	err := s.jabatanRepo.Create(jabatan)
	return jabatan, err
}

func (s *jabatanService) UpdateJabatan(id uint, nama string) (*domain.RefJabatan, error) {
	jabatan, err := s.jabatanRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	jabatan.NamaJabatan = nama
	err = s.jabatanRepo.Update(jabatan)
	return jabatan, err
}

func (s *jabatanService) DeleteJabatan(id uint) error {
	return s.jabatanRepo.Delete(id)
}
