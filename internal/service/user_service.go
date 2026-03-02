package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// CreateUserRequest adalah DTO untuk request pembuatan user baru.
type CreateUserRequest struct {
	NIP          string `json:"nip"`
	Nama         string `json:"nama"`
	Password     string `json:"password"`
	Role         string `json:"role"`
	JabatanID    *uint  `json:"jabatan_id"`
	SupervisorID *uint  `json:"supervisor_id"`
}

// UpdateUserRequest adalah DTO untuk request update user.
type UpdateUserRequest struct {
	NIP          string `json:"nip"`
	Nama         string `json:"nama"`
	Password     string `json:"password"` // Optional, jika kosong password tidak diupdate
	Role         string `json:"role"`
	JabatanID    *uint  `json:"jabatan_id"`
	SupervisorID *uint  `json:"supervisor_id"`
}

// ChangePasswordRequest adalah DTO untuk request ubah password.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

// UserService adalah interface untuk operasi bisnis User.
type UserService interface {
	GetAllUsers() ([]domain.User, error)
	GetUserByID(id uint) (*domain.User, error)
	CreateUser(req CreateUserRequest) (*domain.User, error)
	UpdateUser(id uint, req UpdateUserRequest) (*domain.User, error)
	DeleteUser(id uint) error
	ChangePassword(userID uint, req ChangePasswordRequest) error
	UpdateProfilePhoto(userID uint, fileHeader *multipart.FileHeader) (string, error)
	UpdateFCMToken(userID uint, token string) error
	GetSupervisors(roleFilter string) ([]domain.User, error)
	GetUsersByRoles(roles []string) ([]domain.User, error)
}

// userService adalah implementasi dari UserService.
type userService struct {
	userRepo repository.UserRepository
}

// NewUserService membuat instance baru UserService.
func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

// GetAllUsers mengambil semua user.
func (s *userService) GetAllUsers() ([]domain.User, error) {
	return s.userRepo.FindAll()
}

// GetUsersByRoles mengambil user berdasarkan roles.
func (s *userService) GetUsersByRoles(roles []string) ([]domain.User, error) {
	return s.userRepo.FindByRoles(roles)
}

// GetUserByID mengambil user berdasarkan ID.
func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}
	return user, nil
}

// CreateUser membuat user baru.
func (s *userService) CreateUser(req CreateUserRequest) (*domain.User, error) {
	// Validasi input
	if req.NIP == "" {
		return nil, errors.New("NIP wajib diisi")
	}
	if req.Nama == "" {
		return nil, errors.New("nama wajib diisi")
	}
	if req.Password == "" {
		return nil, errors.New("password wajib diisi")
	}
	if req.Role == "" {
		return nil, errors.New("role wajib diisi")
	}

	// Normalisasi role agar sesuai dengan sistem (sekertaris)
	role := strings.ToLower(req.Role)
	if role == "sekretaris" {
		role = "sekertaris"
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	// Buat domain User
	user := &domain.User{
		NIP:          req.NIP,
		Nama:         req.Nama,
		Password:     string(hashedPassword),
		Role:         role,
		JabatanID:    req.JabatanID,
		SupervisorID: req.SupervisorID,
		CreatedAt:    time.Now(),
	}

	// Simpan ke database
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, errors.New("gagal menyimpan user, pastikan NIP belum digunakan")
	}

	return user, nil
}

// UpdateUser mengupdate user berdasarkan ID.
func (s *userService) UpdateUser(id uint, req UpdateUserRequest) (*domain.User, error) {
	// Cek apakah user ada
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	// Update field yang diisi
	if req.NIP != "" {
		user.NIP = req.NIP
	}
	if req.Nama != "" {
		user.Nama = req.Nama
	}
	if req.Role != "" {
		role := strings.ToLower(req.Role)
		if role == "sekretaris" {
			role = "sekertaris"
		}
		user.Role = role
	}

	// Update JabatanID (bisa null)
	user.JabatanID = req.JabatanID

	// Update SupervisorID (bisa null)
	user.SupervisorID = req.SupervisorID

	// Jika password diisi, hash password baru
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("gagal mengenkripsi password")
		}
		user.Password = string(hashedPassword)
	}

	// Update ke database
	err = s.userRepo.Update(user)
	if err != nil {
		return nil, errors.New("gagal mengupdate user")
	}

	return user, nil
}

// DeleteUser menghapus user berdasarkan ID beserta data terkait dan file fisik.
func (s *userService) DeleteUser(id uint) error {
	// Cek apakah user ada
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// Hapus user dan data terkait di database
	filePaths, err := s.userRepo.DeleteWithCleanup(id)
	if err != nil {
		return errors.New("gagal menghapus user dan data terkait")
	}

	// Hapus file fisik dari disk
	for _, path := range filePaths {
		if path != "" {
			// Pastikan path menggunakan separator yang benar untuk OS
			os.Remove(filepath.FromSlash(path))
		}
	}

	return nil
}

// ChangePassword mengubah password user dengan validasi old password.
func (s *userService) ChangePassword(userID uint, req ChangePasswordRequest) error {
	// 1. Validasi input
	if req.OldPassword == "" {
		return errors.New("password lama wajib diisi")
	}
	if req.NewPassword == "" {
		return errors.New("password baru wajib diisi")
	}
	if len(req.NewPassword) < 8 {
		return errors.New("password baru minimal 8 karakter")
	}

	// 2. Ambil data user dari database
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// 3. Verifikasi password lama dengan bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return errors.New("password lama tidak sesuai")
	}

	// 4. Validasi: password baru tidak boleh sama dengan password lama
	if req.OldPassword == req.NewPassword {
		return errors.New("password baru tidak boleh sama dengan password lama")
	}

	// 5. Hash password baru
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengenkripsi password")
	}

	// 6. Update password di database (hanya field password)
	err = s.userRepo.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		return errors.New("gagal mengubah password")
	}

	return nil
}

// UpdateProfilePhoto mengubah foto profil user.
func (s *userService) UpdateProfilePhoto(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	// 1. Validasi tipe file (hanya jpg/jpeg/png)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return "", errors.New("format file tidak didukung, gunakan JPG/JPEG/PNG")
	}

	// 2. Validasi ukuran file (max 2MB)
	if fileHeader.Size > 2*1024*1024 {
		return "", errors.New("ukuran file maksimal 2MB")
	}

	// 3. Ambil data user untuk cek foto lama
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	// 4. Hapus foto lama jika ada
	if user.FotoPath != nil && *user.FotoPath != "" {
		os.Remove(*user.FotoPath)
	}

	// 5. Simpan file baru ke ./uploads/photos/
	uploadDir := "./uploads/photos"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("gagal membuat direktori upload: %v", err)
	}

	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %v", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("gagal menulis file: %v", err)
	}

	// 6. Update foto_path di database
	err = s.userRepo.UpdateFoto(userID, destPath)
	if err != nil {
		// Hapus file yang baru diupload jika gagal update DB
		os.Remove(destPath)
		return "", errors.New("gagal mengupdate foto profil")
	}

	return destPath, nil
}

// GetSupervisors mengambil daftar atasan secara dinamis berdasarkan query parameter roleFilter.
func (s *userService) GetSupervisors(roleFilter string) ([]domain.User, error) {
	supervisors, err := s.userRepo.FindSupervisors(roleFilter)
	if err != nil {
		return nil, err
	}
	if len(supervisors) == 0 {
		return nil, errors.New("data atasan tidak ditemukan")
	}
	return supervisors, nil
}

// UpdateFCMToken memperbarui token FCM untuk pengguna tertentu
func (s *userService) UpdateFCMToken(userID uint, token string) error {
	// Validasi token tidak boleh kosong
	if token == "" {
		return errors.New("fcm token tidak boleh kosong")
	}

	// Update ke database
	err := s.userRepo.UpdateFCMToken(userID, token)
	if err != nil {
		return errors.New("gagal mengupdate fcm token")
	}

	return nil
}
