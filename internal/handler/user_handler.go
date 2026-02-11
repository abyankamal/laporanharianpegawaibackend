package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// UserResponse adalah struct response untuk user (tanpa password).
type UserResponse struct {
	ID           uint             `json:"id"`
	NIP          string           `json:"nip"`
	Nama         string           `json:"nama"`
	Role         string           `json:"role"`
	JabatanID    *uint            `json:"jabatan_id"`
	SupervisorID *uint            `json:"supervisor_id"`
	Jabatan      *JabatanResponse `json:"jabatan,omitempty"`
	CreatedAt    string           `json:"created_at"`
}

// JabatanResponse adalah struct response untuk jabatan.
type JabatanResponse struct {
	ID          uint   `json:"id"`
	NamaJabatan string `json:"nama_jabatan"`
}

// UserHandler menangani request user management.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler membuat instance baru UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetAll mengambil semua user.
func (h *UserHandler) GetAll(c fiber.Ctx) error {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Map ke response (tanpa password)
	var response []UserResponse
	for _, user := range users {
		userResp := UserResponse{
			ID:           user.ID,
			NIP:          user.NIP,
			Nama:         user.Nama,
			Role:         user.Role,
			JabatanID:    user.JabatanID,
			SupervisorID: user.SupervisorID,
			CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		// Map jabatan jika ada
		if user.Jabatan != nil {
			userResp.Jabatan = &JabatanResponse{
				ID:          user.Jabatan.ID,
				NamaJabatan: user.Jabatan.NamaJabatan,
			}
		}

		response = append(response, userResp)
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Data user berhasil diambil",
		"data":    response,
	})
}

// GetOne mengambil detail user berdasarkan ID.
func (h *UserHandler) GetOne(c fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tidak valid",
		})
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// Map ke response (tanpa password)
	response := UserResponse{
		ID:           user.ID,
		NIP:          user.NIP,
		Nama:         user.Nama,
		Role:         user.Role,
		JabatanID:    user.JabatanID,
		SupervisorID: user.SupervisorID,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if user.Jabatan != nil {
		response.Jabatan = &JabatanResponse{
			ID:          user.Jabatan.ID,
			NamaJabatan: user.Jabatan.NamaJabatan,
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Data user berhasil diambil",
		"data":    response,
	})
}

// Create membuat user baru.
func (h *UserHandler) Create(c fiber.Ctx) error {
	var req service.CreateUserRequest

	// Parse body request
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "User berhasil dibuat",
		"data": fiber.Map{
			"id":   user.ID,
			"nip":  user.NIP,
			"nama": user.Nama,
			"role": user.Role,
		},
	})
}

// Update mengupdate data user.
func (h *UserHandler) Update(c fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tidak valid",
		})
	}

	var req service.UpdateUserRequest

	// Parse body request
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid",
		})
	}

	user, err := h.userService.UpdateUser(uint(id), req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User berhasil diupdate",
		"data": fiber.Map{
			"id":   user.ID,
			"nip":  user.NIP,
			"nama": user.Nama,
			"role": user.Role,
		},
	})
}

// Delete menghapus user.
func (h *UserHandler) Delete(c fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "ID tidak valid",
		})
	}

	err = h.userService.DeleteUser(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User berhasil dihapus",
	})
}
