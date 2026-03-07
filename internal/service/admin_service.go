package service

import (
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

type AdminService interface {
	GetRekapLaporanAdmin(filter repository.AdminReportFilter) ([]domain.Laporan, error)
}

type adminService struct {
	adminRepo repository.AdminRepository
}

func NewAdminService(adminRepo repository.AdminRepository) AdminService {
	return &adminService{adminRepo: adminRepo}
}

func (s *adminService) GetRekapLaporanAdmin(filter repository.AdminReportFilter) ([]domain.Laporan, error) {
	// Panggil repository untuk menjalankan query dengan filter
	laporanList, err := s.adminRepo.GetRekapLaporanAdmin(filter)
	if err != nil {
		return nil, err
	}

	return laporanList, nil
}
