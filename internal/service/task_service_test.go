package service

import (
	"testing"
	"time"

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
		req := CreateOrganizationalTaskRequest{
			TargetUserIDs: []int{2, 3},
			JudulTugas:    "Tugas Organisasi Penting",
			Deskripsi:     "Deskripsi tugas organisasi",
			FileBukti:     "https://example.com/bukti.pdf",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tugas)
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

		req := CreateOrganizationalTaskRequest{
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

		req := CreateOrganizationalTaskRequest{
			TargetUserIDs: []int{}, // kosong
			JudulTugas:    "Tugas tanpa target",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "target_user_ids wajib diisi", err.Error())
	})
}

// = [x] Test GetAllTasks
// ============================================================

func TestGetAllTasks_Success(t *testing.T) {
	t.Run("Sukses: Mengambil semua tugas organisasi", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		expectedTasks := []domain.TugasOrganisasi{
			{JudulTugas: "Tugas 1"},
			{JudulTugas: "Tugas 2"},
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

// ============================================================
// Test UpdateTask — Tugas Organisasi
// ============================================================

func TestUpdateTask_Success_Partial(t *testing.T) {
	t.Run("Sukses: Lurah mengedit hanya judul tugas", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		existingDeadline := time.Now().Add(24 * time.Hour)
		existingTask := &domain.TugasOrganisasi{
			ID:         1,
			JudulTugas: "Judul Lama",
			Deskripsi:  "Deskripsi Lama",
			Deadline:   &existingDeadline,
		}

		mockTaskRepo.On("FindByID", uint(1)).Return(existingTask, nil)
		mockTaskRepo.On("Update", mock.Anything).Return(nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		// Execute: hanya kirim judul baru
		req := UpdateOrganizationalTaskRequest{
			JudulTugas: "Judul Baru",
		}
		updated, err := taskSvc.UpdateTask(1, "lurah", 1, req)

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "Judul Baru", updated.JudulTugas)
		assert.Equal(t, "Deskripsi Lama", updated.Deskripsi) // Tetap yang lama
		assert.Equal(t, &existingDeadline, updated.Deadline) // Tetap yang lama

		mockTaskRepo.AssertExpectations(t)
	})

	t.Run("Sukses: Lurah mengedit hanya assignees", func(t *testing.T) {
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		existingTask := &domain.TugasOrganisasi{
			ID:         1,
			JudulTugas: "Judul",
			Assignees:  []domain.User{{ID: 2}},
		}

		newUser := &domain.User{ID: 3}
		mockTaskRepo.On("FindByID", uint(1)).Return(existingTask, nil)
		mockUserRepo.On("FindByID", uint(3)).Return(newUser, nil)
		mockTaskRepo.On("ReplaceAssignees", uint(1), []domain.User{*newUser}).Return(nil)
		mockTaskRepo.On("Update", mock.Anything).Return(nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := UpdateOrganizationalTaskRequest{
			TargetUserIDs: []int{3},
		}
		updated, err := taskSvc.UpdateTask(1, "lurah", 1, req)

		assert.NoError(t, err)
		assert.Len(t, updated.Assignees, 1)
		assert.Equal(t, uint(3), updated.Assignees[0].ID)
		mockTaskRepo.AssertExpectations(t)
	})
}
