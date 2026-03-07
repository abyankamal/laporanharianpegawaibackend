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

func TestGetHolidays_Success(t *testing.T) {
	mockRepo := new(mocks.HolidayRepositoryMock)
	mockTime, _ := time.Parse("2006-01-02", "2026-08-17")

	expectedList := []domain.Holiday{
		{ID: 1, TanggalMulai: mockTime, TanggalSelesai: mockTime, Keterangan: "Hari Kemerdekaan"},
	}

	mockRepo.On("GetAll").Return(expectedList, nil)

	svc := NewHolidayService(mockRepo)

	holidays, err := svc.GetHolidays()

	assert.NoError(t, err)
	assert.NotNil(t, holidays)
	assert.Len(t, holidays, 1)
	assert.Equal(t, "Hari Kemerdekaan", holidays[0].Keterangan)
	mockRepo.AssertExpectations(t)
}

func TestCreateHoliday_Success(t *testing.T) {
	mockRepo := new(mocks.HolidayRepositoryMock)
	mockRepo.On("Create", mock.AnythingOfType("*domain.Holiday")).Return(nil)

	svc := NewHolidayService(mockRepo)

	holiday, err := svc.CreateHoliday("2026-12-25", "2026-12-25", "Hari Natal")

	assert.NoError(t, err)
	assert.NotNil(t, holiday)
	assert.Equal(t, "Hari Natal", holiday.Keterangan)
	mockRepo.AssertExpectations(t)
}

func TestCreateHoliday_EmptyKeterangan(t *testing.T) {
	mockRepo := new(mocks.HolidayRepositoryMock)
	svc := NewHolidayService(mockRepo)

	holiday, err := svc.CreateHoliday("2026-12-25", "2026-12-25", "")

	assert.Error(t, err)
	assert.Nil(t, holiday)
	assert.Equal(t, "keterangan hari libur wajib diisi", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateHoliday_InvalidDate(t *testing.T) {
	mockRepo := new(mocks.HolidayRepositoryMock)
	svc := NewHolidayService(mockRepo)

	holiday, err := svc.CreateHoliday("25-12-2026", "26-12-2026", "Salah Format")

	assert.Error(t, err)
	assert.Nil(t, holiday)
	assert.Equal(t, "format tanggal tidak valid (gunakan YYYY-MM-DD, contoh: 2026-08-17)", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestDeleteHoliday_Success(t *testing.T) {
	mockRepo := new(mocks.HolidayRepositoryMock)
	mockRepo.On("Delete", uint(1)).Return(nil)

	svc := NewHolidayService(mockRepo)

	err := svc.DeleteHoliday(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteHoliday_Error(t *testing.T) {
	mockRepo := new(mocks.HolidayRepositoryMock)
	mockRepo.On("Delete", uint(1)).Return(errors.New("not found"))

	svc := NewHolidayService(mockRepo)

	err := svc.DeleteHoliday(1)

	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	mockRepo.AssertExpectations(t)
}
