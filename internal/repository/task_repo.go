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
	FindByID(id uint) (*domain.TugasPokok, error)
	Update(task *domain.TugasPokok) error
	Delete(id uint) error
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
	return r.db.Omit("User", "Creator").Create(task).Error
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

// FindByID mengambil tugas pokok berdasarkan ID.
func (r *taskRepository) FindByID(id uint) (*domain.TugasPokok, error) {
	var task domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").
		First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update mengubah data tugas pokok di database.
func (r *taskRepository) Update(task *domain.TugasPokok) error {
	return r.db.Omit("User", "Creator").Save(task).Error
}

// Delete menghapus tugas pokok dari database.
func (r *taskRepository) Delete(id uint) error {
	return r.db.Delete(&domain.TugasPokok{}, id).Error
}
