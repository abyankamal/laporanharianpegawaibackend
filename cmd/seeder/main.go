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
	fmt.Println("   - NIP: 19800101 | Password: 123456 | Role: lurah")
	fmt.Println("   - NIP: 19900202 | Password: 123456 | Role: sekertaris")
	fmt.Println("   - NIP: 20000303 | Password: 123456 | Role: kasi")
	fmt.Println("   - NIP: 20100404 | Password: 123456 | Role: staf")
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

	// Ambil ID jabatan yang diperlukan
	var kasiPemerintahan domain.RefJabatan
	db.Where("nama_jabatan = ?", "Kasi Pemerintahan").First(&kasiPemerintahan)

	var operatorLayananOperasional domain.RefJabatan
	db.Where("nama_jabatan = ?", "Operator Layanan Operasional").First(&operatorLayananOperasional)

	// Data users
	users := []domain.User{
		{
			NIP:       "19800101",
			Nama:      "Iis Yuniawardani",
			Password:  string(hashedPassword),
			Role:      "lurah",
			CreatedAt: time.Now(),
		},
		{
			NIP:       "19900202",
			Nama:      "Bu Sekertaris",
			Password:  string(hashedPassword),
			Role:      "sekertaris",
			CreatedAt: time.Now(),
		},
		{
			NIP:       "20000303",
			Nama:      "Pak Kasi",
			Password:  string(hashedPassword),
			Role:      "kasi",
			JabatanID: &kasiPemerintahan.ID,
			CreatedAt: time.Now(),
		},
		{
			NIP:       "20100404",
			Nama:      "Mas Staf",
			Password:  string(hashedPassword),
			Role:      "staf",
			JabatanID: &operatorLayananOperasional.ID,
			CreatedAt: time.Now(),
		},
	}

	for _, user := range users {
		result := db.Where("nip = ?", user.NIP).FirstOrCreate(&user)
		if result.RowsAffected > 0 {
			fmt.Printf("   ✅ User '%s' (NIP: %s, Role: %s) berhasil ditambahkan\n", user.Nama, user.NIP, user.Role)
		} else {
			fmt.Printf("   ⏭️  User '%s' sudah ada, dilewati\n", user.Nama)
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
