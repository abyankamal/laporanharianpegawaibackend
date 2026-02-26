package service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
)

// ============================================================
// Test CreateTask — Tugas Organisasi
// ============================================================

func TestCreateTask_Organisasi_Success(t *testing.T) {
	t.Run("Sukses: Lurah membuat tugas organisasi ke 2 user", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: target users
		user1 := &domain.User{ID: 2, Nama: "Aep Saepudin", Role: "sekertaris"}
		user2 := &domain.User{ID: 3, Nama: "Mas Kasi", Role: "kasi"}
		mockUserRepo.On("FindByID", uint(2)).Return(user1, nil)
		mockUserRepo.On("FindByID", uint(3)).Return(user2, nil)

		// Mock: simpan tugas berhasil
		mockTaskRepo.On("Create", mock.Anything).Return(nil)
		mockTaskRepo.On("ReplaceAssignees", mock.AnythingOfType("uint"), mock.Anything).Return(nil)

		// Mock: notifikasi berhasil
		mockNotifRepo.On("Create", mock.Anything).Return(nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		// Execute
		req := CreateTaskRequest{
			JenisTugas:    "organisasi",
			TargetUserIDs: []int{2, 3},
			JudulTugas:    "Tugas Organisasi Penting",
			Deskripsi:     "Deskripsi tugas organisasi",
			FileBukti:     "https://example.com/bukti.pdf",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tugas)
		assert.Equal(t, "organisasi", tugas.JenisTugas)
		assert.Equal(t, "Tugas Organisasi Penting", tugas.JudulTugas)
		assert.Len(t, tugas.Assignees, 2)

		mockTaskRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		// Notifikasi dikirim ke 2 assignees
		mockNotifRepo.AssertNumberOfCalls(t, "Create", 2)
	})
}

func TestCreateTask_Organisasi_Fail_NonLurah(t *testing.T) {
	t.Run("Gagal: Sekertaris tidak boleh membuat tugas organisasi", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas:    "organisasi",
			TargetUserIDs: []int{3},
			JudulTugas:    "Tugas dari sekertaris",
		}
		tugas, err := taskSvc.CreateTask(2, "sekertaris", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "hanya Lurah yang boleh membuat tugas organisasi", err.Error())
		mockTaskRepo.AssertNotCalled(t, "Create")
	})
}

func TestCreateTask_Organisasi_Fail_EmptyTargetUserIDs(t *testing.T) {
	t.Run("Gagal: Tugas organisasi tanpa target_user_ids", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas:    "organisasi",
			TargetUserIDs: []int{}, // kosong
			JudulTugas:    "Tugas tanpa target",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "target_user_ids wajib diisi untuk tugas organisasi", err.Error())
	})
}

func TestCreateTask_Organisasi_Fail_TargetNotFound(t *testing.T) {
	t.Run("Gagal: User target tidak ditemukan di database", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		mockUserRepo.On("FindByID", uint(999)).Return(nil, errors.New("record not found"))

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas:    "organisasi",
			TargetUserIDs: []int{999},
			JudulTugas:    "Tugas ke user yang tidak ada",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Contains(t, err.Error(), "tidak ditemukan")
	})
}

// ============================================================
// Test CreateTask — Tugas Individu
// ============================================================

func TestCreateTask_Individu_Success(t *testing.T) {
	t.Run("Sukses: User membuat tugas individu untuk diri sendiri", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		mockTaskRepo.On("Create", mock.Anything).Return(nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas: "individu",
			JudulTugas: "Tugas Mandiri Staf",
			Deskripsi:  "Saya buat tugas sendiri",
		}
		tugas, err := taskSvc.CreateTask(3, "staf", req)

		assert.NoError(t, err)
		assert.NotNil(t, tugas)
		assert.Equal(t, "individu", tugas.JenisTugas)
		assert.Equal(t, uint(3), *tugas.UserID) // UserID = requesterID
		assert.Equal(t, uint(3), *tugas.CreatedBy)
		mockTaskRepo.AssertExpectations(t)
		// Tidak ada notifikasi untuk tugas individu
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}

// ============================================================
// Test CreateTask — Validasi Umum
// ============================================================

func TestCreateTask_Fail_EmptyJudul(t *testing.T) {
	t.Run("Gagal: Judul tugas kosong", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas: "individu",
			JudulTugas: "", // kosong
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "judul_tugas wajib diisi", err.Error())
	})
}

func TestCreateTask_Fail_InvalidJenisTugas(t *testing.T) {
	t.Run("Gagal: Jenis tugas tidak valid", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas: "invalid",
			JudulTugas: "Test",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "jenis_tugas harus 'organisasi' atau 'individu'", err.Error())
	})
}

func TestCreateTask_Fail_DBError(t *testing.T) {
	t.Run("Gagal: Error saat simpan tugas individu", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		mockTaskRepo.On("Create", mock.Anything).Return(errors.New("database error"))

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			JenisTugas: "individu",
			JudulTugas: "Tugas yang gagal disimpan",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Contains(t, err.Error(), "gagal menyimpan tugas")
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}

// ============================================================
// Test GetAllTasks
// ============================================================

func TestGetAllTasks_Success(t *testing.T) {
	t.Run("Sukses: Mengambil semua tugas pokok", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		expectedTasks := []domain.TugasPokok{
			{JudulTugas: "Tugas 1", JenisTugas: "individu"},
			{JudulTugas: "Tugas 2", JenisTugas: "organisasi"},
		}

		mockTaskRepo.On("FindAll").Return(expectedTasks, nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		tasks, err := taskSvc.GetAllTasks()

		assert.NoError(t, err)
		assert.Len(t, tasks, 2)
		assert.Equal(t, "Tugas 1", tasks[0].JudulTugas)
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestGetAllTasks_Fail_DBError(t *testing.T) {
	t.Run("Gagal: Database error saat mengambil semua tugas", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		mockTaskRepo.On("FindAll").Return(nil, errors.New("db disconnect"))

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		tasks, err := taskSvc.GetAllTasks()

		assert.Error(t, err)
		assert.Nil(t, tasks)
		assert.Equal(t, "db disconnect", err.Error())
	})
}
