package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// TaskRepository adalah interface untuk operasi database Tugas Organisasi.
type TaskRepository interface {
	Create(task *domain.TugasOrganisasi) error
	FindByAssigneeID(userID int) ([]domain.TugasOrganisasi, error) // Query tugas organisasi berdasarkan assignee
	FindAll() ([]domain.TugasOrganisasi, error)
	FindByID(id uint) (*domain.TugasOrganisasi, error)
	Update(task *domain.TugasOrganisasi) error
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

// Create menyimpan tugas organisasi baru ke database.
func (r *taskRepository) Create(task *domain.TugasOrganisasi) error {
	return r.db.Omit("Creator", "Assignees").Create(task).Error
}

// FindByAssigneeID mengambil tugas organisasi yang di-assign ke user tertentu (via M2M), yang belum lewat deadline.
func (r *taskRepository) FindByAssigneeID(userID int) ([]domain.TugasOrganisasi, error) {
	var tasks []domain.TugasOrganisasi
	err := r.db.Preload("Creator").Preload("Assignees").Preload("Assignees.Jabatan").
		Joins("JOIN tugas_assignees ON tugas_assignees.tugas_organisasi_id = tugas_organisasi.id").
		Where("tugas_assignees.user_id = ?", userID).
		Where("tugas_organisasi.deadline IS NULL OR tugas_organisasi.deadline >= ?", time.Now()).
		Order("tugas_organisasi.created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindAll mengambil semua tugas organisasi dari database, yang belum lewat deadline.
func (r *taskRepository) FindAll() ([]domain.TugasOrganisasi, error) {
	var tasks []domain.TugasOrganisasi
	err := r.db.Preload("Creator").Preload("Assignees").Preload("Assignees.Jabatan").
		Where("deadline IS NULL OR deadline >= ?", time.Now()).
		Order("created_at DESC").
		Find(&tasks).Error
	return tasks, err
}

// FindByID mengambil tugas organisasi berdasarkan ID.
func (r *taskRepository) FindByID(id uint) (*domain.TugasOrganisasi, error) {
	var task domain.TugasOrganisasi
	err := r.db.Preload("Creator").Preload("Assignees").Preload("Assignees.Jabatan").
		First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// Update mengubah data tugas organisasi di database.
func (r *taskRepository) Update(task *domain.TugasOrganisasi) error {
	return r.db.Omit("Creator", "Assignees").Save(task).Error
}

// ReplaceAssignees mengganti semua assignees untuk tugas tertentu (M2M).
func (r *taskRepository) ReplaceAssignees(taskID uint, users []domain.User) error {
	task := &domain.TugasOrganisasi{ID: taskID}
	return r.db.Model(task).Association("Assignees").Replace(users)
}

// Delete menghapus tugas organisasi dari database beserta relasi M2M-nya.
func (r *taskRepository) Delete(id uint) error {
	// Hapus relasi M2M terlebih dahulu
	task := &domain.TugasOrganisasi{ID: id}
	if err := r.db.Model(task).Association("Assignees").Clear(); err != nil {
		return err
	}
	// Hapus tugas
	return r.db.Delete(&domain.TugasOrganisasi{}, id).Error
}
