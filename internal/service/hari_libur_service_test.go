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

func TestGetHariLibur_Success(t *testing.T) {
	mockRepo := new(mocks.HariLiburRepositoryMock)
	mockTime, _ := time.Parse("2006-01-02", "2026-08-17")

	expectedList := []domain.HariLibur{
		{ID: 1, Tanggal: mockTime, Keterangan: "Hari Kemerdekaan"},
	}

	mockRepo.On("GetAll").Return(expectedList, nil)

	svc := NewHariLiburService(mockRepo)

	liburs, err := svc.GetHariLibur()

	assert.NoError(t, err)
	assert.NotNil(t, liburs)
	assert.Len(t, liburs, 1)
	assert.Equal(t, "Hari Kemerdekaan", liburs[0].Keterangan)
	mockRepo.AssertExpectations(t)
}

func TestCreateHariLibur_Success(t *testing.T) {
	mockRepo := new(mocks.HariLiburRepositoryMock)
	mockRepo.On("Create", mock.AnythingOfType("*domain.HariLibur")).Return(nil)

	svc := NewHariLiburService(mockRepo)

	libur, err := svc.CreateHariLibur("2026-12-25", "Hari Natal")

	assert.NoError(t, err)
	assert.NotNil(t, libur)
	assert.Equal(t, "Hari Natal", libur.Keterangan)
	mockRepo.AssertExpectations(t)
}

func TestCreateHariLibur_EmptyKeterangan(t *testing.T) {
	mockRepo := new(mocks.HariLiburRepositoryMock)
	svc := NewHariLiburService(mockRepo)

	libur, err := svc.CreateHariLibur("2026-12-25", "")

	assert.Error(t, err)
	assert.Nil(t, libur)
	assert.Equal(t, "keterangan hari libur wajib diisi", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestCreateHariLibur_InvalidDate(t *testing.T) {
	mockRepo := new(mocks.HariLiburRepositoryMock)
	svc := NewHariLiburService(mockRepo)

	libur, err := svc.CreateHariLibur("25-12-2026", "Salah Format")

	assert.Error(t, err)
	assert.Nil(t, libur)
	assert.Equal(t, "format tanggal tidak valid (gunakan YYYY-MM-DD, contoh: 2026-08-17)", err.Error())
	mockRepo.AssertNotCalled(t, "Create")
}

func TestDeleteHariLibur_Success(t *testing.T) {
	mockRepo := new(mocks.HariLiburRepositoryMock)
	mockRepo.On("Delete", uint(1)).Return(nil)

	svc := NewHariLiburService(mockRepo)

	err := svc.DeleteHariLibur(1)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestDeleteHariLibur_Error(t *testing.T) {
	mockRepo := new(mocks.HariLiburRepositoryMock)
	mockRepo.On("Delete", uint(1)).Return(errors.New("not found"))

	svc := NewHariLiburService(mockRepo)

	err := svc.DeleteHariLibur(1)

	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
	mockRepo.AssertExpectations(t)
}
