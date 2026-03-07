package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

// AdminReportFilter adalah struct untuk menyimpan parameter pencarian
type AdminReportFilter struct {
	StartDate   string // Format YYYY-MM-DD
	EndDate     string // Format YYYY-MM-DD
	StatusWaktu string // "Tepat Waktu" atau "Lembur"
	Search      string // Pencarian nama pegawai atau NIP
}

type AdminRepository interface {
	GetRekapLaporanAdmin(filter AdminReportFilter) ([]domain.Laporan, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) GetRekapLaporanAdmin(filter AdminReportFilter) ([]domain.Laporan, error) {
	var reports []domain.Laporan

	// 1. Inisialisasi basis query
	// Kita mulai query pada model Laporan
	query := r.db.Model(&domain.Laporan{})

	// Karena kita mungkin butuh mencari berdasarkan nama user (dari tabel users),
	// dan kita selalu ingin me-load data relasinya, kita join sekalian di awal
	// Ini memungkinkan kita memfilter field dari tabel users seperti 'nama' dan 'nip'
	query = query.Joins("LEFT JOIN users ON users.id = laporan.user_id")

	// 2. Dynamic Query / Kondisional WHERE

	// Filter Rentang Tanggal
	if filter.StartDate != "" && filter.EndDate != "" {
		// Jika keduanya ada, gunakan BETWEEN
		query = query.Where("laporan.waktu_pelaporan BETWEEN ? AND ?", filter.StartDate+" 00:00:00", filter.EndDate+" 23:59:59")
	} else if filter.StartDate != "" {
		// Jika hanya start date
		query = query.Where("laporan.waktu_pelaporan >= ?", filter.StartDate+" 00:00:00")
	} else if filter.EndDate != "" {
		// Jika hanya end date
		query = query.Where("laporan.waktu_pelaporan <= ?", filter.EndDate+" 23:59:59")
	}

	// Filter Status Waktu ("Tepat Waktu" atau "Lembur")
	if filter.StatusWaktu != "" {
		query = query.Where("laporan.status_waktu = ?", filter.StatusWaktu)
	}

	// Filter Search (Pencarian Nama atau NIP)
	if filter.Search != "" {
		// Tambahkan % untuk pencarian wildcard (LIKE)
		searchTerm := "%" + filter.Search + "%"
		// Menggunakan kurung (users.nama LIKE ? OR users.nip LIKE ?) agar tidak merusak filter lain
		query = query.Where("(users.nama LIKE ? OR users.nip LIKE ?)", searchTerm, searchTerm)
	}

	// 3. Eksekusi Query dengan Preload
	// Preload berguna agar GORM otomatis mengambilkan data User dan TugasOrganisasi
	// dan memasukkannya ke dalam struct berelasi.
	err := query.
		Preload("User").
		Preload("User.Jabatan").
		Preload("TugasOrganisasi").
		Order("laporan.waktu_pelaporan DESC"). // Urutkan dari yang terbaru
		Find(&reports).Error

	if err != nil {
		return nil, err
	}

	return reports, nil
}
