package service

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
)

// ============================================================
// Test Login (AuthService)
// ============================================================

func TestLogin_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)

	// Set JWT_SECRET agar tidak error
	os.Setenv("JWT_SECRET", "test-secret-key-for-unit-test")
	defer os.Unsetenv("JWT_SECRET")

	// Hash password untuk simulasi data di database (password seeder: 123456)
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	expectedUser := &domain.User{
		ID:       1,
		NIP:      "198106152014102004",
		Nama:     "Iis Yuniawardani, S.IP",
		Password: string(hashedPassword),
		Role:     "lurah",
	}

	// Mock: FindByNIP harus dipanggil dengan NIP yang benar dan mengembalikan user
	mockUserRepo.On("FindByNIP", "198106152014102004").Return(expectedUser, nil)

	// Buat service
	authSvc := NewAuthService(mockUserRepo)

	// Execute
	token, err := authSvc.Login("198106152014102004", "123456")

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, token, "Token JWT harus dikembalikan")
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Fail_NIPNotFound(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)

	// Mock: FindByNIP mengembalikan error (user tidak ditemukan)
	mockUserRepo.On("FindByNIP", "000000000").Return(nil, errors.New("record not found"))

	// Buat service
	authSvc := NewAuthService(mockUserRepo)

	// Execute
	token, err := authSvc.Login("000000000", "123456")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "NIP atau password salah", err.Error())
	mockUserRepo.AssertExpectations(t)
}

func TestLogin_Fail_WrongPassword(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)

	expectedUser := &domain.User{
		ID:       1,
		NIP:      "198106152014102004",
		Nama:     "Iis Yuniawardani, S.IP",
		Password: string(hashedPassword),
		Role:     "lurah",
	}

	// Mock: FindByNIP berhasil menemukan user
	mockUserRepo.On("FindByNIP", "198106152014102004").Return(expectedUser, nil)

	// Buat service
	authSvc := NewAuthService(mockUserRepo)

	// Execute: login dengan password yang salah (password seeder: 123456)
	token, err := authSvc.Login("198106152014102004", "wrongpassword")

	// Assert
	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Equal(t, "NIP atau password salah", err.Error())
	mockUserRepo.AssertExpectations(t)
}

// ============================================================
// Test ChangePassword (UserService)
// ============================================================

func TestChangePassword_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword123"), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:       1,
		NIP:      "198106152014102004",
		Nama:     "Iis Yuniawardani, S.IP",
		Password: string(hashedPassword),
		Role:     "lurah",
	}

	// Mock: FindByID mengembalikan user yang ada
	mockUserRepo.On("FindByID", uint(1)).Return(existingUser, nil)
	// Mock: UpdatePassword berhasil
	mockUserRepo.On("UpdatePassword", uint(1), mock.AnythingOfType("string")).Return(nil)

	// Buat service
	userSvc := NewUserService(mockUserRepo)

	// Execute
	req := ChangePasswordRequest{
		OldPassword: "oldpassword123",
		NewPassword: "newpassword456",
	}
	err := userSvc.ChangePassword(1, req)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestChangePassword_Fail_WrongOldPassword(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("oldpassword123"), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:       1,
		NIP:      "198106152014102004",
		Nama:     "Iis Yuniawardani, S.IP",
		Password: string(hashedPassword),
		Role:     "lurah",
	}

	// Mock: FindByID mengembalikan user yang ada
	mockUserRepo.On("FindByID", uint(1)).Return(existingUser, nil)

	// Buat service
	userSvc := NewUserService(mockUserRepo)

	// Execute: kirim old password yang salah
	req := ChangePasswordRequest{
		OldPassword: "wrongoldpassword",
		NewPassword: "newpassword456",
	}
	err := userSvc.ChangePassword(1, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "password lama tidak sesuai", err.Error())
	mockUserRepo.AssertExpectations(t)
	// UpdatePassword TIDAK boleh dipanggil karena old password salah
	mockUserRepo.AssertNotCalled(t, "UpdatePassword")
}

func TestChangePassword_Fail_EmptyOldPassword(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)
	userSvc := NewUserService(mockUserRepo)

	// Execute: old password kosong
	req := ChangePasswordRequest{
		OldPassword: "",
		NewPassword: "newpassword456",
	}
	err := userSvc.ChangePassword(1, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "password lama wajib diisi", err.Error())
}

func TestChangePassword_Fail_NewPasswordTooShort(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)
	userSvc := NewUserService(mockUserRepo)

	// Execute: password baru kurang dari 8 karakter
	req := ChangePasswordRequest{
		OldPassword: "oldpassword123",
		NewPassword: "short",
	}
	err := userSvc.ChangePassword(1, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "password baru minimal 8 karakter", err.Error())
}

func TestChangePassword_Fail_SameAsOldPassword(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("samepassword123"), bcrypt.DefaultCost)

	existingUser := &domain.User{
		ID:       1,
		Password: string(hashedPassword),
	}

	mockUserRepo.On("FindByID", uint(1)).Return(existingUser, nil)
	userSvc := NewUserService(mockUserRepo)

	// Execute: password baru sama dengan password lama
	req := ChangePasswordRequest{
		OldPassword: "samepassword123",
		NewPassword: "samepassword123",
	}
	err := userSvc.ChangePassword(1, req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "password baru tidak boleh sama dengan password lama", err.Error())
}

// ============================================================
// Test DeleteUser (UserService)
// ============================================================

func TestDeleteUser_Success(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)
	userSvc := NewUserService(mockUserRepo)

	existingUser := &domain.User{ID: 1, NIP: "123"}
	mockUserRepo.On("FindByID", uint(1)).Return(existingUser, nil)

	// Mock returning related file paths
	filePaths := []string{"uploads/photos/test.jpg", "uploads/reports/doc.pdf"}
	mockUserRepo.On("DeleteWithCleanup", uint(1)).Return(filePaths, nil)

	// Execute
	err := userSvc.DeleteUser(1)

	// Assert
	assert.NoError(t, err)
	mockUserRepo.AssertExpectations(t)
}

func TestDeleteUser_Fail_UserNotFound(t *testing.T) {
	// Setup
	mockUserRepo := new(mocks.UserRepositoryMock)
	userSvc := NewUserService(mockUserRepo)

	mockUserRepo.On("FindByID", uint(1)).Return(nil, errors.New("not found"))

	// Execute
	err := userSvc.DeleteUser(1)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "user tidak ditemukan", err.Error())
	mockUserRepo.AssertNotCalled(t, "DeleteWithCleanup")
}
