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
// Test CreateReport (ReportService)
// ============================================================

func TestCreateReport_Success_DocumentMode(t *testing.T) {
	t.Run("Sukses membuat laporan mode dokumen (tanpa lokasi, tanpa file)", func(t *testing.T) {
		// Setup
		mockReportRepo := new(mocks.ReportRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

		// Mock: Bukan hari libur
		mockHolidayRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, nil)
		// Mock: Pengaturan jam kerja default
		mockWorkHourRepo.On("Get").Return(&domain.WorkHour{JamPulang: "16:00"}, nil)
		// Mock: Simpan laporan berhasil
		mockReportRepo.On("Create", mock.Anything).Return(nil)

		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

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
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

		mockHolidayRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, nil)
		mockWorkHourRepo.On("Get").Return(&domain.WorkHour{JamPulang: "16:00"}, nil)
		mockReportRepo.On("Create", mock.Anything).Return(nil)

		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

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
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

		// Mock: Hari ini adalah hari libur
		mockHolidayRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(true, nil)

		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

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
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

		mockHolidayRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, nil)

		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

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
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

		// Mock: CheckIsHoliday mengembalikan error
		mockHolidayRepo.On("CheckIsHoliday", mock.AnythingOfType("time.Time")).Return(false, errors.New("db error"))

		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

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

// ============================================================
// Test EvaluateReport (ReportService)
// ============================================================

func TestEvaluateReport_Success_Lurah(t *testing.T) {
	t.Run("Sukses: Lurah bebas menyetujui laporan siapa saja", func(t *testing.T) {
		mockReportRepo := new(mocks.ReportRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

		laporan := &domain.Laporan{
			ID:     1,
			UserID: func(i uint) *uint { return &i }(3),
			User: &domain.User{
				ID:   3,
				Role: "staf",
			},
		}

		mockReportRepo.On("GetByID", uint(1)).Return(laporan, nil)
		mockReportRepo.On("Update", mock.Anything).Return(nil)

		req := EvaluateReportRequest{
			ReportID: 1,
			Komentar: "Bagus",
		}
		err := reportSvc.EvaluateReport(1, "lurah", req)

		assert.NoError(t, err)
		assert.Equal(t, "sudah_direview", laporan.Status)
		assert.Equal(t, "Bagus", *laporan.KomentarAtasan)
		mockReportRepo.AssertCalled(t, "Update", mock.Anything)
	})
}

func TestEvaluateReport_Success_SekertarisToStaf(t *testing.T) {
	t.Run("Sukses: Sekertaris menilai laporan Staf", func(t *testing.T) {
		mockReportRepo := new(mocks.ReportRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

		laporan := &domain.Laporan{
			ID:     2,
			UserID: func(i uint) *uint { return &i }(4),
			User: &domain.User{
				ID:   4,
				Role: "staf",
			},
		}

		mockReportRepo.On("GetByID", uint(2)).Return(laporan, nil)
		mockReportRepo.On("Update", mock.Anything).Return(nil)

		req := EvaluateReportRequest{
			ReportID: 2,
			Komentar: "Perbaiki format",
		}
		err := reportSvc.EvaluateReport(2, "sekertaris", req)

		assert.NoError(t, err)
		assert.Equal(t, "sudah_direview", laporan.Status)
		assert.Equal(t, "Perbaiki format", *laporan.KomentarAtasan)
	})
}

func TestEvaluateReport_Fail_SekertarisToKasiWithoutSupervisorID(t *testing.T) {
	t.Run("Gagal: Sekertaris menilai Kasi yang bukan bawahannya", func(t *testing.T) {
		mockReportRepo := new(mocks.ReportRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

		laporan := &domain.Laporan{
			ID:     3,
			UserID: func(i uint) *uint { return &i }(5),
			User: &domain.User{
				ID:           5,
				Role:         "kasi",
				SupervisorID: func(i uint) *uint { return &i }(99), // bukan 2 (AssessorID)
			},
		}

		mockReportRepo.On("GetByID", uint(3)).Return(laporan, nil)

		req := EvaluateReportRequest{
			ReportID: 3,
			Komentar: "Mantap",
		}
		err := reportSvc.EvaluateReport(2, "sekertaris", req)

		assert.Error(t, err)
		assert.Equal(t, "Sekertaris hanya memiliki hak untuk mengevaluasi laporan Staf", err.Error())
		mockReportRepo.AssertNotCalled(t, "Update")
	})
}

func TestEvaluateReport_Fail_Kasi(t *testing.T) {
	t.Run("Gagal: Kasi/Staf mencoba melakukan evaluasi", func(t *testing.T) {
		mockReportRepo := new(mocks.ReportRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		reportSvc := NewReportService(mockReportRepo, mockHolidayRepo, mockWorkHourRepo)

		laporan := &domain.Laporan{
			ID:     4,
			UserID: func(i uint) *uint { return &i }(6),
			User: &domain.User{
				ID:   6,
				Role: "staf",
			},
		}

		mockReportRepo.On("GetByID", uint(4)).Return(laporan, nil)

		req := EvaluateReportRequest{
			ReportID: 4,
			Komentar: "Tes",
		}
		err := reportSvc.EvaluateReport(5, "kasi", req)

		assert.Error(t, err)
		assert.Equal(t, "akses ditolak", err.Error())
	})
}
