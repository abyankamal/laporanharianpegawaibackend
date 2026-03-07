package service

import (
	"errors"
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type AdminService interface {
	GetRekapLaporanAdmin(filter repository.AdminReportFilter) (*repository.AdminReportResponse, error)
	GetLaporanExportAdmin(filter repository.AdminReportFilter) ([]domain.Laporan, error)
	GetDashboardSummaryAdmin() (*repository.DashboardSummaryResponse, error)

	// Pegawai Management
	GetPegawaiAdmin(filter repository.AdminPegawaiFilter) (*repository.AdminPegawaiResponse, error)
	GetPegawaiStatistikAdmin() (*repository.PegawaiStatistik, error)
	CreatePegawaiAdmin(req *domain.User) error
	UpdatePegawaiAdmin(userID uint, req *domain.User) error
	DeletePegawaiAdmin(userID uint) error
}

type adminService struct {
	adminRepo repository.AdminRepository
	userRepo  repository.UserRepository // Tambahkan ini
}

func NewAdminService(adminRepo repository.AdminRepository, userRepo repository.UserRepository) AdminService {
	return &adminService{
		adminRepo: adminRepo,
		userRepo:  userRepo,
	}
}

func (s *adminService) GetRekapLaporanAdmin(filter repository.AdminReportFilter) (*repository.AdminReportResponse, error) {
	// Panggil repository untuk menjalankan query dengan filter
	laporanList, err := s.adminRepo.GetRekapLaporanAdmin(filter)
	if err != nil {
		return nil, err
	}

	return laporanList, nil
}

func (s *adminService) GetLaporanExportAdmin(filter repository.AdminReportFilter) ([]domain.Laporan, error) {
	return s.adminRepo.GetLaporanExportAdmin(filter)
}

func (s *adminService) GetDashboardSummaryAdmin() (*repository.DashboardSummaryResponse, error) {
	return s.adminRepo.GetDashboardSummaryAdmin()
}

// ---------------------------------------------------------
// PEGAWAI MANAGEMENT
// ---------------------------------------------------------

func (s *adminService) GetPegawaiAdmin(filter repository.AdminPegawaiFilter) (*repository.AdminPegawaiResponse, error) {
	return s.adminRepo.GetPegawaiAdmin(filter)
}

func (s *adminService) GetPegawaiStatistikAdmin() (*repository.PegawaiStatistik, error) {
	return s.adminRepo.GetPegawaiStatistik()
}

func (s *adminService) CreatePegawaiAdmin(req *domain.User) error {
	// 1. Validasi NIP tidak boleh duplikat (Dicek lewat userRepo)
	existing, _ := s.userRepo.FindByNIP(req.NIP)
	if existing != nil {
		return errors.New("NIP sudah terdaftar")
	}

	// 2. Hash Password (jika tidak kosong)
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("gagal memproses password")
		}
		req.Password = string(hashed)
	}

	// 3. Simpan ke Database
	return s.userRepo.Create(req)
}

func (s *adminService) UpdatePegawaiAdmin(userID uint, req *domain.User) error {
	// 1. Cari data user lama
	existing, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("data pegawai tidak ditemukan")
	}

	// 2. Cek apakah ada perubahan NIP dan validasi duplikasi ifchanged
	if req.NIP != "" && req.NIP != existing.NIP {
		checkDuplicate, _ := s.userRepo.FindByNIP(req.NIP)
		if checkDuplicate != nil {
			return errors.New("NIP sudah digunakan oleh pegawai lain")
		}
		existing.NIP = req.NIP
	}

	// 3. Update field yang diperbolehkan
	if req.Nama != "" {
		existing.Nama = req.Nama
	}
	if req.Role != "" {
		existing.Role = req.Role
	}
	if req.JabatanID != nil {
		existing.JabatanID = req.JabatanID
	}
	if req.FotoPath != nil { // FotoPath sebagai "foto_profil" di struct db
		existing.FotoPath = req.FotoPath
	}

	// 4. Hash password baru jika ada input password
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("gagal memproses password")
		}
		existing.Password = string(hashed)
	}

	// 5. Simpan (Omit relasi dilakukan otomatis di r.userRepo.Update)
	return s.userRepo.Update(existing)
}

func (s *adminService) DeletePegawaiAdmin(userID uint) error {
	// Pengecekan pegawai ada atau tidak
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("data pegawai tidak ditemukan")
	}

	// Gunakan DeleteWithCleanup agar file/relasi jika ada ikut terbersihkan
	// Atau hapus biasa dengan r.userRepo.Delete(userID)
	// Kita gunakan Delete biasa karena requestnya mungkin simple hard delete
	return s.userRepo.Delete(userID)
}
