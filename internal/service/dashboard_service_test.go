package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/repository/mocks"
)

// ============================================================
// Test GetSummary (DashboardService)
// ============================================================

func TestGetSummary_Success_Staf(t *testing.T) {
	t.Run("Sukses: Dashboard staf — tanpa laporan_masuk_hari_ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)

		now := time.Now()

		// Mock: 2 tugas belum dilaporkan hari ini
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(3)).Return(int64(2), nil)
		// Mock: 15 laporan bulan ini
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(3), now.Year(), int(now.Month())).Return(int64(15), nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo)

		// Execute: staf (bukan atasan, jadi LaporanMasukHariIni tidak muncul)
		summary, err := dashboardSvc.GetSummary(3, "staf")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, int64(2), summary.TaskPending)
		assert.Equal(t, int64(15), summary.LaporanBulanIni)
		assert.Nil(t, summary.LaporanMasukHariIni, "Staf tidak boleh melihat laporan_masuk_hari_ini")

		mockDashboardRepo.AssertExpectations(t)
		// CountLaporanHariIni TIDAK boleh dipanggil untuk staf
		mockDashboardRepo.AssertNotCalled(t, "CountLaporanHariIni")
	})
}

func TestGetSummary_Success_Lurah(t *testing.T) {
	t.Run("Sukses: Dashboard lurah — dengan laporan_masuk_hari_ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)

		now := time.Now()

		// Mock: 0 tugas pending (lurah tidak punya tugas pokok)
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(1)).Return(int64(0), nil)
		// Mock: 5 laporan bulan ini
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(1), now.Year(), int(now.Month())).Return(int64(5), nil)
		// Mock: 8 laporan masuk hari ini (semua pegawai)
		mockDashboardRepo.On("CountLaporanHariIni").Return(int64(8), nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo)

		// Execute: lurah (atasan, LaporanMasukHariIni harus muncul)
		summary, err := dashboardSvc.GetSummary(1, "lurah")

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, int64(0), summary.TaskPending)
		assert.Equal(t, int64(5), summary.LaporanBulanIni)
		assert.NotNil(t, summary.LaporanMasukHariIni, "Lurah harus bisa melihat laporan_masuk_hari_ini")
		assert.Equal(t, int64(8), *summary.LaporanMasukHariIni)

		mockDashboardRepo.AssertExpectations(t)
	})
}

func TestGetSummary_Success_Sekertaris(t *testing.T) {
	t.Run("Sukses: Dashboard sekertaris — dengan laporan_masuk_hari_ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)

		now := time.Now()

		mockDashboardRepo.On("CountTugasPendingHariIni", uint(2)).Return(int64(1), nil)
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(2), now.Year(), int(now.Month())).Return(int64(20), nil)
		mockDashboardRepo.On("CountLaporanHariIni").Return(int64(12), nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo)

		summary, err := dashboardSvc.GetSummary(2, "sekertaris")

		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, int64(1), summary.TaskPending)
		assert.Equal(t, int64(20), summary.LaporanBulanIni)
		assert.NotNil(t, summary.LaporanMasukHariIni)
		assert.Equal(t, int64(12), *summary.LaporanMasukHariIni)

		mockDashboardRepo.AssertExpectations(t)
	})
}

func TestGetSummary_Fail_CountPendingError(t *testing.T) {
	t.Run("Gagal: Error saat hitung tugas pending", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)

		mockDashboardRepo.On("CountTugasPendingHariIni", uint(1)).
			Return(int64(0), errors.New("database error"))

		dashboardSvc := NewDashboardService(mockDashboardRepo)

		summary, err := dashboardSvc.GetSummary(1, "lurah")

		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Equal(t, "database error", err.Error())
	})
}

func TestGetSummary_Fail_CountLaporanBulanError(t *testing.T) {
	t.Run("Gagal: Error saat hitung laporan bulan ini", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)

		now := time.Now()

		// CountTugasPendingHariIni berhasil
		mockDashboardRepo.On("CountTugasPendingHariIni", uint(1)).Return(int64(0), nil)
		// CountLaporanByUserAndMonth error
		mockDashboardRepo.On("CountLaporanByUserAndMonth", uint(1), now.Year(), int(now.Month())).
			Return(int64(0), errors.New("query timeout"))

		dashboardSvc := NewDashboardService(mockDashboardRepo)

		summary, err := dashboardSvc.GetSummary(1, "lurah")

		assert.Error(t, err)
		assert.Nil(t, summary)
		assert.Equal(t, "query timeout", err.Error())

		// CountLaporanHariIni tidak boleh terpanggil karena error terjadi sebelumnya
		mockDashboardRepo.AssertNotCalled(t, "CountLaporanHariIni")
	})
}

func TestGetSummary_Success_ZeroValues(t *testing.T) {
	t.Run("Sukses: Semua count bernilai 0 (pegawai baru)", func(t *testing.T) {
		mockDashboardRepo := new(mocks.DashboardRepositoryMock)

		now := time.Now()

		mockDashboardRepo.On("CountTugasPendingHariIni", mock.AnythingOfType("uint")).Return(int64(0), nil)
		mockDashboardRepo.On("CountLaporanByUserAndMonth", mock.AnythingOfType("uint"), now.Year(), int(now.Month())).Return(int64(0), nil)

		dashboardSvc := NewDashboardService(mockDashboardRepo)

		summary, err := dashboardSvc.GetSummary(99, "kasi")

		assert.NoError(t, err)
		assert.NotNil(t, summary)
		assert.Equal(t, int64(0), summary.TaskPending)
		assert.Equal(t, int64(0), summary.LaporanBulanIni)
		assert.Nil(t, summary.LaporanMasukHariIni, "Kasi tidak boleh melihat laporan_masuk_hari_ini")
	})
}
