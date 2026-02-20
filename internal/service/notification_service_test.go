package service

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository/mocks"
)

// ============================================================
// Test GetMyNotifications (NotificationService)
// ============================================================

func TestGetMyNotifications_Success(t *testing.T) {
	t.Run("Sukses: Mengembalikan list notifikasi milik user", func(t *testing.T) {
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		expectedNotifs := []domain.Notification{
			{
				ID:        1,
				UserID:    1,
				Kategori:  "Tugas",
				Judul:     "Tugas Baru Ditetapkan",
				Pesan:     "Anda telah ditugaskan untuk 'Menyusun Laporan Bulanan'.",
				IsRead:    false,
				TerkaitID: 5,
				CreatedAt: time.Now(),
			},
			{
				ID:        2,
				UserID:    1,
				Kategori:  "Penilaian",
				Judul:     "Penilaian Kinerja Baru",
				Pesan:     "Atasan Anda telah memberikan penilaian kinerja untuk periode Bulanan.",
				IsRead:    true,
				TerkaitID: 3,
				CreatedAt: time.Now().Add(-24 * time.Hour),
			},
			{
				ID:        3,
				UserID:    1,
				Kategori:  "Sistem",
				Judul:     "Pengingat Pelaporan",
				Pesan:     "Halo, Anda belum mengisi laporan kinerja untuk hari ini.",
				IsRead:    false,
				TerkaitID: 0,
				CreatedAt: time.Now().Add(-2 * time.Hour),
			},
		}

		mockNotifRepo.On("FindByUserID", 1).Return(expectedNotifs, nil)

		notifSvc := NewNotificationService(mockNotifRepo)

		// Execute
		notifs, err := notifSvc.GetMyNotifications(1)

		// Assert
		assert.NoError(t, err)
		assert.Len(t, notifs, 3)
		assert.Equal(t, "Tugas Baru Ditetapkan", notifs[0].Judul)
		assert.Equal(t, "Penilaian Kinerja Baru", notifs[1].Judul)
		assert.Equal(t, "Pengingat Pelaporan", notifs[2].Judul)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestGetMyNotifications_Success_Empty(t *testing.T) {
	t.Run("Sukses: User belum punya notifikasi (empty list)", func(t *testing.T) {
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		mockNotifRepo.On("FindByUserID", 99).Return([]domain.Notification{}, nil)

		notifSvc := NewNotificationService(mockNotifRepo)

		notifs, err := notifSvc.GetMyNotifications(99)

		assert.NoError(t, err)
		assert.Empty(t, notifs)
		assert.Len(t, notifs, 0)
		mockNotifRepo.AssertExpectations(t)
	})
}

func TestGetMyNotifications_Fail_DBError(t *testing.T) {
	t.Run("Gagal: Error database saat query notifikasi", func(t *testing.T) {
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		mockNotifRepo.On("FindByUserID", 1).Return(nil, errors.New("database connection lost"))

		notifSvc := NewNotificationService(mockNotifRepo)

		notifs, err := notifSvc.GetMyNotifications(1)

		assert.Error(t, err)
		assert.Nil(t, notifs)
		assert.Equal(t, "database connection lost", err.Error())
	})
}

// ============================================================
// Test ReadNotification (NotificationService)
// ============================================================

func TestReadNotification_Success(t *testing.T) {
	t.Run("Sukses: Menandai notifikasi sebagai dibaca", func(t *testing.T) {
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: MarkAsRead berhasil (user 1 menandai notif 5 sebagai dibaca)
		mockNotifRepo.On("MarkAsRead", 5, 1).Return(nil)

		notifSvc := NewNotificationService(mockNotifRepo)

		err := notifSvc.ReadNotification(5, 1)

		assert.NoError(t, err)
		mockNotifRepo.AssertExpectations(t)
		mockNotifRepo.AssertNumberOfCalls(t, "MarkAsRead", 1)
	})
}

func TestReadNotification_Fail_NotFound(t *testing.T) {
	t.Run("Gagal: Notifikasi tidak ditemukan atau bukan milik user", func(t *testing.T) {
		mockNotifRepo := new(mocks.NotificationRepositoryMock)

		// Mock: MarkAsRead mengembalikan error (notif tidak ditemukan / bukan punya user)
		mockNotifRepo.On("MarkAsRead", 999, 1).Return(errors.New("record not found"))

		notifSvc := NewNotificationService(mockNotifRepo)

		err := notifSvc.ReadNotification(999, 1)

		assert.Error(t, err)
		assert.Equal(t, "record not found", err.Error())
	})
}
