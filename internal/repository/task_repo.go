package repository

import (
	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// TaskRepository adalah interface untuk operasi database Tugas Pokok.
type TaskRepository interface {
	Create(task *domain.TugasPokok) error
	FindByUserID(userID int) ([]domain.TugasPokok, error)
	FindAll() ([]domain.TugasPokok, error)
}

// taskRepository adalah implementasi dari TaskRepository.
type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository membuat instance baru TaskRepository.
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

// Create menyimpan tugas pokok baru ke database.
func (r *taskRepository) Create(task *domain.TugasPokok) error {
	return r.db.Create(task).Error
}

// FindByUserID mengambil tugas pokok berdasarkan user_id.
func (r *taskRepository) FindByUserID(userID int) ([]domain.TugasPokok, error) {
	var tasks []domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindAll mengambil semua tugas pokok dari database.
func (r *taskRepository) FindAll() ([]domain.TugasPokok, error) {
	var tasks []domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}
