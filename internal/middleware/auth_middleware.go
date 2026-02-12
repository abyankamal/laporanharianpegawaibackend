package middleware

import (
	"fmt"
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

// AllowRoles adalah middleware RBAC untuk membatasi akses berdasarkan role user.
// Middleware ini HARUS dipanggil SETELAH Protected() agar c.Locals("role") sudah terisi.
func AllowRoles(allowedRoles ...string) fiber.Handler {
	// Buat map untuk lookup O(1)
	roleMap := make(map[string]bool, len(allowedRoles))
	for _, role := range allowedRoles {
		roleMap[role] = true
	}

	return func(c fiber.Ctx) error {
		// 1. Ambil role dari Locals (sudah di-set oleh Protected middleware)
		userRole, ok := c.Locals("role").(string)
		if !ok || userRole == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": "Role user tidak ditemukan",
			})
		}

		// 2. Cek apakah role ada di daftar yang diizinkan
		if !roleMap[userRole] {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Akses ditolak. Role '%s' tidak memiliki izin. Role yang diizinkan: %s", userRole, strings.Join(allowedRoles, ", ")),
			})
		}

		// 3. Role valid, lanjut ke handler berikutnya
		return c.Next()
	}
}
