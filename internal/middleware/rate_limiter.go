package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/limiter"
)

// LoginLimiter mengembalikan middleware rate limiter untuk memproteksi endpoint login dari brute force.
// Membatasi 5 request per 1 menit per IP.
func LoginLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        5,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  "error",
				"message": "Terlalu banyak percobaan login. Silakan coba lagi dalam 1 menit.",
			})
		},
	})
}

// RefreshTokenLimiter mengembalikan middleware rate limiter untuk memproteksi endpoint refresh token.
// Membatasi 15 request per 1 menit per IP.
func RefreshTokenLimiter() fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        15,
		Expiration: 1 * time.Minute,
		LimitReached: func(c fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  "error",
				"message": "Terlalu banyak permintaan refresh token. Silakan coba lagi nanti.",
			})
		},
	})
}
