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
	FotoPath     *string          `json:"foto_path"`
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

// UserModelResponse adalah struct response untuk user (sesuai format frontend).
type UserModelResponse struct {
	ID          uint    `json:"id"`
	NamaLengkap string  `json:"nama_lengkap"`
	Jabatan     string  `json:"jabatan"`
	NIP         string  `json:"nip"`
	FotoUser    *string `json:"foto_user"`
}

// UserHandler menangani request user management.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler membuat instance baru UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ProfileResponse adalah struct response untuk endpoint GET /api/profile.
type ProfileResponse struct {
	ID           uint    `json:"id"`
	NIP          string  `json:"nip"`
	Nama         string  `json:"nama"`
	Role         string  `json:"role"`
	FotoPath     *string `json:"foto_path"`
	JabatanID    *uint   `json:"jabatan_id"`
	NamaJabatan  string  `json:"nama_jabatan"`
	SupervisorID *uint   `json:"supervisor_id"`
	NamaAtasan   string  `json:"nama_atasan"`
	CreatedAt    string  `json:"created_at"`
}

// GetProfile menangani request profil user yang sedang login.
func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := uint(userIDFloat)

	// 2. Query user dari database (dengan preload Jabatan & Supervisor)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak ditemukan",
		})
	}

	// 3. Siapkan response
	profile := ProfileResponse{
		ID:           user.ID,
		NIP:          user.NIP,
		Nama:         user.Nama,
		Role:         user.Role,
		FotoPath:     user.FotoPath,
		JabatanID:    user.JabatanID,
		SupervisorID: user.SupervisorID,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Isi nama jabatan jika ada
	if user.Jabatan != nil {
		profile.NamaJabatan = user.Jabatan.NamaJabatan
	}

	// Isi nama atasan jika ada
	if user.Supervisor != nil {
		profile.NamaAtasan = user.Supervisor.Nama
	}

	// 4. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data profil berhasil diambil",
		"data":    profile,
	})
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
	var response []UserModelResponse
	for _, user := range users {
		jabatanName := ""
		if user.Jabatan != nil {
			jabatanName = user.Jabatan.NamaJabatan
		}

		userResp := UserModelResponse{
			ID:          user.ID,
			NamaLengkap: user.Nama,
			Jabatan:     jabatanName,
			NIP:         user.NIP,
			FotoUser:    user.FotoPath,
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

	jabatanName := ""
	if user.Jabatan != nil {
		jabatanName = user.Jabatan.NamaJabatan
	}

	// Map ke response (tanpa password)
	response := UserModelResponse{
		ID:          user.ID,
		NamaLengkap: user.Nama,
		Jabatan:     jabatanName,
		NIP:         user.NIP,
		FotoUser:    user.FotoPath,
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

// ChangePassword mengubah password user yang sedang login.
func (h *UserHandler) ChangePassword(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token (via Locals dari middleware)
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := uint(userIDFloat)

	// 2. Parse JSON Body
	var req service.ChangePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid: " + err.Error(),
		})
	}

	// 3. Validasi input wajib
	if req.OldPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "password lama wajib diisi",
		})
	}
	if req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "password baru wajib diisi",
		})
	}

	// 4. Panggil service
	err := h.userService.ChangePassword(userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 5. Return response sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Password berhasil diubah",
	})
}

// ChangePhoto mengubah foto profil user yang sedang login.
func (h *UserHandler) ChangePhoto(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := uint(userIDFloat)

	// 2. Ambil file dari form
	fileHeader, err := c.FormFile("foto")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "File foto wajib diupload",
		})
	}

	// 3. Panggil service
	fotoPath, err := h.userService.UpdateProfilePhoto(userID, fileHeader)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 4. Return response sukses
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Foto profil berhasil diubah",
		"data": fiber.Map{
			"foto_path": fotoPath,
		},
	})
}

// GetSupervisors menangani request untuk mengambil daftar atasan yang sesuai berdasarkan role.
func (h *UserHandler) GetSupervisors(c fiber.Ctx) error {
	role := c.Query("role")
	if role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Parameter role wajib diisi",
		})
	}

	supervisors, err := h.userService.GetSupervisorsForRole(role)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	var response []UserModelResponse
	for _, s := range supervisors {
		jabatanName := ""
		if s.Jabatan != nil {
			jabatanName = s.Jabatan.NamaJabatan
		}

		response = append(response, UserModelResponse{
			ID:          s.ID,
			NamaLengkap: s.Nama,
			Jabatan:     jabatanName,
			NIP:         s.NIP,
			FotoUser:    s.FotoPath,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Data atasan berhasil diambil",
		"data":    response,
	})
}
