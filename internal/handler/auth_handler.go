package handler

import (
	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// LoginRequest adalah struktur untuk request body login.
type LoginRequest struct {
	NIP      string `json:"nip"`
	Password string `json:"password"`
}

// AuthHandler menangani request autentikasi.
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler membuat instance baru AuthHandler.
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login menangani proses login user.
func (h *AuthHandler) Login(c fiber.Ctx) error {
	var req LoginRequest

	// Parse body request
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	// Validasi input
	if req.NIP == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "NIP dan password wajib diisi",
		})
	}

	// Panggil service login
	token, err := h.authService.Login(req.NIP, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Return sukses dengan token
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Login berhasil",
		"token":   token,
	})
}
