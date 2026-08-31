package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// JabatanResponse adalah struct response untuk jabatan.
type JabatanResponse struct {
	ID          uint   `json:"id"`
	NamaJabatan string `json:"nama_jabatan"`
}

// UserResponse adalah struct response standar untuk user (tanpa password).
type UserResponse struct {
	ID           uint             `json:"id"`
	NIP          string           `json:"nip"`
	Nama         string           `json:"nama"`
	NamaLengkap  string           `json:"nama_lengkap"` // Alias untuk backward compatibility
	Role         string           `json:"role"`
	FotoPath     *string          `json:"foto_path"`
	FotoUser     *string          `json:"foto_user"` // Alias untuk backward compatibility
	JabatanID    *uint            `json:"jabatan_id"`
	Jabatan      *JabatanResponse `json:"jabatan,omitempty"`
	NamaJabatan  string           `json:"nama_jabatan,omitempty"`
	SupervisorID *uint            `json:"supervisor_id"`
	NamaAtasan   string           `json:"nama_atasan,omitempty"`
	CreatedAt    string           `json:"created_at,omitempty"`
}

// UserModelResponse adalah alias tipe untuk backward compatibility handler/test lama.
type UserModelResponse = UserResponse

// ProfileResponse adalah alias tipe untuk response profil.
type ProfileResponse = UserResponse

// UserHandler menangani request user management.
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler membuat instance baru UserHandler.
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile menangani request profil user yang sedang login.
func (h *UserHandler) GetProfile(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	userID := uint(userIDFloat)

	// 2. Query user dari database (dengan preload Jabatan & Supervisor)
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		return SendError(c, fiber.StatusNotFound, "User tidak ditemukan")
	}

	// 3. Siapkan response
	profile := UserResponse{
		ID:           user.ID,
		NIP:          user.NIP,
		Nama:         user.Nama,
		NamaLengkap:  user.Nama,
		Role:         user.Role,
		FotoPath:     user.FotoPath,
		FotoUser:     user.FotoPath,
		JabatanID:    user.JabatanID,
		SupervisorID: user.SupervisorID,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// Isi nama jabatan jika ada
	if user.Jabatan != nil {
		profile.NamaJabatan = user.Jabatan.NamaJabatan
		profile.Jabatan = &JabatanResponse{
			ID:          user.Jabatan.ID,
			NamaJabatan: user.Jabatan.NamaJabatan,
		}
	}

	// Isi nama atasan jika ada
	if user.Supervisor != nil {
		profile.NamaAtasan = user.Supervisor.Nama
	}

	// 4. Return response
	return SendSuccess(c, fiber.StatusOK, "Data profil berhasil diambil", profile)
}

// GetAll mengambil semua user.
func (h *UserHandler) GetAll(c fiber.Ctx) error {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	// Map ke response (tanpa password)
	var response []UserResponse
	for _, user := range users {
		jabatanName := ""
		var jabatanResp *JabatanResponse
		if user.Jabatan != nil {
			jabatanName = user.Jabatan.NamaJabatan
			jabatanResp = &JabatanResponse{
				ID:          user.Jabatan.ID,
				NamaJabatan: user.Jabatan.NamaJabatan,
			}
		}

		userResp := UserResponse{
			ID:           user.ID,
			NIP:          user.NIP,
			Nama:         user.Nama,
			NamaLengkap:  user.Nama,
			Role:         user.Role,
			FotoPath:     user.FotoPath,
			FotoUser:     user.FotoPath,
			JabatanID:    user.JabatanID,
			Jabatan:      jabatanResp,
			NamaJabatan:  jabatanName,
			SupervisorID: user.SupervisorID,
			CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
		}

		if user.Supervisor != nil {
			userResp.NamaAtasan = user.Supervisor.Nama
		}

		response = append(response, userResp)
	}

	return SendSuccess(c, fiber.StatusOK, "Data user berhasil diambil", response)
}

// GetOne mengambil detail user berdasarkan ID.
func (h *UserHandler) GetOne(c fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	user, err := h.userService.GetUserByID(uint(id))
	if err != nil {
		return SendError(c, fiber.StatusNotFound, err.Error())
	}

	jabatanName := ""
	var jabatanResp *JabatanResponse
	if user.Jabatan != nil {
		jabatanName = user.Jabatan.NamaJabatan
		jabatanResp = &JabatanResponse{
			ID:          user.Jabatan.ID,
			NamaJabatan: user.Jabatan.NamaJabatan,
		}
	}

	// Map ke response (tanpa password)
	response := UserResponse{
		ID:           user.ID,
		NIP:          user.NIP,
		Nama:         user.Nama,
		NamaLengkap:  user.Nama,
		Role:         user.Role,
		FotoPath:     user.FotoPath,
		FotoUser:     user.FotoPath,
		JabatanID:    user.JabatanID,
		Jabatan:      jabatanResp,
		NamaJabatan:  jabatanName,
		SupervisorID: user.SupervisorID,
		CreatedAt:    user.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if user.Supervisor != nil {
		response.NamaAtasan = user.Supervisor.Nama
	}

	return SendSuccess(c, fiber.StatusOK, "Data user berhasil diambil", response)
}

// Create membuat user baru.
func (h *UserHandler) Create(c fiber.Ctx) error {
	var req service.CreateUserRequest

	// Parse body request
	if err := c.Bind().Body(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	user, err := h.userService.CreateUser(req)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	return SendSuccess(c, fiber.StatusCreated, "User berhasil dibuat", UserResponse{
		ID:           user.ID,
		Nama:         user.Nama,
		NamaLengkap:  user.Nama,
		NIP:          user.NIP,
		Role:         user.Role,
		JabatanID:    user.JabatanID,
		SupervisorID: user.SupervisorID,
		FotoPath:     user.FotoPath,
		FotoUser:     user.FotoPath,
	})
}

// Update mengupdate data user.
func (h *UserHandler) Update(c fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	requesterRole, _ := c.Locals("role").(string)

	var req service.UpdateUserRequest

	// Parse body request
	if err := c.Bind().Body(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	user, err := h.userService.UpdateUser(uint(id), req, requesterRole)
	if err != nil {
		return ErrorResponse(c, err)
	}

	return SendSuccess(c, fiber.StatusOK, "User berhasil diupdate", UserResponse{
		ID:           user.ID,
		Nama:         user.Nama,
		NamaLengkap:  user.Nama,
		NIP:          user.NIP,
		Role:         user.Role,
		JabatanID:    user.JabatanID,
		SupervisorID: user.SupervisorID,
		FotoPath:     user.FotoPath,
		FotoUser:     user.FotoPath,
	})
}

// Delete menghapus user.
func (h *UserHandler) Delete(c fiber.Ctx) error {
	// Ambil ID dari parameter
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	requesterRole, _ := c.Locals("role").(string)

	err = h.userService.DeleteUser(uint(id), requesterRole)
	if err != nil {
		return ErrorResponse(c, err)
	}

	return SendSuccess(c, fiber.StatusOK, "User berhasil dihapus", nil)
}

// ChangePassword mengubah password user yang sedang login.
func (h *UserHandler) ChangePassword(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token (via Locals dari middleware)
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	userID := uint(userIDFloat)

	// 2. Parse JSON Body
	var req service.ChangePasswordRequest
	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid: "+err.Error())
	}

	// 3. Validasi input wajib
	if req.OldPassword == "" {
		return SendError(c, fiber.StatusBadRequest, "password lama wajib diisi")
	}
	if req.NewPassword == "" {
		return SendError(c, fiber.StatusBadRequest, "password baru wajib diisi")
	}

	// 4. Panggil service
	err := h.userService.ChangePassword(userID, req)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	// 5. Return response sukses
	return SendSuccess(c, fiber.StatusOK, "Password berhasil diubah", nil)
}

// ResetPassword mereset password user oleh admin/sekertaris.
func (h *UserHandler) ResetPassword(c fiber.Ctx) error {
	idParam := c.Params("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "ID tidak valid")
	}

	requesterRole, _ := c.Locals("role").(string)

	var req struct {
		NewPassword string `json:"new_password"`
	}
	_ = c.Bind().JSON(&req)

	err = h.userService.ResetPasswordByAdmin(uint(id), req.NewPassword, requesterRole)
	if err != nil {
		return ErrorResponse(c, err)
	}

	return SendSuccess(c, fiber.StatusOK, "Password berhasil direset oleh admin", nil)
}

// ChangePhoto mengubah foto profil user yang sedang login.
func (h *UserHandler) ChangePhoto(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	userID := uint(userIDFloat)

	// 2. Ambil file dari form
	fileHeader, err := c.FormFile("foto")
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, "File foto wajib diupload")
	}

	// 3. Panggil service
	fotoPath, err := h.userService.UpdateProfilePhoto(userID, fileHeader)
	if err != nil {
		return SendError(c, fiber.StatusBadRequest, err.Error())
	}

	// 4. Return response sukses
	return SendSuccess(c, fiber.StatusOK, "Foto profil berhasil diubah", fiber.Map{
		"foto_path": fotoPath,
		"foto_user": fotoPath,
	})
}

// GetSupervisors menangani request untuk mengambil daftar atasan secara dinamis.
func (h *UserHandler) GetSupervisors(c fiber.Ctx) error {
	roleFilter := c.Query("role")

	supervisors, err := h.userService.GetSupervisors(roleFilter)
	if err != nil {
		return SendError(c, fiber.StatusNotFound, err.Error())
	}

	var response []UserResponse
	for _, s := range supervisors {
		jabatanName := ""
		var jabatanResp *JabatanResponse
		if s.Jabatan != nil {
			jabatanName = s.Jabatan.NamaJabatan
			jabatanResp = &JabatanResponse{
				ID:          s.Jabatan.ID,
				NamaJabatan: s.Jabatan.NamaJabatan,
			}
		}

		response = append(response, UserResponse{
			ID:           s.ID,
			NIP:          s.NIP,
			Nama:         s.Nama,
			NamaLengkap:  s.Nama,
			Role:         s.Role,
			FotoPath:     s.FotoPath,
			FotoUser:     s.FotoPath,
			JabatanID:    s.JabatanID,
			Jabatan:      jabatanResp,
			NamaJabatan:  jabatanName,
			SupervisorID: s.SupervisorID,
		})
	}

	return SendSuccess(c, fiber.StatusOK, "Data atasan berhasil diambil", response)
}

// UpdateFCMToken memperbarui fcm_token untuk user yang sedang login.
func (h *UserHandler) UpdateFCMToken(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	userID := uint(userIDFloat)

	// 2. Parse request body
	var req struct {
		FCMToken string `json:"fcm_token"`
	}

	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	// 3. Panggil service
	err := h.userService.UpdateFCMToken(userID, req.FCMToken)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, err.Error())
	}

	// 4. Return response
	return SendSuccess(c, fiber.StatusOK, "FCM Token berhasil diperbarui", nil)
}
