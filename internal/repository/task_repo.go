package repository

import (
	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// TaskRepository adalah interface untuk operasi database Tugas Pokok.
type TaskRepository interface {
	Create(task *domain.TugasPokok) error
	FindByUserID(userID int) ([]domain.TugasPokok, error)
	FindByAssigneeID(userID int) ([]domain.TugasPokok, error) // Query tugas organisasi berdasarkan assignee
	FindAll() ([]domain.TugasPokok, error)
	FindByID(id uint) (*domain.TugasPokok, error)
	Update(task *domain.TugasPokok) error
	ReplaceAssignees(taskID uint, users []domain.User) error // Update relasi M2M assignees
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
// Untuk tugas organisasi, simpan assignees M2M secara terpisah setelah Create.
func (r *taskRepository) Create(task *domain.TugasPokok) error {
	return r.db.Omit("User", "Creator", "Assignees").Create(task).Error
}

// FindByUserID mengambil tugas pokok berdasarkan user_id (untuk tugas individu).
func (r *taskRepository) FindByUserID(userID int) ([]domain.TugasPokok, error) {
	var tasks []domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").Preload("Assignees").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindByAssigneeID mengambil tugas organisasi yang di-assign ke user tertentu (via M2M).
func (r *taskRepository) FindByAssigneeID(userID int) ([]domain.TugasPokok, error) {
	var tasks []domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").Preload("Assignees").
		Joins("JOIN tugas_assignees ON tugas_assignees.tugas_pokok_id = tugas_pokok.id").
		Where("tugas_assignees.user_id = ?", userID).
		Order("tugas_pokok.created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindAll mengambil semua tugas pokok dari database.
func (r *taskRepository) FindAll() ([]domain.TugasPokok, error) {
	var tasks []domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").Preload("Assignees").
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindByID mengambil tugas pokok berdasarkan ID.
func (r *taskRepository) FindByID(id uint) (*domain.TugasPokok, error) {
	var task domain.TugasPokok
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Creator").Preload("Assignees").
		First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update mengubah data tugas pokok di database.
func (r *taskRepository) Update(task *domain.TugasPokok) error {
	return r.db.Omit("User", "Creator", "Assignees").Save(task).Error
}

// ReplaceAssignees mengganti semua assignees untuk tugas tertentu (M2M).
func (r *taskRepository) ReplaceAssignees(taskID uint, users []domain.User) error {
	task := &domain.TugasPokok{ID: taskID}
	return r.db.Model(task).Association("Assignees").Replace(users)
}

// Delete menghapus tugas pokok dari database beserta relasi M2M-nya.
func (r *taskRepository) Delete(id uint) error {
	// Hapus relasi M2M terlebih dahulu
	task := &domain.TugasPokok{ID: id}
	if err := r.db.Model(task).Association("Assignees").Clear(); err != nil {
		return err
	}
	// Hapus tugas
	return r.db.Delete(&domain.TugasPokok{}, id).Error
}
