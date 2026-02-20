package service

import (
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// NotificationService adalah interface untuk operasi bisnis Notifikasi.
type NotificationService interface {
	GetMyNotifications(userID int) ([]domain.Notification, error)
	ReadNotification(notifID int, userID int) error
}

// notificationService adalah implementasi dari NotificationService.
type notificationService struct {
	notifRepo repository.NotificationRepository
}

// NewNotificationService membuat instance baru NotificationService.
func NewNotificationService(notifRepo repository.NotificationRepository) NotificationService {
	return &notificationService{notifRepo: notifRepo}
}

// GetMyNotifications mengambil semua notifikasi milik user.
func (s *notificationService) GetMyNotifications(userID int) ([]domain.Notification, error) {
	return s.notifRepo.FindByUserID(userID)
}

// ReadNotification menandai notifikasi sebagai sudah dibaca.
func (s *notificationService) ReadNotification(notifID int, userID int) error {
	return s.notifRepo.MarkAsRead(notifID, userID)
}
