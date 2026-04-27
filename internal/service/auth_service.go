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
	Login(nip string, password string) (map[string]interface{}, error)
	RefreshToken(refreshToken string) (map[string]interface{}, error)
}

// authService adalah implementasi dari AuthService.
type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService membuat instance baru AuthService.
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{userRepo: userRepo}
}

// Login memvalidasi kredensial dan mengembalikan access & refresh token jika berhasil.
func (s *authService) Login(nip string, password string) (map[string]interface{}, error) {
	// 1. Cari user berdasarkan NIP
	user, err := s.userRepo.FindByNIP(nip)
	if err != nil {
		return nil, errors.New("NIP atau password salah")
	}

	// 2. Verifikasi password dengan bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("NIP atau password salah")
	}

	// 3. Generate Token Ganda (Access & Refresh)
	return s.generateTokenPair(user.ID, user.Role, user.JabatanID)
}

// RefreshToken memvalidasi refresh token dan memberikan pasangan token baru.
func (s *authService) RefreshToken(refreshTokenString string) (map[string]interface{}, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET tidak dikonfigurasi")
	}

	// 1. Parse dan validasi refresh token
	token, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("metode signing tidak valid")
		}
		return []byte(jwtSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("refresh token tidak valid atau sudah kadaluarsa")
	}

	// 2. Cek klaim token
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("gagal membaca claims token")
	}

	// 3. Pastikan ini benar-benar refresh token (bukan access token yang dicoba-pakai)
	if claims["token_type"] != "refresh" {
		return nil, errors.New("bukan tipe refresh token")
	}

	// 4. Ambil data user dari claims
	userID := uint(claims["user_id"].(float64))
	role := claims["role"].(string)
	
	var jabatanID *uint
	if claims["jabatan_id"] != nil {
		jID := uint(claims["jabatan_id"].(float64))
		jabatanID = &jID
	}

	// 5. Generate pasangan token baru (Refresh Token Rotation)
	return s.generateTokenPair(userID, role, jabatanID)
}

// generateTokenPair adalah helper untuk membuat Access Token (1 jam) dan Refresh Token (30 hari).
func (s *authService) generateTokenPair(userID uint, role string, jabatanID *uint) (map[string]interface{}, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, errors.New("JWT_SECRET tidak dikonfigurasi")
	}

	now := time.Now()

	// --- 1. Access Token (Umur Pendek: 1 Jam) ---
	accessClaims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		"jabatan_id": jabatanID,
		"token_type": "access", // Penanda akses biasa
		"exp":        now.Add(1 * time.Hour).Unix(),
		"iat":        now.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, errors.New("gagal membuat access token")
	}

	// --- 2. Refresh Token (Umur Panjang: 30 Hari) ---
	refreshClaims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		"jabatan_id": jabatanID,
		"token_type": "refresh", // Penanda khusus refresh
		"exp":        now.Add(30 * 24 * time.Hour).Unix(),
		"iat":        now.Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(jwtSecret))
	if err != nil {
		return nil, errors.New("gagal membuat refresh token")
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    3600, // 1 jam dalam detik
	}, nil
}
