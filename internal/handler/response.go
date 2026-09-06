package handler

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v3"
)

// Pagination mendefinisikan metadata pagination standar.
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
	// Backward compatibility field aliases
	CurrentPage int   `json:"current_page,omitempty"`
	TotalData   int64 `json:"total_data,omitempty"`
}

// Response mendefinisikan standar envelope response JSON.
type Response struct {
	Success    bool        `json:"success"`
	Status     string      `json:"status"` // "success" | "error"
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       *Pagination `json:"meta,omitempty"` // Alias untuk backward compatibility
	Details    interface{} `json:"details,omitempty"`
}

// SendSuccess mengirimkan response sukses standar.
func SendSuccess(c fiber.Ctx, status int, message string, data interface{}) error {
	return c.Status(status).JSON(Response{
		Success: true,
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// SendPaginated mengirimkan response sukses berhalaman standar.
func SendPaginated(c fiber.Ctx, status int, message string, data interface{}, page, limit int, totalItems int64, totalPages int) error {
	p := &Pagination{
		Page:        page,
		Limit:       limit,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		CurrentPage: page,
		TotalData:   totalItems,
	}

	return c.Status(status).JSON(Response{
		Success:    true,
		Status:     "success",
		Message:    message,
		Data:       data,
		Pagination: p,
		Meta:       p,
	})
}

// SendError mengirimkan response error standar.
func SendError(c fiber.Ctx, status int, message string, details ...interface{}) error {
	var detail interface{}
	if len(details) > 0 {
		detail = details[0]
	}

	if status == fiber.StatusInternalServerError {
		lowered := strings.ToLower(message)
		if strings.Contains(lowered, "sql") ||
			strings.Contains(lowered, "mysql") ||
			strings.Contains(lowered, "driver") ||
			strings.Contains(lowered, "gorm") ||
			strings.Contains(lowered, "syntax error") ||
			strings.Contains(lowered, "connection refused") ||
			strings.Contains(lowered, "table ") ||
			strings.Contains(lowered, "column ") {
			message = "Terjadi kesalahan internal server"
		}
		if detail != nil {
			if dStr, ok := detail.(string); ok {
				dLower := strings.ToLower(dStr)
				if strings.Contains(dLower, "sql") ||
					strings.Contains(dLower, "mysql") ||
					strings.Contains(dLower, "driver") ||
					strings.Contains(dLower, "gorm") {
					detail = nil
				}
			}
		}
	}

	return c.Status(status).JSON(Response{
		Success: false,
		Status:  "error",
		Message: message,
		Details: detail,
	})
}

// ParsePagination mengekstrak query parameter 'page' dan 'limit' dengan nilai default dan batas maksimum.
// Default: page=1, limit=10 (atau defaultLimit yang diberikan), limit capped pada 100.
func ParsePagination(c fiber.Ctx, defaultLimits ...int) (page int, limit int, offset int) {
	page = 1
	limit = 10
	if len(defaultLimits) > 0 && defaultLimits[0] > 0 {
		limit = defaultLimits[0]
	}

	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p > 0 {
			page = p
		}
	}

	if lStr := c.Query("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}

	if limit > 100 {
		limit = 100
	}

	offset = (page - 1) * limit
	return page, limit, offset
}

// CalculateTotalPages menghitung total halaman berdasarkan total items dan limit.
func CalculateTotalPages(totalItems int64, limit int) int {
	if limit <= 0 {
		return 1
	}
	totalPages := int((totalItems + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}
	return totalPages
}
