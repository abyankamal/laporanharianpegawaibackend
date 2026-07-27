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

	absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo)

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
	t.Run("Fails if face verification failed", func(t *testing.T) {
		mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
		mockHolidayRepo := new(mocks.HolidayRepositoryMock)
		mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

		absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo)

		// Make sure today is a workday for test or mock CheckIsHoliday
		now := time.Now()
		if now.Weekday() != time.Saturday && now.Weekday() != time.Sunday {
			mockHolidayRepo.On("CheckIsHoliday", mock.Anything).Return(false, nil).Maybe()
			mockAbsensiRepo.On("GetTodayAbsensi", uint(1)).Return(nil, errors.New("not found")).Maybe()

			input := service.AbsensiCheckInInput{
				UserID:       1,
				FaceVerified: false,
			}

			_, err := absensiService.CheckIn(input)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "verifikasi wajah gagal")
		}
	})
}

func TestAbsensiService_GetMonthlyRecap(t *testing.T) {
	mockAbsensiRepo := new(mocks.AbsensiRepositoryMock)
	mockHolidayRepo := new(mocks.HolidayRepositoryMock)
	mockWorkHourRepo := new(mocks.WorkHourRepositoryMock)

	absensiService := service.NewAbsensiService(mockAbsensiRepo, mockHolidayRepo, mockWorkHourRepo)

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
