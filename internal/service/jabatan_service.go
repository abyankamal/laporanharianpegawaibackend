package service

import (
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

type JabatanService interface {
	GetAllJabatan() ([]domain.RefJabatan, error)
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
