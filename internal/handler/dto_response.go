package handler

import (
	"time"

	"laporanharianapi/internal/domain"
)

// FormatISO8601 memformat time.Time ke format ISO 8601 / RFC3339 standar.
func FormatISO8601(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}

// FormatISO8601Ptr memformat pointer time.Time ke pointer string ISO 8601 / RFC3339.
func FormatISO8601Ptr(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}

// =========================================================================
// NOTIFICATION DTO
// =========================================================================

// NotificationResponseDTO mengenkapsulasi data notifikasi yang diekspos ke klien.
type NotificationResponseDTO struct {
	ID        uint   `json:"id"`
	UserID    int    `json:"user_id"`
	Kategori  string `json:"kategori"`
	Judul     string `json:"judul"`
	Pesan     string `json:"pesan"`
	IsRead    bool   `json:"is_read"`
	TerkaitID int    `json:"terkait_id"`
	CreatedAt string `json:"created_at"`
}

// ToNotificationDTO mengonversi domain.Notification ke NotificationResponseDTO.
func ToNotificationDTO(n domain.Notification) NotificationResponseDTO {
	return NotificationResponseDTO{
		ID:        n.ID,
		UserID:    n.UserID,
		Kategori:  n.Kategori,
		Judul:     n.Judul,
		Pesan:     n.Pesan,
		IsRead:    n.IsRead,
		TerkaitID: n.TerkaitID,
		CreatedAt: FormatISO8601(n.CreatedAt),
	}
}

// ToNotificationDTOList mengonversi slice domain.Notification ke slice NotificationResponseDTO.
func ToNotificationDTOList(list []domain.Notification) []NotificationResponseDTO {
	dtos := make([]NotificationResponseDTO, len(list))
	for i, item := range list {
		dtos[i] = ToNotificationDTO(item)
	}
	return dtos
}

// =========================================================================
// ABSENSI DTO
// =========================================================================

// AbsensiResponseDTO mengenkapsulasi data absensi harian yang diekspos ke klien.
type AbsensiResponseDTO struct {
	ID               uint    `json:"id"`
	UserID           uint    `json:"user_id"`
	Tanggal          string  `json:"tanggal"`
	JamMasuk         *string `json:"jam_masuk"`
	SelfieMasukPath  *string `json:"selfie_masuk_path"`
	LokasiMasukLat   *string `json:"lokasi_masuk_lat"`
	LokasiMasukLong  *string `json:"lokasi_masuk_long"`
	JamPulang        *string `json:"jam_pulang"`
	SelfiePulangPath *string `json:"selfie_pulang_path"`
	LokasiPulangLat  *string `json:"lokasi_pulang_lat"`
	LokasiPulangLong *string `json:"lokasi_pulang_long"`
	FaceVerified     bool    `json:"face_verified"`
	Status           string  `json:"status"`
	CreatedAt        string  `json:"created_at"`
}

// ToAbsensiDTO mengonversi pointer domain.Absensi ke pointer AbsensiResponseDTO.
func ToAbsensiDTO(a *domain.Absensi) *AbsensiResponseDTO {
	if a == nil {
		return nil
	}
	return &AbsensiResponseDTO{
		ID:               a.ID,
		UserID:           a.UserID,
		Tanggal:          FormatISO8601(a.Tanggal),
		JamMasuk:         FormatISO8601Ptr(a.JamMasuk),
		SelfieMasukPath:  a.SelfieMasukPath,
		LokasiMasukLat:   a.LokasiMasukLat,
		LokasiMasukLong:  a.LokasiMasukLong,
		JamPulang:        FormatISO8601Ptr(a.JamPulang),
		SelfiePulangPath: a.SelfiePulangPath,
		LokasiPulangLat:  a.LokasiPulangLat,
		LokasiPulangLong: a.LokasiPulangLong,
		FaceVerified:     a.FaceVerified,
		Status:           a.Status,
		CreatedAt:        FormatISO8601(a.CreatedAt),
	}
}

// =========================================================================
// PENGAJUAN IZIN DTO
// =========================================================================

// UserBriefDTO merepresentasikan ringkasan profil user untuk relasi DTO.
type UserBriefDTO struct {
	ID   uint   `json:"id"`
	Nama string `json:"nama"`
	NIP  string `json:"nip"`
	Role string `json:"role"`
}

// PengajuanIzinResponseDTO mengenkapsulasi data pengajuan izin/sakit/cuti yang diekspos ke klien.
type PengajuanIzinResponseDTO struct {
	ID               uint          `json:"id"`
	UserID           uint          `json:"user_id"`
	JenisIzin        string        `json:"jenis_izin"`
	TanggalMulai     string        `json:"tanggal_mulai"`
	TanggalSelesai   string        `json:"tanggal_selesai"`
	Keterangan       string        `json:"keterangan"`
	DokumenPath      *string       `json:"dokumen_path"`
	StatusApproval   string        `json:"status_approval"`
	ApprovedBy       *uint         `json:"approved_by"`
	ApprovedAt       *string       `json:"approved_at"`
	KomentarApprover *string       `json:"komentar_approver"`
	CreatedAt        string        `json:"created_at"`
	User             *UserBriefDTO `json:"user,omitempty"`
	Approver         *UserBriefDTO `json:"approver,omitempty"`
}

// ToPengajuanIzinDTO mengonversi pointer domain.PengajuanIzin ke pointer PengajuanIzinResponseDTO.
func ToPengajuanIzinDTO(p *domain.PengajuanIzin) *PengajuanIzinResponseDTO {
	if p == nil {
		return nil
	}
	dto := &PengajuanIzinResponseDTO{
		ID:               p.ID,
		UserID:           p.UserID,
		JenisIzin:        p.JenisIzin,
		TanggalMulai:     FormatISO8601(p.TanggalMulai),
		TanggalSelesai:   FormatISO8601(p.TanggalSelesai),
		Keterangan:       p.Keterangan,
		DokumenPath:      p.DokumenPath,
		StatusApproval:   p.StatusApproval,
		ApprovedBy:       p.ApprovedBy,
		ApprovedAt:       FormatISO8601Ptr(p.ApprovedAt),
		KomentarApprover: p.KomentarApprover,
		CreatedAt:        FormatISO8601(p.CreatedAt),
	}
	if p.User != nil {
		dto.User = &UserBriefDTO{
			ID:   p.User.ID,
			Nama: p.User.Nama,
			NIP:  p.User.NIP,
			Role: p.User.Role,
		}
	}
	if p.Approver != nil {
		dto.Approver = &UserBriefDTO{
			ID:   p.Approver.ID,
			Nama: p.Approver.Nama,
			NIP:  p.Approver.NIP,
			Role: p.Approver.Role,
		}
	}
	return dto
}

// ToPengajuanIzinDTOList mengonversi slice domain.PengajuanIzin ke slice PengajuanIzinResponseDTO.
func ToPengajuanIzinDTOList(list []domain.PengajuanIzin) []PengajuanIzinResponseDTO {
	dtos := make([]PengajuanIzinResponseDTO, len(list))
	for i, item := range list {
		dto := ToPengajuanIzinDTO(&item)
		if dto != nil {
			dtos[i] = *dto
		}
	}
	return dtos
}
