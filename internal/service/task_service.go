package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// CreateTaskRequest adalah struct input untuk membuat tugas pokok baru.
type CreateTaskRequest struct {
	TargetUserID int    `json:"target_user_id" validate:"required"`
	JudulTugas   string `json:"judul_tugas" validate:"required"`
	Deskripsi    string `json:"deskripsi"`
}

// UpdateTaskRequest adalah struct input untuk mengubah tugas pokok.
type UpdateTaskRequest struct {
	TargetUserID int    `json:"target_user_id" validate:"required"`
	JudulTugas   string `json:"judul_tugas" validate:"required"`
	Deskripsi    string `json:"deskripsi"`
}

// TaskService adalah interface untuk operasi bisnis Tugas Pokok.
type TaskService interface {
	CreateTask(requesterID uint, requesterRole string, req CreateTaskRequest) (*domain.TugasPokok, error)
	GetTasksByUserID(userID int) ([]domain.TugasPokok, error)
	GetAllTasks() ([]domain.TugasPokok, error)
	UpdateTask(requesterID uint, requesterRole string, taskID uint, req UpdateTaskRequest) (*domain.TugasPokok, error)
	DeleteTask(requesterID uint, requesterRole string, taskID uint) error
}

// taskService adalah implementasi dari TaskService.
type taskService struct {
	taskRepo  repository.TaskRepository
	userRepo  repository.UserRepository
	notifRepo repository.NotificationRepository
}

// NewTaskService membuat instance baru TaskService.
func NewTaskService(taskRepo repository.TaskRepository, userRepo repository.UserRepository, notifRepo repository.NotificationRepository) TaskService {
	return &taskService{
		taskRepo:  taskRepo,
		userRepo:  userRepo,
		notifRepo: notifRepo,
	}
}

// CreateTask membuat tugas pokok baru dengan validasi hierarki RBAC.
// - Lurah hanya boleh memberi tugas ke Sekertaris dan Kasi.
// - Sekertaris hanya boleh memberi tugas ke Staf.
// - Kasi dan Staf TIDAK boleh memberi tugas.
func (s *taskService) CreateTask(requesterID uint, requesterRole string, req CreateTaskRequest) (*domain.TugasPokok, error) {
	// 1. Validasi: Hanya lurah dan sekertaris yang boleh memberi tugas
	if requesterRole != "lurah" && requesterRole != "sekertaris" {
		return nil, errors.New("hanya Lurah dan Sekertaris yang boleh memberi tugas")
	}

	// 2. Validasi input
	if req.TargetUserID <= 0 {
		return nil, errors.New("target_user_id wajib diisi dan harus valid")
	}
	if req.JudulTugas == "" {
		return nil, errors.New("judul_tugas wajib diisi")
	}

	// 3. Tidak boleh memberi tugas ke diri sendiri (kecuali Lurah)
	if uint(req.TargetUserID) == requesterID && requesterRole != "lurah" {
		return nil, errors.New("tidak dapat memberi tugas ke diri sendiri")
	}

	// 4. Validasi hierarki: Cek role target user
	targetUser, err := s.userRepo.FindByID(uint(req.TargetUserID))
	if err != nil {
		return nil, errors.New("user target tidak ditemukan")
	}

	switch requesterRole {
	case "lurah":
		// Lurah boleh memberi tugas ke sekertaris, kasi, dan diri sendiri
		if targetUser.Role != "sekertaris" && targetUser.Role != "kasi" && targetUser.Role != "lurah" {
			return nil, errors.New("Lurah hanya boleh memberi tugas ke Sekertaris, Kasi, atau diri sendiri")
		}
	case "sekertaris":
		// Sekertaris hanya boleh memberi tugas ke staf
		if targetUser.Role != "staf" {
			return nil, errors.New("Sekertaris hanya boleh memberi tugas ke Staf")
		}
	}

	// 5. Buat struct TugasPokok
	userID := uint(req.TargetUserID)
	tugas := &domain.TugasPokok{
		UserID:     &userID,
		JudulTugas: req.JudulTugas,
		Deskripsi:  req.Deskripsi,
		CreatedBy:  &requesterID,
		CreatedAt:  time.Now(),
	}

	// 6. Simpan ke database
	err = s.taskRepo.Create(tugas)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan tugas: %v", err)
	}

	// 7. Buat notifikasi untuk target user
	notif := &domain.Notification{
		UserID:    req.TargetUserID,
		Kategori:  "Tugas",
		Judul:     "Tugas Baru Ditetapkan",
		Pesan:     fmt.Sprintf("Anda telah ditugaskan untuk '%s'. Silakan cek detail tugas.", req.JudulTugas),
		TerkaitID: int(tugas.ID),
		CreatedAt: time.Now(),
	}
	if err := s.notifRepo.Create(notif); err != nil {
		log.Printf("⚠️ Gagal membuat notifikasi tugas: %v", err)
	}

	return tugas, nil
}

// GetTasksByUserID mengambil daftar tugas pokok untuk user tertentu.
func (s *taskService) GetTasksByUserID(userID int) ([]domain.TugasPokok, error) {
	return s.taskRepo.FindByUserID(userID)
}

// GetAllTasks mengambil semua tugas pokok (untuk atasan).
func (s *taskService) GetAllTasks() ([]domain.TugasPokok, error) {
	return s.taskRepo.FindAll()
}

// UpdateTask mengubah tugas pokok dengan validasi bahwa hanya pembuat atau lurah yang dapat mengubahnya.
func (s *taskService) UpdateTask(requesterID uint, requesterRole string, taskID uint, req UpdateTaskRequest) (*domain.TugasPokok, error) {
	// 1. Cari tugas berdasarkan ID
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, errors.New("tugas tidak ditemukan")
	}

	// 2. Validasi otorisasi: Hanya pembuat tugas atau Lurah yang boleh mengedit
	if *task.CreatedBy != requesterID && requesterRole != "lurah" {
		return nil, errors.New("anda tidak memiliki akses untuk mengubah tugas ini")
	}

	// 3. Validasi input
	if req.TargetUserID <= 0 {
		return nil, errors.New("target_user_id wajib diisi dan harus valid")
	}
	if req.JudulTugas == "" {
		return nil, errors.New("judul_tugas wajib diisi")
	}

	// 4. Validasi hierarki target user (sama dengan create)
	targetUser, err := s.userRepo.FindByID(uint(req.TargetUserID))
	if err != nil {
		return nil, errors.New("user target tidak ditemukan")
	}

	switch requesterRole {
	case "lurah":
		// Lurah boleh memberi tugas ke sekertaris, kasi, dan diri sendiri
		if targetUser.Role != "sekertaris" && targetUser.Role != "kasi" && targetUser.Role != "lurah" {
			return nil, errors.New("Lurah hanya boleh memberi tugas ke Sekertaris, Kasi, atau diri sendiri")
		}
	case "sekertaris":
		// Sekertaris hanya boleh memberi tugas ke staf
		if targetUser.Role != "staf" {
			return nil, errors.New("Sekertaris hanya boleh memberi tugas ke Staf")
		}
	}

	// 5. Update data
	userID := uint(req.TargetUserID)
	task.UserID = &userID
	task.JudulTugas = req.JudulTugas
	task.Deskripsi = req.Deskripsi

	// 6. Simpan perubahan ke DB
	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("gagal mengubah tugas: %v", err)
	}

	return task, nil
}

// DeleteTask menghapus tugas pokok dengan validasi bahwa hanya pembuat atau lurah yang dapat menghapusnya.
func (s *taskService) DeleteTask(requesterID uint, requesterRole string, taskID uint) error {
	// 1. Cari tugas berdasarkan ID
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	// 2. Validasi otorisasi: Hanya pembuat tugas atau Lurah yang boleh menghapus
	if *task.CreatedBy != requesterID && requesterRole != "lurah" {
		return errors.New("anda tidak memiliki akses untuk menghapus tugas ini")
	}

	// 3. Hapus tugas
	if err := s.taskRepo.Delete(taskID); err != nil {
		return fmt.Errorf("gagal menghapus tugas: %v", err)
	}

	return nil
}
