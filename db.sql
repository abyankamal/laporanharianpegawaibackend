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
CREATE DATABASE IF NOT EXISTS `laporan_harian` /*!40100 DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci */ /*!80016 DEFAULT ENCRYPTION='N' */;
USE `laporan_harian`;

-- Dumping structure for table laporan_harian.file_laporan
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
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.file_laporan: ~0 rows (approximately)

-- Dumping structure for table laporan_harian.holiday
CREATE TABLE IF NOT EXISTS `holiday` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `tanggal` date DEFAULT NULL,
  `keterangan` varchar(255) DEFAULT NULL,
  `tanggal_mulai` date NOT NULL,
  `tanggal_selesai` date NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_holiday_tanggal` (`tanggal`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.holiday: ~0 rows (approximately)
INSERT INTO `holiday` (`id`, `tanggal`, `keterangan`, `tanggal_mulai`, `tanggal_selesai`) VALUES
	(1, NULL, 'Cuti Bersama Idul Fitri', '2026-03-19', '2026-03-26');

-- Dumping structure for table laporan_harian.laporan
CREATE TABLE IF NOT EXISTS `laporan` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `tipe_laporan` tinyint(1) DEFAULT NULL,
  `tugas_organisasi_id` bigint unsigned DEFAULT NULL,
  `judul_kegiatan` varchar(255) DEFAULT NULL,
  `deskripsi_hasil` text,
  `waktu_pelaporan` datetime(3) DEFAULT NULL,
  `is_overtime` tinyint(1) DEFAULT '0',
  `lokasi_lat` varchar(50) DEFAULT NULL,
  `lokasi_long` varchar(50) DEFAULT NULL,
  `alamat_lokasi` text,
  `foto_url` varchar(255) DEFAULT NULL,
  `dokumen_url` varchar(255) DEFAULT NULL,
  `status` varchar(50) DEFAULT 'menunggu_review',
  `jam_kerja` bigint DEFAULT '0',
  `komentar_atasan` text,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_laporan_user` (`user_id`),
  KEY `fk_laporan_tugas_organisasi` (`tugas_organisasi_id`),
  CONSTRAINT `fk_laporan_tugas_organisasi` FOREIGN KEY (`tugas_organisasi_id`) REFERENCES `tugas_organisasi` (`id`),
  CONSTRAINT `fk_laporan_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=6 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.laporan: ~4 rows (approximately)
INSERT INTO `laporan` (`id`, `user_id`, `tipe_laporan`, `tugas_organisasi_id`, `judul_kegiatan`, `deskripsi_hasil`, `waktu_pelaporan`, `is_overtime`, `lokasi_lat`, `lokasi_long`, `alamat_lokasi`, `foto_url`, `dokumen_url`, `status`, `jam_kerja`, `komentar_atasan`, `created_at`) VALUES
	(1, 5, 0, NULL, 'Apalah', 'Apalah', '2026-02-27 16:40:17.000', 0, '-7.2492775', '107.9233245', 'Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat', 'uploads/reports/images/7d729b97-d6b8-42c6-93c9-d624df7f67b9.jpg', 'uploads/reports/documents/36c7d7a6-a9f1-4614-8c78-92d829835ce8.pdf', 'sudah_direview', 0, 'udah bagus bang', '2026-02-27 09:40:56.944'),
	(2, 5, 1, 1, 'Melaksanakan Jumat Bersih', 'Saya sedang memberikan halaman pekarangan rumah', '2026-02-27 17:19:32.000', 0, '-7.2492742', '107.9233157', 'Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat', 'uploads/reports/images/05007224-ce8d-43fb-a169-0a58ff89c6ce.jpg', NULL, 'sudah_direview', 0, 'Fotonya tidak sesuai', '2026-02-27 10:20:04.940'),
	(3, 5, 0, NULL, 'Ngoding', 'Ngoding', '2026-02-27 17:23:42.000', 0, '-7.2492714', '107.9233149', 'Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat', 'uploads/reports/images/9b73686c-d158-4273-848e-e8753fb6cb27.jpg', 'uploads/reports/documents/92545577-a5f4-466e-8b22-32f43fe95d16.pdf', 'menunggu_review', 0, NULL, '2026-02-27 10:24:16.493'),
	(4, 3, 1, 1, 'Melaksanakan Jumat Bersih', 'Jumsih rw. 05', '2026-02-27 19:35:37.000', 0, '-7.2492768', '107.9233247', 'Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat', 'uploads/reports/images/e2f54c32-9651-4290-a40c-b27de146e3a2.jpg', NULL, 'sudah_direview', 0, 'Udah bagus', '2026-02-27 12:37:20.451'),
	(5, 1, 0, NULL, 'Tes laporan ibu lurah', 'Tes laporan ibu lurah', '2026-03-08 23:18:20.000', 1, '-7.1162702', '107.897642', 'VVMX+G3V, Haruman, Kecamatan Leles, Kabupaten Garut, Jawa Barat', 'uploads/reports/images/de6b30b7-74c6-4671-8a2d-a13418b2172a.jpg', NULL, 'menunggu_review', 0, NULL, '2026-03-08 23:18:42.981');

-- Dumping structure for table laporan_harian.notifications
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
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.notifications: ~14 rows (approximately)
INSERT INTO `notifications` (`id`, `user_id`, `kategori`, `judul`, `pesan`, `is_read`, `terkait_id`, `created_at`) VALUES
	(1, 5, 'Penilaian', 'Penilaian Kinerja Baru', 'Atasan Anda telah memberikan penilaian kinerja untuk periode Bulanan.', 0, 1, '2026-02-27 10:02:23.732'),
	(2, 1, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.595'),
	(3, 2, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.598'),
	(4, 3, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.599'),
	(5, 4, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.601'),
	(6, 5, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 1, 1, '2026-02-27 10:18:57.602'),
	(7, 6, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.604'),
	(8, 7, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.605'),
	(9, 8, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.607'),
	(10, 9, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.609'),
	(11, 10, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Jumat Bersih\'. Silakan cek detail tugas.', 0, 1, '2026-02-27 10:18:57.611'),
	(12, 10, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Apel Apalah\'. Silakan cek detail tugas.', 0, 2, '2026-03-08 08:00:31.146'),
	(13, 6, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Apel Apalah\'. Silakan cek detail tugas.', 0, 2, '2026-03-08 08:00:31.148'),
	(14, 4, 'Tugas', 'Tugas Organisasi Baru', 'Anda telah ditugaskan untuk tugas organisasi \'Melaksanakan Apel Apalah\'. Silakan cek detail tugas.', 0, 2, '2026-03-08 08:00:31.149'),
	(15, 7, 'Penilaian', 'Penilaian Kinerja Baru', 'Atasan Anda telah memberikan penilaian kinerja untuk periode Bulanan.', 0, 2, '2026-03-09 05:19:35.043');

-- Dumping structure for table laporan_harian.penilaian
CREATE TABLE IF NOT EXISTS `penilaian` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint unsigned DEFAULT NULL,
  `penilai_id` bigint unsigned DEFAULT NULL,
  `skor_id` bigint unsigned DEFAULT NULL,
  `jenis_periode` varchar(50) DEFAULT NULL,
  `bulan` bigint DEFAULT NULL,
  `tahun` bigint DEFAULT NULL,
  `tanggal_mulai` date DEFAULT NULL,
  `tanggal_selesai` date DEFAULT NULL,
  `catatan` text,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_bulan_tahun` (`user_id`,`bulan`,`tahun`),
  KEY `fk_penilaian_skor` (`skor_id`),
  KEY `fk_penilaian_penilai` (`penilai_id`),
  CONSTRAINT `fk_penilaian_penilai` FOREIGN KEY (`penilai_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_penilaian_skor` FOREIGN KEY (`skor_id`) REFERENCES `ref_skor_penilaian` (`id`),
  CONSTRAINT `fk_penilaian_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.penilaian: ~0 rows (approximately)
INSERT INTO `penilaian` (`id`, `user_id`, `penilai_id`, `skor_id`, `jenis_periode`, `bulan`, `tahun`, `tanggal_mulai`, `tanggal_selesai`, `catatan`, `created_at`, `updated_at`) VALUES
	(1, 5, 2, 2, 'Bulanan', 2, 2026, '2026-02-01', '2026-02-28', 'Udah bagus, tingkatkan', '2026-02-27 10:02:23.732', '2026-02-27 10:02:23.732'),
	(2, 7, 1, 2, 'Bulanan', 1, 2026, '2026-01-01', '2026-01-31', 'tingkatkan', '2026-03-09 05:19:35.043', '2026-03-09 05:19:35.043');

-- Dumping structure for table laporan_harian.ref_jabatan
CREATE TABLE IF NOT EXISTS `ref_jabatan` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nama_jabatan` varchar(255) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_ref_jabatan_nama_jabatan` (`nama_jabatan`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.ref_jabatan: ~10 rows (approximately)
INSERT INTO `ref_jabatan` (`id`, `nama_jabatan`, `created_at`) VALUES
	(1, 'Lurah', '2026-02-27 09:34:55.950'),
	(2, 'Sekertaris', '2026-02-27 09:34:55.962'),
	(3, 'Kasi Pemerintahan', '2026-02-27 09:34:55.964'),
	(4, 'Kasi Kesejahteraan Masyarakat', '2026-02-27 09:34:55.965'),
	(5, 'Kasi Ekonomi dan Pembangunan', '2026-02-27 09:34:55.967'),
	(6, 'Pengadministrasi Perkantoran', '2026-02-27 09:34:55.968'),
	(7, 'Pengelola Aset', '2026-02-27 09:34:55.969'),
	(8, 'Operator Layanan Operasional', '2026-02-27 09:34:55.971'),
	(9, 'Operator DTKS/DTSEN', '2026-02-27 09:34:55.972'),
	(10, 'Penata Kelola Sistem dan Teknologi Informasi', '2026-02-27 09:34:55.973'),
	(11, 'Admin', '2026-03-08 23:25:38.031');

-- Dumping structure for table laporan_harian.ref_skor_penilaian
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

-- Dumping structure for table laporan_harian.tugas_assignees
CREATE TABLE IF NOT EXISTS `tugas_assignees` (
  `tugas_organisasi_id` bigint unsigned NOT NULL,
  `user_id` bigint unsigned NOT NULL,
  PRIMARY KEY (`tugas_organisasi_id`,`user_id`),
  KEY `fk_tugas_assignees_user` (`user_id`),
  CONSTRAINT `fk_tugas_assignees_tugas_organisasi` FOREIGN KEY (`tugas_organisasi_id`) REFERENCES `tugas_organisasi` (`id`),
  CONSTRAINT `fk_tugas_assignees_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.tugas_assignees: ~13 rows (approximately)
INSERT INTO `tugas_assignees` (`tugas_organisasi_id`, `user_id`) VALUES
	(1, 1),
	(1, 2),
	(1, 3),
	(1, 4),
	(2, 4),
	(1, 5),
	(1, 6),
	(2, 6),
	(1, 7),
	(1, 8),
	(1, 9),
	(1, 10),
	(2, 10);

-- Dumping structure for table laporan_harian.tugas_organisasi
CREATE TABLE IF NOT EXISTS `tugas_organisasi` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `judul_tugas` varchar(255) DEFAULT NULL,
  `deskripsi` text,
  `file_bukti` varchar(255) DEFAULT NULL,
  `created_by` bigint unsigned DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `deadline` timestamp NULL DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_tugas_organisasi_creator` (`created_by`),
  CONSTRAINT `fk_tugas_organisasi_creator` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.tugas_organisasi: ~1 rows (approximately)
INSERT INTO `tugas_organisasi` (`id`, `judul_tugas`, `deskripsi`, `file_bukti`, `created_by`, `created_at`, `deadline`) VALUES
	(1, 'Melaksanakan Jumat Bersih', 'Geura beberes kantor', NULL, 1, '2026-02-27 10:18:57.579', '2026-03-12 16:59:00'),
	(2, 'Melaksanakan Apel Apalah', 'TTTT', 'uploads/reports/documents/task-5bc7fb23-e361-43a9-890b-2ad968a07e82.pdf', 1, '2026-03-08 08:00:31.133', '2026-03-26 05:25:00');

-- Dumping structure for table laporan_harian.users
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
  `fcm_token` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_nip` (`nip`),
  KEY `fk_users_jabatan` (`jabatan_id`),
  KEY `fk_users_supervisor` (`supervisor_id`),
  CONSTRAINT `fk_users_jabatan` FOREIGN KEY (`jabatan_id`) REFERENCES `ref_jabatan` (`id`),
  CONSTRAINT `fk_users_supervisor` FOREIGN KEY (`supervisor_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.users: ~10 rows (approximately)
INSERT INTO `users` (`id`, `nip`, `nama`, `password`, `role`, `jabatan_id`, `supervisor_id`, `foto_path`, `created_at`, `fcm_token`) VALUES
	(1, '198106152014102004', 'Iis Yuniawardani, S.IP', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'lurah', 1, NULL, NULL, '2026-02-27 09:34:56.035', NULL),
	(2, '198002012009061001', 'Aep Saepudin, S.Kom', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'sekertaris', 2, 1, NULL, '2026-02-27 09:34:56.039', NULL),
	(3, '197905172014101003', 'Cahyo Dirgantoro Priyawan, A.Md', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'kasi', 5, 1, NULL, '2026-02-27 09:34:56.041', NULL),
	(4, '198102252014111001', 'Budi Budiansyah', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'staf', 6, 2, NULL, '2026-02-27 09:34:56.043', NULL),
	(5, '200112282025041006', 'Muhammad Abyan Kamal, S.Kom', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'staf', 10, 2, 'uploads/photos/a93bd524-3cb0-4ac8-8708-33204263983a.jpg', '2026-02-27 09:34:56.045', NULL),
	(6, '198001022008011003', 'Kustaman, S.E', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'kasi', 3, 1, NULL, '2026-02-27 09:34:56.047', NULL),
	(7, '196904051994031011', 'Agus Haris', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'kasi', 4, 1, NULL, '2026-02-27 09:34:56.049', NULL),
	(8, '198908152025212085', 'Dewi Srimulyati', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'staf', 8, 2, NULL, '2026-02-27 09:34:56.050', NULL),
	(9, '198410022025212046', 'Erlin Wili Aspiantiny', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'staf', 7, 2, NULL, '2026-02-27 09:34:56.052', NULL),
	(10, '198205202025211085', 'Tantan Kustandi', '$2a$10$UcA.Ok4McFfbpn9JFunYzOA25Tn1GkhFSiPpc5Bl1AoKj1wbR0pv.', 'staf', 9, 2, NULL, '2026-02-27 09:34:56.054', NULL),
	(11, '888888888888888888', 'Master Admin SIOPIK', '$2a$10$2OzwbJTA2KtuziYxYFz8SusfMzJoOICsmXdthPpb/ikze0Mq3tB6y', 'admin', 11, NULL, NULL, '2026-03-08 23:25:38.094', NULL);

-- Dumping structure for table laporan_harian.work_hour
CREATE TABLE IF NOT EXISTS `work_hour` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `jam_masuk` varchar(5) DEFAULT NULL,
  `jam_pulang` varchar(5) DEFAULT NULL,
  `jam_masuk_jumat` varchar(5) DEFAULT NULL,
  `jam_pulang_jumat` varchar(5) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=2 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- Dumping data for table laporan_harian.work_hour: ~1 rows (approximately)
INSERT INTO `work_hour` (`id`, `jam_masuk`, `jam_pulang`, `jam_masuk_jumat`, `jam_pulang_jumat`) VALUES
	(1, '06:00', '14:30', '06:00', '16:00');

/*!40103 SET TIME_ZONE=IFNULL(@OLD_TIME_ZONE, 'system') */;
/*!40101 SET SQL_MODE=IFNULL(@OLD_SQL_MODE, '') */;
/*!40014 SET FOREIGN_KEY_CHECKS=IFNULL(@OLD_FOREIGN_KEY_CHECKS, 1) */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40111 SET SQL_NOTES=IFNULL(@OLD_SQL_NOTES, 1) */;
