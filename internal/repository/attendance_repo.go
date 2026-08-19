package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// AbsensiRecapResponse adalah struct untuk response rekap absensi bulanan.
type AbsensiRecapResponse struct {
	TotalHariKerja  int `json:"total_hari_kerja"`
	TotalHadir      int `json:"total_hadir"`
	TotalTerlambat  int `json:"total_terlambat"`
	TotalPulangCepat int `json:"total_pulang_cepat"`
	TotalAlpha      int `json:"total_alpha"`
	TotalIzin       int `json:"total_izin"`
	TotalSakit      int `json:"total_sakit"`
	TotalCuti       int `json:"total_cuti"`
	TotalDinasLuar  int `json:"total_dinas_luar"`
}

// AbsensiRepository adalah interface untuk operasi database Absensi.
type AbsensiRepository interface {
	Create(absensi *domain.Absensi) error
	Update(absensi *domain.Absensi) error
	GetByUserAndDate(userID uint, tanggal time.Time) (*domain.Absensi, error)
	GetByUserAndMonth(userID uint, bulan int, tahun int) ([]domain.Absensi, error)
	GetAllByMonth(bulan int, tahun int) ([]domain.Absensi, error)
	GetTodayAbsensi(userID uint) (*domain.Absensi, error)
	GetAbsensiRecap(userID uint, bulan int, tahun int) (*AbsensiRecapResponse, error)
}

type absensiRepository struct {
	db *gorm.DB
}

// NewAbsensiRepository membuat instance baru AbsensiRepository.
func NewAbsensiRepository(db *gorm.DB) AbsensiRepository {
	return &absensiRepository{db: db}
}

// Create menyimpan data absensi baru.
func (r *absensiRepository) Create(absensi *domain.Absensi) error {
	return r.db.Create(absensi).Error
}

// Update memperbarui data absensi yang sudah ada.
func (r *absensiRepository) Update(absensi *domain.Absensi) error {
	return r.db.Save(absensi).Error
}

// GetByUserAndDate mengambil data absensi berdasarkan user ID dan tanggal.
func (r *absensiRepository) GetByUserAndDate(userID uint, tanggal time.Time) (*domain.Absensi, error) {
	var absensi domain.Absensi
	dateOnly := tanggal.Format("2006-01-02")
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("User.KategoriPegawai").
		Where("user_id = ? AND tanggal = ?", userID, dateOnly).
		First(&absensi).Error
	if err != nil {
		return nil, err
	}
	return &absensi, nil
}

// GetByUserAndMonth mengambil semua data absensi user dalam satu bulan.
func (r *absensiRepository) GetByUserAndMonth(userID uint, bulan int, tahun int) ([]domain.Absensi, error) {
	var list []domain.Absensi
	startDate := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, -1)

	err := r.db.Preload("User").Preload("User.Jabatan").Preload("User.KategoriPegawai").
		Where("user_id = ? AND tanggal BETWEEN ? AND ?", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("tanggal ASC").
		Find(&list).Error
	return list, err
}

// GetAllByMonth mengambil semua data absensi semua pegawai dalam satu bulan.
func (r *absensiRepository) GetAllByMonth(bulan int, tahun int) ([]domain.Absensi, error) {
	var list []domain.Absensi
	startDate := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, -1)

	err := r.db.Preload("User").Preload("User.Jabatan").Preload("User.KategoriPegawai").
		Where("tanggal BETWEEN ? AND ?", startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Order("user_id ASC, tanggal ASC").
		Find(&list).Error
	return list, err
}

// GetTodayAbsensi mengambil data absensi user hari ini.
func (r *absensiRepository) GetTodayAbsensi(userID uint) (*domain.Absensi, error) {
	var absensi domain.Absensi
	today := time.Now().Format("2006-01-02")
	err := r.db.Preload("User").Preload("User.KategoriPegawai").
		Where("user_id = ? AND tanggal = ?", userID, today).
		First(&absensi).Error
	if err != nil {
		return nil, err
	}
	return &absensi, nil
}

// GetAbsensiRecap menghitung rekap statistik absensi per user per bulan.
func (r *absensiRepository) GetAbsensiRecap(userID uint, bulan int, tahun int) (*AbsensiRecapResponse, error) {
	startDate := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.Local)
	endDate := startDate.AddDate(0, 1, -1)

	type CountResult struct {
		Status string
		Total  int
	}

	var results []CountResult
	err := r.db.Table("absensi").
		Select("status, COUNT(*) as total").
		Where("user_id = ? AND tanggal BETWEEN ? AND ?", userID, startDate.Format("2006-01-02"), endDate.Format("2006-01-02")).
		Group("status").
		Scan(&results).Error
	if err != nil {
		return nil, err
	}

	recap := &AbsensiRecapResponse{}
	for _, r := range results {
		switch r.Status {
		case "hadir":
			recap.TotalHadir = r.Total
		case "terlambat":
			recap.TotalTerlambat = r.Total
		case "pulang_cepat":
			recap.TotalPulangCepat = r.Total
		case "alpha":
			recap.TotalAlpha = r.Total
		case "izin":
			recap.TotalIzin = r.Total
		case "sakit":
			recap.TotalSakit = r.Total
		case "cuti":
			recap.TotalCuti = r.Total
		case "dinas_luar":
			recap.TotalDinasLuar = r.Total
		}
	}

	// Hitung total hari kerja (hadir + terlambat + pulang_cepat = hadir di kantor)
	recap.TotalHariKerja = recap.TotalHadir + recap.TotalTerlambat + recap.TotalPulangCepat +
		recap.TotalAlpha + recap.TotalIzin + recap.TotalSakit + recap.TotalCuti + recap.TotalDinasLuar

	return recap, nil
}
