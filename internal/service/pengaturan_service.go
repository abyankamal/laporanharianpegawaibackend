package service

import (
	"errors"
	"regexp"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// PengaturanService adalah interface untuk operasi bisnis Pengaturan.
type PengaturanService interface {
	GetPengaturan() (*domain.Pengaturan, error)
	UpdatePengaturan(jamMasuk, jamPulang string) (*domain.Pengaturan, error)
}

type pengaturanService struct {
	repo repository.PengaturanRepository
}

// NewPengaturanService membuat instance baru PengaturanService.
func NewPengaturanService(repo repository.PengaturanRepository) PengaturanService {
	return &pengaturanService{repo: repo}
}

// GetPengaturan mengambil data pengaturan sistem saat ini.
func (s *pengaturanService) GetPengaturan() (*domain.Pengaturan, error) {
	return s.repo.Get()
}

// UpdatePengaturan memperbarui konfigurasi jam kerja.
func (s *pengaturanService) UpdatePengaturan(jamMasuk, jamPulang string) (*domain.Pengaturan, error) {
	// Validasi format jam (HH:mm)
	timeFormat := regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0-5][0-9]$`)
	if !timeFormat.MatchString(jamMasuk) {
		return nil, errors.New("format jam masuk tidak valid (gunakan HH:mm, contoh: 07:00)")
	}
	if !timeFormat.MatchString(jamPulang) {
		return nil, errors.New("format jam pulang tidak valid (gunakan HH:mm, contoh: 18:00)")
	}

	pengaturan := &domain.Pengaturan{
		ID:        1,
		JamMasuk:  jamMasuk,
		JamPulang: jamPulang,
	}

	err := s.repo.Update(pengaturan)
	if err != nil {
		return nil, err
	}

	return pengaturan, nil
}
