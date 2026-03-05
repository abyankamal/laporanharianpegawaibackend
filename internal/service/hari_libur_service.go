package service

import (
	"errors"
	"time"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// HariLiburService adalah interface untuk operasi bisnis HariLibur.
type HariLiburService interface {
	GetHariLibur() ([]domain.HariLibur, error)
	CreateHariLibur(tanggal, keterangan string) (*domain.HariLibur, error)
	DeleteHariLibur(id uint) error
}

type hariLiburService struct {
	repo repository.HariLiburRepository
}

// NewHariLiburService membuat instance baru HariLiburService.
func NewHariLiburService(repo repository.HariLiburRepository) HariLiburService {
	return &hariLiburService{repo: repo}
}

// GetHariLibur mengambil semua data jadwal hari libur.
func (s *hariLiburService) GetHariLibur() ([]domain.HariLibur, error) {
	return s.repo.GetAll()
}

// CreateHariLibur menambahkan tanggal hari libur baru.
func (s *hariLiburService) CreateHariLibur(tanggalStr, keterangan string) (*domain.HariLibur, error) {
	if keterangan == "" {
		return nil, errors.New("keterangan hari libur wajib diisi")
	}

	// Parsing format tanggal YYYY-MM-DD
	tanggal, err := time.ParseInLocation("2006-01-02", tanggalStr, time.Local)
	if err != nil {
		return nil, errors.New("format tanggal tidak valid (gunakan YYYY-MM-DD, contoh: 2026-08-17)")
	}

	hariLibur := &domain.HariLibur{
		Tanggal:    tanggal,
		Keterangan: keterangan,
	}

	err = s.repo.Create(hariLibur)
	if err != nil {
		return nil, err
	}

	return hariLibur, nil
}

// DeleteHariLibur menghapus data hari libur.
func (s *hariLiburService) DeleteHariLibur(id uint) error {
	return s.repo.Delete(id)
}
