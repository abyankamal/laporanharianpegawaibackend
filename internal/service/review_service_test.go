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
// Test SubmitReview (ReviewService)
// ============================================================

func TestSubmitReview_Success_LurahToSekertaris(t *testing.T) {
	t.Run("Sukses: Lurah menilai Sekertaris + notifikasi terkirim", func(t *testing.T) {
		// Setup
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: target user adalah sekertaris
		targetUser := &domain.User{
			ID:   2,
			Nama: "Aep Saepudin, S.Kom",
			Role: "sekertaris",
		}
		mockUserRepo.On("FindByID", uint(2)).Return(targetUser, nil)

		// Mock: check existing review returns false
		mockReviewRepo.On("CheckExistingReview", uint(2), 2, 2026).Return(false, nil)

		// Mock: simpan penilaian berhasil
		mockReviewRepo.On("Create", mock.Anything).Return(nil)

		// Mock: simpan notifikasi berhasil
		mockNotifRepo.On("Create", mock.Anything).Return(nil)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		// Execute
		req := CreateReviewRequest{
			TargetUserID:   2,
			SkorID:         2,
			JenisPeriode:   "Bulanan",
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-01",
			TanggalSelesai: "2026-02-28",
			Catatan:        "Kinerja baik bulan ini",
		}
		penilaian, err := reviewSvc.SubmitReview(1, "lurah", req)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, penilaian)
		assert.Equal(t, "Bulanan", penilaian.JenisPeriode)
		assert.Equal(t, "Kinerja baik bulan ini", penilaian.Catatan)

		// Verifikasi: semua mock terpanggil
		mockReviewRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockNotifRepo.AssertExpectations(t)

		// Verifikasi: NotificationRepo.Create() dipanggil tepat 1 kali
		mockNotifRepo.AssertNumberOfCalls(t, "Create", 1)
	})
}

func TestSubmitReview_Success_SekertarisToStaf(t *testing.T) {
	t.Run("Sukses: Sekertaris menilai Staf", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{
			ID:   3,
			Nama: "Mas Staf",
			Role: "staf",
		}
		mockUserRepo.On("FindByID", uint(3)).Return(targetUser, nil)
		mockReviewRepo.On("CheckExistingReview", uint(3), 2, 2026).Return(false, nil)
		mockReviewRepo.On("Create", mock.Anything).Return(nil)
		mockNotifRepo.On("Create", mock.Anything).Return(nil)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   3,
			SkorID:         3,
			JenisPeriode:   "Bulanan",
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-20",
			TanggalSelesai: "2026-02-20",
			Catatan:        "Sangat rajin hari ini",
		}

		penilaian, err := reviewSvc.SubmitReview(2, "sekertaris", req)

		assert.NoError(t, err)
		assert.NotNil(t, penilaian)
		mockNotifRepo.AssertNumberOfCalls(t, "Create", 1)
	})
}

func TestSubmitReview_Fail_SelfReview(t *testing.T) {
	t.Run("Gagal: Tidak boleh menilai diri sendiri", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   1, // sama dengan penilaiID
			SkorID:         2,
			JenisPeriode:   "Bulanan",
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-01",
			TanggalSelesai: "2026-02-28",
		}
		penilaian, err := reviewSvc.SubmitReview(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, penilaian)
		assert.Equal(t, "tidak dapat menilai diri sendiri", err.Error())
		mockReviewRepo.AssertNotCalled(t, "Create")
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}

func TestSubmitReview_Fail_InvalidDateRange(t *testing.T) {
	t.Run("Gagal: Tanggal mulai lebih besar dari tanggal selesai", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{
			ID:   2,
			Nama: "Aep Saepudin, S.Kom",
			Role: "sekertaris",
		}
		mockUserRepo.On("FindByID", uint(2)).Return(targetUser, nil)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   2,
			SkorID:         2,
			JenisPeriode:   "Bulanan", // Must be Bulanan to reach Date validation
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-28", // lebih besar
			TanggalSelesai: "2026-02-01", // lebih kecil
		}
		penilaian, err := reviewSvc.SubmitReview(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, penilaian)
		assert.Equal(t, "tanggal_mulai tidak boleh lebih besar dari tanggal_selesai", err.Error())
		mockReviewRepo.AssertNotCalled(t, "Create")
	})
}

func TestSubmitReview_Fail_StafCannotReview(t *testing.T) {
	t.Run("Gagal: Staf tidak boleh melakukan penilaian", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   1,
			SkorID:         2,
			JenisPeriode:   "Bulanan",
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-01",
			TanggalSelesai: "2026-02-28",
		}
		penilaian, err := reviewSvc.SubmitReview(3, "staf", req)

		assert.Error(t, err)
		assert.Nil(t, penilaian)
		assert.Equal(t, "hanya Lurah dan Sekertaris yang boleh melakukan penilaian", err.Error())
	})
}

func TestSubmitReview_Fail_InvalidPeriode(t *testing.T) {
	t.Run("Gagal: JenisPeriode tidak valid", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{ID: 2, Role: "sekertaris"}
		mockUserRepo.On("FindByID", uint(2)).Return(targetUser, nil)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   2,
			SkorID:         2,
			JenisPeriode:   "Tahunan", // tidak valid
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-01",
			TanggalSelesai: "2026-02-28",
		}
		penilaian, err := reviewSvc.SubmitReview(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, penilaian)
		assert.Equal(t, "hanya periode 'Bulanan' yang didukung untuk penilaian kinerja", err.Error())
	})
}

func TestSubmitReview_Fail_LurahToStaf(t *testing.T) {
	t.Run("Gagal: Lurah tidak boleh langsung menilai Staf", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{ID: 3, Nama: "Mas Staf", Role: "staf"}
		mockUserRepo.On("FindByID", uint(3)).Return(targetUser, nil)

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   3,
			SkorID:         2,
			JenisPeriode:   "Bulanan",
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-01",
			TanggalSelesai: "2026-02-28",
		}
		penilaian, err := reviewSvc.SubmitReview(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, penilaian)
		assert.Equal(t, "Lurah hanya boleh menilai Sekertaris dan Kasi", err.Error())
	})
}

func TestSubmitReview_Fail_DBError_NoNotification(t *testing.T) {
	t.Run("Gagal: DB error saat simpan — notifikasi tidak terkirim", func(t *testing.T) {
		mockReviewRepo := new(mocks.ReviewRepositoryMock)
		mockUserRepo := new(mocks.UserRepositoryMock)
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		targetUser := &domain.User{ID: 2, Role: "sekertaris"}
		mockUserRepo.On("FindByID", uint(2)).Return(targetUser, nil)
		mockReviewRepo.On("CheckExistingReview", uint(2), 2, 2026).Return(false, nil)
		mockReviewRepo.On("Create", mock.Anything).Return(errors.New("database error"))

		reviewSvc := NewReviewService(mockReviewRepo, mockUserRepo, mockNotifRepo)

		req := CreateReviewRequest{
			TargetUserID:   2,
			SkorID:         2,
			JenisPeriode:   "Bulanan",
			Bulan:          2,
			Tahun:          2026,
			TanggalMulai:   "2026-02-01",
			TanggalSelesai: "2026-02-28",
		}
		penilaian, err := reviewSvc.SubmitReview(1, "lurah", req)

		assert.Error(t, err)
		assert.Nil(t, penilaian)
		assert.Contains(t, err.Error(), "gagal menyimpan penilaian")
		mockNotifRepo.AssertNotCalled(t, "Create")
	})
}
