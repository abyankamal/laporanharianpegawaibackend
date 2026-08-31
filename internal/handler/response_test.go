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
