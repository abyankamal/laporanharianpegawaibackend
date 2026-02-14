package service

import (
	"time"

	"laporanharianapi/internal/repository"
)

// DashboardSummary adalah struct response untuk dashboard.
type DashboardSummary struct {
	JumlahLaporanBulanIni int64  `json:"jumlah_laporan_bulan_ini"`
	JumlahTugasPokok      int64  `json:"jumlah_tugas_pokok"`
	LaporanMasukHariIni   *int64 `json:"laporan_masuk_hari_ini,omitempty"` // Hanya muncul untuk atasan
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

	// 1. Hitung jumlah laporan user bulan ini
	jumlahLaporan, err := s.dashboardRepo.CountLaporanByUserAndMonth(userID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}

	// 2. Hitung jumlah tugas pokok user
	jumlahTugas, err := s.dashboardRepo.CountTugasPokokByUser(userID)
	if err != nil {
		return nil, err
	}

	summary := &DashboardSummary{
		JumlahLaporanBulanIni: jumlahLaporan,
		JumlahTugasPokok:      jumlahTugas,
	}

	// 3. Jika atasan (lurah/sekertaris), hitung laporan masuk hari ini
	if userRole == "lurah" || userRole == "sekertaris" {
		laporanHariIni, err := s.dashboardRepo.CountLaporanHariIni()
		if err != nil {
			return nil, err
		}
		summary.LaporanMasukHariIni = &laporanHariIni
	}

	return summary, nil
}
