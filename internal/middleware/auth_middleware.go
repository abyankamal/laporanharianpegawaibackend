package middleware

import (
	"os"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/golang-jwt/jwt/v5"
)

// Protected adalah middleware untuk memproteksi route dengan JWT.
func Protected() fiber.Handler {
	return func(c fiber.Ctx) error {
		// 1. Ambil header Authorization
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Unauthorized",
			})
		}

		// 2. Cek format "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Format token tidak valid",
			})
		}
		tokenString := parts[1]

		// 3. Parse dan validasi token
		jwtSecret := os.Getenv("JWT_SECRET")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Pastikan algoritma adalah HS256
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Metode signing tidak valid")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Token tidak valid atau sudah kadaluarsa",
			})
		}

		// 4. Ambil claims dari token
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Gagal membaca claims token",
			})
		}

		// 5. Simpan claims ke Fiber Locals untuk digunakan Handler selanjutnya
		c.Locals("user_id", claims["user_id"])
		c.Locals("role", claims["role"])
		c.Locals("jabatan_id", claims["jabatan_id"])

		// 6. Lanjut ke handler berikutnya
		return c.Next()
	}
}
