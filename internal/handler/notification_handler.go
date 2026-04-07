package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// NotificationHandler menangani request notifikasi.
type NotificationHandler struct {
	notifService service.NotificationService
}

// NewNotificationHandler membuat instance baru NotificationHandler.
func NewNotificationHandler(notifService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{notifService: notifService}
}

// GetMy menangani request untuk mengambil semua notifikasi milik user yang sedang login.
func (h *NotificationHandler) GetMy(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token (via Locals dari middleware)
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := int(userIDFloat)

	// 2. Panggil service
	notifications, err := h.notifService.GetMyNotifications(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil notifikasi: " + err.Error(),
		})
	}

	// 3. Return response sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Notifikasi berhasil diambil",
		"data":    notifications,
	})
}

// GetByID menangani request untuk mengambil satu notifikasi spesifik berdasarkan ID.
// Route: GET /api/notifications/:id
func (h *NotificationHandler) GetByID(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := int(userIDFloat)

	// 2. Ambil ID notifikasi dari parameter URL
	notifID, err := strconv.Atoi(c.Params("id"))
	if err != nil || notifID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID notifikasi tidak valid",
		})
	}

	// 3. Panggil service
	notification, err := h.notifService.GetNotificationByID(notifID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Notifikasi tidak ditemukan atau bukan milik Anda",
		})
	}

	// 4. Return response sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Notifikasi berhasil diambil",
		"data":    notification,
	})
}

// MarkRead menangani request untuk menandai notifikasi sebagai sudah dibaca.
// Route: PUT /api/notifications/:id/read
func (h *NotificationHandler) MarkRead(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := int(userIDFloat)

	// 2. Ambil ID notifikasi dari parameter URL
	notifID, err := strconv.Atoi(c.Params("id"))
	if err != nil || notifID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID notifikasi tidak valid",
		})
	}

	// 3. Panggil service
	err = h.notifService.ReadNotification(notifID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "Notifikasi tidak ditemukan atau bukan milik Anda",
		})
	}

	// 4. Return response sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Notifikasi berhasil ditandai sebagai sudah dibaca",
	})
}
