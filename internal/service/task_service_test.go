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
// Test CreateTask (TaskService)
// ============================================================

func TestCreateTask_Success_LurahToSekertaris(t *testing.T) {
	t.Run("Sukses: Lurah memberi tugas ke Sekertaris + notifikasi terkirim", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: target user adalah sekertaris
		targetUser := &domain.User{
			ID:   2,
			Nama: "Aep Saepudin, S.Kom",
			Role: "sekertaris",
		}
		mockUserRepo.On("FindByID", uint(2)).Return(targetUser, nil)

		// Mock: simpan tugas berhasil
		mockTaskRepo.On("Create", mock.Anything).Return(nil)

		// Mock: simpan notifikasi berhasil
		mockNotifRepo.On("Create", mock.Anything).Return(nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		// Execute
		req := CreateTaskRequest{
			TargetUserID: 2,
			JudulTugas:   "Menyusun Laporan Bulanan",
			Deskripsi:    "Buat laporan bulanan untuk bulan Februari",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tugas)
		assert.Equal(t, "Menyusun Laporan Bulanan", tugas.JudulTugas)

		// Verifikasi: semua mock terpanggil sesuai ekspektasi
		mockTaskRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockNotifRepo.AssertExpectations(t)

		// Verifikasi: NotificationRepo.Create() dipanggil tepat 1 kali
		mockNotifRepo.AssertNumberOfCalls(t, "Create", 1)
	})
}

func TestCreateTask_Success_SekertarisToStaf(t *testing.T) {
	t.Run("Sukses: Sekertaris memberi tugas ke Staf", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{
			ID:   3,
			Nama: "Mas Staf",
			Role: "staf",
		}
		mockUserRepo.On("FindByID", uint(3)).Return(targetUser, nil)
		mockTaskRepo.On("Create", mock.Anything).Return(nil)
		mockNotifRepo.On("Create", mock.Anything).Return(nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		// Execute
		req := CreateTaskRequest{
			TargetUserID: 3,
			JudulTugas:   "Input Data Warga",
			Deskripsi:    "Input data warga baru ke sistem",
		}
		tugas, err := taskSvc.CreateTask(2, "sekertaris", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tugas)
		mockNotifRepo.AssertNumberOfCalls(t, "Create", 1)
		mockTaskRepo.AssertExpectations(t)
	})
}

func TestCreateTask_Fail_StafCannotAssign(t *testing.T) {
	t.Run("Gagal: Staf tidak boleh memberi tugas", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		// Execute
		req := CreateTaskRequest{
			TargetUserID: 1,
			JudulTugas:   "Tugas dari staf",
		}
		tugas, err := taskSvc.CreateTask(3, "staf", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "hanya Lurah dan Sekertaris yang boleh memberi tugas", err.Error())

		// Tidak ada interaksi ke repo
		mockTaskRepo.AssertNotCalled(t, "Create")
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}

func TestCreateTask_Fail_EmptyJudul(t *testing.T) {
	t.Run("Gagal: Judul tugas kosong", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			TargetUserID: 2,
			JudulTugas:   "", // kosong
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "judul_tugas wajib diisi", err.Error())
	})
}

func TestCreateTask_Fail_AssignToSelf(t *testing.T) {
	t.Run("Gagal: Tidak boleh memberi tugas ke diri sendiri", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			TargetUserID: 1, // sama dengan requesterID
			JudulTugas:   "Tugas untuk diri sendiri",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "tidak dapat memberi tugas ke diri sendiri", err.Error())
	})
}

func TestCreateTask_Fail_LurahToStaf(t *testing.T) {
	t.Run("Gagal: Lurah tidak boleh memberi tugas ke Staf (harus lewat Sekertaris)", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: target user adalah staf
		targetUser := &domain.User{
			ID:   3,
			Nama: "Mas Staf",
			Role: "staf",
		}
		mockUserRepo.On("FindByID", uint(3)).Return(targetUser, nil)

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			TargetUserID: 3,
			JudulTugas:   "Tugas langsung ke staf",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "Lurah hanya boleh memberi tugas ke Sekertaris dan Kasi", err.Error())
		mockTaskRepo.AssertNotCalled(t, "Create")
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}

func TestCreateTask_Fail_TargetUserNotFound(t *testing.T) {
	t.Run("Gagal: User target tidak ditemukan di database", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: FindByID mengembalikan error
		mockUserRepo.On("FindByID", uint(999)).Return(nil, errors.New("record not found"))

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			TargetUserID: 999,
			JudulTugas:   "Tugas ke user yang tidak ada",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Equal(t, "user target tidak ditemukan", err.Error())
	})
}

func TestCreateTask_Fail_DBError_StillNoNotification(t *testing.T) {
	t.Run("Gagal: Error saat simpan tugas — notifikasi tidak terkirim", func(t *testing.T) {
		// Setup
		mockTaskRepo := new(mocks.TaskRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{
			ID:   2,
			Nama: "Aep Saepudin, S.Kom",
			Role: "sekertaris",
		}
		mockUserRepo.On("FindByID", uint(2)).Return(targetUser, nil)
		// Mock: simpan tugas gagal
		mockTaskRepo.On("Create", mock.Anything).Return(errors.New("database error"))

		taskSvc := NewTaskService(mockTaskRepo, mockUserRepo, mockNotifRepo)

		req := CreateTaskRequest{
			TargetUserID: 2,
			JudulTugas:   "Tugas yang gagal disimpan",
		}
		tugas, err := taskSvc.CreateTask(1, "lurah", req)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, tugas)
		assert.Contains(t, err.Error(), "gagal menyimpan tugas")
		// Notifikasi TIDAK boleh terkirim jika tugas gagal disimpan
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}
