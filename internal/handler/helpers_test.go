package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/apperror"
)

func TestErrorResponse_StatusMapping(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
		expectedMsg    string
	}{
		{
			name:           "409 Conflict - NIP already exists",
			err:            apperror.ErrNIPAlreadyExists,
			expectedStatus: fiber.StatusConflict,
			expectedMsg:    "NIP sudah terdaftar",
		},
		{
			name:           "409 Conflict - Already checked in",
			err:            apperror.ErrAlreadyCheckedIn,
			expectedStatus: fiber.StatusConflict,
			expectedMsg:    "Anda sudah melakukan absensi masuk hari ini",
		},
		{
			name:           "409 Conflict - String containing duplicate/conflict",
			err:            errors.New("duplicate entry for key 'nip'"),
			expectedStatus: fiber.StatusConflict,
			expectedMsg:    "duplicate entry for key 'nip'",
		},
		{
			name:           "404 Not Found - Task not found",
			err:            apperror.ErrTaskNotFound,
			expectedStatus: fiber.StatusNotFound,
			expectedMsg:    "tugas tidak ditemukan",
		},
		{
			name:           "404 Not Found - Report not found",
			err:            apperror.ErrReportNotFound,
			expectedStatus: fiber.StatusNotFound,
			expectedMsg:    "laporan tidak ditemukan",
		},
		{
			name:           "404 Not Found - User not found",
			err:            apperror.ErrUserNotFound,
			expectedStatus: fiber.StatusNotFound,
			expectedMsg:    "user tidak ditemukan",
		},
		{
			name:           "403 Forbidden - Access denied message",
			err:            errors.New("anda tidak memiliki akses untuk melihat detail tugas ini"),
			expectedStatus: fiber.StatusForbidden,
			expectedMsg:    "anda tidak memiliki akses untuk melihat detail tugas ini",
		},
		{
			name:           "403 Forbidden - Apperror forbidden",
			err:            apperror.ErrForbidden,
			expectedStatus: fiber.StatusForbidden,
			expectedMsg:    "akses ditolak",
		},
		{
			name:           "401 Unauthorized - Invalid token",
			err:            apperror.ErrInvalidToken,
			expectedStatus: fiber.StatusUnauthorized,
			expectedMsg:    "token tidak valid atau sudah kadaluarsa",
		},
		{
			name:           "400 Bad Request - Reason required",
			err:            apperror.ErrReasonRequired,
			expectedStatus: fiber.StatusBadRequest,
			expectedMsg:    "alasan (komentar) wajib diisi jika laporan ditolak",
		},
		{
			name:           "500 Internal Server Error - Sanitized raw DB error",
			err:            errors.New("Error 1064: You have an error in your SQL syntax; check mysql manual"),
			expectedStatus: fiber.StatusInternalServerError,
			expectedMsg:    "Terjadi kesalahan internal server",
		},
		{
			name:           "500 Internal Server Error - Sanitized GORM driver error",
			err:            errors.New("driver connection refused table users not found"),
			expectedStatus: fiber.StatusInternalServerError,
			expectedMsg:    "Terjadi kesalahan internal server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := fiber.New()
			app.Get("/test", func(c fiber.Ctx) error {
				return ErrorResponse(c, tt.err)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, resp.StatusCode)

			body, _ := io.ReadAll(resp.Body)
			var res Response
			err = json.Unmarshal(body, &res)
			assert.NoError(t, err)
			assert.False(t, res.Success)
			assert.Equal(t, "error", res.Status)
			assert.Equal(t, tt.expectedMsg, res.Message)
		})
	}
}
