package handler

import (
	"bytes"
	"errors"
	"io"
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *ReportServiceMock) GetReportRecapAggregated(filter repository.ReportFilter, requesterRole string, requesterID uint) (*repository.ReportRecapResponse, error) {
	args := m.Called(filter, requesterRole, requesterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ReportRecapResponse), args.Error(1)
}

func (m *ReportServiceMock) EvaluateReport(assessorID uint, assessorRole string, req service.EvaluateReportRequest) error {
	args := m.Called(assessorID, assessorRole, req)
	return args.Error(0)
}

func (m *ReportServiceMock) UpdateReport(id uint, judul string, deskripsi string, fileFoto *multipart.FileHeader, requesterID uint, requesterRole string) error {
	args := m.Called(id, judul, deskripsi, fileFoto, requesterID, requesterRole)
	return args.Error(0)
}

func (m *ReportServiceMock) DeleteReport(id uint, requesterID uint, requesterRole string) error {
	args := m.Called(id, requesterID, requesterRole)
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

func (m *UserServiceMock) ResetPasswordByAdmin(targetUserID uint, newPassword string) error {
	args := m.Called(targetUserID, newPassword)
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
				UserID:         &userID,
				JudulKegiatan:  "Test Kegiatan",
				DeskripsiHasil: "Test Hasil",
				WaktuPelaporan: time.Now(),
				TipeLaporan:    false,
			},
		}

		mockUserService.On("GetUserByID", userID).Return(&user, nil)
		mockReportService.On("GetAllReports", mock.Anything, "staf", userID).Return(reports, int64(1), nil)

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
				UserID:         &users[0].ID,
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
				UserID:         &targetUserID,
				JudulKegiatan:  "Staff Activity",
				WaktuPelaporan: time.Now(),
			},
		}

		mockUserService.On("GetUserByID", targetUserID).Return(&targetUser, nil)
		mockReportService.On("GetAllReports", mock.Anything, "sekertaris", requesterID).Return(reports, int64(1), nil)

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

	t.Run("Success Export PDF - Lurah with multi-select user_ids", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)

		requesterID := uint(2)
		role := "lurah"

		user1 := domain.User{ID: 1, Nama: "Staff 1", Role: "staf"}
		user3 := domain.User{ID: 3, Nama: "Staff 3", Role: "staf"}

		reports := []domain.Laporan{
			{
				ID:             1,
				UserID:         &user1.ID,
				JudulKegiatan:  "Staff 1 Activity",
				WaktuPelaporan: time.Now(),
			},
		}

		mockUserService.On("GetUserByID", uint(1)).Return(&user1, nil)
		mockUserService.On("GetUserByID", uint(3)).Return(&user3, nil)
		mockReportService.On("GetAllReports", mock.Anything, "lurah", requesterID).Return(reports, int64(1), nil)

		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(requesterID))
			c.Locals("role", role)
			return c.Next()
		})

		app.Get("/export/pdf", h.ExportReportPDFHandler)

		req := httptest.NewRequest(http.MethodGet, "/export/pdf?user_ids=1,3", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "application/pdf", resp.Header.Get("Content-Type"))

		mockUserService.AssertExpectations(t)
		mockReportService.AssertExpectations(t)
	})

	t.Run("Forbidden Export PDF - Sekertaris with unauthorized multi-select user_ids", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)

		requesterID := uint(3)
		role := "sekertaris"

		user1 := domain.User{ID: 1, Nama: "Staff 1", Role: "staf"}
		user2 := domain.User{ID: 2, Nama: "Lurah", Role: "lurah"} // Forbidden target role for Secretary

		mockUserService.On("GetUserByID", uint(1)).Return(&user1, nil)
		mockUserService.On("GetUserByID", uint(2)).Return(&user2, nil)

		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(requesterID))
			c.Locals("role", role)
			return c.Next()
		})

		app.Get("/export/pdf", h.ExportReportPDFHandler)

		req := httptest.NewRequest(http.MethodGet, "/export/pdf?user_ids=1,2", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)

		mockUserService.AssertExpectations(t)
		mockReportService.AssertExpectations(t)
	})
}

// ============================================================
// Test UpdateReport Handler
// ============================================================

func TestUpdateReportHandler(t *testing.T) {
	t.Run("Update Success", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		userID := uint(1)
		role := "staf"
		reportID := uint(10)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})
		app.Put("/reports/:id", h.Update)

		mockReportService.On("UpdateReport", reportID, "New Title", "New Detail", mock.Anything, userID, role).Return(nil)

		req := httptest.NewRequest(http.MethodPut, "/reports/10", strings.NewReader(`{"judul_kegiatan":"New Title", "deskripsi_hasil":"New Detail"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		mockReportService.AssertExpectations(t)
	})

	t.Run("Update Forbidden", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		userID := uint(2)
		role := "staf"
		reportID := uint(10)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})
		app.Put("/reports/:id", h.Update)

		mockReportService.On("UpdateReport", reportID, "Title", "Detail", mock.Anything, userID, role).Return(errors.New("akses ditolak"))

		req := httptest.NewRequest(http.MethodPut, "/reports/10", strings.NewReader(`{"judul_kegiatan":"Title", "deskripsi_hasil":"Detail"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

// ============================================================
// Test DeleteReport Handler
// ============================================================

func TestDeleteReportHandler(t *testing.T) {
	t.Run("Delete Success (Lurah)", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		userID := uint(1)
		role := "lurah"
		reportID := uint(10)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})
		app.Delete("/reports/:id", h.Delete)

		mockReportService.On("DeleteReport", reportID, userID, role).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/reports/10", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Delete Forbidden (Staff)", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		userID := uint(2)
		role := "staf"
		reportID := uint(10)

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})
		app.Delete("/reports/:id", h.Delete)

		mockReportService.On("DeleteReport", reportID, userID, role).Return(errors.New("akses ditolak"))

		req := httptest.NewRequest(http.MethodDelete, "/reports/10", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})
}

// ============================================================
// Test CreateReport Handler
// ============================================================

func TestCreateReportHandler(t *testing.T) {
	t.Run("Create Success with Offline Sync", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		userID := uint(1)
		role := "staf"

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})
		app.Post("/reports", h.Create)

		mockReportService.On("CreateReport", mock.MatchedBy(func(input service.ReportInput) bool {
			return input.IsOfflineSync == true && input.JudulKegiatan == "Offline Task"
		})).Return(&domain.Laporan{
			ID:         1,
			IsOvertime: false,
		}, nil)

		body := new(bytes.Buffer)
		writer := multipart.NewWriter(body)
		_ = writer.WriteField("judul_kegiatan", "Offline Task")
		_ = writer.WriteField("deskripsi_hasil", "Selesai offline")
		_ = writer.WriteField("is_offline_sync", "true")
		_ = writer.WriteField("waktu_pelaporan", "2024-03-01 10:00:00")
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/reports", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		mockReportService.AssertExpectations(t)
	})
}

// ============================================================
// Test GetOne Report Handler
// ============================================================

func TestGetOneReportHandler(t *testing.T) {
	t.Run("GetOne Success with Location", func(t *testing.T) {
		mockReportService := new(ReportServiceMock)
		mockUserService := new(UserServiceMock)
		app := fiber.New()
		h := NewReportHandler(mockReportService, mockUserService)

		userID := uint(1)
		role := "staf"
		reportID := uint(10)
		lat := "-6.2088"
		long := "106.8456"
		alamat := "Kantor Kelurahan"

		app.Use(func(c fiber.Ctx) error {
			c.Locals("user_id", float64(userID))
			c.Locals("role", role)
			return c.Next()
		})
		app.Get("/reports/:id", h.GetOne)

		expectedReport := &domain.Laporan{
			ID:             reportID,
			UserID:         &userID,
			Status:         "menunggu_review",
			JudulKegiatan:  "Survei Lapangan",
			WaktuPelaporan: time.Now(),
			AlamatLokasi:   &alamat,
			LokasiLat:      &lat,
			LokasiLong:     &long,
			DeskripsiHasil: "Survei selesai",
			User: &domain.User{
				Nama: "Test User",
				NIP:  "12345",
				Role: role,
			},
		}

		mockReportService.On("GetReportDetail", reportID, role, userID).Return(expectedReport, nil)

		req := httptest.NewRequest(http.MethodGet, "/reports/10", nil)
		resp, err := app.Test(req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		bodyStr := string(body)
		assert.Contains(t, bodyStr, `"lokasi_lat":"-6.2088"`)
		assert.Contains(t, bodyStr, `"lokasi_long":"106.8456"`)
		assert.Contains(t, bodyStr, `"lokasi":"Kantor Kelurahan"`)

		mockReportService.AssertExpectations(t)
	})
}

