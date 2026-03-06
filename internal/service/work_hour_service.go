package service

import (
	"errors"
	"regexp"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// WorkHourService adalah interface untuk operasi bisnis WorkHour.
type WorkHourService interface {
	GetWorkHour() (*domain.WorkHour, error)
	UpdateWorkHour(jamMasuk, jamPulang, jamMasukJumat, jamPulangJumat string) (*domain.WorkHour, error)
}

type workHourService struct {
	repo repository.WorkHourRepository
}

// NewWorkHourService membuat instance baru WorkHourService.
func NewWorkHourService(repo repository.WorkHourRepository) WorkHourService {
	return &workHourService{repo: repo}
}

// GetWorkHour mengambil data workHour sistem saat ini.
func (s *workHourService) GetWorkHour() (*domain.WorkHour, error) {
	return s.repo.Get()
}

// UpdateWorkHour memperbarui konfigurasi jam kerja.
func (s *workHourService) UpdateWorkHour(jamMasuk, jamPulang, jamMasukJumat, jamPulangJumat string) (*domain.WorkHour, error) {
	// Validasi format jam (HH:mm)
	timeFormat := regexp.MustCompile(`^([0-1][0-9]|2[0-3]):[0-5][0-9]$`)
	if !timeFormat.MatchString(jamMasuk) {
		return nil, errors.New("format jam masuk tidak valid (gunakan HH:mm, contoh: 07:00)")
	}
	if !timeFormat.MatchString(jamPulang) {
		return nil, errors.New("format jam pulang tidak valid (gunakan HH:mm, contoh: 18:00)")
	}
	if !timeFormat.MatchString(jamMasukJumat) {
		return nil, errors.New("format jam masuk jumat tidak valid (gunakan HH:mm, contoh: 07:00)")
	}
	if !timeFormat.MatchString(jamPulangJumat) {
		return nil, errors.New("format jam pulang jumat tidak valid (gunakan HH:mm, contoh: 16:00)")
	}

	workHour := &domain.WorkHour{
		ID:             1,
		JamMasuk:       jamMasuk,
		JamPulang:      jamPulang,
		JamMasukJumat:  jamMasukJumat,
		JamPulangJumat: jamPulangJumat,
	}

	err := s.repo.Update(workHour)
	if err != nil {
		return nil, err
	}

	return workHour, nil
}
