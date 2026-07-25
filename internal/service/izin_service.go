package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// PengajuanIzinInput adalah struct untuk input pengajuan izin/sakit/cuti.
type PengajuanIzinInput struct {
	UserID         uint
	JenisIzin      string // sakit, cuti, izin, dinas_luar
	TanggalMulai   string // format YYYY-MM-DD
	TanggalSelesai string // format YYYY-MM-DD
	Keterangan     string
	FileDokumen    *multipart.FileHeader // Surat sakit, dll (opsional)
}

// IzinService adalah interface untuk operasi bisnis PengajuanIzin.
type IzinService interface {
	CreatePengajuan(input PengajuanIzinInput) (*domain.PengajuanIzin, error)
	ApprovePengajuan(izinID uint, approverID uint, approved bool, komentar string) error
	GetMyPengajuan(userID uint) ([]domain.PengajuanIzin, error)
	GetPendingApprovals() ([]domain.PengajuanIzin, error)
}

type izinService struct {
	izinRepo    repository.IzinRepository
	absensiRepo repository.AbsensiRepository
}

// NewIzinService membuat instance baru IzinService.
func NewIzinService(izinRepo repository.IzinRepository, absensiRepo repository.AbsensiRepository) IzinService {
	return &izinService{
		izinRepo:    izinRepo,
		absensiRepo: absensiRepo,
	}
}

// CreatePengajuan membuat pengajuan izin/sakit/cuti baru.
func (s *izinService) CreatePengajuan(input PengajuanIzinInput) (*domain.PengajuanIzin, error) {
	// 1. Validasi jenis izin
	jenisLower := strings.ToLower(input.JenisIzin)
	validJenis := map[string]bool{
		"sakit":      true,
		"cuti":       true,
		"izin":       true,
		"dinas_luar": true,
	}
	if !validJenis[jenisLower] {
		return nil, errors.New("jenis izin tidak valid (gunakan: sakit, cuti, izin, atau dinas_luar)")
	}

	// 2. Validasi keterangan
	if strings.TrimSpace(input.Keterangan) == "" {
		return nil, errors.New("keterangan wajib diisi")
	}

	// 3. Parse tanggal
	tanggalMulai, err := time.ParseInLocation("2006-01-02", input.TanggalMulai, time.Local)
	if err != nil {
		return nil, errors.New("format tanggal mulai tidak valid (gunakan YYYY-MM-DD)")
	}

	tanggalSelesai, err := time.ParseInLocation("2006-01-02", input.TanggalSelesai, time.Local)
	if err != nil {
		return nil, errors.New("format tanggal selesai tidak valid (gunakan YYYY-MM-DD)")
	}

	if tanggalSelesai.Before(tanggalMulai) {
		return nil, errors.New("tanggal selesai tidak boleh lebih awal dari tanggal mulai")
	}

	// 4. Simpan dokumen pendukung jika ada
	var dokumenPath *string
	if input.FileDokumen != nil {
		path, err := s.saveDokumenIzin(input.FileDokumen)
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan dokumen: %v", err)
		}
		dokumenPath = &path
	}

	// 5. Buat record pengajuan izin
	izin := &domain.PengajuanIzin{
		UserID:         input.UserID,
		JenisIzin:      jenisLower,
		TanggalMulai:   tanggalMulai,
		TanggalSelesai: tanggalSelesai,
		Keterangan:     input.Keterangan,
		DokumenPath:    dokumenPath,
		StatusApproval: "menunggu",
		CreatedAt:      time.Now(),
	}

	err = s.izinRepo.Create(izin)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan pengajuan izin: %v", err)
	}

	return izin, nil
}

// ApprovePengajuan menyetujui atau menolak pengajuan izin.
// Jika disetujui, otomatis update status absensi pada tanggal yang di-cover.
func (s *izinService) ApprovePengajuan(izinID uint, approverID uint, approved bool, komentar string) error {
	// 1. Ambil data pengajuan
	izin, err := s.izinRepo.GetByID(izinID)
	if err != nil {
		return errors.New("pengajuan izin tidak ditemukan")
	}

	// 2. Cek status
	if izin.StatusApproval != "menunggu" {
		return errors.New("pengajuan ini sudah diproses sebelumnya")
	}

	// 3. Update status
	now := time.Now()
	izin.ApprovedBy = &approverID
	izin.ApprovedAt = &now

	if approved {
		izin.StatusApproval = "disetujui"
		if komentar == "" {
			komentar = "Disetujui"
		}
	} else {
		izin.StatusApproval = "ditolak"
		if strings.TrimSpace(komentar) == "" {
			return errors.New("alasan penolakan wajib diisi")
		}
	}
	izin.KomentarApprover = &komentar

	err = s.izinRepo.Update(izin)
	if err != nil {
		return fmt.Errorf("gagal memproses pengajuan: %v", err)
	}

	// 4. Jika disetujui, update status absensi pada tanggal-tanggal yang di-cover
	if approved {
		s.updateAbsensiForApprovedIzin(izin)
	}

	return nil
}

// updateAbsensiForApprovedIzin mengupdate atau membuat record absensi
// dengan status sesuai jenis izin pada tanggal-tanggal yang di-cover.
func (s *izinService) updateAbsensiForApprovedIzin(izin *domain.PengajuanIzin) {
	current := izin.TanggalMulai
	for !current.After(izin.TanggalSelesai) {
		// Skip weekend
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			current = current.AddDate(0, 0, 1)
			continue
		}

		dateOnly := time.Date(current.Year(), current.Month(), current.Day(), 0, 0, 0, 0, time.Local)

		// Cek apakah sudah ada record absensi
		existing, _ := s.absensiRepo.GetByUserAndDate(izin.UserID, dateOnly)
		if existing != nil {
			// Update status yang sudah ada
			existing.Status = izin.JenisIzin
			s.absensiRepo.Update(existing)
		} else {
			// Buat record baru dengan status izin
			absensi := &domain.Absensi{
				UserID:    izin.UserID,
				Tanggal:   dateOnly,
				Status:    izin.JenisIzin,
				CreatedAt: time.Now(),
			}
			s.absensiRepo.Create(absensi)
		}

		current = current.AddDate(0, 0, 1)
	}
}

// GetMyPengajuan mengambil semua pengajuan izin milik user.
func (s *izinService) GetMyPengajuan(userID uint) ([]domain.PengajuanIzin, error) {
	return s.izinRepo.GetByUserID(userID)
}

// GetPendingApprovals mengambil semua pengajuan yang menunggu approval.
func (s *izinService) GetPendingApprovals() ([]domain.PengajuanIzin, error) {
	return s.izinRepo.GetPendingApprovals()
}

// saveDokumenIzin menyimpan file dokumen pendukung izin.
func (s *izinService) saveDokumenIzin(fileHeader *multipart.FileHeader) (string, error) {
	uploadDir := "./uploads/izin"
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	// Validasi ukuran (max 20MB)
	if fileHeader.Size > 20*1024*1024 {
		return "", errors.New("ukuran dokumen maksimal 20MB")
	}

	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", err
	}

	return filepath.ToSlash(destPath), nil
}
