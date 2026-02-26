package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
)

// ============================================================
// Test GetSummary (DashboardService)
// ============================================================

func TestGetSummary_Success_Staf(t *testing.T) {
	t.Run("Sukses: Dashboard staf — tanpa laporan_masuk_hari_ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		now := time.Now()
		namaUser := "Staf Test"

		// Mock: ambil user
		mockUserRepo.On("FindByID", uint(3)).Return(&domain.User{ID: 3, Nama: namaUser}, nil)
		// Mock: 2 tugas belum dilaporkan hari ini
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(3)).Return(int64(2), nil)
		// Mock: 15 laporan bulan ini
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(3), now.Year(), int(now.Month())).Return(int64(15), nil)
		// Mock: laporan terbaru (kosong)
		mockDashboardRepo.On("GetRecentLaporan", uint(3), 3).Return([]domain.Laporan{}, nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		// Execute: staf (bukan atasan, jadi LaporanMasukHariIni tidak muncul)
		summary, err := dashboardSvc.GetSummary(3, "staf")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, namaUser, summary.NamaUser)
		assert.Equal(t, int64(2), summary.TaskPending)
		assert.Equal(t, int64(15), summary.LaporanBulanIni)
		assert.Nil(t, summary.LaporanMasukHariIni, "Staf tidak boleh melihat laporan_masuk_hari_ini")
		assert.NotNil(t, summary.RiwayatTerakhir)

		mockDashboardRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		// CountLaporanHariIni TIDAK boleh dipanggil untuk staf
		mockDashboardRepo.AssertNotCalled(t, "CountLaporanHariIni")
	})
}

func TestGetSummary_Success_Lurah(t *testing.T) {
	t.Run("Sukses: Dashboard lurah — dengan laporan_masuk_hari_ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		now := time.Now()
		namaUser := "Lurah Test"

		// Mock: ambil user
		mockUserRepo.On("FindByID", uint(1)).Return(&domain.User{ID: 1, Nama: namaUser}, nil)
		// Mock: 0 tugas pending (lurah tidak punya tugas pokok)
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(1)).Return(int64(0), nil)
		// Mock: 5 laporan bulan ini
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(1), now.Year(), int(now.Month())).Return(int64(5), nil)
		// Mock: laporan terbaru
		mockDashboardRepo.On("GetRecentLaporan", uint(1), 3).Return([]domain.Laporan{}, nil)
		// Mock: 8 laporan masuk hari ini (semua pegawai)
		mockDashboardRepo.On("CountLaporanHariIni").Return(int64(8), nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		// Execute: lurah (atasan, LaporanMasukHariIni harus muncul)
		summary, err := dashboardSvc.GetSummary(1, "lurah")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, namaUser, summary.NamaUser)
		assert.Equal(t, int64(0), summary.TaskPending)
		assert.Equal(t, int64(5), summary.LaporanBulanIni)
		assert.NotNil(t, summary.LaporanMasukHariIni, "Lurah harus bisa melihat laporan_masuk_hari_ini")
		assert.Equal(t, int64(8), *summary.LaporanMasukHariIni)

		mockDashboardRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestGetSummary_Success_Sekertaris(t *testing.T) {
	t.Run("Sukses: Dashboard sekertaris — dengan laporan_masuk_hari_ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		now := time.Now()
		namaUser := "Sekertaris Test"

		mockUserRepo.On("FindByID", uint(2)).Return(&domain.User{ID: 2, Nama: namaUser}, nil)
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(2)).Return(int64(1), nil)
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(2), now.Year(), int(now.Month())).Return(int64(20), nil)
		mockDashboardRepo.On("GetRecentLaporan", uint(2), 3).Return([]domain.Laporan{}, nil)
		mockDashboardRepo.On("CountLaporanHariIni").Return(int64(12), nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		summary, err := dashboardSvc.GetSummary(2, "sekertaris")

		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, namaUser, summary.NamaUser)
		assert.Equal(t, int64(1), summary.TaskPending)
		assert.Equal(t, int64(20), summary.LaporanBulanIni)
		assert.NotNil(t, summary.LaporanMasukHariIni)
		assert.Equal(t, int64(12), *summary.LaporanMasukHariIni)

		mockDashboardRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestGetSummary_Fail_UserNotFound(t *testing.T) {
	t.Run("Gagal: User tidak ditemukan", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		mockUserRepo.On("FindByID", uint(99)).Return(nil, errors.New("record not found"))

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		summary, err := dashboardSvc.GetSummary(99, "staf")

		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Equal(t, "record not found", err.Error())
	})
}

func TestGetSummary_Fail_CountPendingError(t *testing.T) {
	t.Run("Gagal: Error saat hitung tugas pending", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		mockUserRepo.On("FindByID", uint(1)).Return(&domain.User{ID: 1, Nama: "Test"}, nil)
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(1)).
			Return(int64(0), errors.New("database error"))

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		summary, err := dashboardSvc.GetSummary(1, "lurah")

		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestGetSummary_Fail_CountLaporanBulanError(t *testing.T) {
	t.Run("Gagal: Error saat hitung laporan bulan ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		now := time.Now()

		mockUserRepo.On("FindByID", uint(1)).Return(&domain.User{ID: 1, Nama: "Test"}, nil)
		// CountTugasPendingHariIni berhasil
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(1)).Return(int64(0), nil)
		// CountLaporanByUserAndMonth error
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(1), now.Year(), int(now.Month())).
			Return(int64(0), errors.New("query timeout"))

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		summary, err := dashboardSvc.GetSummary(1, "lurah")

		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Equal(t, "query timeout", err.Error())

		// CountLaporanHariIni tidak boleh terpanggil karena error terjadi sebelumnya
		mockDashboardRepo.AssertNotCalled(t, "CountLaporanHariIni")
	})
}

func TestGetSummary_Success_WithRiwayat(t *testing.T) {
	t.Run("Sukses: Mengembalikan riwayat laporan terbaru", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		now := time.Now()

		laporan := []domain.Laporan{
			{ID: 10, JudulKegiatan: "Rapat Koordinasi", CreatedAt: now.AddDate(0, 0, -1)},
			{ID: 9, JudulKegiatan: "Pembuatan Laporan", CreatedAt: now.AddDate(0, 0, -2)},
		}

		mockUserRepo.On("FindByID", mock.AnythingOfType("uint")).Return(&domain.User{ID: 5, Nama: "Staf Test"}, nil)
		mockDashboardRepo.On("CountTugasPendingHariIni", mock.AnythingOfType("uint")).Return(int64(0), nil)
		mockDashboardRepo.On("CountLaporanByUserAndMonth", mock.AnythingOfType("uint"), now.Year(), int(now.Month())).Return(int64(2), nil)
		mockDashboardRepo.On("GetRecentLaporan", mock.AnythingOfType("uint"), 3).Return(laporan, nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		summary, err := dashboardSvc.GetSummary(5, "kasi")

		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, 2, len(summary.RiwayatTerakhir))
		assert.Equal(t, "Rapat Koordinasi", summary.RiwayatTerakhir[0].JudulKegiatan)
		assert.Equal(t, uint(10), summary.RiwayatTerakhir[0].ID)
	})
}

func TestGetSummary_Success_ZeroValues(t *testing.T) {
	t.Run("Sukses: Semua count bernilai 0 (pegawai baru)", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		now := time.Now()

		mockUserRepo.On("FindByID", mock.AnythingOfType("uint")).Return(&domain.User{ID: 99, Nama: "Pegawai Baru"}, nil)
		mockDashboardRepo.On("CountTugasPendingHariIni", mock.AnythingOfType("uint")).Return(int64(0), nil)
		mockDashboardRepo.On("CountLaporanByUserAndMonth", mock.AnythingOfType("uint"), now.Year(), int(now.Month())).Return(int64(0), nil)
		mockDashboardRepo.On("GetRecentLaporan", mock.AnythingOfType("uint"), 3).Return([]domain.Laporan{}, nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo, mockUserRepo)

		summary, err := dashboardSvc.GetSummary(99, "kasi")

		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, int64(0), summary.TaskPending)
		assert.Equal(t, int64(0), summary.LaporanBulanIni)
		assert.Nil(t, summary.LaporanMasukHariIni, "Kasi tidak boleh melihat laporan_masuk_hari_ini")
		assert.Equal(t, 0, len(summary.RiwayatTerakhir))
	})
}
