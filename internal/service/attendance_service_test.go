package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
	"laporanharianapi/internal/service"
)

func TestAbsensiService_IsWorkday(t *testing.T) {
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)
	mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
	mockUserRepo := new(mocks.UserRepositoryMock)

	absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo)

	t.Run("Saturday is not workday", func(t *testing.T) {
		saturday := time.Date(2026, 7, 25, 10, 0, 0, 0, time.Local) // 25 July 2026 is Saturday
		isWorkday, err := absensiService.IsWorkday(saturday)
		assert.NoError(t, err)
		assert.False(t, isWorkday)
	})

	t.Run("Sunday is not workday", func(t *testing.T) {
		sunday := time.Date(2026, 7, 26, 10, 0, 0, 0, time.Local)
		isWorkday, err := absensiService.IsWorkday(sunday)
		assert.NoError(t, err)
		assert.False(t, isWorkday)
	})

	t.Run("Monday is workday if not holiday", func(t *testing.T) {
		monday := time.Date(2026, 7, 27, 10, 0, 0, 0, time.Local)
		mockHolidayRepo.On("CheckIsHoliday", monday).Return(false, nil).Once()

		isWorkday, err := absensiService.IsWorkday(monday)
		assert.NoError(t, err)
		assert.True(t, isWorkday)
		mockHolidayRepo.AssertExpectations(t)
	})

	t.Run("Monday is not workday if holiday", func(t *testing.T) {
		monday := time.Date(2026, 7, 27, 10, 0, 0, 0, time.Local)
		mockHolidayRepo.On("CheckIsHoliday", monday).Return(true, nil).Once()

		isWorkday, err := absensiService.IsWorkday(monday)
		assert.NoError(t, err)
		assert.False(t, isWorkday)
		mockHolidayRepo.AssertExpectations(t)
	})
}

func TestAbsensiService_CheckIn(t *testing.T) {
	// Fixed weekday (Monday 2026-07-27 08:00:00)
	mondayTime := func() time.Time {
		return time.Date(2026, 7, 27, 8, 0, 0, 0, time.Local)
	}

	t.Run("Fails if user has no registered profile photo", func(t *testing.T) {
		mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		absensiService := service.NewAbsensiServiceWithClock(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo, mondayTime)

		mockHolidayRepo.On("CheckIsHoliday", mock.Anything).Return(false, nil).Maybe()
		mockAbsensiRepo.On("GetTodayAbsensi", uint(1)).Return(nil, errors.New("not found")).Maybe()
		mockUserRepo.On("FindByID", uint(1)).Return(&domain.User{ID: 1, FotoPath: nil}, nil).Once()

		input := service.AbsensiCheckInInput{
			UserID:       1,
			FaceVerified: true,
		}

		_, err := absensiService.CheckIn(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "belum mendaftarkan foto profil")
	})

	t.Run("Fails if face verification failed", func(t *testing.T) {
		mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		absensiService := service.NewAbsensiServiceWithClock(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo, mondayTime)

		foto := "uploads/photos/foto.jpg"
		mockHolidayRepo.On("CheckIsHoliday", mock.Anything).Return(false, nil).Maybe()
		mockAbsensiRepo.On("GetTodayAbsensi", uint(1)).Return(nil, errors.New("not found")).Maybe()
		mockUserRepo.On("FindByID", uint(1)).Return(&domain.User{ID: 1, FotoPath: &foto}, nil).Maybe()

		input := service.AbsensiCheckInInput{
			UserID:       1,
			FaceVerified: false,
		}

		_, err := absensiService.CheckIn(input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "verifikasi wajah gagal")
	})
}

func TestAbsensiService_GetMonthlyRecap(t *testing.T) {
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)
	mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
	mockUserRepo := new(mocks.UserRepositoryMock)

	absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo)

	t.Run("Returns calendar-aware recap and details with calculated alpha", func(t *testing.T) {
		hadirDate := time.Date(2026, 1, 5, 0, 0, 0, 0, time.Local) // Senin, 5 Jan 2026
		expectedDetails := []domain.Absensi{
			{ID: 1, UserID: 1, Tanggal: hadirDate, Status: "hadir"},
		}

		mockAbsensiRepo.On("GetByUserAndMonth", uint(1), 1, 2026).Return(expectedDetails, nil).Once()
		mockHolidayRepo.On("GetAll").Return([]domain.Holiday{}, nil).Once()

		details, recap, err := absensiService.GetMonthlyRecap(1, 1, 2026)
		assert.NoError(t, err)
		assert.Equal(t, expectedDetails, details)
		assert.Equal(t, 1, recap.TotalHadir)
		// Januari 2026 memiliki 22 hari kerja efektif (Senin-Jumat, 0 hari libur)
		assert.Equal(t, 22, recap.TotalHariKerja)
		// 22 hari kerja - 1 hari hadir = 21 hari alpha
		assert.Equal(t, 21, recap.TotalAlpha)
		mockAbsensiRepo.AssertExpectations(t)
		mockHolidayRepo.AssertExpectations(t)
	})
}

func TestAbsensiService_GetAllMonthlyRecap(t *testing.T) {
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)
	mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
	mockUserRepo := new(mocks.UserRepositoryMock)

	absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo)

	t.Run("Calculates batch recap accurately based on calendar workdays", func(t *testing.T) {
		users := []domain.User{
			{ID: 1, Nama: "Pegawai 1"},
			{ID: 2, Nama: "Pegawai 2"},
		}

		d1 := time.Date(2026, 7, 1, 0, 0, 0, 0, time.Local)
		d2 := time.Date(2026, 7, 2, 0, 0, 0, 0, time.Local)
		d3 := time.Date(2026, 7, 3, 0, 0, 0, 0, time.Local)
		d6 := time.Date(2026, 7, 6, 0, 0, 0, 0, time.Local)
		d7 := time.Date(2026, 7, 7, 0, 0, 0, 0, time.Local)
		d8 := time.Date(2026, 7, 8, 0, 0, 0, 0, time.Local)
		d9 := time.Date(2026, 7, 9, 0, 0, 0, 0, time.Local)
		d10 := time.Date(2026, 7, 10, 0, 0, 0, 0, time.Local)

		allAbsensi := []domain.Absensi{
			{ID: 1, UserID: 1, Tanggal: d1, Status: "hadir"},
			{ID: 2, UserID: 1, Tanggal: d2, Status: "terlambat"},
			{ID: 3, UserID: 1, Tanggal: d3, Status: "pulang_cepat"},
			{ID: 4, UserID: 2, Tanggal: d6, Status: "izin"},
			{ID: 5, UserID: 2, Tanggal: d7, Status: "sakit"},
			{ID: 6, UserID: 2, Tanggal: d8, Status: "alpha"},
			{ID: 7, UserID: 2, Tanggal: d9, Status: "cuti"},
			{ID: 8, UserID: 2, Tanggal: d10, Status: "dinas_luar"},
		}

		mockAbsensiRepo.On("GetAllByMonth", 7, 2026).Return(allAbsensi, nil).Once()
		mockHolidayRepo.On("GetAll").Return([]domain.Holiday{}, nil).Once()

		result, err := absensiService.GetAllMonthlyRecap(7, 2026, users)
		assert.NoError(t, err)
		assert.Len(t, result, 2)

		// Juli 2026 memiliki 23 hari kerja resmi (Senin-Jumat, 0 hari libur)
		// User 1 verification (3 hadir di kantor, 20 hari tanpa keterangan)
		assert.Equal(t, uint(1), result[0].User.ID)
		assert.Len(t, result[0].Details, 3)
		assert.Equal(t, 23, result[0].Recap.TotalHariKerja)
		assert.Equal(t, 1, result[0].Recap.TotalHadir)
		assert.Equal(t, 1, result[0].Recap.TotalTerlambat)
		assert.Equal(t, 1, result[0].Recap.TotalPulangCepat)
		assert.Equal(t, 20, result[0].Recap.TotalAlpha)

		// User 2 verification (5 hari tercatat: 1 izin, 1 sakit, 1 cuti, 1 dinas luar, 1 explicit alpha + 18 missing = 19 alpha)
		assert.Equal(t, uint(2), result[1].User.ID)
		assert.Len(t, result[1].Details, 5)
		assert.Equal(t, 23, result[1].Recap.TotalHariKerja)
		assert.Equal(t, 1, result[1].Recap.TotalIzin)
		assert.Equal(t, 1, result[1].Recap.TotalSakit)
		assert.Equal(t, 19, result[1].Recap.TotalAlpha)
		assert.Equal(t, 1, result[1].Recap.TotalCuti)
		assert.Equal(t, 1, result[1].Recap.TotalDinasLuar)

		mockAbsensiRepo.AssertExpectations(t)
		mockHolidayRepo.AssertExpectations(t)
	})
}

