package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
)

func TestGetWorkHour_Success(t *testing.T) {
	mockRepo := new(mocks.WorkHourRepositoryMock)
	mockRepo.On("Get").Return(&domain.WorkHour{ID: 1, JamMasuk: "07:00", JamPulang: "18:00", JamMasukJumat: "07:00", JamPulangJumat: "16:00"}, nil)

	svc := NewWorkHourService(mockRepo)

	workHour, err := svc.GetWorkHour()

	assert.NoError(t, err)
	assert.NotNil(t, workHour)
	assert.Equal(t, "07:00", workHour.JamMasuk)
	assert.Equal(t, "18:00", workHour.JamPulang)
	assert.Equal(t, "07:00", workHour.JamMasukJumat)
	assert.Equal(t, "16:00", workHour.JamPulangJumat)
	mockRepo.AssertExpectations(t)
}

func TestGetWorkHour_Error(t *testing.T) {
	mockRepo := new(mocks.WorkHourRepositoryMock)
	mockRepo.On("Get").Return(nil, errors.New("db error"))

	svc := NewWorkHourService(mockRepo)

	workHour, err := svc.GetWorkHour()

	assert.Error(t, err)
	assert.Nil(t, workHour)
	mockRepo.AssertExpectations(t)
}

func TestUpdateWorkHour_Success(t *testing.T) {
	mockRepo := new(mocks.WorkHourRepositoryMock)
	mockRepo.On("Update", mock.AnythingOfType("*domain.WorkHour")).Return(nil)

	svc := NewWorkHourService(mockRepo)

	workHour, err := svc.UpdateWorkHour("08:00", "16:00", "07:30", "15:30")

	assert.NoError(t, err)
	assert.NotNil(t, workHour)
	assert.Equal(t, "08:00", workHour.JamMasuk)
	assert.Equal(t, "16:00", workHour.JamPulang)
	assert.Equal(t, "07:30", workHour.JamMasukJumat)
	assert.Equal(t, "15:30", workHour.JamPulangJumat)
	assert.Equal(t, uint(1), workHour.ID)
	mockRepo.AssertExpectations(t)
}

func TestUpdateWorkHour_InvalidJamMasuk(t *testing.T) {
	mockRepo := new(mocks.WorkHourRepositoryMock)

	svc := NewWorkHourService(mockRepo)

	workHour, err := svc.UpdateWorkHour("8:00", "16:00", "07:00", "16:00") // invalid format

	assert.Error(t, err)
	assert.Nil(t, workHour)
	assert.Equal(t, "format jam masuk tidak valid (gunakan HH:mm, contoh: 07:00)", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateWorkHour_InvalidJamPulang(t *testing.T) {
	mockRepo := new(mocks.WorkHourRepositoryMock)

	svc := NewWorkHourService(mockRepo)

	workHour, err := svc.UpdateWorkHour("08:00", "46:00", "07:00", "16:00") // invalid time

	assert.Error(t, err)
	assert.Nil(t, workHour)
	assert.Equal(t, "format jam pulang tidak valid (gunakan HH:mm, contoh: 18:00)", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdateWorkHour_DBError(t *testing.T) {
	mockRepo := new(mocks.WorkHourRepositoryMock)
	mockRepo.On("Update", mock.AnythingOfType("*domain.WorkHour")).Return(errors.New("db error"))

	svc := NewWorkHourService(mockRepo)

	workHour, err := svc.UpdateWorkHour("08:00", "16:00", "07:00", "16:00")

	assert.Error(t, err)
	assert.Nil(t, workHour)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}
