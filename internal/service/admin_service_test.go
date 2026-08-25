package service_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/repository/mocks"
	"laporanharianapi/internal/service"
)

func TestAdminService_Pegawai(t *testing.T) {
	dummyUser := &domain.User{
		ID:       1,
		NIP:      "198501012010011001",
		Nama:     "Pegawai Baru",
		Password: "password123",
		Role:     "staf",
	}

	t.Run("CreatePegawaiAdmin - Success", func(t *testing.T) {
		mockAdminRepo := new(mocks.AdminRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

		mockUserRepo.On("FindByNIP", "198501012010011001").Return(nil, errors.New("not found"))
		mockUserRepo.On("Create", mock.AnythingOfType("*domain.User")).Return(nil)

		adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)
		newUser := *dummyUser
		err := adminService.CreatePegawaiAdmin(&newUser)

		assert.NoError(t, err)
		assert.NoError(t, bcrypt.CompareHashAndPassword([]byte(newUser.Password), []byte("password123")))
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("CreatePegawaiAdmin - Duplicate NIP", func(t *testing.T) {
		mockAdminRepo := new(mocks.AdminRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

		mockUserRepo.On("FindByNIP", "198501012010011001").Return(dummyUser, nil)

		adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)
		err := adminService.CreatePegawaiAdmin(dummyUser)

		assert.Error(t, err)
		assert.Equal(t, "NIP sudah terdaftar", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("UpdatePegawaiAdmin - User Not Found", func(t *testing.T) {
		mockAdminRepo := new(mocks.AdminRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

		mockUserRepo.On("FindByID", uint(99)).Return(nil, errors.New("not found"))

		adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)
		err := adminService.UpdatePegawaiAdmin(99, dummyUser)

		assert.Error(t, err)
		assert.Equal(t, "pegawai tidak ditemukan", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("ResetPasswordPegawaiAdmin - Length Less Than 8", func(t *testing.T) {
		mockAdminRepo := new(mocks.AdminRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

		adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)
		err := adminService.ResetPasswordPegawaiAdmin(1, "123")

		assert.Error(t, err)
		assert.Equal(t, "password baru minimal 8 karakter", err.Error())
	})

	t.Run("ResetPasswordPegawaiAdmin - Success", func(t *testing.T) {
		mockAdminRepo := new(mocks.AdminRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

		mockUserRepo.On("FindByID", uint(1)).Return(dummyUser, nil)
		mockUserRepo.On("UpdatePassword", uint(1), mock.AnythingOfType("string")).Return(nil)

		adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)
		err := adminService.ResetPasswordPegawaiAdmin(1, "newpassword123")

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("DeletePegawaiAdmin - Success", func(t *testing.T) {
		mockAdminRepo := new(mocks.AdminRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

		mockUserRepo.On("FindByID", uint(1)).Return(dummyUser, nil)
		mockUserRepo.On("DeleteWithCleanup", uint(1)).Return([]string{}, nil)

		adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)
		err := adminService.DeletePegawaiAdmin(1)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAdminService_SupervisorLurah(t *testing.T) {
	mockAdminRepo := new(mocks.AdminRepositoryMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

	adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)

	t.Run("UpdateSupervisorLurahAdmin - Empty Validation", func(t *testing.T) {
		err := adminService.UpdateSupervisorLurahAdmin("", "")
		assert.Error(t, err)
		assert.Equal(t, "nama dan NIP atasan tidak boleh kosong", err.Error())
	})

	t.Run("UpdateSupervisorLurahAdmin - Success", func(t *testing.T) {
		mockSupervisorRepo.On("UpdateSupervisor", "Camat Garut", "197001011990011001").Return(nil)

		err := adminService.UpdateSupervisorLurahAdmin("Camat Garut", "197001011990011001")
		assert.NoError(t, err)
		mockSupervisorRepo.AssertExpectations(t)
	})

	t.Run("GetSupervisorLurahAdmin - Success", func(t *testing.T) {
		expected := &domain.LurahSupervisor{
			ID:   1,
			Nama: "Camat Garut",
			NIP:  "197001011990011001",
		}
		mockSupervisorRepo.On("GetSupervisor").Return(expected, nil)

		res, err := adminService.GetSupervisorLurahAdmin()
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		mockSupervisorRepo.AssertExpectations(t)
	})
}

func TestAdminService_ReportsAndDashboard(t *testing.T) {
	mockAdminRepo := new(mocks.AdminRepositoryMock)
	mockUserRepo := new(mocks.UserRepositoryMock)
	mockSupervisorRepo := new(mocks.SupervisorRepositoryMock)

	adminService := service.NewAdminService(mockAdminRepo, mockUserRepo, mockSupervisorRepo)

	t.Run("GetRekapLaporanAdmin", func(t *testing.T) {
		filter := repository.AdminReportFilter{Page: 1, Limit: 10}
		expected := &repository.AdminReportResponse{
			Data:      []domain.Laporan{{ID: 1, JudulKegiatan: "Laporan 1"}},
			TotalData: 1,
		}
		mockAdminRepo.On("GetRekapLaporanAdmin", filter).Return(expected, nil)

		res, err := adminService.GetRekapLaporanAdmin(filter)
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		mockAdminRepo.AssertExpectations(t)
	})

	t.Run("GetDashboardSummaryAdmin", func(t *testing.T) {
		expected := &repository.DashboardSummaryResponse{
			Statistik: repository.StatistikDashboard{
				TotalPegawai: 15,
			},
		}
		mockAdminRepo.On("GetDashboardSummaryAdmin").Return(expected, nil)

		res, err := adminService.GetDashboardSummaryAdmin()
		assert.NoError(t, err)
		assert.Equal(t, expected, res)
		mockAdminRepo.AssertExpectations(t)
	})
}
