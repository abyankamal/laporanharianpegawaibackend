package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"laporanharianapi/internal/domain"
)

func main() {
	fmt.Println("🚀 Memulai Database Seeder...")

	// 1. Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️  File .env tidak ditemukan, menggunakan environment variables sistem")
	}

	// 2. Setup koneksi database
	db := connectDatabase()
	fmt.Println("✅ Koneksi database berhasil")

	// 3. Seed Master Jabatan
	seedJabatan(db)

	// 4. Seed Master Skor Penilaian
	seedSkorPenilaian(db)

	// 5. Seed Users
	seedUsers(db)

	// 6. Seed Tugas Pokok (Dikomentari karena struct masih nonaktif)
	// seedTugasPokok(db)

	fmt.Println("")
	fmt.Println("🎉 ========================================")
	fmt.Println("🎉 Database Seeding Selesai!")
	fmt.Println("🎉 ========================================")
	fmt.Println("")
	fmt.Println("📝 Akun yang tersedia untuk testing:")
	fmt.Println("   - NIP: 198106152014102004 | Password: 123456 | Role: lurah")
	fmt.Println("   - NIP: 198002012009061001 | Password: 123456 | Role: sekertaris")
	fmt.Println("   - NIP: 197905172014101003 | Password: 123456 | Role: kasi")
	fmt.Println("   - NIP: 200112282025041006 | Password: 123456 | Role: staf")
	fmt.Println("   - NIP: 888888888888888888 | Password: 123456 | Role: admin")
}

// connectDatabase membuat koneksi ke MySQL
func connectDatabase() *gorm.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatalf("❌ Gagal terhubung ke database: %v", err)
	}

	return db
}

// seedJabatan mengisi data master jabatan
func seedJabatan(db *gorm.DB) {
	fmt.Println("")
	fmt.Println("📌 Seeding Master Jabatan...")

	jabatanList := []string{
		"Admin",
		"Lurah",
		"Sekertaris",
		"Kasi Pemerintahan",
		"Kasi Kesejahteraan Masyarakat",
		"Kasi Ekonomi dan Pembangunan",
		"Pengadministrasi Perkantoran",
		"Pengelola Aset",
		"Operator Layanan Operasional",
		"Operator DTKS/DTSEN",
		"Penata Kelola Sistem dan Teknologi Informasi",
	}

	for _, nama := range jabatanList {
		jabatan := domain.RefJabatan{
			NamaJabatan: nama,
			CreatedAt:   time.Now(),
		}
		result := db.Where("nama_jabatan = ?", nama).FirstOrCreate(&jabatan)
		if result.RowsAffected > 0 {
			fmt.Printf("   ✅ Jabatan '%s' berhasil ditambahkan\n", nama)
		} else {
			fmt.Printf("   ⏭️  Jabatan '%s' sudah ada, dilewati\n", nama)
		}
	}
}

// seedSkorPenilaian mengisi data master skor penilaian
func seedSkorPenilaian(db *gorm.DB) {
	fmt.Println("")
	fmt.Println("📌 Seeding Master Skor Penilaian...")

	skorList := []domain.RefSkorPenilaian{
		{ID: 1, Keterangan: "Dibawah Ekspektasi", BobotNilai: 1},
		{ID: 2, Keterangan: "Sesuai Ekspektasi", BobotNilai: 2},
		{ID: 3, Keterangan: "Diatas Ekspektasi", BobotNilai: 3},
	}

	for _, skor := range skorList {
		var existing domain.RefSkorPenilaian
		result := db.First(&existing, skor.ID)
		if result.Error != nil {
			// Data belum ada, insert
			db.Create(&skor)
			fmt.Printf("   ✅ Skor '%s' (Bobot: %d) berhasil ditambahkan\n", skor.Keterangan, skor.BobotNilai)
		} else {
			fmt.Printf("   ⏭️  Skor '%s' sudah ada, dilewati\n", skor.Keterangan)
		}
	}
}

// seedUsers mengisi data user untuk testing
func seedUsers(db *gorm.DB) {
	fmt.Println("")
	fmt.Println("📌 Seeding Users...")

	// Hash password default "123456"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("❌ Gagal hash password: %v", err)
	}

	userData := []struct {
		NIP           string
		Nama          string
		Role          string
		JabatanName   string
		SupervisorNIP string
	}{
		{"888888888888888888", "Master Admin SIOPIK", "admin", "Admin", ""},
		{"198106152014102004", "Iis Yuniawardani, S.IP", "lurah", "Lurah", ""},
		{"198002012009061001", "Aep Saepudin, S.Kom", "sekertaris", "Sekertaris", "198106152014102004"},
		{"197905172014101003", "Cahyo Dirgantoro Priyawan, A.Md", "kasi", "Kasi Ekonomi dan Pembangunan", "198106152014102004"},
		{"198102252014111001", "Budi Budiansyah", "staf", "Pengadministrasi Perkantoran", "198002012009061001"},
		{"200112282025041006", "Muhammad Abyan Kamal, S.Kom", "staf", "Penata Kelola Sistem dan Teknologi Informasi", "198002012009061001"},
		{"198001022008011003", "Kustaman, S.E", "kasi", "Kasi Pemerintahan", "198106152014102004"},
		{"196904051994031011", "Agus Haris", "kasi", "Kasi Kesejahteraan Masyarakat", "198106152014102004"},
		{"198908152025212085", "Dewi Srimulyati", "staf", "Operator Layanan Operasional", "198002012009061001"},
		{"198410022025212046", "Erlin Wili Aspiantiny", "staf", "Pengelola Aset", "198002012009061001"},
		{"198205202025211085", "Tantan Kustandi", "staf", "Operator DTKS/DTSEN", "198002012009061001"},
	}

	for _, data := range userData {
		var jab domain.RefJabatan
		db.Where("nama_jabatan = ?", data.JabatanName).First(&jab)

		var supervisorID *uint
		if data.SupervisorNIP != "" {
			var supervisor domain.User
			if err := db.Where("nip = ?", data.SupervisorNIP).First(&supervisor).Error; err == nil {
				supervisorID = &supervisor.ID
			}
		}

		user := domain.User{
			NIP:          data.NIP,
			Nama:         data.Nama,
			Password:     string(hashedPassword),
			Role:         data.Role,
			JabatanID:    &jab.ID,
			SupervisorID: supervisorID,
			CreatedAt:    time.Now(),
		}

		var existing domain.User
		result := db.Where("nip = ?", user.NIP).First(&existing)
		if result.Error != nil {
			// Jika belum ada, buat baru
			db.Create(&user)
			fmt.Printf("   ✅ User '%s' (NIP: %s, Role: %s) berhasil ditambahkan\n", user.Nama, user.NIP, user.Role)
		} else {
			// Jika sudah ada, update datanya
			existing.Nama = user.Nama
			existing.Role = user.Role
			existing.JabatanID = user.JabatanID
			existing.SupervisorID = user.SupervisorID
			db.Save(&existing)
			fmt.Printf("   🔄 User '%s' sudah ada, data diperbarui\n", user.Nama)
		}
	}
}

// seedTugasPokok mengisi data tugas pokok untuk testing
// CATATAN: Aktifkan fungsi ini setelah struct TugasPokok di domain/models.go diaktifkan
/*
func seedTugasPokok(db *gorm.DB) {
	fmt.Println("")
	fmt.Println("📌 Seeding Tugas Pokok...")

	// Ambil user "Mas Staf"
	var masStaf domain.User
	result := db.Where("nip = ?", "20100404").First(&masStaf)
	if result.Error != nil {
		fmt.Println("   ⚠️  User 'Mas Staf' tidak ditemukan, skip seeding tugas pokok")
		return
	}

	tugasList := []domain.TugasPokok{
		{
			UserID:     &masStaf.ID,
			JudulTugas: "Rekap Absensi Harian",
			Deskripsi:  "Melakukan rekap data kehadiran pegawai setiap hari kerja",
			CreatedBy:  &masStaf.ID,
			CreatedAt:  time.Now(),
		},
		{
			UserID:     &masStaf.ID,
			JudulTugas: "Pelayanan Surat Pengantar",
			Deskripsi:  "Membantu warga dalam pembuatan surat pengantar dari kelurahan",
			CreatedBy:  &masStaf.ID,
			CreatedAt:  time.Now(),
		},
	}

	for _, tugas := range tugasList {
		var existing domain.TugasPokok
		result := db.Where("judul_tugas = ? AND user_id = ?", tugas.JudulTugas, tugas.UserID).First(&existing)
		if result.Error != nil {
			db.Create(&tugas)
			fmt.Printf("   ✅ Tugas '%s' berhasil ditambahkan\n", tugas.JudulTugas)
		} else {
			fmt.Printf("   ⏭️  Tugas '%s' sudah ada, dilewati\n", tugas.JudulTugas)
		}
	}
}
*/
