package service

import (
	"errors"
	"time"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// HolidayService adalah interface untuk operasi bisnis Holiday.
type HolidayService interface {
	GetHolidays() ([]domain.Holiday, error)
	CreateHoliday(tanggalMulai, tanggalSelesai, keterangan string) (*domain.Holiday, error)
	DeleteHoliday(id uint) error
}

type holidayService struct {
	repo repository.HolidayRepository
}

// NewHolidayService membuat instance baru HolidayService.
func NewHolidayService(repo repository.HolidayRepository) HolidayService {
	return &holidayService{repo: repo}
}

// GetHolidays mengambil semua data jadwal hari libur.
func (s *holidayService) GetHolidays() ([]domain.Holiday, error) {
	return s.repo.GetAll()
}

// CreateHoliday menambahkan tanggal hari libur baru (mendukung rentang tanggal).
func (s *holidayService) CreateHoliday(tanggalMulaiStr, tanggalSelesaiStr, keterangan string) (*domain.Holiday, error) {
	if keterangan == "" {
		return nil, errors.New("keterangan hari libur wajib diisi")
	}

	// Parsing format tanggal YYYY-MM-DD
	tanggalMulai, err := time.ParseInLocation("2006-01-02", tanggalMulaiStr, time.Local)
	if err != nil {
		return nil, errors.New("format tanggal mulai tidak valid (gunakan YYYY-MM-DD, contoh: 2026-08-17)")
	}

	tanggalSelesai, err := time.ParseInLocation("2006-01-02", tanggalSelesaiStr, time.Local)
	if err != nil {
		return nil, errors.New("format tanggal selesai tidak valid (gunakan YYYY-MM-DD, contoh: 2026-08-17)")
	}

	// Validasi range tanggal
	if tanggalSelesai.Before(tanggalMulai) {
		return nil, errors.New("tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	}

	holiday := &domain.Holiday{
		TanggalMulai:   tanggalMulai,
		TanggalSelesai: tanggalSelesai,
		Keterangan:     keterangan,
	}

	err = s.repo.Create(holiday)
	if err != nil {
		return nil, err
	}

	return holiday, nil
}

// DeleteHoliday menghapus data hari libur.
func (s *holidayService) DeleteHoliday(id uint) error {
	return s.repo.Delete(id)
}
