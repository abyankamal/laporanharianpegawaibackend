package handler

import (
	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// TaskHandler menangani request tugas pokok.
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler membuat instance baru TaskHandler.
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create menangani pembuatan tugas pokok baru oleh atasan.
func (h *TaskHandler) Create(c fiber.Ctx) error {
	// 1. Ambil requester dari JWT Token (via Locals dari middleware)
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	requesterID := uint(requesterIDFloat)

	requesterRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// 2. Parse JSON Body
	var req service.CreateTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid: " + err.Error(),
		})
	}

	// 3. Validasi input wajib
	if req.TargetUserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "target_user_id wajib diisi",
		})
	}
	if req.JudulTugas == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "judul_tugas wajib diisi",
		})
	}

	// 4. Panggil service
	tugas, err := h.taskService.CreateTask(requesterID, requesterRole, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 5. Return response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Tugas pokok berhasil dibuat",
		"data": fiber.Map{
			"id":          tugas.ID,
			"user_id":     tugas.UserID,
			"judul_tugas": tugas.JudulTugas,
			"deskripsi":   tugas.Deskripsi,
			"created_by":  tugas.CreatedBy,
			"created_at":  tugas.CreatedAt,
		},
	})
}

// GetMyTasks menangani request pegawai untuk melihat daftar tugas pokok miliknya.
func (h *TaskHandler) GetMyTasks(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := int(userIDFloat)

	// 2. Panggil service
	tasks, err := h.taskService.GetTasksByUserID(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil daftar tugas: " + err.Error(),
		})
	}

	// 3. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Daftar tugas berhasil diambil",
		"data":    tasks,
	})
}
