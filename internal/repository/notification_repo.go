package repository

import (
	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// NotificationRepository adalah interface untuk operasi database Notification.
type NotificationRepository interface {
	Create(notif *domain.Notification) error
	FindByUserID(userID int) ([]domain.Notification, error)
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

// FindByUserID mengambil semua notifikasi milik user tertentu, diurutkan terbaru di atas.
func (r *notificationRepository) FindByUserID(userID int) ([]domain.Notification, error) {
	var notifications []domain.Notification
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&notifications).Error
	if err != nil {
		return nil, err
	}
	return notifications, nil
}

// MarkAsRead menandai notifikasi sebagai sudah dibaca.
// Hanya bisa menandai notifikasi milik user yang bersangkutan (keamanan).
func (r *notificationRepository) MarkAsRead(notifID int, userID int) error {
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
