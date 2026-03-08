package repository

import (
	"errors"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
)

// UserRepository adalah interface untuk operasi database User.
type UserRepository interface {
	FindByNIP(nip string) (*domain.User, error)
	FindAll() ([]domain.User, error)
	FindByID(id uint) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
	Delete(id uint) error
	DeleteWithCleanup(id uint) ([]string, error) // Transactional delete with related data and file path collection
	UpdatePassword(userID uint, newPasswordHash string) error
	UpdateFoto(userID uint, fotoPath string) error
	UpdateFCMToken(userID uint, token string) error
	FindByRoles(roles []string) ([]domain.User, error)
	FindSupervisors(roleFilter string) ([]domain.User, error)
}

// userRepository adalah implementasi dari UserRepository.
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository membuat instance baru UserRepository.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// FindByNIP mencari user berdasarkan NIP.
func (r *userRepository) FindByNIP(nip string) (*domain.User, error) {
	var user domain.User
	result := r.db.Where("nip = ?", nip).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// FindAll mengambil semua user dengan preload relasi Jabatan.
func (r *userRepository) FindAll() ([]domain.User, error) {
	var users []domain.User
	result := r.db.Preload("Jabatan").Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// FindByID mengambil user berdasarkan ID dengan preload relasi Jabatan.
func (r *userRepository) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	result := r.db.Preload("Jabatan").Preload("Supervisor").First(&user, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// Create menyimpan user baru ke database.
func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

// Update mengupdate data user.
// Kita Omit "Jabatan" dan "Supervisor" agar GORM tidak mencoba mengupdate relasi.
func (r *userRepository) Update(user *domain.User) error {
	return r.db.Model(user).Omit("Jabatan", "Supervisor").Save(user).Error
}

// Delete menghapus user berdasarkan ID.
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&domain.User{}, id).Error
}

// DeleteWithCleanup menghapus user beserta semua data terkait dalam satu transaksi.
// Mengembalikan daftar file path yang perlu dihapus dari filesystem.
func (r *userRepository) DeleteWithCleanup(id uint) ([]string, error) {
	var filePaths []string

	err := r.db.Transaction(func(tx *gorm.DB) error {
		// 1. Ambil info user (untuk foto profil)
		var user domain.User
		if err := tx.First(&user, id).Error; err != nil {
			return err
		}
		if user.FotoPath != nil && *user.FotoPath != "" {
			filePaths = append(filePaths, *user.FotoPath)
		}

		// 2. Clear tugas_assignees (M2M)
		if err := tx.Exec("DELETE FROM tugas_assignees WHERE user_id = ?", id).Error; err != nil {
			return err
		}

		// 3. Clear record terkait di tabel lain
		tx.Where("user_id = ?", id).Delete(&domain.Notification{})
		tx.Where("user_id = ? OR penilai_id = ?", id, id).Delete(&domain.Penilaian{})

		// 4. Koleksi file path dari file_laporan sebelum dihapus
		var reportFiles []domain.FileLaporan
		tx.Joins("JOIN laporan ON laporan.id = file_laporan.laporan_id").
			Where("laporan.user_id = ?", id).
			Find(&reportFiles)

		for _, f := range reportFiles {
			filePaths = append(filePaths, f.FilePath)
		}

		// 5. Hapus file_laporan dan laporan
		// GORM delete by query might not trigger hooks if any, but fine for simple delete
		if err := tx.Exec("DELETE file_laporan FROM file_laporan JOIN laporan ON laporan.id = file_laporan.laporan_id WHERE laporan.user_id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", id).Delete(&domain.Laporan{}).Error; err != nil {
			return err
		}

		// 6. Hapus tugas_organisasi yang dibuat oleh user ini
		// Sebelum itu, ambil file bukti tugas jika ada
		var tugasDocs []domain.TugasOrganisasi
		tx.Where("created_by = ?", id).Find(&tugasDocs)
		for _, t := range tugasDocs {
			if t.FileBukti != nil && *t.FileBukti != "" {
				filePaths = append(filePaths, *t.FileBukti)
			}
			// Hapus assignees untuk tugas ini
			tx.Exec("DELETE FROM tugas_assignees WHERE tugas_organisasi_id = ?", t.ID)
		}
		tx.Where("created_by = ?", id).Delete(&domain.TugasOrganisasi{})

		// 7. Nullify supervisor_id bagi yang dibawahi user ini
		tx.Model(&domain.User{}).Where("supervisor_id = ?", id).Update("supervisor_id", nil)

		// 8. Akhirnya hapus User
		if err := tx.Delete(&domain.User{}, id).Error; err != nil {
			return err
		}

		return nil
	})

	return filePaths, err
}

// UpdatePassword mengupdate password user secara spesifik.
// Method ini hanya mengupdate field password untuk menghindari perubahan field lain.
func (r *userRepository) UpdatePassword(userID uint, newPasswordHash string) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("password", newPasswordHash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

// UpdateFoto mengupdate foto profil user secara spesifik.
func (r *userRepository) UpdateFoto(userID uint, fotoPath string) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("foto_path", fotoPath)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

// UpdateFCMToken mengupdate fcm_token user secara spesifik.
func (r *userRepository) UpdateFCMToken(userID uint, token string) error {
	result := r.db.Model(&domain.User{}).Where("id = ?", userID).Update("fcm_token", token)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user tidak ditemukan")
	}
	return nil
}

// FindByRoles mengambil daftar user berdasarkan beberapa role.
func (r *userRepository) FindByRoles(roles []string) ([]domain.User, error) {
	var users []domain.User
	result := r.db.Preload("Jabatan").Where("role IN ?", roles).Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}
	return users, nil
}

// FindSupervisors mengambil daftar atasan secara dinamis berdasarkan query parameter roleFilter.
func (r *userRepository) FindSupervisors(roleFilter string) ([]domain.User, error) {
	var users []domain.User
	query := r.db.Preload("Jabatan")

	if roleFilter == "" {
		// Jika roleFilter kosong, ambil semua user yang memiliki role 'Atasan' atau 'Admin'
		// (Catatan: disesuaikan dengan role sistem: lurah, sekertaris, kasi)
		query = query.Where("role IN ?", []string{"Atasan", "lurah", "sekertaris", "kasi"})
	} else {
		// Jika roleFilter memiliki isi (misal: "sekertaris"), tambahkan kondisi
		query = query.Where("role = ?", roleFilter)
	}

	result := query.Find(&users)
	if result.Error != nil {
		return nil, result.Error
	}

	return users, nil
}
