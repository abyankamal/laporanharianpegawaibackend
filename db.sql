-- --------------------------------------------------------
-- Host:                         127.0.0.1
-- Server version:               8.0.30 - MySQL Community Server - GPL
-- Server OS:                    Win64
-- HeidiSQL Version:             12.1.0.6537
-- --------------------------------------------------------

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET NAMES utf8 */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;


-- Dumping database structure for laporan_harian
DROP DATABASE IF EXISTS `laporan_harian`;
CREATE DATABASE IF NOT EXISTS `laporan_harian` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `laporan_harian`;

-- Dumping structure for table laporan_harian.file_laporan
DROP TABLE IF EXISTS `file_laporan`;
CREATE TABLE IF NOT EXISTS `file_laporan` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `laporan_id` bigint unsigned DEFAULT NULL,
  `tipe_file` varchar(50) DEFAULT NULL,
  `file_path` varchar(255) DEFAULT NULL,
  `metadata_exif` text,
  `uploaded_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_file_laporan_laporan` (`laporan_id`),
  CONSTRAINT `fk_file_laporan_laporan` FOREIGN KEY (`laporan_id`) REFERENCES `laporan` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.file_laporan: ~3 rows (approximately)
INSERT INTO `file_laporan` (`id`, `laporan_id`, `tipe_file`, `file_path`, `metadata_exif`, `uploaded_at`) VALUES
	(1, 1, 'image', 'uploads\\reports\\656d061e-95ed-4e67-b22d-750f2f6f1584.jpg', '', '2026-02-25 10:05:25.967'),
	(2, 2, 'document', 'uploads\\reports\\b58a4b19-19bd-4aee-8349-73ac34791e37.pdf', '', '2026-02-25 11:07:49.012'),
	(3, 3, 'image', 'uploads\\reports\\60324ba3-dffb-43d9-9ccf-2f43e2f02768.jpg', '', '2026-02-26 10:35:22.719');

-- Dumping structure for table laporan_harian.laporan
DROP TABLE IF EXISTS `laporan`;
CREATE TABLE IF NOT EXISTS `laporan` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `tipe_laporan` tinyint(1) DEFAULT NULL,
  `tugas_pokok_id` bigint unsigned DEFAULT NULL,
  `judul_kegiatan` varchar(255) DEFAULT NULL,
  `deskripsi_hasil` text,
  `waktu_pelaporan` datetime(3) DEFAULT NULL,
  `is_overtime` tinyint(1) DEFAULT '0',
  `lokasi_lat` varchar(50) DEFAULT NULL,
  `lokasi_long` varchar(50) DEFAULT NULL,
  `alamat_lokasi` text,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_laporan_tugas_pokok` (`tugas_pokok_id`),
  KEY `fk_laporan_user` (`user_id`),
  CONSTRAINT `fk_laporan_tugas_pokok` FOREIGN KEY (`tugas_pokok_id`) REFERENCES `tugas_pokok` (`id`),
  CONSTRAINT `fk_laporan_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.laporan: ~3 rows (approximately)
INSERT INTO `laporan` (`id`, `user_id`, `tipe_laporan`, `tugas_pokok_id`, `judul_kegiatan`, `deskripsi_hasil`, `waktu_pelaporan`, `is_overtime`, `lokasi_lat`, `lokasi_long`, `alamat_lokasi`, `created_at`) VALUES
	(1, 6, 1, NULL, 'Apaan ya bingung', 'Lorem ipsum dolor sit amet', '2026-02-25 17:04:32.000', 0, '-7.2492838', '107.9233216', 'Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat', '2026-02-25 10:05:25.967'),
	(2, 6, 0, NULL, 'gabut bang', 'Gitu aja udah', '2026-02-25 18:07:20.000', 0, NULL, NULL, NULL, '2026-02-25 11:07:49.012'),
	(3, 3, 0, NULL, 'Mewakili pimpinan untuk acara taraweh keliling', 'Masjid AT Thurmudiyah', '2026-02-26 17:32:59.000', 0, NULL, NULL, NULL, '2026-02-26 10:35:22.719');

-- Dumping structure for table laporan_harian.notifications
DROP TABLE IF EXISTS `notifications`;
CREATE TABLE IF NOT EXISTS `notifications` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint NOT NULL,
  `kategori` varchar(50) NOT NULL,
  `judul` varchar(255) NOT NULL,
  `pesan` text NOT NULL,
  `is_read` tinyint(1) DEFAULT '0',
  `terkait_id` bigint DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=8 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.notifications: ~6 rows (approximately)
INSERT INTO `notifications` (`id`, `user_id`, `kategori`, `judul`, `pesan`, `is_read`, `terkait_id`, `created_at`) VALUES
	(1, 2, 'Tugas', 'Tugas Baru Ditetapkan', 'Anda telah ditugaskan untuk \'Buat Absensi\'. Silakan cek detail tugas.', 1, 1, '2026-02-24 13:55:34.929'),
	(2, 4, 'Tugas', 'Tugas Baru Ditetapkan', 'Anda telah ditugaskan untuk \'Rekap Keuangan\'. Silakan cek detail tugas.', 0, 2, '2026-02-24 16:50:29.222'),
	(3, 6, 'Tugas', 'Tugas Baru Ditetapkan', 'Anda telah ditugaskan untuk \'Apaan ya bingung\'. Silakan cek detail tugas.', 0, 3, '2026-02-25 09:54:11.995'),
	(4, 2, 'Sistem', 'Pengingat Pelaporan', 'Halo Aep Saepudin, S.Kom, Anda belum mengisi laporan kinerja untuk hari ini. Segera isi sebelum jam 18:00 ya!', 0, 0, '2026-02-25 16:27:14.481'),
	(5, 3, 'Sistem', 'Pengingat Pelaporan', 'Halo Cahyo Dirgantoro Priyawan, A.Md, Anda belum mengisi laporan kinerja untuk hari ini. Segera isi sebelum jam 18:00 ya!', 0, 0, '2026-02-25 16:27:14.535'),
	(6, 4, 'Sistem', 'Pengingat Pelaporan', 'Halo Budi Budiansyah, Anda belum mengisi laporan kinerja untuk hari ini. Segera isi sebelum jam 18:00 ya!', 0, 0, '2026-02-25 16:27:14.571'),
	(7, 6, 'Penilaian', 'Penilaian Kinerja Baru', 'Atasan Anda telah memberikan penilaian kinerja untuk periode Harian.', 0, 1, '2026-02-25 17:36:20.747');

-- Dumping structure for table laporan_harian.penilaian
DROP TABLE IF EXISTS `penilaian`;
CREATE TABLE IF NOT EXISTS `penilaian` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `penilai_id` bigint unsigned DEFAULT NULL,
  `skor_id` bigint unsigned DEFAULT NULL,
  `jenis_periode` varchar(50) DEFAULT NULL,
  `tanggal_mulai` date DEFAULT NULL,
  `tanggal_selesai` date DEFAULT NULL,
  `catatan` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_penilaian_user` (`user_id`),
  KEY `fk_penilaian_penilai` (`penilai_id`),
  KEY `fk_penilaian_skor` (`skor_id`),
  CONSTRAINT `fk_penilaian_penilai` FOREIGN KEY (`penilai_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_penilaian_skor` FOREIGN KEY (`skor_id`) REFERENCES `ref_skor_penilaian` (`id`),
  CONSTRAINT `fk_penilaian_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.penilaian: ~0 rows (approximately)
INSERT INTO `penilaian` (`id`, `user_id`, `penilai_id`, `skor_id`, `jenis_periode`, `tanggal_mulai`, `tanggal_selesai`, `catatan`, `created_at`, `updated_at`) VALUES
	(1, 6, 2, 2, 'Harian', '2026-02-25', '2026-02-25', 'lanjutkan anak muda', '2026-02-25 17:36:20.747', '2026-02-25 17:36:20.747');

-- Dumping structure for table laporan_harian.ref_jabatan
DROP TABLE IF EXISTS `ref_jabatan`;
CREATE TABLE IF NOT EXISTS `ref_jabatan` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nama_jabatan` varchar(255) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_ref_jabatan_nama_jabatan` (`nama_jabatan`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.ref_jabatan: ~10 rows (approximately)
INSERT INTO `ref_jabatan` (`id`, `nama_jabatan`, `created_at`) VALUES
	(1, 'Lurah', '2026-02-24 09:31:58.281'),
	(2, 'Sekertaris', '2026-02-24 09:31:58.293'),
	(3, 'Kasi Pemerintahan', '2026-02-24 09:31:58.295'),
	(4, 'Kasi Kesejahteraan Masyarakat', '2026-02-24 09:31:58.296'),
	(5, 'Kasi Ekonomi dan Pembangunan', '2026-02-24 09:31:58.297'),
	(6, 'Pengadministrasi Perkantoran', '2026-02-24 09:31:58.298'),
	(7, 'Pengelola Aset', '2026-02-24 09:31:58.299'),
	(8, 'Operator Layanan Operasional', '2026-02-24 09:31:58.301'),
	(9, 'Operator DTKS/DTSEN', '2026-02-24 09:31:58.302'),
	(10, 'Penata Kelola Sistem dan Teknologi Informasi', '2026-02-24 09:31:58.303');

-- Dumping structure for table laporan_harian.ref_skor_penilaian
DROP TABLE IF EXISTS `ref_skor_penilaian`;
CREATE TABLE IF NOT EXISTS `ref_skor_penilaian` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `keterangan` varchar(255) DEFAULT NULL,
  `bobot_nilai` bigint DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.ref_skor_penilaian: ~3 rows (approximately)
INSERT INTO `ref_skor_penilaian` (`id`, `keterangan`, `bobot_nilai`) VALUES
	(1, 'Dibawah Ekspektasi', 1),
	(2, 'Sesuai Ekspektasi', 2),
	(3, 'Diatas Ekspektasi', 3);

-- Dumping structure for table laporan_harian.tugas_pokok
DROP TABLE IF EXISTS `tugas_pokok`;
CREATE TABLE IF NOT EXISTS `tugas_pokok` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `judul_tugas` varchar(255) DEFAULT NULL,
  `deskripsi` text,
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_tugas_pokok_user` (`user_id`),
  KEY `fk_tugas_pokok_creator` (`created_by`),
  CONSTRAINT `fk_tugas_pokok_creator` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_tugas_pokok_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.tugas_pokok: ~1 rows (approximately)
INSERT INTO `tugas_pokok` (`id`, `user_id`, `judul_tugas`, `deskripsi`, `created_by`, `created_at`) VALUES
	(1, 2, 'Buat Absensi', 'Buat Absensi', 1, '2026-02-24 13:55:34.923'),
	(3, 6, 'Apaan ya bingung', 'kumaha karep we yan wkwkwk', 2, '2026-02-25 09:54:11.975');

-- Dumping structure for table laporan_harian.users
DROP TABLE IF EXISTS `users`;
CREATE TABLE IF NOT EXISTS `users` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nip` varchar(20) NOT NULL,
  `nama` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `role` varchar(50) NOT NULL,
  `jabatan_id` bigint unsigned DEFAULT NULL,
  `supervisor_id` bigint unsigned DEFAULT NULL,
  `foto_path` varchar(255) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_nip` (`nip`),
  KEY `fk_users_jabatan` (`jabatan_id`),
  KEY `fk_users_supervisor` (`supervisor_id`),
  CONSTRAINT `fk_users_jabatan` FOREIGN KEY (`jabatan_id`) REFERENCES `ref_jabatan` (`id`),
  CONSTRAINT `fk_users_supervisor` FOREIGN KEY (`supervisor_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.users: ~10 rows (approximately)
INSERT INTO `users` (`id`, `nip`, `nama`, `password`, `role`, `jabatan_id`, `supervisor_id`, `foto_path`, `created_at`) VALUES
	(1, '198106152014102004', 'Iis Yuniawardani, S.IP', '$2a$10$fbfsmK6oT1KWVni9qC3WgOFGJpsuoXH2v51XvPdhxALdKPrmWbUwG', 'lurah', 1, NULL, NULL, '2026-02-24 09:31:58.363'),
	(2, '198002012009061001', 'Aep Saepudin, S.Kom', '$2a$10$fbfsmK6oT1KWVni9qC3WgOFGJpsuoXH2v51XvPdhxALdKPrmWbUwG', 'sekertaris', 2, 1, NULL, '2026-02-24 09:31:58.363'),
	(3, '197905172014101003', 'Cahyo Dirgantoro Priyawan, A.Md', '$2a$10$fbfsmK6oT1KWVni9qC3WgOFGJpsuoXH2v51XvPdhxALdKPrmWbUwG', 'kasi', 5, 1, NULL, '2026-02-24 09:31:58.363'),
	(4, '198102252014111001', 'Budi Budiansyah', '$2a$10$fbfsmK6oT1KWVni9qC3WgOFGJpsuoXH2v51XvPdhxALdKPrmWbUwG', 'staf', 6, 2, NULL, '2026-02-24 09:31:58.363'),
	(6, '200112282025041006', 'Muhammad Abyan Kamal, S.Kom', '$2a$10$dl2WQnTODeh8qw4AZQvB3.6t4umqYpUqZEludFXNbMlwy1PIG3gGa', 'staf', 10, 2, 'uploads\\photos\\0b1caa23-1016-426d-a5b7-34229d943b43.jpg', '2026-02-25 08:25:38.026'),
	(7, '198001022008011003', 'Kustaman, S.E', '$2a$10$m2/AUR0qXVxub/AggBYd2Oh2Xv8ZecYTq7MbPtvwzZuWI1atUqf5.', 'kasi', 3, 1, NULL, '2026-02-26 07:43:01.275'),
	(8, '196904051994031011', 'Agus Haris', '$2a$10$XJcdL6b/vuSv4oobXbnJQuV9yjRhiJVFjls2dG4fHe9ckVGf9.sgy', 'kasi', 4, 1, NULL, '2026-02-26 07:43:45.915'),
	(9, '198908152025212085', 'Dewi Srimulyati', '$2a$10$YEsHCWKcX0LxrivXrSnz.ejlisPeAQg4NcWF1EgYKzXiDKvAw6Gle', 'staf', 8, 2, NULL, '2026-02-26 07:44:51.054'),
	(10, '198410022025212046', 'Erlin Wili Aspiantiny', '$2a$10$z1ieFyqQ5msAWc1iw0DlvO1e4NETK33ioPVW1WGs4N8DecUN5ehgC', 'staf', 7, 2, NULL, '2026-02-26 07:45:41.110'),
	(11, '198205202025211085', 'Tantan Kustandi', '$2a$10$poPKmfz16xySvdDsw6uLW.99DVzOn/KHTcasw6jEKL1.Cmw75TKBy', 'staf', 9, 2, NULL, '2026-02-26 07:47:20.214');

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
