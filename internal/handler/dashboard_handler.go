package handler

import (
	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// DashboardHandler menangani request dashboard.
type DashboardHandler struct {
	dashboardService service.DashboardService
}

// NewDashboardHandler membuat instance baru DashboardHandler.
func NewDashboardHandler(dashboardService service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashboardService: dashboardService}
}

// GetSummary menangani request ringkasan dashboard.
func (h *DashboardHandler) GetSummary(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := uint(userIDFloat)

	// 2. Ambil role dari JWT Token
	userRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// 3. Panggil service
	summary, err := h.dashboardService.GetSummary(userID, userRole)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data dashboard: " + err.Error(),
		})
	}

	// 4. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"success": true,
		"status":  "success",
		"message": "Data dashboard berhasil diambil",
		"data":    summary,
	})
}
