package handler

import (
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

	return c.Status(status).JSON(Response{
		Success: false,
		Status:  "error",
		Message: message,
		Details: detail,
	})
}
