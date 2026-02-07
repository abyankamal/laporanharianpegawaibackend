package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/joho/godotenv"

	"laporanharianapi/config"
	"laporanharianapi/internal/handler"
	"laporanharianapi/internal/middleware"
	"laporanharianapi/internal/repository"
	"laporanharianapi/internal/service"
)

func main() {
	// 1. Load environment variables dari file .env
	err := godotenv.Load()
	if err != nil {
		log.Println("Peringatan: File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	// 2. Koneksi ke Database
	config.ConnectDatabase()

	// 3. Dependency Injection (Wiring)
	// Repository
	userRepo := repository.NewUserRepository(config.DB)

	// Service
	authService := service.NewAuthService(userRepo)

	// Handler
	authHandler := handler.NewAuthHandler(authService)

	// 4. Setup Fiber App
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
			})
		},
	})

	// 5. Middleware CORS (agar bisa diakses dari HP/Web)
	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// 6. Setup Routes
	api := app.Group("/api")

	// Route Public (tidak perlu login)
	api.Post("/login", authHandler.Login)

	// Route Protected (perlu login dengan JWT)
	protected := api.Group("/", middleware.Protected())

	// Contoh route protected
	protected.Get("/profile", func(c fiber.Ctx) error {
		userID := c.Locals("user_id")
		role := c.Locals("role")
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Selamat datang di area rahasia",
			"data": fiber.Map{
				"user_id": userID,
				"role":    role,
			},
		})
	})

	// 7. Jalankan server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("Server berjalan di http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
