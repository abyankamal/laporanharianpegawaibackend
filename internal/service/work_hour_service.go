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
	UpdateGeofencing(kantorLat, kantorLong *string, radiusMeter int, geofencingEnabled bool) (*domain.WorkHour, error)
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

// UpdateWorkHour memperbarui konfigurasi jam kerja tanpa menimpa koordinat geofencing.
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

	existing, err := s.repo.Get()
	if err != nil || existing == nil {
		existing = &domain.WorkHour{
			ID:                1,
			RadiusMeter:       60,
			GeofencingEnabled: false,
		}
	}

	existing.ID = 1
	existing.JamMasuk = jamMasuk
	existing.JamPulang = jamPulang
	existing.JamMasukJumat = jamMasukJumat
	existing.JamPulangJumat = jamPulangJumat

	err = s.repo.Update(existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}

// UpdateGeofencing memperbarui konfigurasi geofencing kantor tanpa menimpa jam kerja.
func (s *workHourService) UpdateGeofencing(kantorLat, kantorLong *string, radiusMeter int, geofencingEnabled bool) (*domain.WorkHour, error) {
	if radiusMeter <= 0 {
		radiusMeter = 60
	}

	existing, err := s.repo.Get()
	if err != nil || existing == nil {
		existing = &domain.WorkHour{
			ID:             1,
			JamMasuk:       "07:30",
			JamPulang:      "16:00",
			JamMasukJumat:  "07:30",
			JamPulangJumat: "16:30",
		}
	}

	existing.ID = 1
	existing.KantorLat = kantorLat
	existing.KantorLong = kantorLong
	existing.RadiusMeter = radiusMeter
	existing.GeofencingEnabled = geofencingEnabled

	err = s.repo.Update(existing)
	if err != nil {
		return nil, err
	}

	return existing, nil
}
