package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"laporanharianapi/internal/domain"
)

// DB adalah instance global GORM untuk digunakan di seluruh aplikasi.
var DB *gorm.DB

// ConnectDatabase menginisialisasi koneksi ke database MySQL.
func ConnectDatabase() {
	// Ambil konfigurasi dari environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Buat Data Source Name (DSN) untuk MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName,
	)

	// Buka koneksi ke database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Gagal terhubung ke database: %v", err)
	}

	// Ambil underlying *sql.DB untuk konfigurasi connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Gagal mendapatkan instance sql.DB: %v", err)
	}

	// Connection Pooling (Optimasi untuk VPS 1GB RAM)
	sqlDB.SetMaxIdleConns(10)           // Jumlah koneksi idle maksimum
	sqlDB.SetMaxOpenConns(100)          // Jumlah koneksi terbuka maksimum
	sqlDB.SetConnMaxLifetime(time.Hour) // Masa hidup maksimum koneksi (1 jam)

	// Auto Migration untuk semua model
	err = db.AutoMigrate(
		&domain.User{},
		&domain.RefJabatan{},
		&domain.TugasOrganisasi{}, // Mengaktifkan AutoMigrate — akan membuat tabel tugas_organisasi + tugas_assignees (M2M)
		&domain.Laporan{},
		&domain.FileLaporan{},
		&domain.Penilaian{},
		&domain.RefSkorPenilaian{},
		&domain.Notification{},
		// &domain.HariLibur{}, // dikomen dulu, nanti diaktifkan lagi kalau sudah ada fitur hari libur
	)
	if err != nil {
		log.Fatalf("Gagal melakukan auto migration: %v", err)
	}

	log.Println("Koneksi database berhasil dan migration selesai.")

	// Set variabel global
	DB = db
}
