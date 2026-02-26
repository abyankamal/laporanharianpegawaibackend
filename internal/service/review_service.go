package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// CreateReviewRequest adalah struct input untuk membuat penilaian baru.
type CreateReviewRequest struct {
	TargetUserID   int    `json:"target_user_id" validate:"required"`
	SkorID         int    `json:"skor_id" validate:"required"`
	JenisPeriode   string `json:"jenis_periode" validate:"required"` // Harian, Mingguan, Bulanan, Custom
	Bulan          int    `json:"bulan" validate:"required"`
	Tahun          int    `json:"tahun" validate:"required"`
	TanggalMulai   string `json:"tanggal_mulai" validate:"required"`   // Format YYYY-MM-DD
	TanggalSelesai string `json:"tanggal_selesai" validate:"required"` // Format YYYY-MM-DD
	Catatan        string `json:"catatan"`
}

// ReviewService adalah interface untuk operasi bisnis Penilaian.
type ReviewService interface {
	SubmitReview(penilaiID uint, penilaiRole string, req CreateReviewRequest) (*domain.Penilaian, error)
	GetReviewsByUserID(userID int, limit int, offset int) ([]domain.Penilaian, int64, error)
	GetReviewsByPenilaiID(penilaiID int) ([]domain.Penilaian, error)
}

// reviewService adalah implementasi dari ReviewService.
type reviewService struct {
	reviewRepo repository.ReviewRepository
	userRepo   repository.UserRepository
	notifRepo  repository.NotificationRepository
}

// NewReviewService membuat instance baru ReviewService.
func NewReviewService(reviewRepo repository.ReviewRepository, userRepo repository.UserRepository, notifRepo repository.NotificationRepository) ReviewService {
	return &reviewService{
		reviewRepo: reviewRepo,
		userRepo:   userRepo,
		notifRepo:  notifRepo,
	}
}

// SubmitReview membuat penilaian baru dengan validasi bisnis dan hierarki RBAC.
func (s *reviewService) SubmitReview(penilaiID uint, penilaiRole string, req CreateReviewRequest) (*domain.Penilaian, error) {
	// 1. Validasi: Tidak boleh menilai diri sendiri
	if uint(req.TargetUserID) == penilaiID {
		return nil, errors.New("tidak dapat menilai diri sendiri")
	}

	// 2. Validasi: Hanya lurah dan sekertaris yang boleh menilai
	if penilaiRole != "lurah" && penilaiRole != "sekertaris" {
		return nil, errors.New("hanya Lurah dan Sekertaris yang boleh melakukan penilaian")
	}

	// 3. Validasi hierarki: Cek role target user
	targetUser, err := s.userRepo.FindByID(uint(req.TargetUserID))
	if err != nil {
		return nil, errors.New("user yang akan dinilai tidak ditemukan")
	}

	switch penilaiRole {
	case "lurah":
		// Lurah hanya boleh menilai sekertaris dan kasi
		if targetUser.Role != "sekertaris" && targetUser.Role != "kasi" {
			return nil, errors.New("Lurah hanya boleh menilai Sekertaris dan Kasi")
		}
	case "sekertaris":
		// Sekertaris hanya boleh menilai staf
		if targetUser.Role != "staf" {
			return nil, errors.New("Sekertaris hanya boleh menilai Staf")
		}
	}

	// 4. Validasi: JenisPeriode harus valid
	validPeriode := map[string]bool{
		"Harian":   true,
		"Mingguan": true,
		"Bulanan":  true,
		"Custom":   true,
	}
	if !validPeriode[req.JenisPeriode] {
		return nil, errors.New("jenis_periode tidak valid (pilihan: Harian, Mingguan, Bulanan, Custom)")
	}

	// 5. Parse dan validasi tanggal
	tanggalMulai, err := time.Parse("2006-01-02", req.TanggalMulai)
	if err != nil {
		return nil, errors.New("format tanggal_mulai tidak valid (gunakan: YYYY-MM-DD)")
	}

	tanggalSelesai, err := time.Parse("2006-01-02", req.TanggalSelesai)
	if err != nil {
		return nil, errors.New("format tanggal_selesai tidak valid (gunakan: YYYY-MM-DD)")
	}

	// 6. Validasi: TanggalMulai tidak boleh lebih besar dari TanggalSelesai
	if tanggalMulai.After(tanggalSelesai) {
		return nil, errors.New("tanggal_mulai tidak boleh lebih besar dari tanggal_selesai")
	}

	// 7. Validasi: SkorID harus positif
	if req.SkorID <= 0 {
		return nil, errors.New("skor_id wajib diisi dan harus valid")
	}

	// 8. Validasi: TargetUserID harus positif
	if req.TargetUserID <= 0 {
		return nil, errors.New("target_user_id wajib diisi dan harus valid")
	}

	// 9. Validasi: Satu bulan hanya bisa satu kali penilaian (Monthly Appraisal Constraint)
	exists, err := s.reviewRepo.CheckExistingReview(uint(req.TargetUserID), req.Bulan, req.Tahun)
	if err != nil {
		return nil, fmt.Errorf("gagal mengecek riwayat penilaian: %v", err)
	}
	if exists {
		return nil, errors.New("Pegawai ini sudah dinilai pada bulan tersebut")
	}

	// 10. Buat struct Penilaian
	now := time.Now()
	userID := uint(req.TargetUserID)
	skorID := uint(req.SkorID)
	penilaian := &domain.Penilaian{
		UserID:         &userID,
		PenilaiID:      &penilaiID,
		SkorID:         &skorID,
		JenisPeriode:   req.JenisPeriode,
		Bulan:          req.Bulan,
		Tahun:          req.Tahun,
		TanggalMulai:   tanggalMulai,
		TanggalSelesai: tanggalSelesai,
		Catatan:        req.Catatan,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// 11. Simpan ke database
	err = s.reviewRepo.Create(penilaian)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan penilaian: %v", err)
	}

	// 12. Buat notifikasi untuk target user
	notif := &domain.Notification{
		UserID:    req.TargetUserID,
		Kategori:  "Penilaian",
		Judul:     "Penilaian Kinerja Baru",
		Pesan:     fmt.Sprintf("Atasan Anda telah memberikan penilaian kinerja untuk periode %s.", req.JenisPeriode),
		TerkaitID: int(penilaian.ID),
		CreatedAt: now,
	}
	if err := s.notifRepo.Create(notif); err != nil {
		log.Printf("⚠️ Gagal membuat notifikasi penilaian: %v", err)
	}

	return penilaian, nil
}

// GetReviewsByUserID mengambil riwayat penilaian berdasarkan pegawai (untuk staf melihat nilai sendiri).
func (s *reviewService) GetReviewsByUserID(userID int, limit int, offset int) ([]domain.Penilaian, int64, error) {
	return s.reviewRepo.FindByUserID(userID, limit, offset)
}

// GetReviewsByPenilaiID mengambil history penilaian yang dibuat oleh atasan tertentu.
func (s *reviewService) GetReviewsByPenilaiID(penilaiID int) ([]domain.Penilaian, error) {
	return s.reviewRepo.FindByPenilaiID(penilaiID)
}
