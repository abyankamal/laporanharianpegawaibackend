package scheduler

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// userBasic adalah struct ringan untuk menampung hasil query user yang belum lapor.
type userBasic struct {
	ID   uint
	Nama string
}

// StartDailyReminder menjalankan background job untuk mengirim notifikasi pengingat
// kepada pegawai yang belum membuat laporan kinerja pada hari tersebut.
// Job berjalan setiap hari kerja (Senin-Jumat) pukul 15:00 waktu lokal (tidak aktif pada hari libur).
// Fungsi ini non-blocking karena menggunakan goroutine.
func StartDailyReminder(db *gorm.DB, notifRepo repository.NotificationRepository, workHourRepo repository.WorkHourRepository, holidayRepo repository.HolidayRepository) {
	go func() {
		// Recover dari panic agar goroutine tidak crash
		defer func() {
			if r := recover(); r != nil {
				log.Printf("❌ [Scheduler] PANIC recovered: %v", r)
			}
		}()

		log.Println("⏰ [Scheduler] Daily Reminder started — target jam 15:00 setiap hari kerja")

		for {
			// 1. Hitung durasi menuju target waktu berikutnya (jam 15:00)
			now := time.Now()
			target := nextWeekdayTarget(now, 15, 0, 0)
			duration := target.Sub(now)

			log.Printf("⏰ [Scheduler] Reminder berikutnya: %s (dalam %s)",
				target.Format("2006-01-02 15:04:05"), duration.Round(time.Second))

			// 2. Tidur sampai waktu target
			timer := time.NewTimer(duration)
			<-timer.C

			// 3. Eksekusi: cari user yang belum lapor hari ini dan kirim notifikasi
			func() {
				// Recover per-eksekusi agar loop tidak berhenti jika satu eksekusi panic
				defer func() {
					if r := recover(); r != nil {
						log.Printf("❌ [Scheduler] PANIC saat eksekusi reminder: %v", r)
					}
				}()

				executeReminder(db, notifRepo, workHourRepo, holidayRepo)
			}()
		}
	}()
}

// nextWeekdayTarget menghitung waktu target berikutnya pada hari kerja (Senin-Jumat).
// Jika waktu sekarang sudah lewat jam target hari ini, targetkan hari berikutnya.
// Jika hari target jatuh pada Sabtu/Minggu, majukan ke Senin.
func nextWeekdayTarget(now time.Time, hour, min, sec int) time.Time {
	// Buat target hari ini
	target := time.Date(now.Year(), now.Month(), now.Day(), hour, min, sec, 0, now.Location())

	// Jika sudah lewat jam target, pindah ke besok
	if now.After(target) {
		target = target.Add(24 * time.Hour)
	}

	// Jika jatuh di hari Sabtu, maju ke Senin (+2 hari)
	switch target.Weekday() {
	case time.Saturday:
		target = target.Add(2 * 24 * time.Hour)
	case time.Sunday:
		target = target.Add(1 * 24 * time.Hour)
	}

	return target
}

// executeReminder menjalankan logika utama: query user yang belum lapor, lalu kirim notifikasi.
func executeReminder(db *gorm.DB, notifRepo repository.NotificationRepository, workHourRepo repository.WorkHourRepository, holidayRepo repository.HolidayRepository) {
	now := time.Now()

	// Cek apakah hari ini adalah hari libur di database
	if holidayRepo != nil {
		isHoliday, err := holidayRepo.CheckIsHoliday(now)
		if err == nil && isHoliday {
			log.Printf("🏖️ [Scheduler] Hari ini (%s) adalah hari libur. Pengingat laporan tidak dikirim.", now.Format("2006-01-02"))
			return
		}
	}

	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	todayEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())

	log.Printf("⏰ [Scheduler] Menjalankan reminder untuk tanggal %s", now.Format("2006-01-02"))

	// Query: Cari user yang BELUM membuat laporan hari ini
	// Exclude role 'lurah' & 'admin' (tidak wajib membuat laporan kinerja staf)
	// Exclude user yang sedang izin/sakit/cuti/dinas_luar hari ini
	var users []userBasic
	todayDateOnly := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	err := db.Table("users").
		Select("users.id, users.nama").
		Where("LOWER(users.role) NOT IN (?)", []string{"lurah", "admin"}).
		Where("users.id NOT IN (?)",
			db.Table("laporan").
				Select("laporan.user_id").
				Where("laporan.waktu_pelaporan >= ? AND laporan.waktu_pelaporan <= ?", todayStart, todayEnd).
				Where("laporan.user_id IS NOT NULL"),
		).
		Where("users.id NOT IN (?)",
			db.Table("absensi").
				Select("absensi.user_id").
				Where("DATE(absensi.tanggal) = DATE(?)", todayDateOnly).
				Where("absensi.status IN (?)", []string{"cuti", "sakit", "izin", "dinas_luar"}),
		).
		Find(&users).Error

	if err != nil {
		log.Printf("❌ [Scheduler] Gagal query user yang belum lapor: %v", err)
		return
	}

	if len(users) == 0 {
		log.Println("✅ [Scheduler] Semua pegawai sudah mengisi laporan hari ini. Tidak ada reminder.")
		return
	}

	// Ambil data jam kerja untuk pesan notifikasi
	jamSelesai := "18:00" // Default fallback
	wh, err := workHourRepo.Get()
	if err == nil && wh != nil {
		if now.Weekday() == time.Friday {
			jamSelesai = wh.JamPulangJumat
		} else {
			jamSelesai = wh.JamPulang
		}
	}

	// Kirim notifikasi untuk setiap user yang belum lapor
	sentCount := 0
	for _, user := range users {
		notif := &domain.Notification{
			UserID:    int(user.ID),
			Kategori:  "Sistem",
			Judul:     "Pengingat Pelaporan",
			Pesan:     fmt.Sprintf("Halo %s, Anda belum mengisi laporan kinerja untuk hari ini. Segera isi sebelum jam %s ya!", user.Nama, jamSelesai),
			TerkaitID: 0,
			CreatedAt: time.Now(),
		}

		if err := notifRepo.Create(notif); err != nil {
			log.Printf("⚠️ [Scheduler] Gagal kirim notifikasi ke %s (ID: %d): %v", user.Nama, user.ID, err)
			continue
		}
		sentCount++
	}

	log.Printf("✅ [Scheduler] Reminder berhasil dikirim ke %d/%d pegawai yang belum lapor", sentCount, len(users))
}
