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

type izinServiceMockForHandler struct {
	mock.Mock
}

func (m *izinServiceMockForHandler) CreatePengajuan(input service.PengajuanIzinInput) (*domain.PengajuanIzin, error) {
	return nil, nil
}

func (m *izinServiceMockForHandler) CreateByAdmin(input service.PengajuanIzinInput, adminID uint) (*domain.PengajuanIzin, error) {
	return nil, nil
}

func (m *izinServiceMockForHandler) ApprovePengajuan(izinID uint, approverID uint, approved bool, komentar string) error {
	return nil
}

func (m *izinServiceMockForHandler) GetMyPengajuan(userID uint) ([]domain.PengajuanIzin, error) {
	return nil, nil
}

func (m *izinServiceMockForHandler) GetPendingApprovals() ([]domain.PengajuanIzin, error) {
	return nil, nil
}

func (m *izinServiceMockForHandler) GetAllPengajuan(page, limit int) ([]domain.PengajuanIzin, int64, error) {
	args := m.Called(page, limit)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.PengajuanIzin), args.Get(1).(int64), args.Error(2)
}

func TestIzinHandler_GetAll_Pagination(t *testing.T) {
	mockSvc := new(izinServiceMockForHandler)
	h := NewIzinHandler(mockSvc)

	app := fiber.New()
	app.Get("/izin", h.GetAll)

	expectedList := []domain.PengajuanIzin{
		{ID: 1, UserID: 10, JenisIzin: "cuti", TanggalMulai: time.Now(), TanggalSelesai: time.Now()},
		{ID: 2, UserID: 11, JenisIzin: "sakit", TanggalMulai: time.Now(), TanggalSelesai: time.Now()},
	}
	var totalItems int64 = 42

	// Query page=2&limit=5
	mockSvc.On("GetAllPengajuan", 2, 5).Return(expectedList, totalItems, nil).Once()

	req := httptest.NewRequest("GET", "/izin?page=2&limit=5", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.Equal(t, true, res["success"])
	assert.Equal(t, "success", res["status"])
	assert.Equal(t, "Data pengajuan izin berhasil diambil", res["message"])

	// Check pagination envelope
	pagination, ok := res["pagination"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), pagination["page"])
	assert.Equal(t, float64(5), pagination["limit"])
	assert.Equal(t, float64(42), pagination["total_items"])
	assert.Equal(t, float64(9), pagination["total_pages"]) // ceil(42/5) = 9

	// Check meta alias
	meta, ok := res["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), meta["page"])
	assert.Equal(t, float64(5), meta["limit"])
	assert.Equal(t, float64(42), meta["total_items"])
	assert.Equal(t, float64(9), meta["total_pages"])

	mockSvc.AssertExpectations(t)
}
