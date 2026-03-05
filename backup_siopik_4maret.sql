/*M!999999\- enable the sandbox mode */ 
-- MariaDB dump 10.19  Distrib 10.11.14-MariaDB, for debian-linux-gnu (x86_64)
--
-- Host: localhost    Database: ekinerja_db
-- ------------------------------------------------------
-- Server version	10.11.14-MariaDB-0ubuntu0.24.04.1

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `file_laporan`
--

DROP TABLE IF EXISTS `file_laporan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `file_laporan` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `laporan_id` bigint(20) unsigned DEFAULT NULL,
  `tipe_file` varchar(50) DEFAULT NULL,
  `file_path` varchar(255) DEFAULT NULL,
  `metadata_exif` text DEFAULT NULL,
  `uploaded_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_file_laporan_laporan` (`laporan_id`),
  CONSTRAINT `fk_file_laporan_laporan` FOREIGN KEY (`laporan_id`) REFERENCES `laporan` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `file_laporan`
--

LOCK TABLES `file_laporan` WRITE;
/*!40000 ALTER TABLE `file_laporan` DISABLE KEYS */;
/*!40000 ALTER TABLE `file_laporan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `laporan`
--

DROP TABLE IF EXISTS `laporan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `laporan` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) unsigned DEFAULT NULL,
  `tipe_laporan` tinyint(1) DEFAULT NULL,
  `tugas_organisasi_id` bigint(20) unsigned DEFAULT NULL,
  `judul_kegiatan` varchar(255) DEFAULT NULL,
  `deskripsi_hasil` text DEFAULT NULL,
  `waktu_pelaporan` datetime(3) DEFAULT NULL,
  `is_overtime` tinyint(1) DEFAULT 0,
  `lokasi_lat` varchar(50) DEFAULT NULL,
  `lokasi_long` varchar(50) DEFAULT NULL,
  `alamat_lokasi` text DEFAULT NULL,
  `foto_url` varchar(255) DEFAULT NULL,
  `dokumen_url` varchar(255) DEFAULT NULL,
  `status` varchar(50) DEFAULT 'Menunggu',
  `jam_kerja` bigint(20) DEFAULT 0,
  `komentar_atasan` text DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_laporan_user` (`user_id`),
  KEY `fk_laporan_tugas_organisasi` (`tugas_organisasi_id`),
  CONSTRAINT `fk_laporan_tugas_organisasi` FOREIGN KEY (`tugas_organisasi_id`) REFERENCES `tugas_organisasi` (`id`),
  CONSTRAINT `fk_laporan_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=12 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `laporan`
--

LOCK TABLES `laporan` WRITE;
/*!40000 ALTER TABLE `laporan` DISABLE KEYS */;
INSERT INTO `laporan` VALUES
(1,5,0,NULL,'Tes foto laporan','Tes foto laporan','2026-03-04 05:28:23.000',1,'-7.1162761','107.8976439','VVMX+G3V, Haruman, Kecamatan Leles, Kabupaten Garut, Jawa Barat','uploads/reports/images/77dcb1c5-9c66-4ba0-a6f5-07c1cefeb6ce.jpg',NULL,'Disetujui',0,'tes','2026-03-03 22:28:47.211'),
(2,4,0,NULL,'Koordinasi Tahapan Kegiatan Pendanaan Kelurahan','Pembahasan Persiapan Kegiatan Anggaran, dan realisasi kegiatan anggaran kas semester 1 TA. 2026','2026-03-04 09:47:50.000',0,'-7.2492146','107.9233793','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/1f1275c4-89a5-4625-a6cb-ede6b7b81a18.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 02:52:07.672'),
(3,6,0,NULL,'Briefing pagi','Ada beberapa hal yg diutarakan oleh Bu lurah, diantaranya piket kebersihan ','2026-03-04 09:52:35.000',0,'-7.2492457','107.9233774','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/cbe375cb-eea3-4712-97dd-5c4d6d56d2d4.jpg',NULL,'Disetujui',0,'-','2026-03-04 02:55:40.746'),
(4,2,0,NULL,'Kegiatan briping mingguan','-Membuat jadwal piket harian.\n-Meningkaykan pelayanan   \n  publik terhadap masyarakat \n-pengembangan Aplikasi siopik\n\n','2026-03-04 09:53:40.000',0,'-7.2492354','107.923387','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/305227f5-e270-437a-a557-02d75480356f.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 02:57:29.701'),
(5,3,0,NULL,'Rakor persiapan Dakel','Kegiatan menunggu terbitnya cpcl','2026-03-04 09:58:52.000',0,'-7.2492167','107.9233684','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/d1cbe9cc-e75e-45a6-a7c2-3864e534d056.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:00:18.198'),
(6,10,0,NULL,'Briping bulanan','Rapat koordinasi tentang pembahasan DTKS','2026-03-04 09:55:52.000',0,'-7.249223','107.9233686','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/079917b5-c9f8-4daf-bbb8-894ce931dd41.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:05:00.942'),
(7,9,0,NULL,'Breefing staf kelurahan sukanegala','Pembahasan aset kelurahan sukanegla bulan maret tahun 2026','2026-03-04 10:01:56.000',0,'-7.2492135','107.9233799','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/c51e459f-5785-4499-a6fd-8aa80836a8e7.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:07:19.395'),
(8,8,0,NULL,'Breefing Staf Kelurahan Sukanegla','Pembahasan tentang jenis pelayanan umum kepada masyarakat','2026-03-04 10:01:06.000',0,'-7.249215','107.9233754','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/7aad557c-eb3e-45cc-8de7-64d214617abb.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:10:02.223'),
(9,8,0,NULL,'Pelayanan umum','Melayani Pembuatan SKU atas nama noneng nurlaela','2026-03-04 10:24:03.000',0,'-7.2492131','107.9233743','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/ba526ac6-64dd-44ab-a9fb-8f6600e3dad3.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:31:07.240'),
(10,4,0,NULL,'Mengelola SIPD Penatausahaan','Buku Pembantu Bank Periode 01 Maret 2026 s/d 04 Maret 2026, dengan hasil masih Nol','2026-03-04 10:35:50.000',0,'-7.249288','107.9232986','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/4678a9a6-ecc0-49f5-bbd6-ed64fc34a134.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:39:06.731'),
(11,9,0,NULL,'Tugas harian pelayanan umum','Verifikasi dan registrasi data pertanahan. ','2026-03-04 10:52:55.000',0,'-7.2492278','107.9233706','Jl. Margawati No.911, Sukanegla, Kecamatan Garut Kota, Kabupaten Garut, Jawa Barat','uploads/reports/images/0db45abe-a351-403c-be1e-22d6864c0e51.jpg',NULL,'Menunggu',0,NULL,'2026-03-04 03:57:21.567');
/*!40000 ALTER TABLE `laporan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `notifications`
--

DROP TABLE IF EXISTS `notifications`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `notifications` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) NOT NULL,
  `kategori` varchar(50) NOT NULL,
  `judul` varchar(255) NOT NULL,
  `pesan` text NOT NULL,
  `is_read` tinyint(1) DEFAULT 0,
  `terkait_id` bigint(20) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `notifications`
--

LOCK TABLES `notifications` WRITE;
/*!40000 ALTER TABLE `notifications` DISABLE KEYS */;
/*!40000 ALTER TABLE `notifications` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `penilaian`
--

DROP TABLE IF EXISTS `penilaian`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `penilaian` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `user_id` bigint(20) unsigned DEFAULT NULL,
  `penilai_id` bigint(20) unsigned DEFAULT NULL,
  `skor_id` bigint(20) unsigned DEFAULT NULL,
  `jenis_periode` varchar(50) DEFAULT NULL,
  `bulan` bigint(20) DEFAULT NULL,
  `tahun` bigint(20) DEFAULT NULL,
  `tanggal_mulai` date DEFAULT NULL,
  `tanggal_selesai` date DEFAULT NULL,
  `catatan` text DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  `updated_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_user_bulan_tahun` (`user_id`,`bulan`,`tahun`),
  KEY `fk_penilaian_skor` (`skor_id`),
  KEY `fk_penilaian_penilai` (`penilai_id`),
  CONSTRAINT `fk_penilaian_penilai` FOREIGN KEY (`penilai_id`) REFERENCES `users` (`id`),
  CONSTRAINT `fk_penilaian_skor` FOREIGN KEY (`skor_id`) REFERENCES `ref_skor_penilaian` (`id`),
  CONSTRAINT `fk_penilaian_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `penilaian`
--

LOCK TABLES `penilaian` WRITE;
/*!40000 ALTER TABLE `penilaian` DISABLE KEYS */;
/*!40000 ALTER TABLE `penilaian` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ref_jabatan`
--

DROP TABLE IF EXISTS `ref_jabatan`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `ref_jabatan` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `nama_jabatan` varchar(255) NOT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_ref_jabatan_nama_jabatan` (`nama_jabatan`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ref_jabatan`
--

LOCK TABLES `ref_jabatan` WRITE;
/*!40000 ALTER TABLE `ref_jabatan` DISABLE KEYS */;
INSERT INTO `ref_jabatan` VALUES
(1,'Lurah','2026-03-03 15:14:28.549'),
(2,'Sekertaris','2026-03-03 15:14:28.562'),
(3,'Kasi Pemerintahan','2026-03-03 15:14:28.576'),
(4,'Kasi Kesejahteraan Masyarakat','2026-03-03 15:14:28.582'),
(5,'Kasi Ekonomi dan Pembangunan','2026-03-03 15:14:28.585'),
(6,'Pengadministrasi Perkantoran','2026-03-03 15:14:28.590'),
(7,'Pengelola Aset','2026-03-03 15:14:28.595'),
(8,'Operator Layanan Operasional','2026-03-03 15:14:28.598'),
(9,'Operator DTKS/DTSEN','2026-03-03 15:14:28.601'),
(10,'Penata Kelola Sistem dan Teknologi Informasi','2026-03-03 15:14:28.606');
/*!40000 ALTER TABLE `ref_jabatan` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `ref_skor_penilaian`
--

DROP TABLE IF EXISTS `ref_skor_penilaian`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `ref_skor_penilaian` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `keterangan` varchar(255) DEFAULT NULL,
  `bobot_nilai` bigint(20) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=4 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `ref_skor_penilaian`
--

LOCK TABLES `ref_skor_penilaian` WRITE;
/*!40000 ALTER TABLE `ref_skor_penilaian` DISABLE KEYS */;
INSERT INTO `ref_skor_penilaian` VALUES
(1,'Dibawah Ekspektasi',1),
(2,'Sesuai Ekspektasi',2),
(3,'Diatas Ekspektasi',3);
/*!40000 ALTER TABLE `ref_skor_penilaian` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tugas_assignees`
--

DROP TABLE IF EXISTS `tugas_assignees`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `tugas_assignees` (
  `tugas_organisasi_id` bigint(20) unsigned NOT NULL,
  `user_id` bigint(20) unsigned NOT NULL,
  PRIMARY KEY (`tugas_organisasi_id`,`user_id`),
  KEY `fk_tugas_assignees_user` (`user_id`),
  CONSTRAINT `fk_tugas_assignees_tugas_organisasi` FOREIGN KEY (`tugas_organisasi_id`) REFERENCES `tugas_organisasi` (`id`),
  CONSTRAINT `fk_tugas_assignees_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tugas_assignees`
--

LOCK TABLES `tugas_assignees` WRITE;
/*!40000 ALTER TABLE `tugas_assignees` DISABLE KEYS */;
/*!40000 ALTER TABLE `tugas_assignees` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `tugas_organisasi`
--

DROP TABLE IF EXISTS `tugas_organisasi`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `tugas_organisasi` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `judul_tugas` varchar(255) DEFAULT NULL,
  `deskripsi` text DEFAULT NULL,
  `file_bukti` varchar(255) DEFAULT NULL,
  `created_by` bigint(20) unsigned DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  KEY `fk_tugas_organisasi_creator` (`created_by`),
  CONSTRAINT `fk_tugas_organisasi_creator` FOREIGN KEY (`created_by`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `tugas_organisasi`
--

LOCK TABLES `tugas_organisasi` WRITE;
/*!40000 ALTER TABLE `tugas_organisasi` DISABLE KEYS */;
/*!40000 ALTER TABLE `tugas_organisasi` ENABLE KEYS */;
UNLOCK TABLES;

--
-- Table structure for table `users`
--

DROP TABLE IF EXISTS `users`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!40101 SET character_set_client = utf8mb4 */;
CREATE TABLE `users` (
  `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
  `nip` varchar(20) NOT NULL,
  `nama` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `role` varchar(50) NOT NULL,
  `jabatan_id` bigint(20) unsigned DEFAULT NULL,
  `supervisor_id` bigint(20) unsigned DEFAULT NULL,
  `foto_path` varchar(255) DEFAULT NULL,
  `fcm_token` varchar(255) DEFAULT NULL,
  `created_at` datetime(3) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uni_users_nip` (`nip`),
  KEY `fk_users_jabatan` (`jabatan_id`),
  KEY `fk_users_supervisor` (`supervisor_id`),
  CONSTRAINT `fk_users_jabatan` FOREIGN KEY (`jabatan_id`) REFERENCES `ref_jabatan` (`id`),
  CONSTRAINT `fk_users_supervisor` FOREIGN KEY (`supervisor_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=11 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Dumping data for table `users`
--

LOCK TABLES `users` WRITE;
/*!40000 ALTER TABLE `users` DISABLE KEYS */;
INSERT INTO `users` VALUES
(1,'198106152014102004','Iis Yuniawardani, S.IP','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','lurah',1,NULL,NULL,NULL,'2026-03-03 15:14:28.743'),
(2,'198002012009061001','Aep Saepudin, S.Kom','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','sekertaris',2,1,NULL,NULL,'2026-03-03 15:14:28.750'),
(3,'197905172014101003','Cahyo Dirgantoro Priyawan, A.Md','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','kasi',5,1,'uploads/photos/ac6ee7df-75ed-4b23-a2f9-7e9bfcec6494.jpg',NULL,'2026-03-03 15:14:28.766'),
(4,'198102252014111001','Budi Budiansyah','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','staf',6,2,NULL,NULL,'2026-03-03 15:14:28.773'),
(5,'200112282025041006','Muhammad Abyan Kamal, S.Kom','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','staf',10,2,'uploads/photos/fc26666e-cb55-4347-aaf0-7367e1a79221.jpg',NULL,'2026-03-03 15:14:28.780'),
(6,'198001022008011003','Kustaman, S.E','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','kasi',3,1,'uploads/photos/ba816404-297b-47bb-bf55-adccc54c104d.jpg',NULL,'2026-03-03 15:14:28.791'),
(7,'196904051994031011','Agus Haris','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','kasi',4,1,NULL,NULL,'2026-03-03 15:14:28.798'),
(8,'198908152025212085','Dewi Srimulyati','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','staf',8,2,NULL,NULL,'2026-03-03 15:14:28.803'),
(9,'198410022025212046','Erlin Wili Aspiantiny','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','staf',7,2,NULL,NULL,'2026-03-03 15:14:28.808'),
(10,'198205202025211085','Tantan Kustandi','$2a$10$ALZ4cDi5.cdGxVu3T3GB8.NNhKymy9b9tnU4AwXVgYZXfQvovZkCe','staf',9,2,NULL,NULL,'2026-03-03 15:14:28.813');
/*!40000 ALTER TABLE `users` ENABLE KEYS */;
UNLOCK TABLES;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2026-03-04  4:07:15
