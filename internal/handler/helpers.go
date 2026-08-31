package handler

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	
	"laporanharianapi/internal/apperror"
)

// GetRequesterFromCtx mengambil userID dan role dari JWT claims yang ada di context.
func GetRequesterFromCtx(c fiber.Ctx) (userID uint, role string, err error) {
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return 0, "", errors.New("user tidak terautentikasi")
	}

	role, ok = c.Locals("role").(string)
	if !ok {
		return 0, "", errors.New("role tidak ditemukan")
	}

	return uint(requesterIDFloat), role, nil
}

// ParseDateRange memparsing parameter start_date dan end_date dari query.
// Jika kosong, defaultnya adalah hari pertama dan hari terakhir bulan ini.
func ParseDateRange(c fiber.Ctx) (startDate, endDate time.Time, err error) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")

	if startStr != "" {
		startDate, err = time.ParseInLocation("2006-01-02", startStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("format start_date tidak valid")
		}
	} else {
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)
	}

	if endStr != "" {
		endDate, err = time.ParseInLocation("2006-01-02", endStr, loc)
		if err != nil {
			return time.Time{}, time.Time{}, errors.New("format end_date tidak valid")
		}
	} else {
		// Hari terakhir bulan ini
		endDate = time.Date(now.Year(), now.Month()+1, 0, 23, 59, 59, 0, loc)
	}

	// Pastikan end date menutupi sampai jam 23:59:59
	endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 0, loc)

	return startDate, endDate, nil
}

// ErrorResponse memetakan error internal ke response HTTP yang sesuai dengan envelope standar.
func ErrorResponse(c fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError
	message := err.Error()

	switch {
	case errors.Is(err, apperror.ErrReportNotFound),
		errors.Is(err, apperror.ErrUserNotFound),
		strings.Contains(strings.ToLower(message), "tidak ditemukan"):
		status = fiber.StatusNotFound

	case errors.Is(err, apperror.ErrForbidden),
		errors.Is(err, apperror.ErrSecretaryStaffOnly),
		errors.Is(err, apperror.ErrOnlyLurahCanDeleteReport),
		errors.Is(err, apperror.ErrOnlyOwnReportAllowed),
		errors.Is(err, apperror.ErrOnlyStaffOrOwnAllowed),
		errors.Is(err, apperror.ErrOnlyOwnReportModifiable),
		strings.Contains(strings.ToLower(message), "akses ditolak"):
		status = fiber.StatusForbidden

	case errors.Is(err, apperror.ErrUnauthorized),
		errors.Is(err, apperror.ErrInvalidToken):
		status = fiber.StatusUnauthorized

	case errors.Is(err, apperror.ErrBadRequest),
		errors.Is(err, apperror.ErrInvalidEvaluationStatus),
		errors.Is(err, apperror.ErrReasonRequired),
		errors.Is(err, apperror.ErrReportAlreadyReviewed),
		errors.Is(err, apperror.ErrReportAlreadyApproved),
		errors.Is(err, apperror.ErrNIPAlreadyExists),
		errors.Is(err, apperror.ErrOldPasswordMismatch),
		errors.Is(err, apperror.ErrSamePassword):
		status = fiber.StatusBadRequest
	}

	return SendError(c, status, message)
}
