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
)

type adminServiceMockForHandler struct {
	mock.Mock
}

func (m *adminServiceMockForHandler) GetRekapLaporanAdmin(filter repository.AdminReportFilter) (*repository.AdminReportResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AdminReportResponse), args.Error(1)
}

func (m *adminServiceMockForHandler) GetLaporanExportAdmin(filter repository.AdminReportFilter) ([]domain.Laporan, error) {
	return nil, nil
}

func (m *adminServiceMockForHandler) GetDashboardSummaryAdmin() (*repository.DashboardSummaryResponse, error) {
	return nil, nil
}

func (m *adminServiceMockForHandler) GetPegawaiAdmin(filter repository.AdminPegawaiFilter) (*repository.AdminPegawaiResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AdminPegawaiResponse), args.Error(1)
}

func (m *adminServiceMockForHandler) GetPegawaiStatistikAdmin() (*repository.PegawaiStatistik, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PegawaiStatistik), args.Error(1)
}

func (m *adminServiceMockForHandler) CreatePegawaiAdmin(req *domain.User) error {
	return nil
}

func (m *adminServiceMockForHandler) UpdatePegawaiAdmin(userID uint, req *domain.User) error {
	return nil
}

func (m *adminServiceMockForHandler) DeletePegawaiAdmin(userID uint) error {
	return nil
}

func (m *adminServiceMockForHandler) ResetPasswordPegawaiAdmin(userID uint, newPassword string) error {
	return nil
}

func (m *adminServiceMockForHandler) GetPengumumanAdmin(filter repository.AdminPengumumanFilter) (*repository.AdminPengumumanResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AdminPengumumanResponse), args.Error(1)
}

func (m *adminServiceMockForHandler) GetPengumumanStatistikAdmin() (*repository.PengumumanStatistik, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PengumumanStatistik), args.Error(1)
}

func (m *adminServiceMockForHandler) CreatePengumumanAdmin(pengumuman *domain.Notification) error {
	return nil
}

func (m *adminServiceMockForHandler) UpdatePengumumanAdmin(id uint, pengumuman *domain.Notification) error {
	return nil
}

func (m *adminServiceMockForHandler) DeletePengumumanAdmin(id uint) error {
	return nil
}

func (m *adminServiceMockForHandler) GetSupervisorLurahAdmin() (*domain.LurahSupervisor, error) {
	return nil, nil
}

func (m *adminServiceMockForHandler) UpdateSupervisorLurahAdmin(nama, nip string) error {
	return nil
}

func TestAdminHandler_GetPegawai_PaginationHarmonization(t *testing.T) {
	mockSvc := new(adminServiceMockForHandler)
	h := NewAdminHandler(mockSvc)

	app := fiber.New()
	app.Get("/api/admin/pegawai", h.GetPegawai)

	filter := repository.AdminPegawaiFilter{
		Search: "budi",
		Role:   "staf",
		Page:   2,
		Limit:  15,
	}

	mockPegawaiResp := &repository.AdminPegawaiResponse{
		Data: []domain.User{
			{ID: 1, Nama: "Budi", NIP: "19800101", Role: "staf"},
		},
		TotalData:   45,
		TotalPage:   3,
		CurrentPage: 2,
	}

	mockStats := &repository.PegawaiStatistik{
		TotalPegawai: 45,
	}

	mockSvc.On("GetPegawaiAdmin", filter).Return(mockPegawaiResp, nil).Once()
	mockSvc.On("GetPegawaiStatistikAdmin").Return(mockStats, nil).Once()

	req := httptest.NewRequest("GET", "/api/admin/pegawai?search=budi&role=staf&page=2&limit=15", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.Equal(t, true, res["success"])
	assert.Equal(t, "success", res["status"])

	// 1. Verifikasi root level 'pagination'
	pagination, ok := res["pagination"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), pagination["page"])
	assert.Equal(t, float64(15), pagination["limit"])
	assert.Equal(t, float64(45), pagination["total_items"])
	assert.Equal(t, float64(3), pagination["total_pages"])
	assert.Equal(t, float64(2), pagination["current_page"])
	assert.Equal(t, float64(45), pagination["total_data"])
	assert.Equal(t, float64(3), pagination["total_page"])

	// 2. Verifikasi root level 'meta' alias
	meta, ok := res["meta"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), meta["page"])
	assert.Equal(t, float64(15), meta["limit"])
	assert.Equal(t, float64(45), meta["total_items"])
	assert.Equal(t, float64(3), meta["total_pages"])
	assert.Equal(t, float64(2), meta["current_page"])
	assert.Equal(t, float64(45), meta["total_data"])
	assert.Equal(t, float64(3), meta["total_page"])

	// 3. Verifikasi nested 'data.pagination' untuk Svelte legacy
	dataMap, ok := res["data"].(map[string]interface{})
	assert.True(t, ok)
	nestedPagination, ok := dataMap["pagination"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(45), nestedPagination["total_data"])
	assert.Equal(t, float64(3), nestedPagination["total_page"])
	assert.Equal(t, float64(3), nestedPagination["total_pages"])
	assert.Equal(t, float64(45), nestedPagination["total_items"])
	assert.Equal(t, float64(2), nestedPagination["current_page"])
	assert.Equal(t, float64(2), nestedPagination["page"])
	assert.Equal(t, float64(15), nestedPagination["limit"])

	mockSvc.AssertExpectations(t)
}

func TestAdminHandler_GetPengumuman_PaginationHarmonization(t *testing.T) {
	mockSvc := new(adminServiceMockForHandler)
	h := NewAdminHandler(mockSvc)

	app := fiber.New()
	app.Get("/api/admin/pengumuman", h.GetPengumuman)

	filter := repository.AdminPengumumanFilter{
		Search: "kerja",
		Page:   1,
		Limit:  10,
	}

	mockNotifResp := &repository.AdminPengumumanResponse{
		Data: []domain.Notification{
			{ID: 1, Judul: "Kerja Bakti", Pesan: "Besok pagi", UserID: 0, CreatedAt: time.Now()},
		},
		TotalData:   12,
		TotalPage:   2,
		CurrentPage: 1,
	}

	mockStats := &repository.PengumumanStatistik{
		TotalPengumuman: 12,
	}

	mockSvc.On("GetPengumumanAdmin", filter).Return(mockNotifResp, nil).Once()
	mockSvc.On("GetPengumumanStatistikAdmin").Return(mockStats, nil).Once()

	req := httptest.NewRequest("GET", "/api/admin/pengumuman?search=kerja", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.Equal(t, true, res["success"])
	assert.Equal(t, "success", res["status"])

	// Root pagination & meta
	pagination := res["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(12), pagination["total_items"])
	assert.Equal(t, float64(2), pagination["total_pages"])
	assert.Equal(t, float64(2), pagination["total_page"])

	meta := res["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(12), meta["total_items"])
	assert.Equal(t, float64(2), meta["total_pages"])
	assert.Equal(t, float64(2), meta["total_page"])

	// Nested data.pagination
	dataMap := res["data"].(map[string]interface{})
	nestedPagination := dataMap["pagination"].(map[string]interface{})
	assert.Equal(t, float64(12), nestedPagination["total_data"])
	assert.Equal(t, float64(2), nestedPagination["total_page"])

	mockSvc.AssertExpectations(t)
}

func TestAdminHandler_GetRekapLaporan_PaginationHarmonization(t *testing.T) {
	mockSvc := new(adminServiceMockForHandler)
	h := NewAdminHandler(mockSvc)

	app := fiber.New()
	app.Get("/api/admin/rekap-laporan", h.GetRekapLaporan)

	filter := repository.AdminReportFilter{
		Page:  1,
		Limit: 10,
	}

	mockResp := &repository.AdminReportResponse{
		Data:        []domain.Laporan{},
		TotalData:   100,
		TotalPage:   10,
		CurrentPage: 1,
	}

	mockSvc.On("GetRekapLaporanAdmin", filter).Return(mockResp, nil).Once()

	req := httptest.NewRequest("GET", "/api/admin/rekap-laporan", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res map[string]interface{}
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.Equal(t, true, res["success"])
	assert.Equal(t, "success", res["status"])

	pagination := res["pagination"].(map[string]interface{})
	assert.Equal(t, float64(1), pagination["page"])
	assert.Equal(t, float64(10), pagination["limit"])
	assert.Equal(t, float64(100), pagination["total_items"])
	assert.Equal(t, float64(10), pagination["total_pages"])
	assert.Equal(t, float64(10), pagination["total_page"])

	meta := res["meta"].(map[string]interface{})
	assert.Equal(t, float64(1), meta["page"])
	assert.Equal(t, float64(10), meta["limit"])
	assert.Equal(t, float64(100), meta["total_items"])
	assert.Equal(t, float64(10), meta["total_pages"])
	assert.Equal(t, float64(10), meta["total_page"])

	mockSvc.AssertExpectations(t)
}
