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
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	penilaiID := uint(penilaiIDFloat)

	// Ambil role dari JWT
	penilaiRole, ok := c.Locals("role").(string)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "Role tidak ditemukan")
	}

	// 2. Parse JSON Body
	var req service.CreateReviewRequest
	if err := c.Bind().JSON(&req); err != nil {
		return SendError(c, fiber.StatusBadRequest, "Format request tidak valid")
	}

	// 3. Validasi input wajib
	if req.TargetUserID == 0 {
		return SendError(c, fiber.StatusBadRequest, "target_user_id wajib diisi")
	}
	if req.SkorID == 0 {
		return SendError(c, fiber.StatusBadRequest, "skor_id wajib diisi")
	}
	if req.JenisPeriode == "" {
		return SendError(c, fiber.StatusBadRequest, "jenis_periode wajib diisi (Harian/Mingguan/Bulanan/Custom)")
	}
	if req.TanggalMulai == "" || req.TanggalSelesai == "" {
		return SendError(c, fiber.StatusBadRequest, "tanggal_mulai dan tanggal_selesai wajib diisi")
	}
	if req.Bulan <= 0 || req.Bulan > 12 || req.Tahun <= 0 {
		return SendError(c, fiber.StatusBadRequest, "bulan (1-12) dan tahun wajib diisi dengan valid")
	}

	// 4. Panggil service
	penilaian, err := h.reviewService.SubmitReview(penilaiID, penilaiRole, req)
	if err != nil {
		return ErrorResponse(c, err)
	}

	// 5. Return response sukses
	return SendSuccess(c, fiber.StatusCreated, "Penilaian berhasil disimpan", fiber.Map{
		"id":              penilaian.ID,
		"user_id":         penilaian.UserID,
		"penilai_id":      penilaian.PenilaiID,
		"skor_id":         penilaian.SkorID,
		"jenis_periode":   penilaian.JenisPeriode,
		"bulan":           penilaian.Bulan,
		"tahun":           penilaian.Tahun,
		"tanggal_mulai":   penilaian.TanggalMulai.Format("2006-01-02"),
		"tanggal_selesai": penilaian.TanggalSelesai.Format("2006-01-02"),
		"created_at":      penilaian.CreatedAt,
	})
}

// GetMyReviews menangani request staf untuk melihat nilai diri sendiri.
func (h *ReviewHandler) GetMyReviews(c fiber.Ctx) error {
	// 1. Ambil user_id dari JWT Token
	userIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	userID := int(userIDFloat)

	// 2. Parse query param pagination
	limit, _ := strconv.Atoi(c.Query("limit"))
	page, _ := strconv.Atoi(c.Query("page"))

	if limit <= 0 {
		limit = 10
	} else if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// 3. Panggil service
	reviews, total, err := h.reviewService.GetReviewsByUserID(userID, limit, offset)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data penilaian")
	}

	// 4. Hitung total halaman
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	// 5. Return response
	return SendPaginated(c, fiber.StatusOK, "Data penilaian berhasil diambil", reviews, page, limit, total, totalPages)
}

// GetMySubmittedReviews menangani request atasan untuk melihat history penilaian yang pernah dibuat.
func (h *ReviewHandler) GetMySubmittedReviews(c fiber.Ctx) error {
	// 1. Ambil penilai_id dari JWT Token
	penilaiIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return SendError(c, fiber.StatusUnauthorized, "User tidak terautentikasi")
	}
	penilaiID := int(penilaiIDFloat)

	// 2. Panggil service
	reviews, err := h.reviewService.GetReviewsByPenilaiID(penilaiID)
	if err != nil {
		return SendError(c, fiber.StatusInternalServerError, "Gagal mengambil data penilaian")
	}

	// 3. Return response
	return SendSuccess(c, fiber.StatusOK, "Data penilaian berhasil diambil", reviews)
}
