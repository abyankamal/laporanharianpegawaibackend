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
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"
)

type mockAbsensiServiceForHandler struct {
	mock.Mock
}

func (m *mockAbsensiServiceForHandler) CheckIn(input service.AbsensiCheckInInput) (*domain.Absensi, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Absensi), args.Error(1)
}

func (m *mockAbsensiServiceForHandler) CheckOut(input service.AbsensiCheckOutInput) (*domain.Absensi, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Absensi), args.Error(1)
}

func (m *mockAbsensiServiceForHandler) GetTodayStatus(userID uint) (*domain.Absensi, bool, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*domain.Absensi), args.Bool(1), args.Error(2)
}

func (m *mockAbsensiServiceForHandler) GetMonthlyRecap(userID uint, bulan, tahun int) ([]domain.Absensi, *repository.AbsensiRecapResponse, error) {
	return nil, nil, nil
}

func (m *mockAbsensiServiceForHandler) GetAllMonthlyRecap(bulan, tahun int, users []domain.User) ([]service.UserAbsensiRecap, error) {
	return nil, nil
}

func (m *mockAbsensiServiceForHandler) IsWorkday(date time.Time) (bool, error) {
	return true, nil
}

func (m *mockAbsensiServiceForHandler) GetEffectiveWorkdays(bulan, tahun int) ([]time.Time, error) {
	return nil, nil
}

func TestAbsensiHandler_GetTodayStatus_DTO(t *testing.T) {
	mockAbsensi := new(mockAbsensiServiceForHandler)
	h := NewAbsensiHandler(mockAbsensi, nil)

	app := fiber.New()
	app.Get("/today", func(c fiber.Ctx) error {
		c.Locals("user_id", float64(10))
		return h.GetTodayStatus(c)
	})

	tgl := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	jamMasuk := time.Date(2026, 9, 6, 7, 45, 0, 0, time.UTC)
	absensi := &domain.Absensi{
		ID:           1,
		UserID:       10,
		Tanggal:      tgl,
		JamMasuk:     &jamMasuk,
		FaceVerified: true,
		Status:       "hadir",
		CreatedAt:    tgl,
	}

	mockAbsensi.On("GetTodayStatus", uint(10)).Return(absensi, true, nil).Once()

	req := httptest.NewRequest("GET", "/today", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.Equal(t, true, res["success"])
	assert.Equal(t, "success", res["status"])

	dataMap := res["data"].(map[string]interface{})
	assert.Equal(t, true, dataMap["is_workday"])

	absensiMap := dataMap["absensi"].(map[string]interface{})
	assert.Equal(t, float64(1), absensiMap["id"])
	assert.Equal(t, float64(10), absensiMap["user_id"])
	assert.Equal(t, "2026-09-06T00:00:00Z", absensiMap["tanggal"])
	assert.Equal(t, "2026-09-06T07:45:00Z", absensiMap["jam_masuk"])
	assert.Equal(t, "hadir", absensiMap["status"])

	mockAbsensi.AssertExpectations(t)
}
