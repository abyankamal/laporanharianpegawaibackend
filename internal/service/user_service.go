package service

import (
	"errors"
	"time"

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

// UserService adalah interface untuk operasi bisnis User.
type UserService interface {
	GetAllUsers() ([]domain.User, error)
	GetUserByID(id uint) (*domain.User, error)
	CreateUser(req CreateUserRequest) (*domain.User, error)
	UpdateUser(id uint, req UpdateUserRequest) (*domain.User, error)
	DeleteUser(id uint) error
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
		Role:         req.Role,
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
		user.Role = req.Role
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

// DeleteUser menghapus user berdasarkan ID.
func (s *userService) DeleteUser(id uint) error {
	// Cek apakah user ada
	_, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user tidak ditemukan")
	}

	// Hapus user
	err = s.userRepo.Delete(id)
	if err != nil {
		return errors.New("gagal menghapus user")
	}

	return nil
}
