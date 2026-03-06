package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/service"
)

// TaskHandler menangani request tugas organisasi.
type TaskHandler struct {
	taskService service.TaskService
}

// NewTaskHandler membuat instance baru TaskHandler.
func NewTaskHandler(taskService service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create menangani pembuatan tugas organisasi baru (Lurah only).
func (h *TaskHandler) Create(c fiber.Ctx) error {
	// 1. Ambil requester dari JWT Token
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

	// 2. Parse Body (Mendukung JSON dan Multipart/Form)
	var req service.CreateOrganizationalTaskRequest
	contentType := string(c.Get("Content-Type"))

	if strings.Contains(contentType, "application/json") {
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format request tidak valid: " + err.Error(),
			})
		}
	} else {
		// Multipart/Form data
		req.JudulTugas = c.FormValue("judul_tugas")
		req.Deskripsi = c.FormValue("deskripsi")
		req.Deadline = c.FormValue("deadline")
		req.FileBukti = c.FormValue("file_bukti")

		// Handle TargetUserIDs (slice)
		form, err := c.MultipartForm()
		if err == nil {
			ids := form.Value["target_user_ids"]
			for _, idStr := range ids {
				if id, err := strconv.Atoi(idStr); err == nil {
					req.TargetUserIDs = append(req.TargetUserIDs, id)
				}
			}
		}

		// Handle File
		file, _ := c.FormFile("file_bukti")
		req.FileHeader = file
	}

	// 3. Validasi input wajib
	if req.JudulTugas == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "judul_tugas wajib diisi",
		})
	}
	if len(req.TargetUserIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "target_user_ids wajib diisi",
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

	// 5. Susun response
	responseData := fiber.Map{
		"id":          tugas.ID,
		"judul_tugas": tugas.JudulTugas,
		"deskripsi":   tugas.Deskripsi,
		"file_bukti":  tugas.FileBukti,
		"created_by":  tugas.CreatedBy,
		"created_at":  tugas.CreatedAt,
	}

	var assigneeList []fiber.Map
	for _, a := range tugas.Assignees {
		assigneeList = append(assigneeList, fiber.Map{
			"id":   a.ID,
			"nama": a.Nama,
			"nip":  a.NIP,
		})
	}
	responseData["assignees"] = assigneeList

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Tugas organisasi berhasil dibuat",
		"data":    responseData,
	})
}

// GetMyTasks menangani pegawai untuk melihat tugas organisasi yang di-assign padanya.
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
	tasks, err := h.taskService.GetMyTasks(userID)
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
		"message": "Daftar tugas organisasi berhasil diambil",
		"data":    responseData,
	})
}

// GetAll menangani Lurah untuk melihat seluruh tugas organisasi.
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
		"message": "Daftar seluruh tugas organisasi berhasil diambil",
		"data":    responseData,
	})
}

// GetByID menangani pengambilan detail tugas organisasi berdasarkan ID.
func (h *TaskHandler) GetByID(c fiber.Ctx) error {
	taskID, err := strconv.Atoi(c.Params("id"))
	if err != nil || taskID <= 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tugas tidak valid",
		})
	}

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

	task, err := h.taskService.GetTaskByID(requesterID, requesterRole, uint(taskID))
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Format response detail penugasan
	creatorName := ""
	if task.Creator != nil {
		creatorName = task.Creator.Nama
	}

	var deadline interface{}
	if task.Deadline != nil {
		deadline = task.Deadline.Format("2006-01-02 15:04:05")
	} else {
		deadline = nil
	}

	responseData := fiber.Map{
		"judul_tugas":        task.JudulTugas,
		"deskripsi":          task.Deskripsi,
		"file_pendukung":     task.FileBukti,
		"deadline":           deadline,
		"nama_pemberi_tugas": creatorName,
	}

	// Optional: add assignees to detail for completeness (though not explicitly requested, helpful for tracing)
	var assigneeList []fiber.Map
	for _, a := range task.Assignees {
		assigneeList = append(assigneeList, fiber.Map{
			"id":   a.ID,
			"nama": a.Nama,
			"nip":  a.NIP,
		})
	}
	responseData["assignees"] = assigneeList

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Detail tugas berhasil diambil",
		"data":    responseData,
	})
}

// Update menangani perubahaan tugas organisasi (Lurah only).
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

	// 3. Parse Body (Mendukung JSON dan Multipart/Form)
	var req service.UpdateOrganizationalTaskRequest
	contentType := string(c.Get("Content-Type"))

	if strings.Contains(contentType, "application/json") {
		if err := c.Bind().JSON(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Format request tidak valid: " + err.Error(),
			})
		}
	} else {
		// Multipart/Form data
		req.JudulTugas = c.FormValue("judul_tugas")
		req.Deskripsi = c.FormValue("deskripsi")
		req.Deadline = c.FormValue("deadline")
		req.FileBukti = c.FormValue("file_bukti")

		// Handle TargetUserIDs (slice)
		form, err := c.MultipartForm()
		if err == nil {
			ids := form.Value["target_user_ids"]
			for _, idStr := range ids {
				if id, err := strconv.Atoi(idStr); err == nil {
					req.TargetUserIDs = append(req.TargetUserIDs, id)
				}
			}
		}

		// Handle File
		file, _ := c.FormFile("file_bukti")
		req.FileHeader = file
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
		"judul_tugas": updatedTask.JudulTugas,
		"deskripsi":   updatedTask.Deskripsi,
		"file_bukti":  updatedTask.FileBukti,
		"created_by":  updatedTask.CreatedBy,
	}

	var assigneeList []fiber.Map
	for _, a := range updatedTask.Assignees {
		assigneeList = append(assigneeList, fiber.Map{
			"id":   a.ID,
			"nama": a.Nama,
			"nip":  a.NIP,
		})
	}
	responseData["assignees"] = assigneeList

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Tugas organisasi berhasil diperbarui",
		"data":    responseData,
	})
}

// Delete menangani penghapusan tugas organisasi (Lurah only).
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
		"message": "Tugas organisasi berhasil dihapus",
	})
}

// mapTasksToResponse mengonversi slice TugasOrganisasi ke format response JSON.
func (h *TaskHandler) mapTasksToResponse(tasks []domain.TugasOrganisasi) []fiber.Map {
	var responseData []fiber.Map
	for _, t := range tasks {
		var deadline interface{}
		if t.Deadline != nil {
			deadline = t.Deadline.Format("2006-01-02 15:04:05")
		} else {
			deadline = nil
		}

		item := fiber.Map{
			"id":             t.ID,
			"judul_tugas":    t.JudulTugas,
			"deskripsi":      t.Deskripsi,
			"file_pendukung": t.FileBukti, // Adjust name as requested for detail API, or keep both
			"file_bukti":     t.FileBukti, // keep for backward compatibility
			"deadline":       deadline,
			"created_by":     t.CreatedBy,
			"date":           t.CreatedAt.Format("2006-01-02"),
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

		// Sertakan daftar assignees
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

		responseData = append(responseData, item)
	}
	return responseData
}
