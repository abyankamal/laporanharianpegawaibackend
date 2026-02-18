package service

import (
	"time"

	"laporanharianapi/internal/repository"
)

// DashboardSummary adalah struct response untuk dashboard.
type DashboardSummary struct {
	TaskPending         int64  `json:"task_pending"`                     // Tugas pokok belum dilaporkan hari ini
	LaporanBulanIni     int64  `json:"laporan_bulan_ini"`                // Total laporan user bulan ini
	LaporanMasukHariIni *int64 `json:"laporan_masuk_hari_ini,omitempty"` // Hanya muncul untuk atasan
}

// DashboardService adalah interface untuk operasi bisnis Dashboard.
type DashboardService interface {
	GetSummary(userID uint, userRole string) (*DashboardSummary, error)
}

// dashboardService adalah implementasi dari DashboardService.
type dashboardService struct {
	dashboardRepo repository.DashboardRepository
}

// NewDashboardService membuat instance baru DashboardService.
func NewDashboardService(dashboardRepo repository.DashboardRepository) DashboardService {
	return &dashboardService{dashboardRepo: dashboardRepo}
}

// GetSummary mengambil ringkasan dashboard berdasarkan role user.
func (s *dashboardService) GetSummary(userID uint, userRole string) (*DashboardSummary, error) {
	now := time.Now()

	// 1. Hitung tugas pokok yang belum dilaporkan hari ini
	taskPending, err := s.dashboardRepo.CountTugasPendingHariIni(userID)
	if err != nil {
		return nil, err
	}

	// 2. Hitung jumlah laporan user bulan ini
	laporanBulanIni, err := s.dashboardRepo.CountLaporanByUserAndMonth(userID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}

	summary := &DashboardSummary{
		TaskPending:     taskPending,
		LaporanBulanIni: laporanBulanIni,
	}

	// 4. Jika atasan (lurah/sekertaris), hitung laporan masuk hari ini
	if userRole == "lurah" || userRole == "sekertaris" {
		laporanHariIni, err := s.dashboardRepo.CountLaporanHariIni()
		if err != nil {
			return nil, err
		}
		summary.LaporanMasukHariIni = &laporanHariIni
	}

	return summary, nil
}
