package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
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

	// Pengumuman Management
	GetPengumumanAdmin(filter repository.AdminPengumumanFilter) (*repository.AdminPengumumanResponse, error)
	GetPengumumanStatistikAdmin() (*repository.PengumumanStatistik, error)
	CreatePengumumanAdmin(pengumuman *domain.Notification) error
	UpdatePengumumanAdmin(id uint, pengumuman *domain.Notification) error
	DeletePengumumanAdmin(id uint) error
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
	// 1. Validasi NIP Unique
	exists, err := s.userRepo.FindByNIP(req.NIP)
	if err == nil && exists != nil {
		return errors.New("NIP sudah terdaftar")
	}

	// 2. Hash Password (jika ada)
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("gagal mengenkripsi kata sandi")
		}
		req.Password = string(hashedPassword)
	}

	// 3. Simpan ke database
	return s.userRepo.Create(req)
}

func (s *adminService) UpdatePegawaiAdmin(userID uint, req *domain.User) error {
	// 1. Cari user lama
	existingUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("pegawai tidak ditemukan")
	}

	// 2. Cek NIP duplikat (jika NIP diubah)
	if existingUser.NIP != req.NIP {
		nipOwner, err := s.userRepo.FindByNIP(req.NIP)
		if err == nil && nipOwner != nil {
			return errors.New("NIP sudah terdaftar pada pengguna lain")
		}
	}

	// 3. Update field
	existingUser.NIP = req.NIP
	existingUser.Nama = req.Nama
	existingUser.Role = req.Role
	// Jika jabatan dan foto ingin diupdate juga, masukkan di sini
	existingUser.JabatanID = req.JabatanID
	existingUser.FotoPath = req.FotoPath

	// 4. Update Password (Hanya jika diisi/dikirim di request)
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return errors.New("gagal mengenkripsi kata sandi")
		}
		existingUser.Password = string(hashedPassword)
	}

	// 5. Simpan Perubahan
	return s.userRepo.Update(existingUser)
}

func (s *adminService) DeletePegawaiAdmin(userID uint) error {
	// Validasi eksistensi sebelum delete
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("pegawai tidak ditemukan")
	}
	return s.userRepo.Delete(userID)
}

// ---------------------------------------------------------
// PENGUMUMAN MANAGEMENT
// ---------------------------------------------------------

func (s *adminService) GetPengumumanAdmin(filter repository.AdminPengumumanFilter) (*repository.AdminPengumumanResponse, error) {
	return s.adminRepo.GetPengumumanAdmin(filter)
}

func (s *adminService) GetPengumumanStatistikAdmin() (*repository.PengumumanStatistik, error) {
	return s.adminRepo.GetPengumumanStatistikAdmin()
}

func (s *adminService) CreatePengumumanAdmin(pengumuman *domain.Notification) error {
	return s.adminRepo.CreatePengumumanAdmin(pengumuman)
}

func (s *adminService) UpdatePengumumanAdmin(id uint, pengumuman *domain.Notification) error {
	return s.adminRepo.UpdatePengumumanAdmin(id, pengumuman)
}

func (s *adminService) DeletePengumumanAdmin(id uint) error {
	return s.adminRepo.DeletePengumumanAdmin(id)
}
