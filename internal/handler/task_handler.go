package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/domain"
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

// Create menangani pembuatan tugas pokok baru.
// Mendukung dua jenis tugas: "organisasi" (Lurah, multi-assign) dan "individu" (self-assign).
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
	if req.JenisTugas == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "jenis_tugas wajib diisi ('organisasi' atau 'individu')",
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

	// 5. Susun response berdasarkan jenis tugas
	responseData := fiber.Map{
		"id":          tugas.ID,
		"jenis_tugas": tugas.JenisTugas,
		"judul_tugas": tugas.JudulTugas,
		"deskripsi":   tugas.Deskripsi,
		"created_by":  tugas.CreatedBy,
		"created_at":  tugas.CreatedAt,
	}

	if tugas.JenisTugas == "organisasi" {
		// Sertakan daftar assignees
		var assigneeList []fiber.Map
		for _, a := range tugas.Assignees {
			assigneeList = append(assigneeList, fiber.Map{
				"id":   a.ID,
				"nama": a.Nama,
				"nip":  a.NIP,
			})
		}
		responseData["assignees"] = assigneeList
		responseData["file_bukti"] = tugas.FileBukti
	} else {
		responseData["user_id"] = tugas.UserID
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Tugas pokok berhasil dibuat",
		"data":    responseData,
	})
}

// GetMyTasks menangani request pegawai untuk melihat daftar tugas pokok miliknya.
// Menggabungkan tugas individu dan tugas organisasi yang di-assign ke user.
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

	// 3. Map ke format response
	responseData := h.mapTasksToResponse(tasks)

	// 4. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Daftar tugas berhasil diambil",
		"data":    responseData,
	})
}

// GetAll menangani request atasan untuk melihat seluruh daftar tugas pokok.
func (h *TaskHandler) GetAll(c fiber.Ctx) error {
	tasks, err := h.taskService.GetAllTasks()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil daftar tugas: " + err.Error(),
		})
	}

	responseData := h.mapTasksToResponse(tasks)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Daftar tugas berhasil diambil",
		"data":    responseData,
	})
}

// Update menangani request untuk mengubah tugas pokok.
func (h *TaskHandler) Update(c fiber.Ctx) error {
	// 1. Ambil task ID dari URL parameter
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil || taskID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tugas tidak valid",
		})
	}

	// 2. Ambil requester dari JWT Token
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

	// 3. Parse JSON Body
	var req service.UpdateTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid: " + err.Error(),
		})
	}

	// 4. Panggil service
	updatedTask, err := h.taskService.UpdateTask(requesterID, requesterRole, uint(taskID), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 5. Susun response
	responseData := fiber.Map{
		"id":          updatedTask.ID,
		"jenis_tugas": updatedTask.JenisTugas,
		"judul_tugas": updatedTask.JudulTugas,
		"deskripsi":   updatedTask.Deskripsi,
		"created_by":  updatedTask.CreatedBy,
		"file_bukti":  updatedTask.FileBukti,
	}

	if updatedTask.JenisTugas == "organisasi" {
		var assigneeList []fiber.Map
		for _, a := range updatedTask.Assignees {
			assigneeList = append(assigneeList, fiber.Map{
				"id":   a.ID,
				"nama": a.Nama,
				"nip":  a.NIP,
			})
		}
		responseData["assignees"] = assigneeList
	} else {
		responseData["user_id"] = updatedTask.UserID
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Tugas pokok berhasil diperbarui",
		"data":    responseData,
	})
}

// Delete menangani request untuk menghapus tugas pokok.
func (h *TaskHandler) Delete(c fiber.Ctx) error {
	// 1. Ambil task ID dari URL parameter
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil || taskID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tugas tidak valid",
		})
	}

	// 2. Ambil requester dari JWT Token
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

	// 3. Panggil service
	err = h.taskService.DeleteTask(requesterID, requesterRole, uint(taskID))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Tugas pokok berhasil dihapus",
	})
}

// mapTasksToResponse mengonversi slice TugasPokok ke format response JSON yang flat.
func (h *TaskHandler) mapTasksToResponse(tasks []domain.TugasPokok) []fiber.Map {
	var responseData []fiber.Map
	for _, t := range tasks {
		item := fiber.Map{
			"id":          t.ID,
			"jenis_tugas": t.JenisTugas,
			"judul_tugas": t.JudulTugas,
			"deskripsi":   t.Deskripsi,
			"file_bukti":  t.FileBukti,
			"created_by":  t.CreatedBy,
			"date":        t.CreatedAt.Format("2006-01-02"),
		}

		// Info creator
		if t.Creator != nil {
			item["creator_name"] = t.Creator.Nama
			item["creator_nip"] = t.Creator.NIP
			if t.Creator.FotoPath != nil {
				item["creator_avatar"] = *t.Creator.FotoPath
			} else {
				item["creator_avatar"] = ""
			}
		}

		if t.JenisTugas == "organisasi" {
			// Tugas organisasi: sertakan daftar assignees
			var assigneeList []fiber.Map
			for _, a := range t.Assignees {
				aMap := fiber.Map{
					"id":   a.ID,
					"nama": a.Nama,
					"nip":  a.NIP,
				}
				if a.FotoPath != nil {
					aMap["avatar"] = *a.FotoPath
				} else {
					aMap["avatar"] = ""
				}
				assigneeList = append(assigneeList, aMap)
			}
			item["assignees"] = assigneeList
		} else {
			// Tugas individu: sertakan info user tunggal
			item["user_id"] = t.UserID
			if t.User != nil {
				item["assigned_to_name"] = t.User.Nama
				item["assigned_to_nip"] = t.User.NIP
				if t.User.FotoPath != nil {
					item["assigned_to_avatar"] = *t.User.FotoPath
				} else {
					item["assigned_to_avatar"] = ""
				}
			}
		}

		responseData = append(responseData, item)
	}
	return responseData
}
