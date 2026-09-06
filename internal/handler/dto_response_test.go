package handler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"laporanharianapi/internal/domain"
)

func TestDTO_Notification(t *testing.T) {
	fixedTime := time.Date(2026, 9, 6, 10, 30, 0, 0, time.UTC)

	notif := domain.Notification{
		ID:        5,
		UserID:    12,
		Kategori:  "Tugas",
		Judul:     "Tugas Baru",
		Pesan:     "Harap selesaikan",
		IsRead:    false,
		TerkaitID: 101,
		CreatedAt: fixedTime,
	}

	dto := ToNotificationDTO(notif)
	assert.Equal(t, uint(5), dto.ID)
	assert.Equal(t, 12, dto.UserID)
	assert.Equal(t, "Tugas", dto.Kategori)
	assert.Equal(t, "Tugas Baru", dto.Judul)
	assert.Equal(t, "Harap selesaikan", dto.Pesan)
	assert.False(t, dto.IsRead)
	assert.Equal(t, 101, dto.TerkaitID)
	assert.Equal(t, "2026-09-06T10:30:00Z", dto.CreatedAt)

	list := []domain.Notification{notif}
	dtos := ToNotificationDTOList(list)
	assert.Len(t, dtos, 1)
	assert.Equal(t, dto, dtos[0])
}

func TestDTO_Absensi(t *testing.T) {
	tanggal := time.Date(2026, 9, 6, 0, 0, 0, 0, time.UTC)
	jamMasuk := time.Date(2026, 9, 6, 7, 55, 0, 0, time.UTC)
	jamPulang := time.Date(2026, 9, 6, 16, 5, 0, 0, time.UTC)
	lat := "-6.1234"
	long := "106.1234"
	selfie := "selfie.jpg"

	absensi := &domain.Absensi{
		ID:               1,
		UserID:           10,
		Tanggal:          tanggal,
		JamMasuk:         &jamMasuk,
		SelfieMasukPath:  &selfie,
		LokasiMasukLat:   &lat,
		LokasiMasukLong:  &long,
		JamPulang:        &jamPulang,
		SelfiePulangPath: &selfie,
		LokasiPulangLat:  &lat,
		LokasiPulangLong: &long,
		FaceVerified:     true,
		Status:           "hadir",
		CreatedAt:        tanggal,
	}

	dto := ToAbsensiDTO(absensi)
	assert.NotNil(t, dto)
	assert.Equal(t, uint(1), dto.ID)
	assert.Equal(t, uint(10), dto.UserID)
	assert.Equal(t, "2026-09-06T00:00:00Z", dto.Tanggal)
	assert.Equal(t, "2026-09-06T07:55:00Z", *dto.JamMasuk)
	assert.Equal(t, "2026-09-06T16:05:00Z", *dto.JamPulang)
	assert.True(t, dto.FaceVerified)
	assert.Equal(t, "hadir", dto.Status)

	// Test nil handling
	assert.Nil(t, ToAbsensiDTO(nil))
}

func TestDTO_PengajuanIzin(t *testing.T) {
	mulai := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
	selesai := time.Date(2026, 9, 12, 0, 0, 0, 0, time.UTC)
	approvedAt := time.Date(2026, 9, 9, 14, 0, 0, 0, time.UTC)
	approverID := uint(2)
	doc := "dokumen.pdf"
	komentar := "Disetujui silakan cuti"

	izin := &domain.PengajuanIzin{
		ID:               99,
		UserID:           5,
		JenisIzin:        "cuti",
		TanggalMulai:     mulai,
		TanggalSelesai:   selesai,
		Keterangan:       "Cuti tahunan",
		DokumenPath:      &doc,
		StatusApproval:   "disetujui",
		ApprovedBy:       &approverID,
		ApprovedAt:       &approvedAt,
		KomentarApprover: &komentar,
		CreatedAt:        mulai,
		User: &domain.User{
			ID:   5,
			Nama: "Pegawai Satu",
			NIP:  "19850101",
			Role: "staf",
		},
		Approver: &domain.User{
			ID:   2,
			Nama: "Pak Lurah",
			NIP:  "19700101",
			Role: "lurah",
		},
	}

	dto := ToPengajuanIzinDTO(izin)
	assert.NotNil(t, dto)
	assert.Equal(t, uint(99), dto.ID)
	assert.Equal(t, "cuti", dto.JenisIzin)
	assert.Equal(t, "2026-09-10T00:00:00Z", dto.TanggalMulai)
	assert.Equal(t, "2026-09-12T00:00:00Z", dto.TanggalSelesai)
	assert.Equal(t, "2026-09-09T14:00:00Z", *dto.ApprovedAt)
	assert.NotNil(t, dto.User)
	assert.Equal(t, "Pegawai Satu", dto.User.Nama)
	assert.NotNil(t, dto.Approver)
	assert.Equal(t, "Pak Lurah", dto.Approver.Nama)

	list := ToPengajuanIzinDTOList([]domain.PengajuanIzin{*izin})
	assert.Len(t, list, 1)
	assert.Equal(t, dto.ID, list[0].ID)

	assert.Nil(t, ToPengajuanIzinDTO(nil))
}
