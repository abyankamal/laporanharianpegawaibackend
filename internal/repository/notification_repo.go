package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// NotificationRepository adalah interface untuk operasi database Notification.
type NotificationRepository interface {
	Create(notif *domain.Notification) error
	FindByUserID(userID int) ([]domain.Notification, error)
	FindByID(id int, userID int) (*domain.Notification, error)
	MarkAsRead(notifID int, userID int) error
}

// notificationRepository adalah implementasi dari NotificationRepository.
type notificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository membuat instance baru NotificationRepository.
func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

// Create menyimpan notifikasi baru ke database.
func (r *notificationRepository) Create(notif *domain.Notification) error {
	return r.db.Create(notif).Error
}

// FindByUserID mengambil semua notifikasi milik user tertentu atau pengumuman global (user_id = 0), diurutkan terbaru di atas.
// Untuk pengumuman global (user_id = 0), status is_read ditentukan dari apakah user sudah memiliki catatan di notification_reads.
func (r *notificationRepository) FindByUserID(userID int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Table("notifications").
		Select("notifications.id, notifications.user_id, notifications.kategori, notifications.judul, notifications.pesan, CASE WHEN notifications.user_id = 0 THEN (CASE WHEN nr.id IS NOT NULL THEN 1 ELSE 0 END) ELSE notifications.is_read END AS is_read, notifications.terkait_id, notifications.created_at").
		Joins("LEFT JOIN notification_reads nr ON nr.notification_id = notifications.id AND nr.user_id = ?", userID).
		Where("notifications.user_id = ? OR notifications.user_id = 0", userID).
		Order("notifications.created_at DESC").
		Scan(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// FindByID mengambil satu notifikasi spesifik milik user tertentu atau pengumuman global.
// Untuk pengumuman global (user_id = 0), status is_read ditentukan dari record notification_reads untuk user terkait.
func (r *notificationRepository) FindByID(id int, userID int) (*domain.Notification, error) {
	var notif domain.Notification
	err := r.db.Table("notifications").
		Select("notifications.id, notifications.user_id, notifications.kategori, notifications.judul, notifications.pesan, CASE WHEN notifications.user_id = 0 THEN (CASE WHEN nr.id IS NOT NULL THEN 1 ELSE 0 END) ELSE notifications.is_read END AS is_read, notifications.terkait_id, notifications.created_at").
		Joins("LEFT JOIN notification_reads nr ON nr.notification_id = notifications.id AND nr.user_id = ?", userID).
		Where("notifications.id = ? AND (notifications.user_id = ? OR notifications.user_id = 0)", id, userID).
		Take(&notif).Error
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

// MarkAsRead menandai notifikasi sebagai sudah dibaca.
// Untuk pengumuman global (user_id = 0), status dibaca dicatat ke tabel notification_reads per-user agar tidak mempengaruhi user lain.
// Untuk notifikasi personal, field is_read di tabel notifications di-update secara langsung.
func (r *notificationRepository) MarkAsRead(notifID int, userID int) error {
	var notif domain.Notification
	err := r.db.Where("id = ? AND (user_id = ? OR user_id = 0)", notifID, userID).Take(&notif).Error
	if err != nil {
		return err
	}

	if notif.UserID == 0 {
		readRecord := domain.NotificationRead{
			UserID:         userID,
			NotificationID: notif.ID,
			ReadAt:         time.Now(),
		}
		return r.db.Where(domain.NotificationRead{
			UserID:         userID,
			NotificationID: notif.ID,
		}).Assign(domain.NotificationRead{ReadAt: time.Now()}).FirstOrCreate(&readRecord).Error
	}

	result := r.db.Model(&domain.Notification{}).
		Where("id = ? AND user_id = ?", notifID, userID).
		Update("is_read", true)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
