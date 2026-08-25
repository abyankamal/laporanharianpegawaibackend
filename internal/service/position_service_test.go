package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
	"laporanharianapi/internal/service"
)

func TestJabatanService(t *testing.T) {
	dummyJabatan := &domain.RefJabatan{
		ID:          1,
		NamaJabatan: "Pranata Komputer",
	}

	t.Run("GetAllJabatan - Success", func(t *testing.T) {
		mockRepo := new(mocks.JabatanRepositoryMock)
		mockRepo.On("GetAll").Return([]domain.RefJabatan{*dummyJabatan}, nil)

		svc := service.NewJabatanService(mockRepo)
		res, err := svc.GetAllJabatan()

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "Pranata Komputer", res[0].NamaJabatan)
		mockRepo.AssertExpectations(t)
	})

	t.Run("GetJabatanByID - Success", func(t *testing.T) {
		mockRepo := new(mocks.JabatanRepositoryMock)
		mockRepo.On("GetByID", uint(1)).Return(dummyJabatan, nil)

		svc := service.NewJabatanService(mockRepo)
		res, err := svc.GetJabatanByID(1)

		assert.NoError(t, err)
		assert.Equal(t, dummyJabatan, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("CreateJabatan - Success", func(t *testing.T) {
		mockRepo := new(mocks.JabatanRepositoryMock)
		mockRepo.On("Create", mock.AnythingOfType("*domain.RefJabatan")).Return(nil)

		svc := service.NewJabatanService(mockRepo)
		res, err := svc.CreateJabatan("Pranata Komputer")

		assert.NoError(t, err)
		assert.Equal(t, "Pranata Komputer", res.NamaJabatan)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateJabatan - Not Found", func(t *testing.T) {
		mockRepo := new(mocks.JabatanRepositoryMock)
		mockRepo.On("GetByID", uint(99)).Return(nil, errors.New("not found"))

		svc := service.NewJabatanService(mockRepo)
		res, err := svc.UpdateJabatan(99, "Updated")

		assert.Error(t, err)
		assert.Nil(t, res)
		mockRepo.AssertExpectations(t)
	})

	t.Run("UpdateJabatan - Success", func(t *testing.T) {
		mockRepo := new(mocks.JabatanRepositoryMock)
		mockRepo.On("GetByID", uint(1)).Return(dummyJabatan, nil)
		mockRepo.On("Update", mock.AnythingOfType("*domain.RefJabatan")).Return(nil)

		svc := service.NewJabatanService(mockRepo)
		res, err := svc.UpdateJabatan(1, "Updated Name")

		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", res.NamaJabatan)
		mockRepo.AssertExpectations(t)
	})

	t.Run("DeleteJabatan - Success", func(t *testing.T) {
		mockRepo := new(mocks.JabatanRepositoryMock)
		mockRepo.On("Delete", uint(1)).Return(nil)

		svc := service.NewJabatanService(mockRepo)
		err := svc.DeleteJabatan(1)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})
}
