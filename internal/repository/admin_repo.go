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

// ---------------------------------------------------------
// DATA STRUCTURES: PEGAWAI MANAGEMENT
// ---------------------------------------------------------

type AdminPegawaiFilter struct {
	Search string // NIP, Nama, atau Jabatan
	Page   int
	Limit  int
}

type AdminPegawaiResponse struct {
	Data        []domain.User `json:"data"`
	TotalData   int64         `json:"total_data"`
	TotalPage   int           `json:"total_page"`
	CurrentPage int           `json:"current_page"`
}

type PegawaiStatistik struct {
	TotalPegawai int64 `json:"total_pegawai"` // Role != 'admin'
}

// ---------------------------------------------------------
// REPOSITORY INTERFACE & SETUP
// ---------------------------------------------------------

// Struct untuk Response JSON berlapis Dashboard Summary
type DashboardSummaryResponse struct {
	Statistik      StatistikDashboard       `json:"statistik"`
	LaporanTerbaru []domain.Laporan         `json:"laporan_terbaru"`
	Notifikasi     *domain.Notification     `json:"notifikasi"` // Pointer agar bisa nil jika kosong
	Agenda         []domain.TugasOrganisasi `json:"agenda"`
}

type StatistikDashboard struct {
	TotalPegawai int64 `json:"total_pegawai"`
	LaporanMasuk int64 `json:"laporan_masuk"`
	TepatWaktu   int64 `json:"tepat_waktu"`
	Lembur       int64 `json:"lembur"`
}

type AdminRepository interface {
	GetRekapLaporanAdmin(filter AdminReportFilter) ([]domain.Laporan, error)
	GetDashboardSummaryAdmin() (*DashboardSummaryResponse, error)
	GetPegawaiAdmin(filter AdminPegawaiFilter) (*AdminPegawaiResponse, error)
	GetPegawaiStatistik() (*PegawaiStatistik, error)
}

type adminRepository struct {
	db *gorm.DB
}

func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{db: db}
}

// ---------------------------------------------------------
// PEGAWAI MANAGEMENT
// ---------------------------------------------------------

func (r *adminRepository) GetPegawaiAdmin(filter AdminPegawaiFilter) (*AdminPegawaiResponse, error) {
	var users []domain.User
	var totalData int64

	// Base query: join users dengan ref_jabatan agar bisa search berdasarkan nama_jabatan
	query := r.db.Model(&domain.User{}).
		Preload("Jabatan").
		Joins("LEFT JOIN ref_jabatan ON users.jabatan_id = ref_jabatan.id")

	// Filter Search (NIP, Nama, Jabatan)
	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query = query.Where("(users.nama LIKE ? OR users.nip LIKE ? OR ref_jabatan.nama_jabatan LIKE ?)", searchPattern, searchPattern, searchPattern)
	}

	// 1. Hitung total data (sebelum pagination/limit/offset diaplikasikan)
	if err := query.Count(&totalData).Error; err != nil {
		return nil, err
	}

	// 2. Kalkulasi Pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit
	totalPage := int((totalData + int64(filter.Limit) - 1) / int64(filter.Limit))

	// 3. Ambil data dengan Limit dan Offset
	if err := query.Order("users.created_at DESC").Limit(filter.Limit).Offset(offset).Find(&users).Error; err != nil {
		return nil, err
	}

	return &AdminPegawaiResponse{
		Data:        users,
		TotalData:   totalData,
		TotalPage:   totalPage,
		CurrentPage: filter.Page,
	}, nil
}

func (r *adminRepository) GetPegawaiStatistik() (*PegawaiStatistik, error) {
	var stats PegawaiStatistik
	err := r.db.Model(&domain.User{}).Where("role != 'admin'").Count(&stats.TotalPegawai).Error
	return &stats, err
}

// ---------------------------------------------------------
// DASHBOARD SUMMARY
// ---------------------------------------------------------
// Fitur 1: Mengambil Ringkasan Dashboard (Widget Data)
// Menggunakan query goroutine-safe dari GORM dan filter time.Now()
func (r *adminRepository) GetDashboardSummaryAdmin() (*DashboardSummaryResponse, error) {
	var summary DashboardSummaryResponse

	// A. STATISTIK HARI INI
	// 1. Total Pegawai (role = 'pegawai' atau 'lurah' atau 'staf' atau 'kasi')
	// Asumsinya kita tidak menghitung 'admin' atau 'sekertaris' (bisa disesuaikan isinya)
	r.db.Model(&domain.User{}).Where("role NOT IN ?", []string{"admin", "sekertaris"}).Count(&summary.Statistik.TotalPegawai)

	// Karena kita akan sering filter *hari ini*, mari gunakan helper SQL Current Date bawaan
	// (Aman untuk berbagai timezones jika database setup-nya benar, atau kita bisa passing string date)
	todayStr := "CURDATE()" // Syntax MariaDB/MySQL untuk mendapatkan hari ini YYYY-MM-DD

	// 2. Laporan Masuk Hari Ini
	r.db.Model(&domain.Laporan{}).Where("DATE(waktu_pelaporan) = " + todayStr).Count(&summary.Statistik.LaporanMasuk)

	// 3. Tepat Waktu Hari Ini (Asumsi nama statusnya "Jam Kerja")
	r.db.Model(&domain.Laporan{}).Where("DATE(waktu_pelaporan) = " + todayStr + " AND status_waktu = 'Jam Kerja'").Count(&summary.Statistik.TepatWaktu)

	// 4. Lembur Hari Ini (Lembur biasa atau hari libur)
	r.db.Model(&domain.Laporan{}).Where("DATE(waktu_pelaporan) = " + todayStr + " AND status_waktu LIKE '%Lembur%'").Count(&summary.Statistik.Lembur)

	// B. LAPORAN TERBARU
	// Ambil 5 teratas untuk hari ini
	r.db.Model(&domain.Laporan{}).
		Preload("User"). // Ambil juga relasi data pengirimnya
		Where("DATE(waktu_pelaporan) = " + todayStr).
		Order("waktu_pelaporan DESC").
		Limit(5).
		Find(&summary.LaporanTerbaru)

	// C. NOTIFIKASI AKTIF (Pengumuman Sistem)
	// Kita ambil 1 notifikasi kategori "Sistem" terbaru yang mungkin belum dibaca secara luas
	// Ini bertindak sebagai pengumuman global
	var notif domain.Notification
	err := r.db.Model(&domain.Notification{}).
		Where("kategori = 'Sistem'").
		Order("created_at DESC").
		First(&notif).Error

	if err == nil {
		summary.Notifikasi = &notif // Set jika ketemu
	}

	// D. AGENDA / TUGAS ORGANISASI
	// Ambil dari TugasOrganisasi yang belum kedaluwarsa (misal deadline mulai dari hari ini ke depan)
	r.db.Model(&domain.TugasOrganisasi{}).
		Where("DATE(deadline) >= " + todayStr).
		Order("deadline ASC").
		Limit(5).
		Find(&summary.Agenda)

	return &summary, nil
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
