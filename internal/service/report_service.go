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

	"laporanharianapi/internal/apperror"
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
	"laporanharianapi/pkg/fcm"
)

// ReportInput adalah struct untuk input pembuatan laporan.
type ReportInput struct {
	UserID            uint
	UserRole          string
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
	IsOfflineSync     bool                  // Penanda sinkronisasi offline
}

// EvaluateReportRequest adalah struct untuk input evaluasi laporan.
type EvaluateReportRequest struct {
	ReportID uint   `json:"report_id" validate:"required"`
	Status   string `json:"status" validate:"required"` // "disetujui" atau "ditolak"
	Komentar string `json:"komentar"`                   // opsional jika disetujui, wajib jika ditolak
}

// ReportService adalah interface untuk operasi bisnis Laporan.
type ReportService interface {
	CreateReport(input ReportInput) (*domain.Laporan, error)
	GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error)
	GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error)
	GetReportRecap(userID uint, startDate, endDate time.Time) (*repository.ReportRecapResponse, error)
	GetReportRecapAggregated(filter repository.ReportFilter, requesterRole string, requesterID uint) (*repository.ReportRecapResponse, error)
	EvaluateReport(assessorID uint, assessorRole string, req EvaluateReportRequest) error
	UpdateReport(id uint, judul string, deskripsi string, fileFoto *multipart.FileHeader, requesterID uint, requesterRole string) error
	DeleteReport(id uint, requesterID uint, requesterRole string) error
}

// reportService adalah implementasi dari ReportService.
type reportService struct {
	reportRepo     repository.ReportRepository
	holidayRepo    repository.HolidayRepository
	workHourRepo   repository.WorkHourRepository
	supervisorRepo repository.SupervisorRepository
	userRepo       repository.UserRepository
	notifRepo      repository.NotificationRepository
}

// NewReportService membuat instance baru ReportService.
func NewReportService(
	reportRepo repository.ReportRepository,
	holidayRepo repository.HolidayRepository,
	workHourRepo repository.WorkHourRepository,
	supervisorRepo repository.SupervisorRepository,
	userRepo repository.UserRepository,
	notifRepo repository.NotificationRepository,
) ReportService {
	return &reportService{
		reportRepo:     reportRepo,
		holidayRepo:    holidayRepo,
		workHourRepo:   workHourRepo,
		supervisorRepo: supervisorRepo,
		userRepo:       userRepo,
		notifRepo:      notifRepo,
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

	// 0. Validasi Role: Hanya 'admin' & 'lurah' yang boleh melapor tidak real-time (backdating)
	// Untuk role lain, batasi selisih waktu maksimal 30 menit dari sekarang,
	// KECUALI laporan dikirim via sinkronisasi offline (IsOfflineSync == true)
	roleLower := strings.ToLower(input.UserRole)
	if roleLower != "admin" && roleLower != "lurah" && !input.IsOfflineSync {
		diff := now.Sub(input.WaktuPelaporan)
		if diff < 0 {
			diff = -diff // abs
		}
		if diff > 30*time.Minute {
			return nil, errors.New("pelaporan non-real-time (backdating) hanya diperbolehkan untuk role Lurah")
		}
	}

	// 0.1 Idempotency Guard: Cegah duplikasi laporan saat retry sinkronisasi offline
	if input.IsOfflineSync {
		existingReports, _, errFind := s.reportRepo.GetAll(repository.ReportFilter{
			UserID:    int(input.UserID),
			StartDate: input.WaktuPelaporan.Format("2006-01-02"),
			EndDate:   input.WaktuPelaporan.Format("2006-01-02"),
			Limit:     50,
		})
		if errFind == nil {
			for _, r := range existingReports {
				if r.DeskripsiHasil == input.DeskripsiHasil && r.WaktuPelaporan.Equal(input.WaktuPelaporan) {
					// Mengembalikan laporan yang sudah tersimpan sebelumnya secara idempoten
					return &r, nil
				}
			}
		}
	}

	// 1. Cek apakah hari pelaporan adalah hari libur atau akhir pekan
	isHoliday, err := s.holidayRepo.CheckIsHoliday(input.WaktuPelaporan)
	if err != nil {
		return nil, errors.New("gagal mengecek status hari libur")
	}

	weekday := input.WaktuPelaporan.Weekday()
	isWeekend := weekday == time.Saturday || weekday == time.Sunday

	// Validasi input berdasarkan tipe laporan
	// true = pokok (pelaporan tugas pokok)
	// false = tambahan (kegiatan tambahan, wajib ada judul)
	if !input.TipeLaporan && input.JudulKegiatan == "" {
		return nil, errors.New("judul kegiatan wajib diisi untuk laporan tambahan")
	}

	// Validasi Aturan Lampiran (Wajib minimal satu: foto atau dokumen)
	if input.FileFoto == nil && input.FileDokumen == nil {
		return nil, errors.New("pelaporan gagal: wajib menyertakan setidaknya satu lampiran (foto atau dokumen)")
	}

	// Cek jam kerja dari tabel WorkHour
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
	// Jika hari libur atau akhir pekan, otomatis lembur
	if isHoliday || isWeekend {
		isOvertime = true
	} else if errParse == nil {
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

	// Default status
	status := "menunggu_review"
	if strings.ToLower(input.UserRole) == "lurah" {
		status = "disetujui" // Auto-approve untuk Lurah
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
		Status:            status,
		CreatedAt:         now,
	}

	// 7. Simpan laporan ke database
	err = s.reportRepo.Create(laporan)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan laporan: %v", err)
	}

	// 8. Kirim Notifikasi ke Atasan jika user memiliki supervisor
	if s.userRepo != nil && s.notifRepo != nil {
		user, errUser := s.userRepo.FindByID(input.UserID)
		if errUser == nil && user != nil && user.SupervisorID != nil {
			supervisor, errSup := s.userRepo.FindByID(*user.SupervisorID)
			if errSup == nil && supervisor != nil {
				judulNotif := "Laporan Harian Baru"
				pesanNotif := fmt.Sprintf("Pegawai %s telah mengirimkan laporan kegiatan: '%s'", user.Nama, input.JudulKegiatan)
				if input.JudulKegiatan == "" {
					pesanNotif = fmt.Sprintf("Pegawai %s telah mengirimkan laporan tugas pokok baru.", user.Nama)
				}
				notif := &domain.Notification{
					UserID:    int(supervisor.ID),
					Kategori:  "Laporan",
					Judul:     judulNotif,
					Pesan:     pesanNotif,
					TerkaitID: int(laporan.ID),
					CreatedAt: now,
				}
				if errNotif := s.notifRepo.Create(notif); errNotif != nil {
					log.Printf("⚠️ Gagal membuat notifikasi laporan ke atasan: %v", errNotif)
				}

				if supervisor.FCMToken != nil && *supervisor.FCMToken != "" {
					fcmToken := *supervisor.FCMToken
					go func() {
						defer func() {
							if r := recover(); r != nil {
								log.Printf("⚠️ Recovered from panic in FCM goroutine: %v", r)
							}
						}()
						fcm.SendPushNotification(fcmToken, judulNotif, pesanNotif)
					}()
				}
			}
		}
	}

	return laporan, nil
}

// GetAllReports mengambil laporan dengan filter berdasarkan role requester (RBAC).
// - Lurah: Boleh melihat SEMUA laporan.
// - Sekertaris: HANYA boleh melihat laporan milik Staf.
// - Kasi & Staf: HANYA boleh melihat laporan DIRI SENDIRI.
func (s *reportService) GetAllReports(filter repository.ReportFilter, requesterRole string, requesterID uint) ([]domain.Laporan, int64, error) {
	switch requesterRole {
	case "admin", "lurah":
		// Admin & Lurah boleh melihat semua laporan — tidak ada filter tambahan
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

	var lurahSupervisor *domain.User
	for i := range reports {
		if reports[i].User != nil && (strings.ToLower(reports[i].User.Role) == "lurah" || (reports[i].User.Jabatan != nil && strings.ToLower(reports[i].User.Jabatan.NamaJabatan) == "lurah")) {
			if reports[i].User.Supervisor == nil {
				if lurahSupervisor == nil {
					lurahSupervisor = s.getLurahSupervisorUser()
				}
				reports[i].User.Supervisor = lurahSupervisor
			}
		}
	}

	return reports, total, nil
}

// GetReportDetail mengambil detail satu laporan.
func (s *reportService) GetReportDetail(id uint, requesterRole string, requesterID uint) (*domain.Laporan, error) {
	// 1. Ambil data laporan
	laporan, err := s.reportRepo.GetByID(id)
	if err != nil {
		return nil, apperror.ErrReportNotFound
	}

	// 2. Terapkan RBAC
	switch requesterRole {
	case "admin", "lurah":
		// Bebas akses
	case "sekertaris":
		// Sekertaris boleh melihat laporan miliknya sendiri ATAU milik staf
		isOwnReport := laporan.UserID != nil && *laporan.UserID == requesterID
		isStaffReport := laporan.User != nil && laporan.User.Role == "staf"
		if !isOwnReport && !isStaffReport {
			return nil, apperror.ErrOnlyStaffOrOwnAllowed
		}
	case "kasi", "staf":
		if laporan.UserID != nil && *laporan.UserID != requesterID {
			return nil, apperror.ErrOnlyOwnReportAllowed
		}
	default:
		return nil, apperror.ErrForbidden
	}

	s.fillLurahSupervisor(laporan)

	return laporan, nil
}

// getLurahSupervisorUser mengambil data atasan lurah untuk mengisi field Supervisor pada user Lurah.
func (s *reportService) getLurahSupervisorUser() *domain.User {
	if s.supervisorRepo != nil {
		supervisorData, err := s.supervisorRepo.GetSupervisor()
		if err == nil && supervisorData != nil {
			return &domain.User{
				Nama: supervisorData.Nama,
				NIP:  supervisorData.NIP,
			}
		}
	}
	// Fallback jika gagal ambil dari DB atau nil
	return &domain.User{
		Nama: "Atasan Lurah",
		NIP:  "-",
	}
}

// fillLurahSupervisor mengisi data pejabat penilai secara dinamis jika user adalah Lurah (karena atasan lurah ada di tingkat kecamatan).
func (s *reportService) fillLurahSupervisor(laporan *domain.Laporan) {
	if laporan.User != nil && (strings.ToLower(laporan.User.Role) == "lurah" || (laporan.User.Jabatan != nil && strings.ToLower(laporan.User.Jabatan.NamaJabatan) == "lurah")) {
		if laporan.User.Supervisor == nil {
			laporan.User.Supervisor = s.getLurahSupervisorUser()
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
		// Validasi ekstensi dokumen (mencegah Stored XSS / arbitrary executable upload)
		allowedDocExts := map[string]bool{
			".pdf":  true,
			".doc":  true,
			".docx": true,
			".xls":  true,
			".xlsx": true,
			".zip":  true,
			".csv":  true,
			".jpg":  true,
			".jpeg": true,
			".png":  true,
		}
		if !allowedDocExts[extLower] {
			return "", errors.New("format dokumen tidak didukung. Gunakan PDF, DOC/DOCX, XLS/XLSX, ZIP, CSV, atau Gambar")
		}
		// Dokumen: Check size (max 200MB)
		if fileHeader.Size > 200*1024*1024 {
			return "", errors.New("ukuran dokumen maksimal 200MB")
		}
	}

	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("gagal membuka file: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file tujuan: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("gagal menyalin isi file: %w", err)
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

// GetReportRecapAggregated menghitung agregasi status dan total jam kerja laporan untuk banyak user (RBAC applied).
func (s *reportService) GetReportRecapAggregated(filter repository.ReportFilter, requesterRole string, requesterID uint) (*repository.ReportRecapResponse, error) {
	switch strings.ToLower(requesterRole) {
	case "admin", "lurah":
		// Admin & Lurah boleh melihat semua laporan.
		// Jika UserRole kosong atau "semua", biarkan kosong agar melihat semua.
		if strings.ToLower(filter.UserRole) == "semua" {
			filter.UserRole = ""
		}
	case "sekertaris", "sekretaris":
		// Sekertaris hanya boleh melihat staf dan diri sendiri
		if filter.UserRole == "" || strings.ToLower(filter.UserRole) == "semua" {
			filter.UserRole = "staf"
			filter.OwnID = int(requesterID)
		} else if strings.ToLower(filter.UserRole) == "staf" {
			filter.UserRole = "staf"
			filter.OwnID = 0 // Hanya staf
		} else if strings.ToLower(filter.UserRole) == "sekertaris" || strings.ToLower(filter.UserRole) == "sekretaris" {
			filter.UserRole = ""
			filter.UserID = int(requesterID) // Hanya diri sendiri
		} else {
			// Jika mencoba filter role lain, paksa ke diri sendiri
			filter.UserRole = ""
			filter.UserID = int(requesterID)
		}
	case "kasi", "staf", "pegawai":
		// Hanya boleh melihat diri sendiri
		filter.UserRole = ""
		filter.UserID = int(requesterID)
	default:
		return nil, errors.New("role tidak dikenali")
	}

	// Jangan gunakan OwnID jika UserID diset spesifik
	if filter.UserID > 0 {
		filter.OwnID = 0
		filter.UserRole = ""
	}

	if filter.UserRole != "" {
		filter.UserRole = strings.ToLower(filter.UserRole)
	}

	rekap, err := s.reportRepo.GetReportRecapAggregated(filter)
	if err != nil {
		return nil, err
	}
	// Sync alias untuk compatibility frontend lama
	rekap.TotalDisetujui = rekap.TotalSudahDireview
	return rekap, nil
}

// EvaluateReport mengevaluasi laporan (Memberikan Masukan) berdasarkan RBAC.
func (s *reportService) EvaluateReport(assessorID uint, assessorRole string, req EvaluateReportRequest) error {

	// Validasi status
	req.Status = strings.ToLower(req.Status)
	if req.Status != "disetujui" && req.Status != "ditolak" {
		return apperror.ErrInvalidEvaluationStatus
	}

	if req.Status == "ditolak" && strings.TrimSpace(req.Komentar) == "" {
		return apperror.ErrReasonRequired
	}
	if req.Status == "disetujui" && strings.TrimSpace(req.Komentar) == "" {
		req.Komentar = "Laporan disetujui"
	}

	// Ambil data laporan beserta relasi User pengirimnya
	laporan, err := s.reportRepo.GetByID(req.ReportID)
	if err != nil {
		return apperror.ErrReportNotFound
	}

	targetUser := laporan.User
	if targetUser == nil {
		return errors.New("data user pemilik laporan tidak valid")
	}

	// Cek apakah laporan sudah dievaluasi sebelumnya (final)
	if laporan.Status == "disetujui" || laporan.Status == "sudah_direview" {
		return apperror.ErrReportAlreadyReviewed
	}

	// Terapkan RBAC Hierarki Penilaian
	switch strings.ToLower(assessorRole) {
	case "sekertaris", "sekretaris":
		// Sekertaris HANYA boleh menilai staf (Permintaan User: Staf dikomentari Sekertaris & Lurah)
		if targetUser.Role != "staf" {
			return apperror.ErrSecretaryStaffOnly
		}
	case "admin", "lurah":
		// Admin & Lurah boleh menilai semua role
	case "kasi", "staf":
		// Kasi / Staf tidak punya hak approve laporan general
		return apperror.ErrForbidden
	default:
		return apperror.ErrForbidden
	}

	// 3. Update field
	laporan.Status = req.Status
	laporan.KomentarAtasan = &req.Komentar

	// 4. Save ke database
	err = s.reportRepo.Update(laporan)
	if err != nil {
		return fmt.Errorf("gagal mengevaluasi laporan: %v", err)
	}

	// 5. In-App Notification & FCM Push Notification ke pembuat laporan
	title := "Status Laporan Diperbarui"
	statusMsg := "Disetujui"
	if req.Status == "ditolak" {
		statusMsg = "Ditolak"
	}
	body := fmt.Sprintf("Laporan %s. Masukan: %s", statusMsg, req.Komentar)

	if s.notifRepo != nil {
		notif := &domain.Notification{
			UserID:    int(targetUser.ID),
			Kategori:  "Laporan",
			Judul:     title,
			Pesan:     body,
			TerkaitID: int(laporan.ID),
			CreatedAt: time.Now(),
		}
		if errNotif := s.notifRepo.Create(notif); errNotif != nil {
			log.Printf("⚠️ Gagal membuat in-app notifikasi evaluasi laporan: %v", errNotif)
		}
	}

	if targetUser.FCMToken != nil && *targetUser.FCMToken != "" {
		fcmToken := *targetUser.FCMToken
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("⚠️ Recovered from panic in FCM goroutine: %v", r)
				}
			}()
			fcm.SendPushNotification(fcmToken, title, body)
		}()
	}

	return nil
}

// UpdateReport memperbarui judul, detail, dan foto laporan sesuai statusnya.
func (s *reportService) UpdateReport(id uint, judul string, deskripsi string, fileFoto *multipart.FileHeader, requesterID uint, requesterRole string) error {
	// 1. Ambil data laporan
	laporan, err := s.reportRepo.GetByID(id)
	if err != nil {
		return apperror.ErrReportNotFound
	}

	// 2. RBAC: User lain hanya milik sendiri. Admin/Sekertaris tidak bisa mengedit laporan orang lain secara detail dari UI (biasanya hanya approve).
	roleBase := strings.ToLower(requesterRole)
	if roleBase != "admin" {
		if laporan.UserID == nil || *laporan.UserID != requesterID {
			return apperror.ErrOnlyOwnReportModifiable
		}
	}

	// Pengecekan status
	// Lurah diperbolehkan mengedit laporannya sendiri meskipun statusnya sudah disetujui (karena auto-approve).
	isLurahEditingOwn := roleBase == "lurah" && laporan.UserID != nil && *laporan.UserID == requesterID

	if (laporan.Status == "disetujui" || laporan.Status == "sudah_direview") && !isLurahEditingOwn {
		return apperror.ErrReportAlreadyApproved
	}

	if laporan.Status == "menunggu_review" {
		// Hanya boleh update judul dan deskripsi
		if fileFoto != nil {
			return errors.New("laporan yang masih menunggu review tidak dapat mengubah foto")
		}
	}

	// 3. Update field yang diperbolehkan
	if judul != "" {
		laporan.JudulKegiatan = judul
	}
	if deskripsi != "" {
		laporan.DeskripsiHasil = deskripsi
	}

	// Jika status ditolak, izinkan update foto dan kembalikan ke menunggu_review
	if laporan.Status == "ditolak" {
		if fileFoto != nil {
			uploadedPath, err := s.saveFile(fileFoto, "images")
			if err != nil {
				return fmt.Errorf("gagal menyimpan file foto baru: %v", err)
			}

			// Hapus foto lama jika ada
			if laporan.FotoURL != nil && *laporan.FotoURL != "" {
				oldPath := filepath.Join(".", *laporan.FotoURL)
				os.Remove(oldPath) // ignore error
			}

			laporan.FotoURL = &uploadedPath
		}

		// Reset status agar dinilai ulang
		laporan.Status = "menunggu_review"
		// Opsional: hapus komentar atasan sebelumnya
		laporan.KomentarAtasan = nil
	}

	// 4. Simpan perubahan
	err = s.reportRepo.Update(laporan)
	if err != nil {
		return fmt.Errorf("gagal memperbarui laporan: %v", err)
	}

	return nil
}

// DeleteReport menghapus laporan (Hanya Lurah).
func (s *reportService) DeleteReport(id uint, requesterID uint, requesterRole string) error {
	// 1. Ambil data laporan
	_, err := s.reportRepo.GetByID(id)
	if err != nil {
		return apperror.ErrReportNotFound
	}

	// 2. RBAC: Hanya Admin & Lurah yang boleh menghapus
	if strings.ToLower(requesterRole) != "admin" && strings.ToLower(requesterRole) != "lurah" {
		return apperror.ErrOnlyLurahCanDeleteReport
	}

	// 3. Hapus laporan
	err = s.reportRepo.Delete(id)
	if err != nil {
		return fmt.Errorf("gagal menghapus laporan: %v", err)
	}

	return nil
}
