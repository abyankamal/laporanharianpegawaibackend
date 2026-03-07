package repository

import (
	"laporanharianapi/internal/domain"

	"gorm.io/gorm"
)

// AdminReportFilter adalah struct untuk menyimpan parameter pencarian
// AdminReportFilter adalah struct untuk menyimpan parameter pencarian
type AdminReportFilter struct {
	StartDate    string // Format YYYY-MM-DD
	EndDate      string // Format YYYY-MM-DD
	StatusWaktu  string // "Tepat Waktu" atau "Lembur"
	StatusReview string // "menunggu_review" atau "sudah_direview"
	Search       string // Pencarian nama pegawai atau NIP atau Jabatan
	Page         int
	Limit        int
}

// Struct untuk mengembalikan respon paginasi rekap laporan
type AdminReportResponse struct {
	Data        []domain.Laporan `json:"data"`
	TotalData   int64            `json:"total_data"`
	TotalPage   int              `json:"total_page"`
	CurrentPage int              `json:"current_page"`
}

// ---------------------------------------------------------
// DATA STRUCTURES: PEGAWAI MANAGEMENT
// ---------------------------------------------------------

type AdminPegawaiFilter struct {
	Search string // NIP, Nama, atau Jabatan
	Role   string // 'lurah', 'sekertaris', 'kasi', 'staf', etc.
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
// DATA STRUCTURES: PENGUMUMAN MANAGEMENT
// ---------------------------------------------------------

type AdminPengumumanFilter struct {
	Search string // Pencarian Judul
	Page   int
	Limit  int
}

type AdminPengumumanResponse struct {
	Data        []domain.Notification `json:"data"`
	TotalData   int64                 `json:"total_data"`
	TotalPage   int                   `json:"total_page"`
	CurrentPage int                   `json:"current_page"`
}

type PengumumanStatistik struct {
	TotalPengumuman int64 `json:"total_pengumuman"` // Total pengumuman sistem
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
	GetRekapLaporanAdmin(filter AdminReportFilter) (*AdminReportResponse, error)
	GetLaporanExportAdmin(filter AdminReportFilter) ([]domain.Laporan, error)
	GetDashboardSummaryAdmin() (*DashboardSummaryResponse, error)

	GetPegawaiAdmin(filter AdminPegawaiFilter) (*AdminPegawaiResponse, error)
	GetPegawaiStatistik() (*PegawaiStatistik, error)

	GetPengumumanAdmin(filter AdminPengumumanFilter) (*AdminPengumumanResponse, error)
	GetPengumumanStatistikAdmin() (*PengumumanStatistik, error)
	CreatePengumumanAdmin(pengumuman *domain.Notification) error
	UpdatePengumumanAdmin(id uint, pengumuman *domain.Notification) error
	DeletePengumumanAdmin(id uint) error
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

	// Filter Role
	if filter.Role != "" && filter.Role != "Semua" {
		query = query.Where("users.role = ?", filter.Role)
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
	// Gunakan Select eksplisit untuk menghindari error jika ada kolom baru (seperti fcm_token) yang belum dimigrasi di DB
	err := query.Select("users.id, users.nip, users.nama, users.role, users.jabatan_id, users.supervisor_id, users.foto_path, users.created_at").
		Order("users.created_at DESC").
		Limit(filter.Limit).
		Offset(offset).
		Find(&users).Error

	if err != nil {
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

	// 1. Total Pegawai
	if err := r.db.Model(&domain.User{}).Where("role != 'admin'").Count(&stats.TotalPegawai).Error; err != nil {
		return nil, err
	}

	return &stats, nil
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

// helper untuk membangun query dinamis rekap laporan agar konsisten antara tabel dan export
func buildRekapLaporanQuery(db *gorm.DB, filter AdminReportFilter) *gorm.DB {
	// 1. Inisialisasi basis query pada model Laporan
	query := db.Model(&domain.Laporan{})

	// 2. Joins
	// Karena butuh filter/pencarian berelasi ke Nama, NIP, maupun Nama Jabatan:
	// INNER JOIN ke users (karena laporan pasti ada usernya), lalu LEFT JOIN ke ref_jabatan
	// Ini menjamin "users.nama" dan "ref_jabatan.nama_jabatan" tersedia untuk difilter 'LIKE' di klausul 'Where'
	query = query.Joins("INNER JOIN users ON users.id = laporan.user_id").
		Joins("LEFT JOIN ref_jabatan ON users.jabatan_id = ref_jabatan.id")

	// 3. Dynamic Query / Kondisional WHERE

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

	// Filter Status Waktu ("Tepat Waktu" atau "Lembur") (Bukan "Semua")
	if filter.StatusWaktu != "" && filter.StatusWaktu != "Semua" {
		query = query.Where("laporan.status_waktu = ?", filter.StatusWaktu)
	}

	// Filter Status Review ("menunggu_review" atau "sudah_direview") (Bukan "Semua")
	if filter.StatusReview != "" && filter.StatusReview != "Semua" {
		query = query.Where("laporan.status = ?", filter.StatusReview)
	}

	// Filter Search (Pencarian Nama, NIP, ATAU Jabatan)
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		// Menggunakan kurung penting agar kondisi OR ini tidak merusak filter AND sebelumnya
		query = query.Where("(users.nama LIKE ? OR users.nip LIKE ? OR ref_jabatan.nama_jabatan LIKE ?)", searchTerm, searchTerm, searchTerm)
	}

	return query
}

// GetRekapLaporanAdmin mengambil data laporan untuk pagination di dashboard admin
func (r *adminRepository) GetRekapLaporanAdmin(filter AdminReportFilter) (*AdminReportResponse, error) {
	var reports []domain.Laporan
	var totalData int64

	// Bangun base query filter
	query := buildRekapLaporanQuery(r.db, filter)

	// Hitung total data berdasarkan filter
	// Count tidak mempedulikan Order/Preload sehingga lebih efisien ditaruh di sini
	if err := query.Count(&totalData).Error; err != nil {
		return nil, err
	}

	// Kalkulasi Pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10 // Default limit 10
	}
	offset := (filter.Page - 1) * filter.Limit
	totalPage := int((totalData + int64(filter.Limit) - 1) / int64(filter.Limit))

	// Eksekusi Query akhir dengan Offset, Limit, Preload, dan Order
	// Preload membantu populate struct berelasi secara otomatis pasca retrieval
	err := query.
		Preload("User").
		Preload("User.Jabatan").
		Preload("TugasOrganisasi").
		Order("laporan.waktu_pelaporan DESC").
		Offset(offset).
		Limit(filter.Limit).
		Find(&reports).Error

	if err != nil {
		return nil, err
	}

	return &AdminReportResponse{
		Data:        reports,
		TotalData:   totalData,
		TotalPage:   totalPage,
		CurrentPage: filter.Page,
	}, nil
}

// GetLaporanExportAdmin mirip GetRekap namun tanpa Pagination (Tarik semua data terfilter)
// Nanti bisa dimanfaatkan service lain untuk merender array Reports menjadi Excel/PDF
func (r *adminRepository) GetLaporanExportAdmin(filter AdminReportFilter) ([]domain.Laporan, error) {
	var reports []domain.Laporan

	// Bangun base query filter yang identik
	query := buildRekapLaporanQuery(r.db, filter)

	// Lansung Preload, Order, dan Tarik data tanpa Offset dan Limit
	err := query.
		Preload("User").
		Preload("User.Jabatan").
		Preload("TugasOrganisasi").
		Order("laporan.waktu_pelaporan DESC").
		Find(&reports).Error

	if err != nil {
		return nil, err
	}

	return reports, nil
}

// ---------------------------------------------------------
// PENGUMUMAN MANAGEMENT
// ---------------------------------------------------------

func (r *adminRepository) GetPengumumanAdmin(filter AdminPengumumanFilter) (*AdminPengumumanResponse, error) {
	var notifs []domain.Notification
	var totalData int64

	// Hanya ambil yang kategori 'Sistem'
	query := r.db.Model(&domain.Notification{}).Where("kategori = 'Sistem'")

	// Pencarian berdasarkan judul
	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Where("judul LIKE ?", searchTerm)
	}

	if err := query.Count(&totalData).Error; err != nil {
		return nil, err
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 10
	}
	offset := (filter.Page - 1) * filter.Limit
	totalPage := int((totalData + int64(filter.Limit) - 1) / int64(filter.Limit))

	if err := query.Order("created_at DESC").Limit(filter.Limit).Offset(offset).Find(&notifs).Error; err != nil {
		return nil, err
	}

	return &AdminPengumumanResponse{
		Data:        notifs,
		TotalData:   totalData,
		TotalPage:   totalPage,
		CurrentPage: filter.Page,
	}, nil
}

func (r *adminRepository) GetPengumumanStatistikAdmin() (*PengumumanStatistik, error) {
	var stats PengumumanStatistik
	err := r.db.Model(&domain.Notification{}).Where("kategori = 'Sistem'").Count(&stats.TotalPengumuman).Error
	return &stats, err
}

func (r *adminRepository) CreatePengumumanAdmin(pengumuman *domain.Notification) error {
	// Pastikan kategori selalu Sistem untuk pengumuman global
	pengumuman.Kategori = "Sistem"
	return r.db.Create(pengumuman).Error
}

func (r *adminRepository) UpdatePengumumanAdmin(id uint, pengumuman *domain.Notification) error {
	// Hanya update field yang diizinkan (Judul, Pesan, UserID)
	return r.db.Model(&domain.Notification{}).Where("id = ? AND kategori = 'Sistem'", id).Updates(map[string]interface{}{
		"judul":   pengumuman.Judul,
		"pesan":   pengumuman.Pesan,
		"user_id": pengumuman.UserID, // Digunakan sebagai representasi audience
	}).Error
}

func (r *adminRepository) DeletePengumumanAdmin(id uint) error {
	return r.db.Where("id = ? AND kategori = 'Sistem'", id).Delete(&domain.Notification{}).Error
}
