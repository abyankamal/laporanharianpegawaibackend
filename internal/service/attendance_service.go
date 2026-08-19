package service

import (
	"errors"
	"fmt"
	"math"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valyala/fasthttp"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// AbsensiCheckInInput adalah struct untuk input check-in absensi.
type AbsensiCheckInInput struct {
	UserID       uint
	LokasiLat    string
	LokasiLong   string
	FaceVerified bool
	FileSelfie   *multipart.FileHeader
}

// AbsensiCheckOutInput adalah struct untuk input check-out absensi.
type AbsensiCheckOutInput struct {
	UserID       uint
	LokasiLat    string
	LokasiLong   string
	FaceVerified bool
	FileSelfie   *multipart.FileHeader
}

// UserAbsensiRecap adalah struct untuk rekap absensi per user.
type UserAbsensiRecap struct {
	User    domain.User                  `json:"user"`
	Recap   *repository.AbsensiRecapResponse `json:"recap"`
	Details []domain.Absensi             `json:"details"`
}

// AbsensiService adalah interface untuk operasi bisnis Absensi.
type AbsensiService interface {
	CheckIn(input AbsensiCheckInInput) (*domain.Absensi, error)
	CheckOut(input AbsensiCheckOutInput) (*domain.Absensi, error)
	GetTodayStatus(userID uint) (*domain.Absensi, bool, error)
	GetMonthlyRecap(userID uint, bulan, tahun int) ([]domain.Absensi, *repository.AbsensiRecapResponse, error)
	GetAllMonthlyRecap(bulan, tahun int, users []domain.User) ([]UserAbsensiRecap, error)
	IsWorkday(date time.Time) (bool, error)
}

type absensiService struct {
	absensiRepo  repository.AbsensiRepository
	holidayRepo  repository.HolidayRepository
	workHourRepo repository.WorkHourRepository
	userRepo     repository.UserRepository
}

// NewAbsensiService membuat instance baru AbsensiService.
func NewAbsensiService(
	absensiRepo repository.AbsensiRepository,
	holidayRepo repository.HolidayRepository,
	workHourRepo repository.WorkHourRepository,
	userRepo repository.UserRepository,
) AbsensiService {
	return &absensiService{
		absensiRepo:  absensiRepo,
		holidayRepo:  holidayRepo,
		workHourRepo: workHourRepo,
		userRepo:     userRepo,
	}
}

// IsWorkday mengecek apakah tanggal tertentu adalah hari kerja (bukan weekend & bukan hari libur).
func (s *absensiService) IsWorkday(date time.Time) (bool, error) {
	// Cek weekend
	weekday := date.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false, nil
	}

	// Cek hari libur nasional
	isHoliday, err := s.holidayRepo.CheckIsHoliday(date)
	if err != nil {
		return false, err
	}
	if isHoliday {
		return false, nil
	}

	return true, nil
}

// CheckIn memproses absensi masuk pegawai.
func (s *absensiService) CheckIn(input AbsensiCheckInInput) (*domain.Absensi, error) {
	now := time.Now()

	// 1. Validasi hari kerja
	isWorkday, err := s.IsWorkday(now)
	if err != nil {
		return nil, errors.New("gagal mengecek status hari kerja")
	}
	if !isWorkday {
		return nil, errors.New("absensi tidak tersedia pada hari libur atau akhir pekan")
	}

	// 2. Cek apakah sudah absen masuk hari ini
	existing, _ := s.absensiRepo.GetTodayAbsensi(input.UserID)
	if existing != nil && existing.JamMasuk != nil {
		return nil, errors.New("Anda sudah melakukan absensi masuk hari ini")
	}

	// 2.5 Validasi foto profil terdaftar untuk verifikasi wajah
	user, err := s.userRepo.FindByID(input.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user tidak ditemukan")
	}
	if user.FotoPath == nil || *user.FotoPath == "" {
		return nil, errors.New("Anda belum mendaftarkan foto profil. Silakan unggah foto profil terlebih dahulu untuk verifikasi absensi wajah")
	}

	// 3. Validasi face verification
	if !input.FaceVerified {
		return nil, errors.New("verifikasi wajah gagal, silakan coba lagi")
	}

	// 4. Validasi geofencing jika diaktifkan
	err = s.validateGeofencing(input.LokasiLat, input.LokasiLong)
	if err != nil {
		return nil, err
	}

	// 5. Simpan selfie
	var selfiePath *string
	if input.FileSelfie != nil {
		path, err := s.saveSelfie(input.FileSelfie, "masuk")
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan selfie: %v", err)
		}
		selfiePath = &path
	}

	// 6. Tentukan status berdasarkan jam masuk
	status := s.determineCheckInStatus(now)

	// 7. Buat atau update record absensi
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if existing != nil {
		// Update record yang sudah ada (misalnya dibuat oleh sistem sebagai alpha)
		existing.JamMasuk = &now
		existing.SelfieMasukPath = selfiePath
		existing.LokasiMasukLat = toStrPtr(input.LokasiLat)
		existing.LokasiMasukLong = toStrPtr(input.LokasiLong)
		existing.FaceVerified = input.FaceVerified
		existing.Status = status

		err = s.absensiRepo.Update(existing)
		if err != nil {
			return nil, fmt.Errorf("gagal memperbarui absensi: %v", err)
		}
		return existing, nil
	}

	// Buat record baru
	absensi := &domain.Absensi{
		UserID:          input.UserID,
		Tanggal:         today,
		JamMasuk:        &now,
		SelfieMasukPath: selfiePath,
		LokasiMasukLat:  toStrPtr(input.LokasiLat),
		LokasiMasukLong: toStrPtr(input.LokasiLong),
		FaceVerified:    input.FaceVerified,
		Status:          status,
		CreatedAt:       now,
	}

	err = s.absensiRepo.Create(absensi)
	if err != nil {
		return nil, fmt.Errorf("gagal menyimpan absensi: %v", err)
	}

	return absensi, nil
}

// CheckOut memproses absensi pulang pegawai.
func (s *absensiService) CheckOut(input AbsensiCheckOutInput) (*domain.Absensi, error) {
	now := time.Now()

	// 1. Cek apakah sudah absen masuk hari ini
	existing, err := s.absensiRepo.GetTodayAbsensi(input.UserID)
	if err != nil || existing == nil {
		return nil, errors.New("Anda belum melakukan absensi masuk hari ini")
	}
	if existing.JamMasuk == nil {
		return nil, errors.New("Anda belum melakukan absensi masuk hari ini")
	}
	if existing.JamPulang != nil {
		return nil, errors.New("Anda sudah melakukan absensi pulang hari ini")
	}

	// 1.5 Validasi foto profil terdaftar untuk verifikasi wajah
	user, err := s.userRepo.FindByID(input.UserID)
	if err != nil || user == nil {
		return nil, errors.New("user tidak ditemukan")
	}
	if user.FotoPath == nil || *user.FotoPath == "" {
		return nil, errors.New("Anda belum mendaftarkan foto profil. Silakan unggah foto profil terlebih dahulu untuk verifikasi absensi wajah")
	}

	// 2. Validasi face verification
	if !input.FaceVerified {
		return nil, errors.New("verifikasi wajah gagal, silakan coba lagi")
	}

	// 3. Validasi geofencing jika diaktifkan
	err = s.validateGeofencing(input.LokasiLat, input.LokasiLong)
	if err != nil {
		return nil, err
	}

	// 4. Simpan selfie
	var selfiePath *string
	if input.FileSelfie != nil {
		path, err := s.saveSelfie(input.FileSelfie, "pulang")
		if err != nil {
			return nil, fmt.Errorf("gagal menyimpan selfie: %v", err)
		}
		selfiePath = &path
	}

	// 5. Update status jika pulang cepat
	status := s.determineCheckOutStatus(now, existing.Status)

	// 6. Update record
	existing.JamPulang = &now
	existing.SelfiePulangPath = selfiePath
	existing.LokasiPulangLat = toStrPtr(input.LokasiLat)
	existing.LokasiPulangLong = toStrPtr(input.LokasiLong)
	existing.Status = status

	err = s.absensiRepo.Update(existing)
	if err != nil {
		return nil, fmt.Errorf("gagal memperbarui absensi: %v", err)
	}

	return existing, nil
}

// GetTodayStatus mengambil status absensi hari ini beserta info apakah hari kerja.
func (s *absensiService) GetTodayStatus(userID uint) (*domain.Absensi, bool, error) {
	now := time.Now()

	isWorkday, err := s.IsWorkday(now)
	if err != nil {
		return nil, false, err
	}

	if !isWorkday {
		return nil, false, nil // Bukan hari kerja, UI absensi menghilang
	}

	absensi, _ := s.absensiRepo.GetTodayAbsensi(userID)
	return absensi, true, nil
}

// GetMonthlyRecap mengambil data absensi dan statistik rekap per user per bulan.
func (s *absensiService) GetMonthlyRecap(userID uint, bulan, tahun int) ([]domain.Absensi, *repository.AbsensiRecapResponse, error) {
	details, err := s.absensiRepo.GetByUserAndMonth(userID, bulan, tahun)
	if err != nil {
		return nil, nil, err
	}

	recap, err := s.absensiRepo.GetAbsensiRecap(userID, bulan, tahun)
	if err != nil {
		return nil, nil, err
	}

	return details, recap, nil
}

// GetAllMonthlyRecap mengambil rekap absensi semua pegawai dalam satu bulan.
func (s *absensiService) GetAllMonthlyRecap(bulan, tahun int, users []domain.User) ([]UserAbsensiRecap, error) {
	allAbsensi, err := s.absensiRepo.GetAllByMonth(bulan, tahun)
	if err != nil {
		return nil, err
	}

	// Group absensi per user
	absensiMap := make(map[uint][]domain.Absensi)
	for _, a := range allAbsensi {
		absensiMap[a.UserID] = append(absensiMap[a.UserID], a)
	}

	var result []UserAbsensiRecap
	for _, user := range users {
		details := absensiMap[user.ID]

		recap, err := s.absensiRepo.GetAbsensiRecap(user.ID, bulan, tahun)
		if err != nil {
			recap = &repository.AbsensiRecapResponse{}
		}

		result = append(result, UserAbsensiRecap{
			User:    user,
			Recap:   recap,
			Details: details,
		})
	}

	return result, nil
}

// determineCheckInStatus menentukan status absensi berdasarkan jam check-in.
func (s *absensiService) determineCheckInStatus(checkInTime time.Time) string {
	workHour, err := s.workHourRepo.Get()
	if err != nil {
		return "hadir" // Default jika gagal ambil jam kerja
	}

	// Ambil jam masuk yang sesuai hari
	jamMasukStr := workHour.JamMasuk
	if checkInTime.Weekday() == time.Friday {
		jamMasukStr = workHour.JamMasukJumat
	}

	// Parse jam masuk
	parts := strings.Split(jamMasukStr, ":")
	if len(parts) != 2 {
		return "hadir"
	}

	hour, errH := strconv.Atoi(parts[0])
	minute, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil {
		return "hadir"
	}

	// Buat batas jam masuk pada tanggal yang sama
	batasJamMasuk := time.Date(checkInTime.Year(), checkInTime.Month(), checkInTime.Day(),
		hour, minute, 0, 0, time.Local)

	// Tanpa toleransi — lewat jam masuk langsung terlambat
	if checkInTime.After(batasJamMasuk) {
		return "terlambat"
	}

	return "hadir"
}

// determineCheckOutStatus menentukan apakah pegawai pulang cepat.
func (s *absensiService) determineCheckOutStatus(checkOutTime time.Time, currentStatus string) string {
	workHour, err := s.workHourRepo.Get()
	if err != nil {
		return currentStatus // Pertahankan status check-in
	}

	// Ambil jam pulang yang sesuai hari
	jamPulangStr := workHour.JamPulang
	if checkOutTime.Weekday() == time.Friday {
		jamPulangStr = workHour.JamPulangJumat
	}

	// Parse jam pulang
	parts := strings.Split(jamPulangStr, ":")
	if len(parts) != 2 {
		return currentStatus
	}

	hour, errH := strconv.Atoi(parts[0])
	minute, errM := strconv.Atoi(parts[1])
	if errH != nil || errM != nil {
		return currentStatus
	}

	batasJamPulang := time.Date(checkOutTime.Year(), checkOutTime.Month(), checkOutTime.Day(),
		hour, minute, 0, 0, time.Local)

	if checkOutTime.Before(batasJamPulang) {
		return "pulang_cepat"
	}

	return currentStatus // Pertahankan status check-in (hadir/terlambat)
}

// validateGeofencing memvalidasi lokasi user terhadap radius kantor.
func (s *absensiService) validateGeofencing(latStr, longStr string) error {
	workHour, err := s.workHourRepo.Get()
	if err != nil || !workHour.GeofencingEnabled {
		return nil // Geofencing nonaktif, skip validasi
	}

	if workHour.KantorLat == nil || workHour.KantorLong == nil {
		return nil // Koordinat kantor belum diatur
	}

	if latStr == "" || longStr == "" {
		return errors.New("lokasi GPS wajib diaktifkan untuk absensi")
	}

	// Parse koordinat
	userLat, err1 := strconv.ParseFloat(latStr, 64)
	userLong, err2 := strconv.ParseFloat(longStr, 64)
	kantorLat, err3 := strconv.ParseFloat(*workHour.KantorLat, 64)
	kantorLong, err4 := strconv.ParseFloat(*workHour.KantorLong, 64)

	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return errors.New("format koordinat tidak valid")
	}

	// Hitung jarak menggunakan Haversine formula
	distance := haversineDistance(userLat, userLong, kantorLat, kantorLong)
	if distance > float64(workHour.RadiusMeter) {
		return fmt.Errorf("lokasi Anda di luar radius kantor (%.0f meter dari kantor, maksimal %d meter)", distance, workHour.RadiusMeter)
	}

	return nil
}

// haversineDistance menghitung jarak antara dua titik koordinat dalam meter.
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000 // meter

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// saveSelfie menyimpan file selfie absensi.
func (s *absensiService) saveSelfie(fileHeader *multipart.FileHeader, subDir string) (string, error) {
	uploadDir := filepath.Join("./uploads/attendance", subDir)
	err := os.MkdirAll(uploadDir, os.ModePerm)
	if err != nil {
		return "", err
	}

	ext := filepath.Ext(fileHeader.Filename)
	extLower := strings.ToLower(ext)

	// Validasi ekstensi
	if extLower != ".jpg" && extLower != ".jpeg" && extLower != ".png" && extLower != ".webp" && extLower != ".heic" {
		return "", errors.New("format file selfie tidak didukung, gunakan JPG/JPEG/PNG/WEBP/HEIC")
	}

	// Validasi ukuran (max 10MB)
	if fileHeader.Size > 10*1024*1024 {
		return "", errors.New("ukuran selfie maksimal 10MB")
	}

	newFileName := uuid.New().String() + ext
	destPath := filepath.Join(uploadDir, newFileName)

	err = fasthttp.SaveMultipartFile(fileHeader, destPath)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan selfie: %w", err)
	}

	return filepath.ToSlash(destPath), nil
}

// toStrPtr mengkonversi string ke pointer, nil jika kosong.
func toStrPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
