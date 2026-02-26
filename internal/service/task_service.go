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
	JenisTugas    string `json:"jenis_tugas" validate:"required"` // "organisasi" | "individu"
	TargetUserIDs []int  `json:"target_user_ids"`                 // Wajib untuk organisasi (banyak user)
	JudulTugas    string `json:"judul_tugas" validate:"required"`
	Deskripsi     string `json:"deskripsi"`
	FileBukti     string `json:"file_bukti"` // Opsional, URL dokumen pendukung untuk tugas organisasi
}

// UpdateTaskRequest adalah struct input untuk mengubah tugas pokok.
type UpdateTaskRequest struct {
	JenisTugas    string `json:"jenis_tugas" validate:"required"` // "organisasi" | "individu"
	TargetUserIDs []int  `json:"target_user_ids"`                 // Wajib untuk organisasi
	JudulTugas    string `json:"judul_tugas" validate:"required"`
	Deskripsi     string `json:"deskripsi"`
	FileBukti     string `json:"file_bukti"` // Opsional
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

// CreateTask membuat tugas pokok baru dengan dua jalur logika:
// - ORGANISASI: Dibuat oleh Lurah, bisa di-assign ke banyak user via M2M.
// - INDIVIDU: Dibuat mandiri oleh user yang bersangkutan.
func (s *taskService) CreateTask(requesterID uint, requesterRole string, req CreateTaskRequest) (*domain.TugasPokok, error) {
	// 1. Validasi input dasar
	if req.JudulTugas == "" {
		return nil, errors.New("judul_tugas wajib diisi")
	}
	if req.JenisTugas != "organisasi" && req.JenisTugas != "individu" {
		return nil, errors.New("jenis_tugas harus 'organisasi' atau 'individu'")
	}

	switch req.JenisTugas {
	case "organisasi":
		return s.createOrganisasiTask(requesterID, requesterRole, req)
	case "individu":
		return s.createIndividuTask(requesterID, req)
	default:
		return nil, errors.New("jenis_tugas tidak valid")
	}
}

// createOrganisasiTask membuat tugas organisasi (hanya boleh oleh Lurah).
// Menerima multiple TargetUserIDs dan menyimpan sebagai M2M assignees.
func (s *taskService) createOrganisasiTask(requesterID uint, requesterRole string, req CreateTaskRequest) (*domain.TugasPokok, error) {
	// 1. Validasi: Hanya Lurah yang boleh membuat tugas organisasi
	if requesterRole != "lurah" {
		return nil, errors.New("hanya Lurah yang boleh membuat tugas organisasi")
	}

	// 2. Validasi: TargetUserIDs wajib diisi untuk tugas organisasi
	if len(req.TargetUserIDs) == 0 {
		return nil, errors.New("target_user_ids wajib diisi untuk tugas organisasi")
	}

	// 3. Validasi semua target user ada di database
	var assignees []domain.User
	for _, uid := range req.TargetUserIDs {
		user, err := s.userRepo.FindByID(uint(uid))
		if err != nil {
			return nil, fmt.Errorf("user target dengan ID %d tidak ditemukan", uid)
		}
		assignees = append(assignees, *user)
	}

	// 4. Buat struct TugasPokok (tanpa UserID karena M2M)
	var fileBukti *string
	if req.FileBukti != "" {
		fileBukti = &req.FileBukti
	}

	tugas := &domain.TugasPokok{
		JenisTugas: "organisasi",
		FileBukti:  fileBukti,
		JudulTugas: req.JudulTugas,
		Deskripsi:  req.Deskripsi,
		CreatedBy:  &requesterID,
		CreatedAt:  time.Now(),
	}

	// 5. Simpan tugas ke database
	if err := s.taskRepo.Create(tugas); err != nil {
		return nil, fmt.Errorf("gagal menyimpan tugas: %v", err)
	}

	// 6. Simpan relasi M2M assignees
	if err := s.taskRepo.ReplaceAssignees(tugas.ID, assignees); err != nil {
		return nil, fmt.Errorf("gagal menyimpan assignees: %v", err)
	}

	// 7. Set assignees di response
	tugas.Assignees = assignees

	// 8. Kirim notifikasi ke semua assignees
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
	}

	return tugas, nil
}

// createIndividuTask membuat tugas individu secara mandiri oleh user yang bersangkutan.
// Tidak memerlukan validasi atasan — UserID = RequesterID.
func (s *taskService) createIndividuTask(requesterID uint, req CreateTaskRequest) (*domain.TugasPokok, error) {
	// 1. Buat struct TugasPokok — UserID = requesterID (diri sendiri)
	tugas := &domain.TugasPokok{
		UserID:     &requesterID,
		JenisTugas: "individu",
		JudulTugas: req.JudulTugas,
		Deskripsi:  req.Deskripsi,
		CreatedBy:  &requesterID,
		CreatedAt:  time.Now(),
	}

	// 2. Simpan ke database
	if err := s.taskRepo.Create(tugas); err != nil {
		return nil, fmt.Errorf("gagal menyimpan tugas: %v", err)
	}

	return tugas, nil
}

// GetTasksByUserID mengambil daftar tugas pokok untuk user tertentu.
// Menggabungkan tugas individu (via user_id) dan tugas organisasi (via M2M assignees).
func (s *taskService) GetTasksByUserID(userID int) ([]domain.TugasPokok, error) {
	// 1. Ambil tugas individu
	individuTasks, err := s.taskRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Ambil tugas organisasi yang di-assign ke user ini
	orgTasks, err := s.taskRepo.FindByAssigneeID(userID)
	if err != nil {
		return nil, err
	}

	// 3. Gabungkan dan deduplicate (menghindari duplikasi jika ada)
	taskMap := make(map[uint]domain.TugasPokok)
	for _, t := range individuTasks {
		taskMap[t.ID] = t
	}
	for _, t := range orgTasks {
		taskMap[t.ID] = t
	}

	// 4. Convert map back to slice
	var result []domain.TugasPokok
	for _, t := range taskMap {
		result = append(result, t)
	}

	return result, nil
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
	if req.JudulTugas == "" {
		return nil, errors.New("judul_tugas wajib diisi")
	}
	if req.JenisTugas != "organisasi" && req.JenisTugas != "individu" {
		return nil, errors.New("jenis_tugas harus 'organisasi' atau 'individu'")
	}

	// 4. Update field utama
	task.JenisTugas = req.JenisTugas
	task.JudulTugas = req.JudulTugas
	task.Deskripsi = req.Deskripsi

	if req.FileBukti != "" {
		task.FileBukti = &req.FileBukti
	} else {
		task.FileBukti = nil
	}

	// 5. Handle berdasarkan jenis tugas
	if req.JenisTugas == "organisasi" {
		// Validasi: Hanya Lurah yang boleh membuat tugas organisasi
		if requesterRole != "lurah" {
			return nil, errors.New("hanya Lurah yang boleh mengubah tugas menjadi jenis organisasi")
		}

		// Validasi TargetUserIDs
		if len(req.TargetUserIDs) == 0 {
			return nil, errors.New("target_user_ids wajib diisi untuk tugas organisasi")
		}

		// Validasi semua target user
		var assignees []domain.User
		for _, uid := range req.TargetUserIDs {
			user, err := s.userRepo.FindByID(uint(uid))
			if err != nil {
				return nil, fmt.Errorf("user target dengan ID %d tidak ditemukan", uid)
			}
			assignees = append(assignees, *user)
		}

		// Update assignees M2M
		if err := s.taskRepo.ReplaceAssignees(taskID, assignees); err != nil {
			return nil, fmt.Errorf("gagal mengubah assignees: %v", err)
		}

		// Clear UserID karena organisasi pakai M2M
		task.UserID = nil
		task.Assignees = assignees
	} else {
		// Tugas individu: UserID tetap requesterID atau target
		if len(req.TargetUserIDs) > 0 {
			uid := uint(req.TargetUserIDs[0])
			task.UserID = &uid
		}

		// Clear M2M assignees
		if err := s.taskRepo.ReplaceAssignees(taskID, []domain.User{}); err != nil {
			return nil, fmt.Errorf("gagal membersihkan assignees: %v", err)
		}
	}

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

	// 3. Hapus tugas (akan otomatis clear M2M di repository)
	if err := s.taskRepo.Delete(taskID); err != nil {
		return fmt.Errorf("gagal menghapus tugas: %v", err)
	}

	return nil
}
