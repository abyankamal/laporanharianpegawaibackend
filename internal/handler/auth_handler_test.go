package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================
// Mock AuthService
// ============================================================

// AuthServiceMock adalah mock untuk service.AuthService.
type AuthServiceMock struct {
	mock.Mock
}

func (m *AuthServiceMock) Login(nip string, password string) (map[string]interface{}, error) {
	args := m.Called(nip, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *AuthServiceMock) RefreshToken(token string) (map[string]interface{}, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

// ============================================================
// Test Login Handler
// ============================================================

func TestLoginHandler_Success(t *testing.T) {
	t.Run("HTTP 200 - Login berhasil, mendapat pasangan token", func(t *testing.T) {
		// 1. Setup mock
		mockResponse := map[string]interface{}{
			"access_token":  "access.token.dummy",
			"refresh_token": "refresh.token.dummy",
			"expires_in":    3600,
		}
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("Login", "198106152014102004", "123456").
			Return(mockResponse, nil)

		// 2. Setup Fiber app + route
		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		// 3. Buat HTTP request
		body := `{"nip": "198106152014102004", "password": "123456"}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// 4. Eksekusi
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// 5. Assert status code
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// 6. Parse response body
		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		err = json.Unmarshal(respBody, &result)
		assert.NoError(t, err)

		// 7. Assert response body
		assert.Equal(t, "success", result["status"])
		assert.Equal(t, true, result["success"])

		data := result["data"].(map[string]interface{})
		assert.Equal(t, "access.token.dummy", data["access_token"])
		assert.Equal(t, "refresh.token.dummy", data["refresh_token"])
		assert.Equal(t, float64(3600), data["expires_in"])

		mockAuthService.AssertExpectations(t)
	})
}

func TestLoginHandler_Fail_BadRequest_EmptyBody(t *testing.T) {
	t.Run("HTTP 400 - Body JSON kosong", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)
		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		mockAuthService.AssertNotCalled(t, "Login")
	})
}

func TestLoginHandler_Fail_BadRequest_MissingFields(t *testing.T) {
	t.Run("HTTP 400 - NIP dan Password kosong dalam JSON", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)
		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		body := `{"nip": "", "password": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		mockAuthService.AssertNotCalled(t, "Login")
	})
}

func TestLoginHandler_Fail_Unauthorized(t *testing.T) {
	t.Run("HTTP 401 - Kredensial salah", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("Login", "198106152014102004", "wrong").
			Return(nil, errors.New("NIP atau password salah"))

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		body := `{"nip": "198106152014102004", "password": "wrong"}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockAuthService.AssertExpectations(t)
	})
}

// ============================================================
// Test Refresh Token Handler
// ============================================================

func TestRefreshTokenHandler_Success(t *testing.T) {
	t.Run("HTTP 200 - Refresh token berhasil", func(t *testing.T) {
		mockResponse := map[string]interface{}{
			"access_token":  "new.access.token",
			"refresh_token": "new.refresh.token",
			"expires_in":    3600,
		}
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("RefreshToken", "old.refresh.token").
			Return(mockResponse, nil)

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/refresh", authHandler.RefreshToken)

		body := `{"refresh_token": "old.refresh.token"}`
		req := httptest.NewRequest(http.MethodPost, "/api/refresh", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		assert.Equal(t, "success", result["status"])
		data := result["data"].(map[string]interface{})
		assert.Equal(t, "new.access.token", data["access_token"])

		mockAuthService.AssertExpectations(t)
	})
}

func TestRefreshTokenHandler_Fail_InvalidToken(t *testing.T) {
	t.Run("HTTP 401 - Refresh token tidak valid", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("RefreshToken", "invalid.token").
			Return(nil, errors.New("refresh token tidak valid"))

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/refresh", authHandler.RefreshToken)

		body := `{"refresh_token": "invalid.token"}`
		req := httptest.NewRequest(http.MethodPost, "/api/refresh", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
		mockAuthService.AssertExpectations(t)
	})
}
