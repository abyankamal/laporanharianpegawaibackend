package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/middleware"
)

func generateTestJWT(userID uint, role string, tokenType string, expDuration time.Duration, secret string) string {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"role":       role,
		"jabatan_id": uint(1),
		"token_type": tokenType,
		"exp":        time.Now().Add(expDuration).Unix(),
		"iat":        time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestMiddleware_Protected(t *testing.T) {
	secret := "secret-jwt-test-123"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	app := fiber.New()
	app.Get("/test-protected", middleware.Protected(), func(c fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"user_id": c.Locals("user_id"),
			"role":    c.Locals("role"),
		})
	})

	t.Run("No Auth Header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid Format Header (No Bearer)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
		req.Header.Set("Authorization", "Basic 12345")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		token := generateTestJWT(1, "staf", "access", time.Hour, "wrong-secret")
		req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Refresh Token Used as Access Token", func(t *testing.T) {
		token := generateTestJWT(1, "staf", "refresh", time.Hour, secret)
		req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Valid Access Token", func(t *testing.T) {
		token := generateTestJWT(5, "staf", "access", time.Hour, secret)
		req := httptest.NewRequest(http.MethodGet, "/test-protected", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})
}

func TestMiddleware_AllowRoles(t *testing.T) {
	secret := "secret-jwt-test-123"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	app := fiber.New()
	app.Get("/lurah-only", middleware.Protected(), middleware.AllowRoles("lurah", "sekertaris"), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	t.Run("Allowed Role (lurah)", func(t *testing.T) {
		token := generateTestJWT(1, "lurah", "access", time.Hour, secret)
		req := httptest.NewRequest(http.MethodGet, "/lurah-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Forbidden Role (staf)", func(t *testing.T) {
		token := generateTestJWT(5, "staf", "access", time.Hour, secret)
		req := httptest.NewRequest(http.MethodGet, "/lurah-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}

func TestMiddleware_AdminOnly(t *testing.T) {
	secret := "secret-jwt-test-123"
	os.Setenv("JWT_SECRET", secret)
	defer os.Unsetenv("JWT_SECRET")

	app := fiber.New()
	app.Get("/admin-only", middleware.Protected(), middleware.AdminOnly(), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	t.Run("Allowed Admin", func(t *testing.T) {
		token := generateTestJWT(11, "admin", "access", time.Hour, secret)
		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	})

	t.Run("Forbidden Non-Admin (staf)", func(t *testing.T) {
		token := generateTestJWT(5, "staf", "access", time.Hour, secret)
		req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
	})
}
