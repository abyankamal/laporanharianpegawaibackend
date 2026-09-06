package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
)

type dummyNotifService struct {
	mock.Mock
}

func (m *dummyNotifService) GetMyNotifications(userID int, page, limit int) ([]domain.Notification, int64, error) {
	args := m.Called(userID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.Notification), args.Get(1).(int64), args.Error(2)
}

func (m *dummyNotifService) GetNotificationByID(id int, userID int) (*domain.Notification, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
}

func (m *dummyNotifService) ReadNotification(notifID int, userID int) error {
	args := m.Called(notifID, userID)
	return args.Error(0)
}

func TestNotificationHandler_GetMy_Pagination(t *testing.T) {
	mockSvc := new(dummyNotifService)
	h := NewNotificationHandler(mockSvc)

	app := fiber.New()
	app.Get("/notifications", func(c fiber.Ctx) error {
		c.Locals("user_id", float64(1))
		return h.GetMy(c)
	})

	expectedList := []domain.Notification{
		{
			ID:        1,
			UserID:    1,
			Judul:     "Tugas Baru",
			Pesan:     "Ada tugas baru",
			CreatedAt: time.Now(),
		},
	}

	mockSvc.On("GetMyNotifications", 1, 2, 5).Return(expectedList, int64(15), nil)

	req := httptest.NewRequest(http.MethodGet, "/notifications?page=2&limit=5", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.True(t, res.Success)
	assert.Equal(t, "success", res.Status)
	assert.NotNil(t, res.Pagination)
	assert.Equal(t, 2, res.Pagination.Page)
	assert.Equal(t, 5, res.Pagination.Limit)
	assert.Equal(t, int64(15), res.Pagination.TotalItems)
	assert.Equal(t, 3, res.Pagination.TotalPages)

	// Backward compatibility verification
	assert.NotNil(t, res.Meta)
	assert.Equal(t, 2, res.Meta.CurrentPage)
	assert.Equal(t, int64(15), res.Meta.TotalData)

	mockSvc.AssertExpectations(t)
}

func TestNotificationHandler_GetMy_DBError(t *testing.T) {
	mockSvc := new(dummyNotifService)
	h := NewNotificationHandler(mockSvc)

	app := fiber.New()
	app.Get("/notifications", func(c fiber.Ctx) error {
		c.Locals("user_id", float64(1))
		return h.GetMy(c)
	})

	mockSvc.On("GetMyNotifications", 1, 1, 10).Return(nil, int64(0), errors.New("db down"))

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	mockSvc.AssertExpectations(t)
}
