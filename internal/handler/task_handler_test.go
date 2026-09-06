package handler

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/service"
)

type taskServiceMockForHandler struct {
	mock.Mock
}

func (m *taskServiceMockForHandler) CreateTask(requesterID uint, requesterRole string, req service.CreateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error) {
	return nil, nil
}

func (m *taskServiceMockForHandler) GetMyTasks(userID int, page, limit int) ([]domain.TugasOrganisasi, int64, error) {
	args := m.Called(userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.TugasOrganisasi), args.Get(1).(int64), args.Error(2)
}

func (m *taskServiceMockForHandler) GetAllTasks(page, limit int) ([]domain.TugasOrganisasi, int64, error) {
	args := m.Called(page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.TugasOrganisasi), args.Get(1).(int64), args.Error(2)
}

func (m *taskServiceMockForHandler) GetTaskByID(requesterID uint, requesterRole string, taskID uint) (*domain.TugasOrganisasi, error) {
	return nil, nil
}

func (m *taskServiceMockForHandler) UpdateTask(requesterID uint, requesterRole string, taskID uint, req service.UpdateOrganizationalTaskRequest) (*domain.TugasOrganisasi, error) {
	return nil, nil
}

func (m *taskServiceMockForHandler) DeleteTask(requesterID uint, requesterRole string, taskID uint) error {
	return nil
}

func TestTaskHandler_GetMyTasks_Pagination(t *testing.T) {
	mockSvc := new(taskServiceMockForHandler)
	h := NewTaskHandler(mockSvc)

	app := fiber.New()
	app.Get("/my-tasks", func(c fiber.Ctx) error {
		c.Locals("user_id", float64(2))
		return h.GetMyTasks(c)
	})

	now := time.Now()
	expectedTasks := []domain.TugasOrganisasi{
		{
			ID:         1,
			JudulTugas: "Tugas A",
			CreatedAt:  now,
		},
	}

	mockSvc.On("GetMyTasks", 2, 1, 5).Return(expectedTasks, int64(12), nil)

	req := httptest.NewRequest(http.MethodGet, "/my-tasks?page=1&limit=5", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.True(t, res.Success)
	assert.NotNil(t, res.Pagination)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 5, res.Pagination.Limit)
	assert.Equal(t, int64(12), res.Pagination.TotalItems)
	assert.Equal(t, 3, res.Pagination.TotalPages)

	mockSvc.AssertExpectations(t)
}

func TestTaskHandler_GetAll_Pagination(t *testing.T) {
	mockSvc := new(taskServiceMockForHandler)
	h := NewTaskHandler(mockSvc)

	app := fiber.New()
	app.Get("/tasks", h.GetAll)

	now := time.Now()
	expectedTasks := []domain.TugasOrganisasi{
		{
			ID:         10,
			JudulTugas: "Tugas Organisasi",
			CreatedAt:  now,
		},
	}

	mockSvc.On("GetAllTasks", 2, 10).Return(expectedTasks, int64(25), nil)

	req := httptest.NewRequest(http.MethodGet, "/tasks?page=2&limit=10", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.True(t, res.Success)
	assert.NotNil(t, res.Pagination)
	assert.Equal(t, 2, res.Pagination.Page)
	assert.Equal(t, 10, res.Pagination.Limit)
	assert.Equal(t, int64(25), res.Pagination.TotalItems)
	assert.Equal(t, 3, res.Pagination.TotalPages)

	mockSvc.AssertExpectations(t)
}
