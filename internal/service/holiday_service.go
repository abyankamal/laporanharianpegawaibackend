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
	GetHolidayByID(id uint) (*domain.Holiday, error)
	CreateHoliday(tanggalMulai, tanggalSelesai, keterangan string) (*domain.Holiday, error)
	UpdateHoliday(id uint, tanggalMulai, tanggalSelesai, keterangan string) (*domain.Holiday, error)
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

// GetHolidayByID mengambil satu data hari libur berdasarkan ID.
func (s *holidayService) GetHolidayByID(id uint) (*domain.Holiday, error) {
	return s.repo.GetByID(id)
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

// UpdateHoliday memperbarui data hari libur yang sudah ada (Mendukung Partial Update).
func (s *holidayService) UpdateHoliday(id uint, tanggalMulaiStr, tanggalSelesaiStr, keterangan string) (*domain.Holiday, error) {
	holiday, err := s.repo.GetByID(id)
	if err != nil {
		return nil, errors.New("hari libur tidak ditemukan")
	}

	// 1. Update Keterangan jika diisi
	if keterangan != "" {
		holiday.Keterangan = keterangan
	}

	// 2. Parsing Tanggal Mulai jika diisi
	if tanggalMulaiStr != "" {
		tanggalMulai, err := time.ParseInLocation("2006-01-02", tanggalMulaiStr, time.Local)
		if err != nil {
			return nil, errors.New("format tanggal mulai tidak valid (gunakan YYYY-MM-DD)")
		}
		holiday.TanggalMulai = tanggalMulai
	}

	// 3. Parsing Tanggal Selesai jika diisi
	if tanggalSelesaiStr != "" {
		tanggalSelesai, err := time.ParseInLocation("2006-01-02", tanggalSelesaiStr, time.Local)
		if err != nil {
			return nil, errors.New("format tanggal selesai tidak valid (gunakan YYYY-MM-DD)")
		}
		holiday.TanggalSelesai = tanggalSelesai
	}

	// 4. Validasi ulang range tanggal jika ada perubahan salah satu atau keduanya
	if holiday.TanggalSelesai.Before(holiday.TanggalMulai) {
		return nil, errors.New("tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	}

	err = s.repo.Update(holiday)
	if err != nil {
		return nil, err
	}

	return holiday, nil
}

// DeleteHoliday menghapus data hari libur.
func (s *holidayService) DeleteHoliday(id uint) error {
	return s.repo.Delete(id)
}
