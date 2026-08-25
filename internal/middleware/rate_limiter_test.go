package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/middleware"
)

func TestMiddleware_RateLimiter(t *testing.T) {
	app := fiber.New()
	app.Post("/login", middleware.LoginLimiter(), func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	// 5 request pertama harus berhasil (200 OK)
	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, resp.StatusCode, "Request ke-%d harus sukses", i)
	}

	// Request ke-6 harus ditolak oleh Rate Limiter (429 Too Many Requests)
	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusTooManyRequests, resp.StatusCode, "Request ke-6 harus kena rate limit 429")
}
