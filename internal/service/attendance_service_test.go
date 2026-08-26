package service_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
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
	t.Run("Fails if user has no registered profile photo", func(t *testing.T) {
		mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)

		absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo)

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

		absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo)

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

	t.Run("Returns recap and details", func(t *testing.T) {
		expectedDetails := []domain.Absensi{
			{ID: 1, UserID: 1, Status: "hadir"},
		}
		expectedRecap := &repository.AbsensiRecapResponse{
			TotalHariKerja: 1,
			TotalHadir:     1,
		}

		mockAbsensiRepo.On("GetByUserAndMonth", uint(1), 1, 2026).Return(expectedDetails, nil).Once()
		mockAbsensiRepo.On("GetAbsensiRecap", uint(1), 1, 2026).Return(expectedRecap, nil).Once()

		details, recap, err := absensiService.GetMonthlyRecap(1, 1, 2026)
		assert.NoError(t, err)
		assert.Equal(t, expectedDetails, details)
		assert.Equal(t, expectedRecap, recap)
		mockAbsensiRepo.AssertExpectations(t)
	})
}

func TestAbsensiService_GetAllMonthlyRecap(t *testing.T) {
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)
	mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)
	mockUserRepo := new(mocks.UserRepositoryMock)

	absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo, mockUserRepo)

	t.Run("Calculates batch recap accurately in-memory without extra DB queries", func(t *testing.T) {
		users := []domain.User{
			{ID: 1, Nama: "Pegawai 1"},
			{ID: 2, Nama: "Pegawai 2"},
		}

		allAbsensi := []domain.Absensi{
			{ID: 1, UserID: 1, Status: "hadir"},
			{ID: 2, UserID: 1, Status: "terlambat"},
			{ID: 3, UserID: 1, Status: "pulang_cepat"},
			{ID: 4, UserID: 2, Status: "izin"},
			{ID: 5, UserID: 2, Status: "sakit"},
			{ID: 6, UserID: 2, Status: "alpha"},
			{ID: 7, UserID: 2, Status: "cuti"},
			{ID: 8, UserID: 2, Status: "dinas_luar"},
		}

		// Only GetAllByMonth should be called, NOT GetAbsensiRecap (eliminating N+1)
		mockAbsensiRepo.On("GetAllByMonth", 7, 2026).Return(allAbsensi, nil).Once()

		result, err := absensiService.GetAllMonthlyRecap(7, 2026, users)
		assert.NoError(t, err)
		assert.Len(t, result, 2)

		// User 1 verification
		assert.Equal(t, uint(1), result[0].User.ID)
		assert.Len(t, result[0].Details, 3)
		assert.Equal(t, 3, result[0].Recap.TotalHariKerja)
		assert.Equal(t, 1, result[0].Recap.TotalHadir)
		assert.Equal(t, 1, result[0].Recap.TotalTerlambat)
		assert.Equal(t, 1, result[0].Recap.TotalPulangCepat)
		assert.Equal(t, 0, result[0].Recap.TotalAlpha)

		// User 2 verification
		assert.Equal(t, uint(2), result[1].User.ID)
		assert.Len(t, result[1].Details, 5)
		assert.Equal(t, 5, result[1].Recap.TotalHariKerja)
		assert.Equal(t, 1, result[1].Recap.TotalIzin)
		assert.Equal(t, 1, result[1].Recap.TotalSakit)
		assert.Equal(t, 1, result[1].Recap.TotalAlpha)
		assert.Equal(t, 1, result[1].Recap.TotalCuti)
		assert.Equal(t, 1, result[1].Recap.TotalDinasLuar)

		mockAbsensiRepo.AssertExpectations(t)
	})
}

