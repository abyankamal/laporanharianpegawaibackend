package middleware_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/middleware"
)

func TestMiddleware_IdempotencyGuard(t *testing.T) {
	app := fiber.New()

	var counter int64
	store := middleware.NewMemoryIdempotencyStore()

	app.Post("/create-resource", middleware.IdempotencyGuard(middleware.IdempotencyConfig{
		Store: store,
	}), func(c fiber.Ctx) error {
		val := atomic.AddInt64(&counter, 1)
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"success": true,
			"status":  "success",
			"message": "Resource created",
			"data": fiber.Map{
				"count": val,
			},
		})
	})

	t.Run("Tanpa Idempotency-Key (Pass-through normal)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create-resource", bytes.NewBufferString(`{"title":"Task 1"}`))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		// Request kedua tanpa key akan memicu handler lagi
		req2 := httptest.NewRequest(http.MethodPost, "/create-resource", bytes.NewBufferString(`{"title":"Task 1"}`))
		req2.Header.Set("Content-Type", "application/json")
		resp2, err := app.Test(req2)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp2.StatusCode)

		// Counter harus bertambah 2 kali
		assert.Equal(t, int64(2), atomic.LoadInt64(&counter))
	})

	t.Run("Dengan Idempotency-Key (Pertama kali sukses dibuat)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create-resource", bytes.NewBufferString(`{"title":"Unique Task"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "idemp-key-12345")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), `"count":3`)
	})

	t.Run("Replay dengan Idempotency-Key dan Payload Identik (Mendapatkan cached response tanpa eksekusi ulang handler)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create-resource", bytes.NewBufferString(`{"title":"Unique Task"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "idemp-key-12345")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
		assert.Equal(t, "HIT - Idempotent Replay", resp.Header.Get("X-Cache-Lookup"))

		bodyBytes, _ := io.ReadAll(resp.Body)
		// Harus tetap count: 3 (tidak naik menjadi 4)
		assert.Contains(t, string(bodyBytes), `"count":3`)
		assert.Equal(t, int64(3), atomic.LoadInt64(&counter))
	})

	t.Run("Replay dengan Idempotency-Key yang sama tetapi Payload Berbeda (422 Unprocessable Entity)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/create-resource", bytes.NewBufferString(`{"title":"Different Payload"}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Idempotency-Key", "idemp-key-12345")

		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

		bodyBytes, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(bodyBytes), "Idempotency key telah digunakan dengan payload request yang berbeda")

		// Counter tidak boleh berubah
		assert.Equal(t, int64(3), atomic.LoadInt64(&counter))
	})
}
