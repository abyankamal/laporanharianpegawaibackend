package repository

import (
	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// ReviewRepository adalah interface untuk operasi database Penilaian.
type ReviewRepository interface {
	Create(review *domain.Penilaian) error
	FindByUserID(userID int, limit int, offset int) ([]domain.Penilaian, int64, error)
	FindByPenilaiID(penilaiID int) ([]domain.Penilaian, error)
}

// reviewRepository adalah implementasi dari ReviewRepository.
type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository membuat instance baru ReviewRepository.
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

// Create menyimpan data penilaian baru ke database.
func (r *reviewRepository) Create(review *domain.Penilaian) error {
	return r.db.Create(review).Error
}

// FindByUserID mengambil riwayat penilaian berdasarkan pegawai yang dinilai (dengan pagination).
func (r *reviewRepository) FindByUserID(userID int, limit int, offset int) ([]domain.Penilaian, int64, error) {
	var reviews []domain.Penilaian
	var total int64

	query := r.db.Model(&domain.Penilaian{}).Where("user_id = ?", userID)

	// Hitung total data untuk pagination metadata
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Default limit
	if limit <= 0 {
		limit = 10
	}

	err := query.
		Preload("User").
		Preload("Penilai").
		Preload("Skor").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&reviews).Error

	if err != nil {
		return nil, 0, err
	}

	return reviews, total, nil
}

// FindByPenilaiID mengambil semua penilaian yang pernah dibuat oleh atasan tertentu.
func (r *reviewRepository) FindByPenilaiID(penilaiID int) ([]domain.Penilaian, error) {
	var reviews []domain.Penilaian

	err := r.db.
		Where("penilai_id = ?", penilaiID).
		Preload("User").
		Preload("User.Jabatan").
		Preload("Skor").
		Order("created_at DESC").
		Find(&reviews).Error

	if err != nil {
		return nil, err
	}

	return reviews, nil
}
