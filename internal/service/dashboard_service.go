package service

import (
	"time"

	"laporanharianapi/internal/repository"
)

// RiwayatLaporan adalah struct untuk item riwayat laporan terbaru di dashboard.
type RiwayatLaporan struct {
	ID            uint      `json:"id"`
	JudulKegiatan string    `json:"judul_kegiatan"`
	CreatedAt     time.Time `json:"created_at"`
}

// DashboardSummary adalah struct response untuk dashboard.
type DashboardSummary struct {
	NamaUser            string           `json:"nama_user"`
	FotoPath            *string          `json:"foto_path"`
	TaskPending         int64            `json:"task_pending"`
	LaporanBulanIni     int64            `json:"laporan_bulan_ini"`
	LaporanMasukHariIni *int64           `json:"laporan_masuk_hari_ini,omitempty"` // Hanya muncul untuk atasan
	RiwayatTerakhir     []RiwayatLaporan `json:"riwayat_terakhir"`
}

// DashboardService adalah interface untuk operasi bisnis Dashboard.
type DashboardService interface {
	GetSummary(userID uint, userRole string) (*DashboardSummary, error)
}

// dashboardService adalah implementasi dari DashboardService.
type dashboardService struct {
	dashboardRepo repository.DashboardRepository
	userRepo      repository.UserRepository
}

// NewDashboardService membuat instance baru DashboardService.
func NewDashboardService(dashboardRepo repository.DashboardRepository, userRepo repository.UserRepository) DashboardService {
	return &dashboardService{
		dashboardRepo: dashboardRepo,
		userRepo:      userRepo,
	}
}

// GetSummary mengambil ringkasan dashboard berdasarkan role user.
func (s *dashboardService) GetSummary(userID uint, userRole string) (*DashboardSummary, error) {
	now := time.Now()

	// 1. Ambil nama user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	// 2. Hitung tugas pokok yang belum dilaporkan hari ini
	taskPending, err := s.dashboardRepo.CountTugasPendingHariIni(userID)
	if err != nil {
		return nil, err
	}

	// 3. Hitung jumlah laporan user bulan ini
	laporanBulanIni, err := s.dashboardRepo.CountLaporanByUserAndMonth(userID, now.Year(), int(now.Month()))
	if err != nil {
		return nil, err
	}

	// 4. Ambil riwayat laporan terbaru (5 terakhir)
	recentLaporan, err := s.dashboardRepo.GetRecentLaporan(userID, 3)
	if err != nil {
		return nil, err
	}

	// 5. Map ke RiwayatLaporan
	riwayat := make([]RiwayatLaporan, 0, len(recentLaporan))
	for _, l := range recentLaporan {
		riwayat = append(riwayat, RiwayatLaporan{
			ID:            l.ID,
			JudulKegiatan: l.JudulKegiatan,
			CreatedAt:     l.CreatedAt,
		})
	}

	summary := &DashboardSummary{
		NamaUser:        user.Nama,
		FotoPath:        user.FotoPath,
		TaskPending:     taskPending,
		LaporanBulanIni: laporanBulanIni,
		RiwayatTerakhir: riwayat,
	}

	// 6. Jika atasan (lurah/sekertaris), hitung laporan masuk hari ini
	if userRole == "lurah" || userRole == "sekertaris" {
		laporanHariIni, err := s.dashboardRepo.CountLaporanHariIni()
		if err != nil {
			return nil, err
		}
		summary.LaporanMasukHariIni = &laporanHariIni
	}

	return summary, nil
}
