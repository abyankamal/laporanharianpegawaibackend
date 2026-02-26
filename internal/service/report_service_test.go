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
// Test CreateReport (ReportService)
// ============================================================

func TestCreateReport_Success_DocumentMode(t *testing.T) {
	t.Run("Sukses membuat laporan mode dokumen (tanpa lokasi, tanpa file)", func(t *testing.T) {
		// Setup
		mockReportRepo := new(mocks.ReportRepositoryMock)

		// Mock: Bukan hari libur
		mockReportRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, nil)
		// Mock: Simpan laporan berhasil
		mockReportRepo.On("Create", mock.Anything).Return(nil)

		reportSvc := NewReportService(mockReportRepo)

		// Execute: input tanpa lokasi (mode dokumen dari meja kantor)
		input := ReportInput{
			UserID:         1,
			TipeLaporan:    true, // pokok
			JudulKegiatan:  "Menyusun Laporan Bulanan",
			DeskripsiHasil: "Laporan bulanan telah diselesaikan",
			WaktuPelaporan: time.Now(),
			LokasiLat:      "", // kosong (opsional)
			LokasiLong:     "", // kosong (opsional)
			AlamatLokasi:   "", // kosong (opsional)
			FileFoto:       nil,
			FileDokumen:    nil,
		}

		laporan, err := reportSvc.CreateReport(input)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, laporan)
		assert.Nil(t, laporan.LokasiLat, "LokasiLat harus nil karena kosong")
		assert.Nil(t, laporan.LokasiLong, "LokasiLong harus nil karena kosong")
		assert.Nil(t, laporan.AlamatLokasi, "AlamatLokasi harus nil karena kosong")
		mockReportRepo.AssertExpectations(t)
	})
}

func TestCreateReport_Success_WithLocation(t *testing.T) {
	t.Run("Sukses membuat laporan dengan data lokasi", func(t *testing.T) {
		// Setup
		mockReportRepo := new(mocks.ReportRepositoryMock)

		mockReportRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, nil)
		mockReportRepo.On("Create", mock.Anything).Return(nil)

		reportSvc := NewReportService(mockReportRepo)

		// Execute: input dengan lokasi lengkap
		input := ReportInput{
			UserID:         1,
			TipeLaporan:    false, // tambahan
			JudulKegiatan:  "Survei Lapangan",
			DeskripsiHasil: "Survei lokasi telah dilakukan",
			WaktuPelaporan: time.Now(),
			LokasiLat:      "-6.2088",
			LokasiLong:     "106.8456",
			AlamatLokasi:   "Kantor Kelurahan",
			FileFoto:       nil,
			FileDokumen:    nil,
		}

		laporan, err := reportSvc.CreateReport(input)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, laporan)
		assert.NotNil(t, laporan.LokasiLat, "LokasiLat harus terisi")
		assert.Equal(t, "-6.2088", *laporan.LokasiLat)
		mockReportRepo.AssertExpectations(t)
	})
}

func TestCreateReport_Fail_Holiday(t *testing.T) {
	t.Run("Gagal membuat laporan pada hari libur", func(t *testing.T) {
		// Setup
		mockReportRepo := new(mocks.ReportRepositoryMock)

		// Mock: Hari ini adalah hari libur
		mockReportRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(true, nil)

		reportSvc := NewReportService(mockReportRepo)

		input := ReportInput{
			UserID:         1,
			TipeLaporan:    true,
			DeskripsiHasil: "Laporan hari libur",
			WaktuPelaporan: time.Now(),
		}

		laporan, err := reportSvc.CreateReport(input)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, laporan)
		assert.Equal(t, "laporan tidak dapat dibuat pada hari libur", err.Error())
		// Create TIDAK boleh dipanggil
		mockReportRepo.AssertNotCalled(t, "Create")
	})
}

func TestCreateReport_Fail_MissingJudulForTambahan(t *testing.T) {
	t.Run("Gagal membuat laporan tambahan tanpa judul kegiatan", func(t *testing.T) {
		// Setup
		mockReportRepo := new(mocks.ReportRepositoryMock)

		mockReportRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, nil)

		reportSvc := NewReportService(mockReportRepo)

		// Execute: tipe laporan = tambahan (false) tapi judul kegiatan kosong
		input := ReportInput{
			UserID:         1,
			TipeLaporan:    false, // tambahan
			JudulKegiatan:  "",    // kosong — harus gagal
			DeskripsiHasil: "Deskripsi tanpa judul",
			WaktuPelaporan: time.Now(),
		}

		laporan, err := reportSvc.CreateReport(input)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, laporan)
		assert.Equal(t, "judul kegiatan wajib diisi untuk laporan tambahan", err.Error())
		mockReportRepo.AssertNotCalled(t, "Create")
	})
}

func TestCreateReport_Fail_CheckHolidayError(t *testing.T) {
	t.Run("Gagal ketika pengecekan hari libur error", func(t *testing.T) {
		// Setup
		mockReportRepo := new(mocks.ReportRepositoryMock)

		// Mock: CheckIsHoliday mengembalikan error
		mockReportRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, errors.New("db error"))

		reportSvc := NewReportService(mockReportRepo)

		input := ReportInput{
			UserID:         1,
			TipeLaporan:    true,
			DeskripsiHasil: "Laporan test",
			WaktuPelaporan: time.Now(),
		}

		laporan, err := reportSvc.CreateReport(input)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, laporan)
		assert.Equal(t, "gagal mengecek status hari libur", err.Error())
	})
}
