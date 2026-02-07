package service

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"laporanharianapi/internal/repository"
)

// AuthService adalah interface untuk operasi autentikasi.
type AuthService interface {
	Login(nip string, password string) (string, error)
}

// authService adalah implementasi dari AuthService.
type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService membuat instance baru AuthService.
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

// Login memvalidasi kredensial dan mengembalikan JWT token jika berhasil.
func (s *authService) Login(nip string, password string) (string, error) {
	// 1. Cari user berdasarkan NIP
	user, err := s.userRepo.FindByNIP(nip)
	if err != nil {
		return "", errors.New("NIP atau password salah")
	}

	// 2. Verifikasi password dengan bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", errors.New("NIP atau password salah")
	}

	// 3. Generate JWT Token
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", errors.New("JWT_SECRET tidak dikonfigurasi")
	}

	// 4. Buat claims dengan data user
	claims := jwt.MapClaims{
		"user_id":    user.ID,
		"role":       user.Role,
		"jabatan_id": user.JabatanID,
		"exp":        time.Now().Add(24 * time.Hour).Unix(), // Token berlaku 24 jam
		"iat":        time.Now().Unix(),
	}

	// 5. Buat token dengan algoritma HS256
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 6. Sign token dengan secret key
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", errors.New("gagal membuat token")
	}

	return tokenString, nil
}
