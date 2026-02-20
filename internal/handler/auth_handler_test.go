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

func (m *AuthServiceMock) Login(nip string, password string) (string, error) {
	args := m.Called(nip, password)
	return args.String(0), args.Error(1)
}

// ============================================================
// Test Login Handler
// ============================================================

func TestLoginHandler_Success(t *testing.T) {
	t.Run("HTTP 200 - Login berhasil, mendapat token JWT", func(t *testing.T) {
		// 1. Setup mock
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("Login", "198106152014102004", "123456").
			Return("eyJhbG.dummy.token", nil)

		// 2. Setup Fiber app + route
		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		// 3. Buat HTTP request (NIP & password sesuai data seeder)
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
		assert.Equal(t, "Login berhasil", result["message"])
		assert.NotEmpty(t, result["token"], "Response harus mengandung token")
		assert.Equal(t, "eyJhbG.dummy.token", result["token"])

		mockAuthService.AssertExpectations(t)
	})
}

func TestLoginHandler_Fail_BadRequest_EmptyBody(t *testing.T) {
	t.Run("HTTP 400 - Body JSON kosong", func(t *testing.T) {
		// 1. Setup (tidak perlu mock karena handler gagal sebelum memanggil service)
		mockAuthService := new(AuthServiceMock)

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		// 2. Kirim body kosong
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")

		// 3. Eksekusi
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// 4. Assert status code 400
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		// 5. Parse response body
		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		assert.Equal(t, "error", result["status"])

		// Login service tidak boleh dipanggil
		mockAuthService.AssertNotCalled(t, "Login")
	})
}

func TestLoginHandler_Fail_BadRequest_MissingFields(t *testing.T) {
	t.Run("HTTP 400 - NIP dan Password kosong dalam JSON", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		// Kirim JSON valid tapi NIP dan password kosong
		body := `{"nip": "", "password": ""}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Assert status code 400
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "NIP dan password wajib diisi", result["message"])

		mockAuthService.AssertNotCalled(t, "Login")
	})
}

func TestLoginHandler_Fail_BadRequest_InvalidJSON(t *testing.T) {
	t.Run("HTTP 400 - Format JSON tidak valid", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		// Kirim body yang bukan JSON valid
		body := `{invalid json!!!}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// Assert status code 400
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		assert.Equal(t, "error", result["status"])

		mockAuthService.AssertNotCalled(t, "Login")
	})
}

func TestLoginHandler_Fail_Unauthorized(t *testing.T) {
	t.Run("HTTP 401 - Kredensial salah (password salah)", func(t *testing.T) {
		// 1. Setup mock: Login mengembalikan error
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("Login", "198106152014102004", "wrongpassword").
			Return("", errors.New("NIP atau password salah"))

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		// 2. Kirim request dengan password salah
		body := `{"nip": "198106152014102004", "password": "wrongpassword"}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// 3. Eksekusi
		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		// 4. Assert status code 401
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		// 5. Parse response body
		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "NIP atau password salah", result["message"])

		mockAuthService.AssertExpectations(t)
	})
}

func TestLoginHandler_Fail_Unauthorized_NIPNotFound(t *testing.T) {
	t.Run("HTTP 401 - NIP tidak ditemukan", func(t *testing.T) {
		mockAuthService := new(AuthServiceMock)
		mockAuthService.On("Login", "000000000", "123456").
			Return("", errors.New("NIP atau password salah"))

		app := fiber.New()
		authHandler := NewAuthHandler(mockAuthService)
		app.Post("/api/login", authHandler.Login)

		body := `{"nip": "000000000", "password": "123456"}`
		req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

		respBody, _ := io.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(respBody, &result)

		assert.Equal(t, "error", result["status"])
		assert.Equal(t, "NIP atau password salah", result["message"])

		mockAuthService.AssertExpectations(t)
	})
}
