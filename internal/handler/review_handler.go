package handler

import (
	"math"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/service"
)

// ReviewHandler menangani request penilaian kinerja.
type ReviewHandler struct {
	reviewService service.ReviewService
}

// NewReviewHandler membuat instance baru ReviewHandler.
func NewReviewHandler(reviewService service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// Create menangani pembuatan penilaian baru oleh atasan.
func (h *ReviewHandler) Create(c fiber.Ctx) error {
	// 1. Ambil penilai_id dari JWT Token (via Locals dari middleware)
	penilaiIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	penilaiID := uint(penilaiIDFloat)

	// Ambil role dari JWT
	penilaiRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "Role tidak ditemukan",
		})
	}

	// 2. Parse JSON Body
	var req service.CreateReviewRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "Format request tidak valid: " + err.Error(),
		})
	}

	// 3. Validasi input wajib
	if req.TargetUserID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "target_user_id wajib diisi",
		})
	}
	if req.SkorID == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "skor_id wajib diisi",
		})
	}
	if req.JenisPeriode == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "jenis_periode wajib diisi (Harian/Mingguan/Bulanan/Custom)",
		})
	}
	if req.TanggalMulai == "" || req.TanggalSelesai == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": "tanggal_mulai dan tanggal_selesai wajib diisi",
		})
	}

	// 4. Panggil service
	penilaian, err := h.reviewService.SubmitReview(penilaiID, penilaiRole, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
		})
	}

	// 5. Return response sukses
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "Penilaian berhasil disimpan",
		"data": fiber.Map{
			"id":              penilaian.ID,
			"user_id":         penilaian.UserID,
			"penilai_id":      penilaian.PenilaiID,
			"skor_id":         penilaian.SkorID,
			"jenis_periode":   penilaian.JenisPeriode,
			"tanggal_mulai":   penilaian.TanggalMulai.Format("2006-01-02"),
			"tanggal_selesai": penilaian.TanggalSelesai.Format("2006-01-02"),
			"created_at":      penilaian.CreatedAt,
		},
	})
}

// GetMyReviews menangani request staf untuk melihat nilai diri sendiri.
func (h *ReviewHandler) GetMyReviews(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	userID := int(userIDFloat)

	// 2. Parse query param pagination
	limit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))

	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 3. Panggil service
	reviews, total, err := h.reviewService.GetReviewsByUserID(userID, limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data penilaian: " + err.Error(),
		})
	}

	// 4. Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data penilaian berhasil diambil",
		"data":    reviews,
		"meta": fiber.Map{
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": totalPages,
		},
	})
}

// GetMySubmittedReviews menangani request atasan untuk melihat history penilaian yang pernah dibuat.
func (h *ReviewHandler) GetMySubmittedReviews(c fiber.Ctx) error {
	// 1. Ambil penilai_id dari JWT Token
	penilaiIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"status":  "error",
			"message": "User tidak terautentikasi",
		})
	}
	penilaiID := int(penilaiIDFloat)

	// 2. Panggil service
	reviews, err := h.reviewService.GetReviewsByPenilaiID(penilaiID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Gagal mengambil data penilaian: " + err.Error(),
		})
	}

	// 3. Return response
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Data penilaian berhasil diambil",
		"data":    reviews,
	})
}
