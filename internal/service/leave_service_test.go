package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
	"laporanharianapi/internal/service"
)

func TestIzinService_CreatePengajuan(t *testing.T) {
	mockIzinRepo := new(mocks.IzinRepositoryMock)
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)

	izinService := service.NewIzinService(mockIzinRepo, mockAbsensiRepo, mockHolidayRepo)

	t.Run("Fails if jenis izin invalid", func(t *testing.T) {
		input := service.PengajuanIzinInput{
			UserID:         1,
			JenisIzin:      "liburan",
			TanggalMulai:   "2026-07-27",
			TanggalSelesai: "2026-07-28",
			Keterangan:     "Acara keluarga",
		}

		_, err := izinService.CreatePengajuan(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "jenis izin tidak valid")
	})

	t.Run("Fails if keterangan empty", func(t *testing.T) {
		input := service.PengajuanIzinInput{
			UserID:         1,
			JenisIzin:      "sakit",
			TanggalMulai:   "2026-07-27",
			TanggalSelesai: "2026-07-28",
			Keterangan:     "   ",
		}

		_, err := izinService.CreatePengajuan(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "keterangan wajib diisi")
	})

	t.Run("Fails if tanggal selesai before mulai", func(t *testing.T) {
		input := service.PengajuanIzinInput{
			UserID:         1,
			JenisIzin:      "cuti",
			TanggalMulai:   "2026-07-28",
			TanggalSelesai: "2026-07-27",
			Keterangan:     "Cuti tahunan",
		}

		_, err := izinService.CreatePengajuan(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "tanggal selesai tidak boleh lebih awal")
	})

	t.Run("CreateByAdmin fails if user_id is 0", func(t *testing.T) {
		input := service.PengajuanIzinInput{
			UserID:         0,
			JenisIzin:      "cuti",
			TanggalMulai:   "2026-07-27",
			TanggalSelesai: "2026-07-28",
			Keterangan:     "Cuti tahunan",
		}

		_, err := izinService.CreateByAdmin(input, 100)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pegawai (user_id) wajib dipilih")
	})
}

func TestIzinService_GetMyPengajuan(t *testing.T) {
	mockIzinRepo := new(mocks.IzinRepositoryMock)
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)

	izinService := service.NewIzinService(mockIzinRepo, mockAbsensiRepo, mockHolidayRepo)

	t.Run("Returns user pengajuan list", func(t *testing.T) {
		expected := []domain.PengajuanIzin{
			{ID: 1, UserID: 1, JenisIzin: "sakit", StatusApproval: "menunggu", CreatedAt: time.Now()},
		}

		mockIzinRepo.On("GetByUserID", uint(1)).Return(expected, nil).Once()

		res, err := izinService.GetMyPengajuan(1)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		mockIzinRepo.AssertExpectations(t)
	})
}

func TestIzinService_ApprovePengajuan_SkipsHoliday(t *testing.T) {
	t.Run("Approved leave creates absensi records on workdays but skips public holiday", func(t *testing.T) {
		mockIzinRepo := new(mocks.IzinRepositoryMock)
		mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)

		izinService := service.NewIzinService(mockIzinRepo, mockAbsensiRepo, mockHolidayRepo)

		// Sen, 17 Agu 2026 (Hari Kemerdekaan RI - Libur Nasional)
		// Sel, 18 Agu 2026 (Hari Kerja Biasa)
		senin, _ := time.Parse("2006-01-02", "2026-08-17")
		selasa, _ := time.Parse("2006-01-02", "2026-08-18")

		izin := domain.PengajuanIzin{
			ID:             10,
			UserID:         5,
			JenisIzin:      "cuti",
			TanggalMulai:   senin,
			TanggalSelesai: selasa,
			StatusApproval: "menunggu",
		}

		mockIzinRepo.On("GetByID", uint(10)).Return(&izin, nil).Once()
		mockIzinRepo.On("Update", &izin).Return(nil).Once()

		// 17 Agu adalah libur nasional -> CheckIsHoliday returns true
		mockHolidayRepo.On("CheckIsHoliday", senin).Return(true, nil)
		// 18 Agu bukan libur -> CheckIsHoliday returns false
		mockHolidayRepo.On("CheckIsHoliday", selasa).Return(false, nil)

		// Hanya 18 Agu yang dicatat sebagai absensi cuti
		selasaDateOnly := time.Date(selasa.Year(), selasa.Month(), selasa.Day(), 0, 0, 0, 0, time.Local)
		mockAbsensiRepo.On("GetByUserAndDate", uint(5), selasaDateOnly).Return(nil, nil).Once()
		mockAbsensiRepo.On("Create", mock.AnythingOfType("*domain.Absensi")).Return(nil).Once()

		err := izinService.ApprovePengajuan(10, 1, true, "Disetujui")
		assert.NoError(t, err)
		mockAbsensiRepo.AssertNumberOfCalls(t, "Create", 1)
	})
}
