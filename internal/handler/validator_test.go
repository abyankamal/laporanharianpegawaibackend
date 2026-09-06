package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

type DummySampleRequest struct {
	Nama     string   `json:"nama" validate:"required,min=3"`
	Email    string   `json:"email" validate:"required"`
	Age      int      `json:"age" validate:"min=18,max=60"`
	Role     string   `json:"role" validate:"oneof=staf lurah sekertaris"`
	Optional string   `json:"optional,omitempty"`
	Items    []string `json:"items" validate:"min=1"`
}

func TestValidateStruct_Success(t *testing.T) {
	req := DummySampleRequest{
		Nama:  "Budi",
		Email: "budi@example.com",
		Age:   25,
		Role:  "staf",
		Items: []string{"item1"},
	}

	errs := ValidateStruct(&req)
	assert.Empty(t, errs)
}

func TestValidateStruct_Failures(t *testing.T) {
	req := DummySampleRequest{
		Nama:  "Al", // Kurang dari min=3
		Email: "",   // Required
		Age:   10,   // Kurang dari min=18
		Role:  "superadmin", // Bukan salah satu dari oneof
		Items: []string{},   // Kurang dari min=1
	}

	errs := ValidateStruct(&req)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs, "nama")
	assert.Contains(t, errs, "email")
	assert.Contains(t, errs, "age")
	assert.Contains(t, errs, "role")
	assert.Contains(t, errs, "items")

	assert.Equal(t, "email wajib diisi", errs["email"])
	assert.Equal(t, "nama minimal 3 karakter", errs["nama"])
	assert.Equal(t, "age minimal bernilai 18", errs["age"])
	assert.Equal(t, "role harus salah satu dari: staf, lurah, sekertaris", errs["role"])
	assert.Equal(t, "items minimal 1 item", errs["items"])
}

func TestBindAndValidate_Success(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c fiber.Ctx) error {
		var req DummySampleRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return SendSuccess(c, fiber.StatusOK, "Berhasil validasi", req)
	})

	bodyPayload := DummySampleRequest{
		Nama:  "Santoso",
		Email: "santoso@example.com",
		Age:   30,
		Role:  "Lurah",
		Items: []string{"penugasan1"},
	}
	bodyBytes, _ := json.Marshal(bodyPayload)

	httpReq := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(respBody, &res)
	assert.NoError(t, err)
	assert.True(t, res.Success)
	assert.Equal(t, "Berhasil validasi", res.Message)
}

func TestBindAndValidate_ValidationFailure_422(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c fiber.Ctx) error {
		var req DummySampleRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return SendSuccess(c, fiber.StatusOK, "Success", req)
	})

	bodyPayload := map[string]interface{}{
		"nama":  "A", // Too short
		"email": "",  // Empty
		"age":   15,  // Under 18
	}
	bodyBytes, _ := json.Marshal(bodyPayload)

	httpReq := httptest.NewRequest("POST", "/test", bytes.NewReader(bodyBytes))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusUnprocessableEntity, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(respBody, &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "error", res.Status)
	assert.Equal(t, "Validasi input gagal", res.Message)

	details, ok := res.Details.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, details["nama"])
	assert.NotEmpty(t, details["email"])
	assert.NotEmpty(t, details["age"])
}

func TestBindAndValidate_InvalidJSON_400(t *testing.T) {
	app := fiber.New()
	app.Post("/test", func(c fiber.Ctx) error {
		var req DummySampleRequest
		if !BindAndValidate(c, &req) {
			return nil
		}
		return SendSuccess(c, fiber.StatusOK, "Success", req)
	})

	httpReq := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte("{invalid-json")))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(httpReq)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var res Response
	err = json.Unmarshal(respBody, &res)
	assert.NoError(t, err)
	assert.False(t, res.Success)
	assert.Equal(t, "Format request tidak valid", res.Message)
}
