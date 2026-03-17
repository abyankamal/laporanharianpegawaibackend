package handler

import (
	"io"
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================
// Mock Services
// ============================================================

type ReportServiceMock struct {
	mock.Mock
}

func (m *ReportServiceMock) CreateReport(input service.ReportInput) (*domain.Laporan, error) {
	args := m.Called(input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Laporan), args.Error(1)
}

func (m *ReportServiceMock) GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error) {
	args := m.Called(filter, requesterRole, requesterID)
	return args.Get(0).([]domain.Laporan), args.Get(1).(int64), args.Error(2)
}

func (m *ReportServiceMock) GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error) {
	args := m.Called(id, requesterRole, requesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Laporan), args.Error(1)
}

func (m *ReportServiceMock) GetReportRecap(userID uint, startDate, endDate time.Time) (*repository.ReportRecapResponse, error) {
	args := m.Called(userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ReportRecapResponse), args.Error(1)
}

func (m *ReportServiceMock) EvaluateReport(assessorID uint, assessorRole string, req service.EvaluateReportRequest) error {
	args := m.Called(assessorID, assessorRole, req)
	return args.Error(0)
}

type UserServiceMock struct {
	mock.Mock
}

func (m *UserServiceMock) GetAllUsers() ([]domain.User, error) {
	args := m.Called()
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *UserServiceMock) GetUserByID(id uint) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserServiceMock) CreateUser(req service.CreateUserRequest) (*domain.User, error) {
	args := m.Called(req)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserServiceMock) UpdateUser(id uint, req service.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(id, req)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserServiceMock) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserServiceMock) ChangePassword(userID uint, req service.ChangePasswordRequest) error {
	args := m.Called(userID, req)
	return args.Error(0)
}

func (m *UserServiceMock) UpdateProfilePhoto(userID uint, fileHeader *multipart.FileHeader) (string, error) {
	args := m.Called(userID, fileHeader)
	return args.String(0), args.Error(1)
}

func (m *UserServiceMock) UpdateFCMToken(userID uint, token string) error {
	args := m.Called(userID, token)
	return args.Error(0)
}

func (m *UserServiceMock) GetSupervisors(roleFilter string) ([]domain.User, error) {
	args := m.Called(roleFilter)
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *UserServiceMock) GetUsersByRoles(roles []string) ([]domain.User, error) {
	args := m.Called(roles)
	return args.Get(0).([]domain.User), args.Error(1)
}

// ============================================================
// Test ExportReportPDFHandler
// ============================================================

func TestExportReportPDFHandler_Success(t *testing.T) {
	t.Run("Success Export PDF - Staff User", func(t *testing.T) {
		// 1. Setup mocks
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)

		userID := uint(1)
		role := "staf"
		user := domain.User{
			ID:   userID,
			Nama: "Test User",
			Role: role,
			Jabatan: &domain.RefJabatan{
				NamaJabatan: "Staff",
			},
		}

		reports := []domain.Laporan{
			{
				ID:             1,
				JudulKegiatan:  "Test Kegiatan",
				DeskripsiHasil: "Test Hasil",
				WaktuPelaporan: time.Now(),
				TipeLaporan:    false,
			},
		}

		mockUserService.On("GetUserByID", userID).Return(&user, nil)
		mockReportService.On("GetAllReports", mock.Anything, "lurah", userID).Return(reports, int64(1), nil)

		// 2. Setup Fiber
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		// Middleware for mock auth
		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})

		app.Get("/export/pdf", h.ExportReportPDFHandler)

		// 3. Request
		req := httptest.NewRequest(http.MethodGet, "/export/pdf?start_date=2024-03-01&end_date=2024-03-31", nil)
		resp, err := app.Test(req)

		// 4. Assert
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
		
		body, _ := io.ReadAll(resp.Body)
		assert.NotEmpty(t, body)
		
		mockUserService.AssertExpectations(t)
		mockReportService.AssertExpectations(t)
	})

	t.Run("Success Export PDF - Lurah User (All Users)", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)

		userID := uint(2)
		role := "lurah"
		
		users := []domain.User{
			{ID: 1, Nama: "Staff 1", Role: "staf"},
			{ID: 2, Nama: "Lurah", Role: "lurah"},
		}

		reports := []domain.Laporan{
			{
				ID:             1,
				JudulKegiatan:  "Staff Activity",
				WaktuPelaporan: time.Now(),
			},
		}

		mockUserService.On("GetAllUsers").Return(users, nil)
		// It will call GetAllReports for each user in targetUsers (Staff 1 and Lurah)
		// We expect two calls, one for userID 1 and one for userID 2
		mockReportService.On("GetAllReports", mock.Anything, "lurah", userID).Return(reports, int64(1), nil).Twice()

		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})

		app.Get("/export/pdf", h.ExportReportPDFHandler)

		req := httptest.NewRequest(http.MethodGet, "/export/pdf", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
		
		mockUserService.AssertExpectations(t)
	})

	t.Run("Success Export PDF - Sekertaris with target user_id", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)

		requesterID := uint(3)
		targetUserID := uint(1)
		role := "sekertaris"
		
		targetUser := domain.User{ID: targetUserID, Nama: "Staff 1", Role: "staf"}

		reports := []domain.Laporan{
			{
				ID:             1,
				JudulKegiatan:  "Staff Activity",
				WaktuPelaporan: time.Now(),
			},
		}

		mockUserService.On("GetUserByID", targetUserID).Return(&targetUser, nil)
		mockReportService.On("GetAllReports", mock.Anything, "lurah", requesterID).Return(reports, int64(1), nil)

		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(requesterID))
			c.Locals("role", role)
			return c.Next()
		})

		app.Get("/export/pdf", h.ExportReportPDFHandler)

		req := httptest.NewRequest(http.MethodGet, "/export/pdf?user_id=1", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))
		
		mockUserService.AssertExpectations(t)
		mockReportService.AssertExpectations(t)
	})
}
