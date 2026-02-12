package repository

import (
	"errors"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// UserRepository adalah interface untuk operasi database User.
type UserRepository interface {
	FindByNIP(nip string) (*domain.User, error)
	FindAll() ([]domain.User, error)
	FindByID(id uint) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
	Delete(id uint) error
	UpdatePassword(userID uint, newPasswordHash string) error
}

// userRepository adalah implementasi dari UserRepository.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository membuat instance baru UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// FindByNIP mencari user berdasarkan NIP.
func (r *userRepository) FindByNIP(nip string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("nip = ?", nip).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindAll mengambil semua user dengan preload relasi Jabatan.
func (r *userRepository) FindAll() ([]domain.User, error) {
	var users []domain.User
	result := r.db.Preload("Jabatan").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// FindByID mengambil user berdasarkan ID dengan preload relasi Jabatan.
func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	result := r.db.Preload("Jabatan").Preload("Supervisor").First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create menyimpan user baru ke database.
func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// Update mengupdate data user.
func (r *userRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}

// Delete menghapus user berdasarkan ID.
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

// UpdatePassword mengupdate password user secara spesifik.
// Method ini hanya mengupdate field password untuk menghindari perubahan field lain.
func (r *userRepository) UpdatePassword(userID uint, newPasswordHash string) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("password", newPasswordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}
