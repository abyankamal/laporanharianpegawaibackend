package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
)

func TestGetPengaturan_Success(t *testing.T) {
	mockRepo := new(mocks.PengaturanRepositoryMock)
	mockRepo.On("Get").Return(&domain.Pengaturan{ID: 1, JamMasuk: "07:00", JamPulang: "18:00"}, nil)

	svc := NewPengaturanService(mockRepo)

	pengaturan, err := svc.GetPengaturan()

	assert.NoError(t, err)
	assert.NotNil(t, pengaturan)
	assert.Equal(t, "07:00", pengaturan.JamMasuk)
	assert.Equal(t, "18:00", pengaturan.JamPulang)
	mockRepo.AssertExpectations(t)
}

func TestGetPengaturan_Error(t *testing.T) {
	mockRepo := new(mocks.PengaturanRepositoryMock)
	mockRepo.On("Get").Return(nil, errors.New("db error"))

	svc := NewPengaturanService(mockRepo)

	pengaturan, err := svc.GetPengaturan()

	assert.Error(t, err)
	assert.Nil(t, pengaturan)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePengaturan_Success(t *testing.T) {
	mockRepo := new(mocks.PengaturanRepositoryMock)
	mockRepo.On("Update", mock.AnythingOfType("*domain.Pengaturan")).Return(nil)

	svc := NewPengaturanService(mockRepo)

	pengaturan, err := svc.UpdatePengaturan("08:00", "16:00")

	assert.NoError(t, err)
	assert.NotNil(t, pengaturan)
	assert.Equal(t, "08:00", pengaturan.JamMasuk)
	assert.Equal(t, "16:00", pengaturan.JamPulang)
	assert.Equal(t, uint(1), pengaturan.ID)
	mockRepo.AssertExpectations(t)
}

func TestUpdatePengaturan_InvalidJamMasuk(t *testing.T) {
	mockRepo := new(mocks.PengaturanRepositoryMock)

	svc := NewPengaturanService(mockRepo)

	pengaturan, err := svc.UpdatePengaturan("8:00", "16:00") // invalid format

	assert.Error(t, err)
	assert.Nil(t, pengaturan)
	assert.Equal(t, "format jam masuk tidak valid (gunakan HH:mm, contoh: 07:00)", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdatePengaturan_InvalidJamPulang(t *testing.T) {
	mockRepo := new(mocks.PengaturanRepositoryMock)

	svc := NewPengaturanService(mockRepo)

	pengaturan, err := svc.UpdatePengaturan("08:00", "46:00") // invalid time

	assert.Error(t, err)
	assert.Nil(t, pengaturan)
	assert.Equal(t, "format jam pulang tidak valid (gunakan HH:mm, contoh: 18:00)", err.Error())
	mockRepo.AssertNotCalled(t, "Update")
}

func TestUpdatePengaturan_DBError(t *testing.T) {
	mockRepo := new(mocks.PengaturanRepositoryMock)
	mockRepo.On("Update", mock.AnythingOfType("*domain.Pengaturan")).Return(errors.New("db error"))

	svc := NewPengaturanService(mockRepo)

	pengaturan, err := svc.UpdatePengaturan("08:00", "16:00")

	assert.Error(t, err)
	assert.Nil(t, pengaturan)
	assert.Equal(t, "db error", err.Error())
	mockRepo.AssertExpectations(t)
}
