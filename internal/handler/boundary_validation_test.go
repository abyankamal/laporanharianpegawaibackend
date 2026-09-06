package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/service"
)

// Mock UserService
type dummyUserService struct {
	mock.Mock
}

func (m *dummyUserService) GetAllUsers() ([]domain.User, error) { return nil, nil }
func (m *dummyUserService) GetUserByID(id uint) (*domain.User, error) { return nil, nil }
func (m *dummyUserService) CreateUser(req service.CreateUserRequest) (*domain.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *dummyUserService) UpdateUser(id uint, req service.UpdateUserRequest, requesterRole string) (*domain.User, error) {
	return nil, nil
}
func (m *dummyUserService) DeleteUser(id uint, requesterRole string) error { return nil }
func (m *dummyUserService) ChangePassword(userID uint, req service.ChangePasswordRequest) error {
	return nil
}
func (m *dummyUserService) ResetPasswordByAdmin(targetUserID uint, newPassword string, requesterRole string) error {
	return nil
}
func (m *dummyUserService) UpdateProfilePhoto(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	return "", nil
}
func (m *dummyUserService) UpdateFCMToken(userID uint, token string) error { return nil }
func (m *dummyUserService) GetSupervisors(roleFilter string) ([]domain.User, error) { return nil, nil }
func (m *dummyUserService) GetUsersByRoles(roles []string) ([]domain.User, error) { return nil, nil }

// Mock TaskService
type dummyTaskService struct {
	mock.Mock
}

func (m *dummyTaskService) CreateTask(requesterID uint, requesterRole string, req service.CreateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error) {
	args := m.Called(requesterID, requesterRole, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TugasOrganisasi), args.Error(1)
}
func (m *dummyTaskService) GetMyTasks(userID int, page, limit int) ([]domain.TugasOrganisasi, int64, error) {
	return nil, 0, nil
}
func (m *dummyTaskService) GetAllTasks(page, limit int) ([]domain.TugasOrganisasi, int64, error) {
	return nil, 0, nil
}
func (m *dummyTaskService) GetTaskByID(requesterID uint, requesterRole string, taskID uint) (*domain.TugasOrganisasi, error) {
	return nil, nil
}
func (m *dummyTaskService) UpdateTask(requesterID uint, requesterRole string, taskID uint, req service.UpdateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error) {
	return nil, nil
}
func (m *dummyTaskService) DeleteTask(requesterID uint, requesterRole string, taskID uint) error {
	return nil
}

// Mock HolidayService
type dummyHolidayService struct {
	mock.Mock
}

func (m *dummyHolidayService) GetHolidays() ([]domain.Holiday, error)           { return nil, nil }
func (m *dummyHolidayService) GetHolidayByID(id uint) (*domain.Holiday, error) { return nil, nil }
func (m *dummyHolidayService) CreateHoliday(tanggalMulaiStr, tanggalSelesaiStr, keterangan string) (*domain.Holiday, error) {
	args := m.Called(tanggalMulaiStr, tanggalSelesaiStr, keterangan)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Holiday), args.Error(1)
}
func (m *dummyHolidayService) UpdateHoliday(id uint, tanggalMulaiStr, tanggalSelesaiStr, keterangan string) (*domain.Holiday, error) {
	return nil, nil
}
func (m *dummyHolidayService) DeleteHoliday(id uint) error { return nil }

// Mock WorkHourService
type dummyWorkHourService struct {
	mock.Mock
}

func (m *dummyWorkHourService) GetWorkHour() (*domain.WorkHour, error) { return nil, nil }
func (m *dummyWorkHourService) UpdateWorkHour(jamMasuk, jamPulang, jamMasukJumat, jamPulangJumat string) (*domain.WorkHour, error) {
	args := m.Called(jamMasuk, jamPulang, jamMasukJumat, jamPulangJumat)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkHour), args.Error(1)
}
func (m *dummyWorkHourService) UpdateGeofencing(kantorLat, kantorLong *string, radiusMeter int, geofencingEnabled bool) (*domain.WorkHour, error) {
	return nil, nil
}

// Mock JabatanService
type dummyJabatanService struct {
	mock.Mock
}

func (m *dummyJabatanService) GetAllJabatan() ([]domain.RefJabatan, error)          { return nil, nil }
func (m *dummyJabatanService) GetJabatanByID(id uint) (*domain.RefJabatan, error) { return nil, nil }
func (m *dummyJabatanService) CreateJabatan(nama string) (*domain.RefJabatan, error) {
	args := m.Called(nama)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefJabatan), args.Error(1)
}
func (m *dummyJabatanService) UpdateJabatan(id uint, nama string) (*domain.RefJabatan, error) {
	return nil, nil
}
func (m *dummyJabatanService) DeleteJabatan(id uint) error { return nil }

// ============================================================
// TESTS
// ============================================================

func TestUserHandler_Create_BoundaryValidation_422(t *testing.T) {
	mockUserSvc := new(dummyUserService)
	h := NewUserHandler(mockUserSvc)

	app := fiber.New()
	app.Post("/users", h.Create)

	// Payload kosong: semua required field tidak ada
	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	_ = json.Unmarshal(body, &res)

	assert.False(t, res.Success)
	assert.Equal(t, "error", res.Status)
	assert.Equal(t, "Validasi input gagal", res.Message)

	details, ok := res.Details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, details, "nip")
	assert.Contains(t, details, "nama")
	assert.Contains(t, details, "password")
	assert.Contains(t, details, "role")

	mockUserSvc.AssertNotCalled(t, "CreateUser")
}

func TestTaskHandler_Create_BoundaryValidation_422(t *testing.T) {
	mockTaskSvc := new(dummyTaskService)
	h := NewTaskHandler(mockTaskSvc)

	app := fiber.New()
	app.Post("/tasks", func(c fiber.Ctx) error {
		c.Locals("user_id", float64(1))
		c.Locals("role", "lurah")
		return h.Create(c)
	})

	// Payload tanpa judul_tugas dan target_user_ids
	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	_ = json.Unmarshal(body, &res)

	details, ok := res.Details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, details, "judul_tugas")
	assert.Contains(t, details, "target_user_ids")
	assert.Contains(t, details, "deadline")

	mockTaskSvc.AssertNotCalled(t, "CreateTask")
}

func TestHolidayHandler_CreateHoliday_BoundaryValidation_422(t *testing.T) {
	mockHolidaySvc := new(dummyHolidayService)
	h := NewHolidayHandler(mockHolidaySvc)

	app := fiber.New()
	app.Post("/holidays", h.CreateHoliday)

	// Payload kosong
	req := httptest.NewRequest(http.MethodPost, "/holidays", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	_ = json.Unmarshal(body, &res)

	details, ok := res.Details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, details, "tanggal_mulai")
	assert.Contains(t, details, "tanggal_selesai")
	assert.Contains(t, details, "keterangan")

	mockHolidaySvc.AssertNotCalled(t, "CreateHoliday")
}

func TestWorkHourHandler_UpdateWorkHour_BoundaryValidation_422(t *testing.T) {
	mockWorkHourSvc := new(dummyWorkHourService)
	h := NewWorkHourHandler(mockWorkHourSvc)

	app := fiber.New()
	app.Put("/work-hour", h.UpdateWorkHour)

	// Format jam salah (kurang dari 5 karakter, e.g. "7:00")
	payload := map[string]string{
		"jam_masuk":        "7:00",
		"jam_pulang":       "18:00",
		"jam_masuk_jumat":  "07:00",
		"jam_pulang_jumat": "16:00",
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPut, "/work-hour", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	_ = json.Unmarshal(body, &res)

	details, ok := res.Details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, details, "jam_masuk")

	mockWorkHourSvc.AssertNotCalled(t, "UpdateWorkHour")
}

func TestJabatanHandler_Create_BoundaryValidation_422(t *testing.T) {
	mockJabatanSvc := new(dummyJabatanService)
	h := NewJabatanHandler(mockJabatanSvc)

	app := fiber.New()
	app.Post("/jabatan", h.Create)

	req := httptest.NewRequest(http.MethodPost, "/jabatan", bytes.NewReader([]byte(`{"nama": ""}`)))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	_ = json.Unmarshal(body, &res)

	details, ok := res.Details.(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, details, "nama")

	mockJabatanSvc.AssertNotCalled(t, "CreateJabatan")
}
