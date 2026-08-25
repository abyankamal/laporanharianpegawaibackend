package service_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
	"laporanharianapi/internal/service"
)

func TestAuthService_Login(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret123")
	defer os.Unsetenv("JWT_SECRET")

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)
	dummyJabatanID := uint(3)

	dummyUser := &domain.User{
		ID:        1,
		NIP:       "198501012010011001",
		Nama:      "Budi Santoso",
		Password:  string(hashedPassword),
		Role:      "staf",
		JabatanID: &dummyJabatanID,
	}

	t.Run("Success Login", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockUserRepo.On("FindByNIP", "198501012010011001").Return(dummyUser, nil)

		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.Login("198501012010011001", "secret123")

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res["access_token"])
		assert.NotEmpty(t, res["refresh_token"])
		assert.Equal(t, 86400, res["expires_in"])
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockUserRepo.On("FindByNIP", "999999").Return(nil, errors.New("not found"))

		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.Login("999999", "secret123")

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "NIP atau password salah", err.Error())
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Wrong Password", func(t *testing.T) {
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockUserRepo.On("FindByNIP", "198501012010011001").Return(dummyUser, nil)

		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.Login("198501012010011001", "wrongpassword")

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "NIP atau password salah", err.Error())
		mockUserRepo.AssertExpectations(t)
	})
}

func TestAuthService_RefreshToken(t *testing.T) {
	secret := "testsecret123"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	dummyJabatanID := uint(3)
	dummyUser := &domain.User{
		ID:        1,
		NIP:       "198501012010011001",
		Nama:      "Budi Santoso",
		Role:      "staf",
		JabatanID: &dummyJabatanID,
	}

	// Helper to generate test tokens
	generateTestToken := func(userID uint, tokenType string, expDuration time.Duration, signingKey string) string {
		claims := jwt.MapClaims{
			"user_id":    userID,
			"role":       "staf",
			"jabatan_id": dummyJabatanID,
			"token_type": tokenType,
			"exp":        time.Now().Add(expDuration).Unix(),
			"iat":        time.Now().Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenStr, _ := token.SignedString([]byte(signingKey))
		return tokenStr
	}

	t.Run("Success Refresh Token", func(t *testing.T) {
		refreshToken := generateTestToken(1, "refresh", 30*24*time.Hour, secret)

		mockUserRepo := new(mocks.UserRepositoryMock)
		mockUserRepo.On("FindByID", uint(1)).Return(dummyUser, nil)

		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.RefreshToken(refreshToken)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res["access_token"])
		assert.NotEmpty(t, res["refresh_token"])
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Invalid Token - Signed with Wrong Secret", func(t *testing.T) {
		refreshToken := generateTestToken(1, "refresh", 30*24*time.Hour, "wrong_secret")

		mockUserRepo := new(mocks.UserRepositoryMock)
		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.RefreshToken(refreshToken)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "refresh token tidak valid atau sudah kadaluarsa", err.Error())
	})

	t.Run("Invalid Token - Wrong token_type (Access Token supplied)", func(t *testing.T) {
		accessToken := generateTestToken(1, "access", 24*time.Hour, secret)

		mockUserRepo := new(mocks.UserRepositoryMock)
		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.RefreshToken(accessToken)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "bukan tipe refresh token", err.Error())
	})

	t.Run("User Not Found or Inactive", func(t *testing.T) {
		refreshToken := generateTestToken(99, "refresh", 30*24*time.Hour, secret)

		mockUserRepo := new(mocks.UserRepositoryMock)
		mockUserRepo.On("FindByID", uint(99)).Return(nil, errors.New("user not found"))

		authService := service.NewAuthService(mockUserRepo)
		res, err := authService.RefreshToken(refreshToken)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "user tidak ditemukan atau akun tidak aktif", err.Error())
		mockUserRepo.AssertExpectations(t)
	})
}
