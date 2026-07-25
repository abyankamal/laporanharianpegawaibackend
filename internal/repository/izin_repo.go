package repository

import (
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// IzinRepository adalah interface untuk operasi database PengajuanIzin.
type IzinRepository interface {
	Create(izin *domain.PengajuanIzin) error
	Update(izin *domain.PengajuanIzin) error
	GetByID(id uint) (*domain.PengajuanIzin, error)
	GetByUserID(userID uint) ([]domain.PengajuanIzin, error)
	GetPendingApprovals() ([]domain.PengajuanIzin, error)
	GetApprovedByUserAndDateRange(userID uint, start, end time.Time) ([]domain.PengajuanIzin, error)
}

type izinRepository struct {
	db *gorm.DB
}

// NewIzinRepository membuat instance baru IzinRepository.
func NewIzinRepository(db *gorm.DB) IzinRepository {
	return &izinRepository{db: db}
}

// Create menyimpan pengajuan izin baru.
func (r *izinRepository) Create(izin *domain.PengajuanIzin) error {
	return r.db.Create(izin).Error
}

// Update memperbarui pengajuan izin yang sudah ada.
func (r *izinRepository) Update(izin *domain.PengajuanIzin) error {
	return r.db.Save(izin).Error
}

// GetByID mengambil satu pengajuan izin berdasarkan ID.
func (r *izinRepository) GetByID(id uint) (*domain.PengajuanIzin, error) {
	var izin domain.PengajuanIzin
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Approver").
		First(&izin, id).Error
	if err != nil {
		return nil, err
	}
	return &izin, nil
}

// GetByUserID mengambil semua pengajuan izin milik user tertentu, diurutkan terbaru.
func (r *izinRepository) GetByUserID(userID uint) ([]domain.PengajuanIzin, error) {
	var list []domain.PengajuanIzin
	err := r.db.Preload("User").Preload("User.Jabatan").Preload("Approver").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&list).Error
	return list, err
}

// GetPendingApprovals mengambil semua pengajuan izin yang menunggu approval.
func (r *izinRepository) GetPendingApprovals() ([]domain.PengajuanIzin, error) {
	var list []domain.PengajuanIzin
	err := r.db.Preload("User").Preload("User.Jabatan").
		Where("status_approval = ?", "menunggu").
		Order("created_at ASC").
		Find(&list).Error
	return list, err
}

// GetApprovedByUserAndDateRange mengambil pengajuan izin yang sudah disetujui untuk user
// dalam rentang tanggal tertentu (digunakan untuk menentukan status absensi).
func (r *izinRepository) GetApprovedByUserAndDateRange(userID uint, start, end time.Time) ([]domain.PengajuanIzin, error) {
	var list []domain.PengajuanIzin
	err := r.db.Where("user_id = ? AND status_approval = ? AND tanggal_mulai <= ? AND tanggal_selesai >= ?",
		userID, "disetujui", end.Format("2006-01-02"), start.Format("2006-01-02")).
		Find(&list).Error
	return list, err
}
