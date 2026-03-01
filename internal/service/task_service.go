package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/pkg/fcm"
)

// CreateOrganizationalTaskRequest adalah struct input untuk membuat tugas organisasi baru.
type CreateOrganizationalTaskRequest struct {
	TargetUserIDs []int  `json:"target_user_ids" validate:"required"`
	JudulTugas    string `json:"judul_tugas" validate:"required"`
	Deskripsi     string `json:"deskripsi"`
	FileBukti     string `json:"file_bukti"` // Opsional, URL dokumen pendukung
}

// UpdateOrganizationalTaskRequest adalah struct input untuk mengubah tugas organisasi.
type UpdateOrganizationalTaskRequest struct {
	TargetUserIDs []int  `json:"target_user_ids" validate:"required"`
	JudulTugas    string `json:"judul_tugas" validate:"required"`
	Deskripsi     string `json:"deskripsi"`
	FileBukti     string `json:"file_bukti"` // Opsional
}

// TaskService adalah interface untuk operasi bisnis Tugas Organisasi.
type TaskService interface {
	CreateTask(requesterID uint, requesterRole string, req CreateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error)
	GetMyTasks(userID int) ([]domain.TugasOrganisasi, error)
	GetAllTasks() ([]domain.TugasOrganisasi, error)
	UpdateTask(requesterID uint, requesterRole string, taskID uint, req UpdateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error)
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

// CreateTask membuat tugas organisasi baru (hanya boleh oleh Lurah).
func (s *taskService) CreateTask(requesterID uint, requesterRole string, req CreateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error) {
	// 1. Validasi: Hanya Lurah yang boleh membuat tugas organisasi
	if requesterRole != "lurah" {
		return nil, errors.New("hanya Lurah yang boleh membuat tugas organisasi")
	}

	// 2. Validasi input dasar
	if req.JudulTugas == "" {
		return nil, errors.New("judul_tugas wajib diisi")
	}

	// 3. Validasi: TargetUserIDs wajib diisi
	if len(req.TargetUserIDs) == 0 {
		return nil, errors.New("target_user_ids wajib diisi")
	}

	// 4. Validasi semua target user ada di database
	var assignees []domain.User
	for _, uid := range req.TargetUserIDs {
		user, err := s.userRepo.FindByID(uint(uid))
		if err != nil {
			return nil, fmt.Errorf("user target dengan ID %d tidak ditemukan", uid)
		}
		assignees = append(assignees, *user)
	}

	// 5. Buat struct TugasOrganisasi
	var fileBukti *string
	if req.FileBukti != "" {
		fileBukti = &req.FileBukti
	}

	tugas := &domain.TugasOrganisasi{
		FileBukti:  fileBukti,
		JudulTugas: req.JudulTugas,
		Deskripsi:  req.Deskripsi,
		CreatedBy:  &requesterID,
		CreatedAt:  time.Now(),
	}

	// 6. Simpan tugas ke database
	if err := s.taskRepo.Create(tugas); err != nil {
		return nil, fmt.Errorf("gagal menyimpan tugas: %v", err)
	}

	// 7. Simpan relasi M2M assignees
	if err := s.taskRepo.ReplaceAssignees(tugas.ID, assignees); err != nil {
		return nil, fmt.Errorf("gagal menyimpan assignees: %v", err)
	}

	// 8. Set assignees di response
	tugas.Assignees = assignees

	// 9. Kirim notifikasi ke semua assignees
	for _, user := range assignees {
		notif := &domain.Notification{
			UserID:    int(user.ID),
			Kategori:  "Tugas",
			Judul:     "Tugas Organisasi Baru",
			Pesan:     fmt.Sprintf("Anda telah ditugaskan untuk tugas organisasi '%s'. Silakan cek detail tugas.", req.JudulTugas),
			TerkaitID: int(tugas.ID),
			CreatedAt: time.Now(),
		}
		if err := s.notifRepo.Create(notif); err != nil {
			log.Printf("⚠️ Gagal membuat notifikasi untuk user %d: %v", user.ID, err)
		}

		// Trigger FCM Push Notification
		if user.FCMToken != nil && *user.FCMToken != "" {
			go fcm.SendPushNotification(*user.FCMToken, notif.Judul, notif.Pesan)
		}
	}

	return tugas, nil
}

// GetMyTasks mengambil daftar tugas organisasi yang di-assign ke user tertentu.
func (s *taskService) GetMyTasks(userID int) ([]domain.TugasOrganisasi, error) {
	return s.taskRepo.FindByAssigneeID(userID)
}

// GetAllTasks mengambil semua tugas organisasi (untuk Lurah).
func (s *taskService) GetAllTasks() ([]domain.TugasOrganisasi, error) {
	return s.taskRepo.FindAll()
}

// UpdateTask mengubah tugas organisasi. Hanya Lurah yang dapat mengubahnya.
func (s *taskService) UpdateTask(requesterID uint, requesterRole string, taskID uint, req UpdateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error) {
	// 1. Validasi otorisasi: Hanya Lurah yang boleh mengedit
	if requesterRole != "lurah" {
		return nil, errors.New("hanya Lurah yang memiliki akses untuk mengubah tugas organisasi")
	}

	// 2. Cari tugas berdasarkan ID
	task, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return nil, errors.New("tugas tidak ditemukan")
	}

	// 3. Validasi input
	if req.JudulTugas == "" {
		return nil, errors.New("judul_tugas wajib diisi")
	}

	// 4. Update field utama
	task.JudulTugas = req.JudulTugas
	task.Deskripsi = req.Deskripsi

	if req.FileBukti != "" {
		task.FileBukti = &req.FileBukti
	} else {
		task.FileBukti = nil
	}

	// 5. Validasi TargetUserIDs
	if len(req.TargetUserIDs) == 0 {
		return nil, errors.New("target_user_ids wajib diisi")
	}

	// 6. Validasi semua target user
	var assignees []domain.User
	for _, uid := range req.TargetUserIDs {
		user, err := s.userRepo.FindByID(uint(uid))
		if err != nil {
			return nil, fmt.Errorf("user target dengan ID %d tidak ditemukan", uid)
		}
		assignees = append(assignees, *user)
	}

	// 7. Update assignees M2M
	if err := s.taskRepo.ReplaceAssignees(taskID, assignees); err != nil {
		return nil, fmt.Errorf("gagal mengubah assignees: %v", err)
	}

	task.Assignees = assignees

	// 8. Simpan perubahan ke DB
	if err := s.taskRepo.Update(task); err != nil {
		return nil, fmt.Errorf("gagal mengubah tugas: %v", err)
	}

	return task, nil
}

// DeleteTask menghapus tugas organisasi. Hanya Lurah yang dapat menghapusnya.
func (s *taskService) DeleteTask(requesterID uint, requesterRole string, taskID uint) error {
	// 1. Validasi otorisasi: Hanya Lurah yang boleh menghapus
	if requesterRole != "lurah" {
		return errors.New("hanya Lurah yang memiliki akses untuk menghapus tugas organisasi")
	}

	// 2. Cari tugas berdasarkan ID
	_, err := s.taskRepo.FindByID(taskID)
	if err != nil {
		return errors.New("tugas tidak ditemukan")
	}

	// 3. Hapus tugas (akan otomatis clear M2M di repository)
	if err := s.taskRepo.Delete(taskID); err != nil {
		return fmt.Errorf("gagal menghapus tugas: %v", err)
	}

	return nil
}
