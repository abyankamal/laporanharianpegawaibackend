package domain

import "strings"

// ==========================================
// 1. Role Constants & Helpers
// ==========================================

const (
	RoleAdmin      = "admin"
	RoleLurah      = "lurah"
	RoleSekertaris = "sekertaris"
	RoleKasi       = "kasi"
	RoleStaf       = "staf"
)

// AllRoles berisi seluruh role yang valid dalam sistem.
var AllRoles = []string{
	RoleAdmin,
	RoleLurah,
	RoleSekertaris,
	RoleKasi,
	RoleStaf,
}

// IsValidRole memeriksa apakah role string terdaftar dalam sistem.
func IsValidRole(role string) bool {
	r := strings.ToLower(strings.TrimSpace(role))
	switch r {
	case RoleAdmin, RoleLurah, RoleSekertaris, "sekretaris", RoleKasi, RoleStaf:
		return true
	default:
		return false
	}
}

// NormalizeRole menstandarisasi variasi role (misal: "sekretaris" -> "sekertaris", huruf besar/kecil).
func NormalizeRole(role string) string {
	r := strings.ToLower(strings.TrimSpace(role))
	if r == "sekretaris" {
		return RoleSekertaris
	}
	return r
}

// Helper pengecekan role
func IsAdmin(role string) bool {
	return NormalizeRole(role) == RoleAdmin
}

func IsLurah(role string) bool {
	return NormalizeRole(role) == RoleLurah
}

func IsSekertaris(role string) bool {
	return NormalizeRole(role) == RoleSekertaris
}

func IsKasi(role string) bool {
	return NormalizeRole(role) == RoleKasi
}

func IsStaf(role string) bool {
	return NormalizeRole(role) == RoleStaf
}

// ==========================================
// 2. Status Laporan Constants
// ==========================================

const (
	StatusLaporanMenungguReview = "menunggu_review"
	StatusLaporanDisetujui      = "disetujui"
	StatusLaporanDitolak        = "ditolak"
	StatusLaporanSudahDireview  = "sudah_direview"
)

// ==========================================
// 3. Status Absensi Constants
// ==========================================

const (
	StatusAbsensiHadir       = "hadir"
	StatusAbsensiTerlambat   = "terlambat"
	StatusAbsensiPulangCepat = "pulang_cepat"
	StatusAbsensiAlpha       = "alpha"
	StatusAbsensiIzin        = "izin"
	StatusAbsensiSakit       = "sakit"
	StatusAbsensiCuti        = "cuti"
	StatusAbsensiDinasLuar   = "dinas_luar"
)

// ==========================================
// 4. Pengajuan Izin Constants
// ==========================================

const (
	JenisIzinSakit     = "sakit"
	JenisIzinCuti      = "cuti"
	JenisIzinIzin      = "izin"
	JenisIzinDinasLuar = "dinas_luar"

	StatusIzinMenunggu  = "menunggu"
	StatusIzinDisetujui = "disetujui"
	StatusIzinDitolak   = "ditolak"
)

// ==========================================
// 5. Notifikasi Kategori Constants
// ==========================================

const (
	NotifKategoriTugas     = "Tugas"
	NotifKategoriLaporan   = "Laporan"
	NotifKategoriPenilaian = "Penilaian"
	NotifKategoriSistem    = "Sistem"
)
