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
	ResetPasswordByAdmin(targetUserID uint, newPassword string) error
	UpdateProfilePhoto(userID uint, fileHeader *multipart.FileHeader) (string, error)
	UpdateFCMToken(userID uint, token string) error
	GetSupervisors(roleFilter string) ([]domain.User, error)
	GetUsersByRoles(roles []string) ([]domain.User, error)
}

// userService adalah implementasi dari UserService.
type userService struct {
	userRepo       repository.UserRepository
	supervisorRepo repository.SupervisorRepository
}

// NewUserService membuat instance baru UserService.
func NewUserService(userRepo repository.UserRepository, supervisorRepo repository.SupervisorRepository) UserService {
	return &userService{
		userRepo:       userRepo,
		supervisorRepo: supervisorRepo,
	}
}

// GetAllUsers mengambil semua user.
func (s *userService) GetAllUsers() ([]domain.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		return nil, err
	}
	for i := range users {
		s.fillLurahSupervisor(&users[i])
	}
	return users, nil
}

// GetUsersByRoles mengambil user berdasarkan roles.
func (s *userService) GetUsersByRoles(roles []string) ([]domain.User, error) {
	users, err := s.userRepo.FindByRoles(roles)
	if err != nil {
		return nil, err
	}
	for i := range users {
		s.fillLurahSupervisor(&users[i])
	}
	return users, nil
}

// GetUserByID mengambil user berdasarkan ID.
func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}
	s.fillLurahSupervisor(user)
	return user, nil
}

func (s *userService) fillLurahSupervisor(user *domain.User) {
	if user != nil && (strings.ToLower(user.Role) == "lurah" || (user.Jabatan != nil && strings.ToLower(user.Jabatan.NamaJabatan) == "lurah")) {
		if user.Supervisor == nil {
			if s.supervisorRepo != nil {
				supervisorData, err := s.supervisorRepo.GetSupervisor()
				if err == nil && supervisorData != nil {
					user.Supervisor = &domain.User{
						Nama: supervisorData.Nama,
						NIP:  supervisorData.NIP,
					}
					return
				}
			}
			user.Supervisor = &domain.User{
				Nama: "Atasan Lurah",
				NIP:  "-",
			}
		}
	}
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
	if len(req.Password) < 8 {
		return nil, errors.New("password minimal 8 karakter")
	}
	if req.Role == "" {
		return nil, errors.New("role wajib diisi")
	}

	// Cek apakah NIP sudah terdaftar
	existingUser, _ := s.userRepo.FindByNIP(req.NIP)
	if existingUser != nil {
		return nil, errors.New("NIP sudah terdaftar")
	}

	// Hash password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("gagal mengenkripsi password")
	}

	// Buat objek user
	user := &domain.User{
		NIP:          req.NIP,
		Nama:         req.Nama,
		Password:     string(hashedPassword),
		Role:         req.Role,
		JabatanID:    req.JabatanID,
		SupervisorID: req.SupervisorID,
		CreatedAt:    time.Now(),
	}

	// Simpan ke database
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, errors.New("gagal membuat user")
	}

	// Return user yang baru dibuat
	createdUser, err := s.userRepo.FindByID(user.ID)
	if err != nil {
		return user, nil
	}

	return createdUser, nil
}

// UpdateUser memperbarui data user.
func (s *userService) UpdateUser(id uint, req UpdateUserRequest) (*domain.User, error) {
	// Cek apakah user ada
	existingUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user tidak ditemukan")
	}

	// Validasi NIP unik jika diubah
	if req.NIP != "" && req.NIP != existingUser.NIP {
		userWithNIP, _ := s.userRepo.FindByNIP(req.NIP)
		if userWithNIP != nil {
			return nil, errors.New("NIP sudah digunakan oleh user lain")
		}
		existingUser.NIP = req.NIP
	}

	// Update fields jika disediakan
	if req.Nama != "" {
		existingUser.Nama = req.Nama
	}
	if req.Role != "" {
		existingUser.Role = req.Role
	}
	if req.JabatanID != nil {
		existingUser.JabatanID = req.JabatanID
	}
	if req.SupervisorID != nil {
		existingUser.SupervisorID = req.SupervisorID
	}

	// Update password jika disediakan
	if req.Password != "" {
		if len(req.Password) < 8 {
			return nil, errors.New("password minimal 8 karakter")
		}
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("gagal mengenkripsi password")
		}
		existingUser.Password = string(hashedPassword)
	}

	// Simpan perubahan ke database
	err = s.userRepo.Update(existingUser)
	if err != nil {
		return nil, errors.New("gagal memperbarui data user")
	}

	// Return user yang sudah diupdate
	updatedUser, err := s.userRepo.FindByID(id)
	if err != nil {
		return existingUser, nil
	}

	return updatedUser, nil
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
	// 1. Validasi input tidak boleh kosong
	if req.OldPassword == "" {
		return errors.New("password lama wajib diisi")
	}
	if req.NewPassword == "" {
		return errors.New("password baru wajib diisi")
	}

	// 2. Validasi panjang password baru minimal 8 karakter
	if len(req.NewPassword) < 8 {
		return errors.New("password baru minimal 8 karakter")
	}

	// 3. Ambil data user dari database
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// 4. Verifikasi old_password cocok dengan hash di database
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		return errors.New("password lama tidak sesuai")
	}

	// Validasi password baru tidak boleh sama dengan password lama
	if req.OldPassword == req.NewPassword {
		return errors.New("password baru tidak boleh sama dengan password lama")
	}

	// 5. Hash new_password menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengenkripsi password baru")
	}

	// 6. Update password di database (hanya field password)
	err = s.userRepo.UpdatePassword(userID, string(hashedPassword))
	if err != nil {
		return errors.New("gagal mengubah password")
	}

	return nil
}

// ResetPasswordByAdmin mereset password user oleh admin/sekertaris tanpa memerlukan password lama.
func (s *userService) ResetPasswordByAdmin(targetUserID uint, newPassword string) error {
	// 1. Validasi input: Jika kosong, gunakan default password
	if newPassword == "" {
		newPassword = "password123"
	}

	if len(newPassword) < 8 {
		return errors.New("password baru minimal 8 karakter")
	}

	// 2. Cek keberadaan user
	_, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// 3. Hash password baru menggunakan bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("gagal mengenkripsi password baru")
	}

	// 4. Update password di database
	err = s.userRepo.UpdatePassword(targetUserID, string(hashedPassword))
	if err != nil {
		return errors.New("gagal mereset password")
	}

	return nil
}

// UpdateProfilePhoto mengubah foto profil user.
func (s *userService) UpdateProfilePhoto(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	// 1. Validasi tipe file (termasuk format kamera HP modern)
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".heic" {
		return "", errors.New("format file tidak didukung, gunakan JPG/JPEG/PNG/WEBP/HEIC")
	}

	// 3. Ambil data user untuk cek foto lama
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return "", errors.New("user tidak ditemukan")
	}

	var oldPhotoPath string
	if user.FotoPath != nil && *user.FotoPath != "" {
		oldPhotoPath = filepath.FromSlash(*user.FotoPath)
	}

	// 4. Simpan file baru ke ./uploads/photos/
	uploadDir := "./uploads/photos"
	err = os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", fmt.Errorf("gagal membuat direktori upload: %v", err)
	}

	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file foto: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file tujuan foto: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("gagal menyalin isi file foto: %w", err)
	}

	destPathSlash := filepath.ToSlash(destPath)

	// 5. Update foto_path di database
	err = s.userRepo.UpdateFoto(userID, destPathSlash)
	if err != nil {
		// Hapus file yang baru diupload jika gagal update DB
		os.Remove(destPath)
		return "", errors.New("gagal mengupdate foto profil")
	}

	// 6. Hapus foto lama setelah foto baru sukses disimpan dan DB ter-update
	if oldPhotoPath != "" {
		os.Remove(oldPhotoPath)
	}

	return destPathSlash, nil
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
