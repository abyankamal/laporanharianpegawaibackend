package handler

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestSendSuccess(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SendSuccess(c, fiber.StatusOK, "Berhasil mengambil data", map[string]string{"foo": "bar"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.True(t, res.Success)
	assert.Equal(t, "success", res.Status)
	assert.Equal(t, "Berhasil mengambil data", res.Message)
	dataMap, ok := res.Data.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "bar", dataMap["foo"])
}

func TestSendPaginated(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		items := []string{"item1", "item2"}
		return SendPaginated(c, fiber.StatusOK, "Data list", items, 1, 10, int64(20), 2)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.True(t, res.Success)
	assert.Equal(t, "success", res.Status)
	assert.Equal(t, "Data list", res.Message)
	assert.NotNil(t, res.Pagination)
	assert.Equal(t, 1, res.Pagination.Page)
	assert.Equal(t, 10, res.Pagination.Limit)
	assert.Equal(t, int64(20), res.Pagination.TotalItems)
	assert.Equal(t, 2, res.Pagination.TotalPages)
	// Backward compatibility verification
	assert.NotNil(t, res.Meta)
	assert.Equal(t, 1, res.Pagination.CurrentPage)
	assert.Equal(t, int64(20), res.Pagination.TotalData)
	assert.Equal(t, 2, res.Pagination.TotalPage)
	assert.Equal(t, res.Pagination.Page, res.Meta.Page)
	assert.Equal(t, res.Pagination.TotalPages, res.Meta.TotalPages)
	assert.Equal(t, res.Pagination.TotalPage, res.Meta.TotalPage)
}

func TestSendError(t *testing.T) {
	app := fiber.New()
	app.Get("/test", func(c fiber.Ctx) error {
		return SendError(c, fiber.StatusBadRequest, "Parameter tidak valid", "detail error info")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(body, &res)
	assert.NoError(t, err)

	assert.False(t, res.Success)
	assert.Equal(t, "error", res.Status)
	assert.Equal(t, "Parameter tidak valid", res.Message)
	assert.Equal(t, "detail error info", res.Details)
}

func TestParsePagination(t *testing.T) {
	app := fiber.New()
	var parsedPage, parsedLimit, parsedOffset int

	app.Get("/test-page", func(c fiber.Ctx) error {
		parsedPage, parsedLimit, parsedOffset = ParsePagination(c, 15)
		return c.SendStatus(fiber.StatusOK)
	})

	t.Run("Default values when empty", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-page", nil)
		_, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 1, parsedPage)
		assert.Equal(t, 15, parsedLimit)
		assert.Equal(t, 0, parsedOffset)
	})

	t.Run("Custom page and limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-page?page=3&limit=25", nil)
		_, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 3, parsedPage)
		assert.Equal(t, 25, parsedLimit)
		assert.Equal(t, 50, parsedOffset)
	})

	t.Run("Cap limit to 100", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-page?page=2&limit=500", nil)
		_, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 2, parsedPage)
		assert.Equal(t, 100, parsedLimit)
		assert.Equal(t, 100, parsedOffset)
	})

	t.Run("Invalid parameters fallback to defaults", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test-page?page=-5&limit=abc", nil)
		_, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 1, parsedPage)
		assert.Equal(t, 15, parsedLimit)
		assert.Equal(t, 0, parsedOffset)
	})
}

func TestCalculateTotalPages(t *testing.T) {
	assert.Equal(t, 1, CalculateTotalPages(0, 10))
	assert.Equal(t, 1, CalculateTotalPages(5, 10))
	assert.Equal(t, 1, CalculateTotalPages(10, 10))
	assert.Equal(t, 2, CalculateTotalPages(11, 10))
	assert.Equal(t, 5, CalculateTotalPages(50, 10))
	assert.Equal(t, 6, CalculateTotalPages(51, 10))
	assert.Equal(t, 1, CalculateTotalPages(10, 0))
}
