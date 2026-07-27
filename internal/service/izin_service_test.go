package service_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
	"laporanharianapi/internal/service"
)

func TestIzinService_CreatePengajuan(t *testing.T) {
	mockIzinRepo := new(mocks.IzinRepositoryMock)
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)

	izinService := service.NewIzinService(mockIzinRepo, mockAbsensiRepo)

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

	izinService := service.NewIzinService(mockIzinRepo, mockAbsensiRepo)

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
