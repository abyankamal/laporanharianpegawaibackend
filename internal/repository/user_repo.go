package repository

import (
	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// UserRepository adalah interface untuk operasi database User.
type UserRepository interface {
	FindByNIP(nip string) (*domain.User, error)
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
