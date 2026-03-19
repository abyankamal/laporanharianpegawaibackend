package service

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/pkg/fcm"
)

// ReportInput adalah struct untuk input pembuatan laporan.
type ReportInput struct {
	UserID            uint
	TipeLaporan       bool  // true = Pokok (linked or manual), false = Tambahan
	TugasOrganisasiID *uint // ID tugas organisasi (optional, only for linked tasks)
	JudulKegiatan     string
	DeskripsiHasil    string
	WaktuPelaporan    time.Time
	LokasiLat         string                // opsional, bisa kosong
	LokasiLong        string                // opsional, bisa kosong
	AlamatLokasi      string                // opsional, bisa kosong
	FileFoto          *multipart.FileHeader // File foto lampiran (opsional)
	FileDokumen       *multipart.FileHeader // File dokumen lampiran (opsional)
}

// EvaluateReportRequest adalah struct untuk input evaluasi laporan.
type EvaluateReportRequest struct {
	ReportID uint   `json:"report_id" validate:"required"`
	Komentar string `json:"komentar" validate:"required"`
}

// ReportService adalah interface untuk operasi bisnis Laporan.
type ReportService interface {
	CreateReport(input ReportInput) (*domain.Laporan, error)
	GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error)
	GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error)
	GetReportRecap(userID uint, startDate, endDate time.Time) (*repository.ReportRecapResponse, error)
	EvaluateReport(assessorID uint, assessorRole string, req EvaluateReportRequest) error
}

// reportService adalah implementasi dari ReportService.
type reportService struct {
	reportRepo   repository.ReportRepository
	holidayRepo  repository.HolidayRepository
	workHourRepo repository.WorkHourRepository
}

// NewReportService membuat instance baru ReportService.
func NewReportService(
	reportRepo repository.ReportRepository,
	holidayRepo repository.HolidayRepository,
	workHourRepo repository.WorkHourRepository,
) ReportService {
	return &reportService{
		reportRepo:   reportRepo,
		holidayRepo:  holidayRepo,
		workHourRepo: workHourRepo,
	}
}

// toStringPtr mengkonversi string ke pointer. Mengembalikan nil jika string kosong.
func toStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// CreateReport membuat laporan baru dengan validasi bisnis.
func (s *reportService) CreateReport(input ReportInput) (*domain.Laporan, error) {
	now := time.Now()

	// 1. Validasi: Cek apakah hari ini adalah hari libur di tabel Holiday
	isHoliday, err := s.holidayRepo.CheckIsHoliday(now)
	if err != nil {
		return nil, errors.New("gagal mengecek status hari libur")
	}
	if isHoliday {
		return nil, errors.New("laporan tidak dapat dibuat pada hari libur")
	}

	// 2. Validasi input berdasarkan tipe laporan
	// true = pokok (pelaporan tugas pokok)
	// false = tambahan (kegiatan tambahan, wajib ada judul)
	if !input.TipeLaporan && input.JudulKegiatan == "" {
		return nil, errors.New("judul kegiatan wajib diisi untuk laporan tambahan")
	}

	// 3. Cek jam kerja dari tabel WorkHour
	workHour, err := s.workHourRepo.Get()
	if err != nil {
		return nil, errors.New("gagal mengambil data pengaturan jam kerja")
	}

	// Parse jam_pulang dari form (misal "18:00")
	// Kita gunakan Dummy Date untuk diparse bersama jam agar bisa membandingkan jamnya
	dummyDate := "2006-01-02"
	formatParsing := "2006-01-02 15:04"

	// Default jam pulang jika gagal parse
	jamPulangStr := workHour.JamPulang
	if input.WaktuPelaporan.Weekday() == time.Friday {
		jamPulangStr = workHour.JamPulangJumat
	}

	parsedJamPulang, errParse := time.Parse(formatParsing, dummyDate+" "+jamPulangStr)

	isOvertime := false
	if errParse == nil {
		// Buat objek waktu WaktuPelaporan dengan tanggal dummy yang sama
		jamKirimDummy, _ := time.Parse(formatParsing, fmt.Sprintf("%s %02d:%02d", dummyDate, input.WaktuPelaporan.Hour(), input.WaktuPelaporan.Minute()))

		// Jika WaktuPelaporan (jam kirim) LEBIH BESAR DARI setting jam pulang -> lembur
		if jamKirimDummy.After(parsedJamPulang) {
			isOvertime = true
		}
	} else {
		// Fallback seperti aturan lama jika parse gagal
		currentHour := input.WaktuPelaporan.Hour()
		if currentHour < 7 || currentHour >= 16 {
			isOvertime = true
		}
	}

	// 4. Proses upload file foto jika ada
	var fotoURL *string
	if input.FileFoto != nil {
		uploadedPath, err := s.saveFile(input.FileFoto, "images")
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan file foto: %v", err)
		}
		fotoURL = &uploadedPath
	}

	// 5. Proses upload file dokumen jika ada
	var dokumenURL *string
	if input.FileDokumen != nil {
		uploadedPath, err := s.saveFile(input.FileDokumen, "documents")
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan file dokumen: %v", err)
		}
		dokumenURL = &uploadedPath
	}

	// 6. Buat struct Laporan
	userID := input.UserID
	laporan := &domain.Laporan{
		UserID:            &userID,
		TipeLaporan:       input.TipeLaporan,
		TugasOrganisasiID: input.TugasOrganisasiID,
		JudulKegiatan:     input.JudulKegiatan,
		DeskripsiHasil:    input.DeskripsiHasil,
		WaktuPelaporan:    input.WaktuPelaporan,
		IsOvertime:        isOvertime,
		LokasiLat:         toStringPtr(input.LokasiLat),
		LokasiLong:        toStringPtr(input.LokasiLong),
		AlamatLokasi:      toStringPtr(input.AlamatLokasi),
		FotoURL:           fotoURL,
		DokumenURL:        dokumenURL,
		CreatedAt:         now,
	}

	// 7. Simpan laporan ke database
	err = s.reportRepo.Create(laporan)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan laporan: %v", err)
	}

	return laporan, nil
}

// GetAllReports mengambil laporan dengan filter berdasarkan role requester (RBAC).
// - Lurah: Boleh melihat SEMUA laporan.
// - Sekertaris: HANYA boleh melihat laporan milik Staf.
// - Kasi & Staf: HANYA boleh melihat laporan DIRI SENDIRI.
func (s *reportService) GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error) {
	switch requesterRole {
	case "lurah":
		// Lurah boleh melihat semua laporan — tidak ada filter tambahan
	case "sekertaris":
		// Sekertaris boleh melihat laporan miliknya sendiri ATAU milik staf
		filter.UserRole = "staf"
		filter.OwnID = int(requesterID)
	case "kasi", "staf":
		// Kasi & Staf hanya boleh melihat laporan diri sendiri
		filter.UserID = int(requesterID)
	default:
		return nil, 0, errors.New("role tidak dikenali")
	}

	reports, total, err := s.reportRepo.GetAll(filter)
	if err != nil {
		return nil, 0, err
	}

	for i := range reports {
		s.fillLurahSupervisor(&reports[i])
	}

	return reports, total, nil
}

// GetReportDetail mengambil detail satu laporan.
func (s *reportService) GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error) {
	// 1. Ambil data laporan
	laporan, err := s.reportRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("laporan tidak ditemukan")
	}

	// 2. Terapkan RBAC
	switch requesterRole {
	case "lurah":
		// Bebas akses
	case "sekertaris":
		// Sekertaris boleh melihat laporan miliknya sendiri ATAU milik staf
		isOwnReport := laporan.UserID != nil && *laporan.UserID == requesterID
		isStaffReport := laporan.User != nil && laporan.User.Role == "staf"
		if !isOwnReport && !isStaffReport {
			return nil, errors.New("akses ditolak: hanya dapat melihat laporan staf atau milik sendiri")
		}
	case "kasi", "staf":
		if laporan.UserID != nil && *laporan.UserID != requesterID {
			return nil, errors.New("akses ditolak: hanya dapat melihat laporan milik sendiri")
		}
	default:
		return nil, errors.New("role tidak dikenali")
	}

	s.fillLurahSupervisor(laporan)

	return laporan, nil
}

// fillLurahSupervisor mengisi data pejabat penilai secara hardcoded jika user adalah Lurah (karena atasan lurah ada di tingkat kecamatan).
func (s *reportService) fillLurahSupervisor(laporan *domain.Laporan) {
	if laporan.User != nil && (strings.ToLower(laporan.User.Role) == "lurah" || (laporan.User.Jabatan != nil && strings.ToLower(laporan.User.Jabatan.NamaJabatan) == "lurah")) {
		if laporan.User.Supervisor == nil {
			laporan.User.Supervisor = &domain.User{
				Nama: "Rena Sudrajat, S.Sos., M.Si",
				NIP:  "197208241992031003",
			}
		}
	}
}

// saveFile menyimpan file ke subfolder uploads/reports/<subDir>.
// Akan otomatis dikompres jika > 5MB dan tipenya gambar (disimpan dalam folder images).
// Akan melempar error jika dokument (selain gambar) > 200MB.
func (s *reportService) saveFile(fileHeader *multipart.FileHeader, subDir string) (string, error) {
	// Pastikan folder uploads/reports/<subDir> ada
	uploadDir := filepath.Join("./uploads/reports", subDir)
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(fileHeader.Filename)
	extLower := strings.ToLower(ext)

	if subDir == "images" {
		// Validasi ekstensi (termasuk format kamera HP modern)
		if extLower != ".jpg" && extLower != ".jpeg" && extLower != ".png" && extLower != ".webp" && extLower != ".heic" {
			return "", errors.New("format file foto tidak didukung, gunakan JPG/JPEG/PNG/WEBP/HEIC")
		}
		// Validasi ukuran foto (max 50MB)
		if fileHeader.Size > 50*1024*1024 {
			return "", errors.New("ukuran foto maksimal 50MB")
		}
	} else {
		// Dokumen/Lainnya: Check size (max 200MB)
		if fileHeader.Size > 200*1024*1024 {
			return "", errors.New("ukuran dokumen maksimal 200MB")
		}
	}

	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	if subDir == "images" {
		// Simpan langsung menggunakan fasthttp.SaveMultipartFile
		err = fasthttp.SaveMultipartFile(fileHeader, destPath)
		if err != nil {
			return "", fmt.Errorf("gagal menyimpan file foto: %w", err)
		}
	} else {
		// Copy file biasa tanpa diproses image kompresi
		src, err := fileHeader.Open()
		if err != nil {
			return "", err
		}
		defer src.Close()

		dst, err := os.Create(destPath)
		if err != nil {
			return "", err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return "", err
		}
	}

	return filepath.ToSlash(destPath), nil
}

// GetReportRecap menghitung agregasi status dan total jam kerja laporan untuk rentang waktu tertentu.
func (s *reportService) GetReportRecap(userID uint, startDate, endDate time.Time) (*repository.ReportRecapResponse, error) {
	rekap, err := s.reportRepo.GetReportRecap(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	// Sync alias untuk compatibility frontend lama
	rekap.TotalDisetujui = rekap.TotalSudahDireview
	return rekap, nil
}

// EvaluateReport mengevaluasi laporan (Memberikan Masukan) berdasarkan RBAC.
func (s *reportService) EvaluateReport(assessorID uint, assessorRole string, req EvaluateReportRequest) error {

	// Ambil data laporan beserta relasi User pengirimnya
	laporan, err := s.reportRepo.GetByID(req.ReportID)
	if err != nil {
		return errors.New("laporan tidak ditemukan")
	}

	targetUser := laporan.User
	if targetUser == nil {
		return errors.New("data user pemilik laporan tidak valid")
	}

	// Cek apakah laporan sudah dievaluasi sebelumnya
	if laporan.Status == "sudah_direview" {
		return errors.New("Laporan ini sudah dievaluasi dan tidak dapat diubah")
	}

	// Terapkan RBAC Hierarki Penilaian
	switch assessorRole {
	case "sekertaris":
		// Sekertaris HANYA boleh menilai staf (Permintaan User: Staf dikomentari Sekertaris & Lurah)
		if targetUser.Role != "staf" {
			return errors.New("Sekertaris hanya memiliki hak untuk mengevaluasi laporan Staf")
		}
	case "lurah":
		// Lurah boleh menilai semua role
	case "kasi", "staf":
		// Kasi / Staf tidak punya hak approve laporan general
		return errors.New("akses ditolak")
	default:
		return errors.New("role tidak dikenali")
	}

	// 3. Update field
	laporan.Status = "sudah_direview"
	laporan.KomentarAtasan = &req.Komentar

	// 4. Save ke database
	err = s.reportRepo.Update(laporan)
	if err != nil {
		return fmt.Errorf("gagal mengevaluasi laporan: %v", err)
	}

	// 5. Trigger FCM Push Notification ke pembuat laporan
	if targetUser.FCMToken != nil && *targetUser.FCMToken != "" {
		title := "Laporan Sudah Direview"
		body := fmt.Sprintf("Masukan Atasan: %s", req.Komentar)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Recovered from panic in FCM goroutine: %v", r)
				}
			}()
			fcm.SendPushNotification(*targetUser.FCMToken, title, body)
		}()
	}

	return nil
}
