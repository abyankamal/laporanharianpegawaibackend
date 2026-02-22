# API Documentation - Laporan Harian Pegawai

Dokumentasi lengkap API Laporan Harian Pegawai.
**Base URL**: `http://localhost:5000/api`

## Daftar Isi
1. [Authentication](#1-authentication)
2. [User Profile](#2-user-profile)
3. [Dashboard](#3-dashboard)
4. [Laporan (Reports)](#4-laporan-reports)
5. [Penilaian (Reviews)](#5-penilaian-reviews)
6. [Tugas Pokok (Tasks)](#6-tugas-pokok-tasks)
7. [User Management](#7-user-management-sekertaris-only)

---

## 1. Authentication

### Login
Login untuk mendapatkan JWT Token yang digunakan untuk akses endpoint lainnya.

- **Endpoint**: `/login`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "nip": "198501012010011001",
    "password": "password123"
  }
  ```
- **Response**:
  ```json
  {
    "status": "success",
    "message": "Login berhasil",
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
  ```

---

## 2. User Profile

### Get Profile
Mendapatkan informasi user yang sedang login.

- **Endpoint**: `/profile`
- **Method**: `GET`
- **Header**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "status": "success",
    "message": "Data profil berhasil diambil",
    "data": {
      "id": 1,
      "nip": "198501012010011001",
      "nama": "Budi Santoso",
      "role": "staf",
      "jabatan_id": 5,
      "nama_jabatan": "Staf Pelaksana Teknis",
      "supervisor_id": 2,
      "nama_atasan": "Siti Aminah",
      "created_at": "2024-01-01 10:00:00"
    }
  }
  ```

### Change Password
Mengubah password user yang sedang login.

- **Endpoint**: `/profile/change-password`
- **Method**: `PUT`
- **Header**: `Authorization: Bearer <token>`
- **Body**:
  ```json
  {
    "old_password": "passwordLama",
    "new_password": "passwordBaru123"
  }
  ```
- **Response**:
  ```json
  {
    "status": "success",
    "message": "Password berhasil diubah"
  }
  ```

---

## 3. Dashboard

### Get Summary
Statistik untuk dashboard.

- **Endpoint**: `/dashboard/summary`
- **Method**: `GET`
- **Header**: `Authorization: Bearer <token>`
- **Response**:
  ```json
  {
    "status": "success",
    "message": "Data dashboard berhasil diambil",
    "data": {
      "task_pending": 0,          // Tugas pokok belum dilaporkan hari ini
      "laporan_bulan_ini": 25,    // Total laporan bulan ini
      "laporan_masuk_hari_ini": 10 // (Hanya Role Atasan)
    }
  }
  ```

---

## 4. Laporan (Reports)

### Create Report
Membuat laporan kinerja baru (Harian).

- **Endpoint**: `/reports`
- **Method**: `POST`
- **Header**: `Authorization: Bearer <token>`
- **Content-Type**: `multipart/form-data`
- **Body (Form Data)**:
  - `tipe_laporan`: `pokok` atau `tambahan`
  - `judul_kegiatan`: String
  - `deskripsi_hasil`: String
  - `waktu_mulai`: `YYYY-MM-DD HH:mm:ss`
  - `waktu_selesai`: `YYYY-MM-DD HH:mm:ss`
  - `lokasi_lat`: String (Latitude)
  - `lokasi_long`: String (Longitude)
  - `alamat_lokasi`: String
  - `file_bukti`: File (Image/PDF) - Optional

- **Response**:
  ```json
  {
    "status": "success",
    "message": "Laporan berhasil dibuat",
    "data": {
      "id": 101,
      "is_overtime": false,
      "created_at": "2026-02-18 10:00:00"
    }
  }
  ```

### Get All Reports
Melihat daftar laporan (User biasa lihat punya sendiri, Atasan lihat bawahan sesuai hierarki).

- **Endpoint**: `/reports`
- **Method**: `GET`
- **Query Params**:
  - `page`: int (Default 1)
  - `limit`: int (Default 10)
  - `start_date`: `YYYY-MM-DD`
  - `end_date`: `YYYY-MM-DD`
  - `user_id`: int (Filter by user ID)
- **Response**:
  ```json
  {
    "status": "success",
    "data": [ ... ],
    "meta": {
      "total": 50,
      "page": 1,
      "limit": 10,
      "total_pages": 5
    }
  }
  ```

---

## 5. Penilaian (Reviews)

### Create Review (Atasan Only)
Atasan (Lurah/Sekertaris) memberikan penilaian ke bawahan.

- **Endpoint**: `/reviews`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "target_user_id": 5,
    "skor_id": 3,
    "jenis_periode": "Bulanan",
    "tanggal_mulai": "2026-02-01",
    "tanggal_selesai": "2026-02-28",
    "catatan": "Kinerja cukup baik"
  }
  ```

### Get My Reviews
User melihat penilaian yang diberikan kepadanya.

- **Endpoint**: `/reviews`
- **Method**: `GET`

### Get My Submitted Reviews (Atasan Only)
Atasan melihat history penilaian yang pernah dia buat.

- **Endpoint**: `/reviews/my-submissions`
- **Method**: `GET`

---

## 6. Tugas Pokok (Tasks)

### Create Task (Atasan Only)
Atasan membuat master tugas pokok untuk bawahan.

- **Endpoint**: `/tasks`
- **Method**: `POST`
- **Body**:
  ```json
  {
    "target_user_id": 5,
    "judul_tugas": "Mengelola Arsip Surat",
    "deskripsi": "Menginput surat masuk dan keluar"
  }
  ```

### Get My Tasks
User melihat daftar tugas pokok miliknya (biasanya untuk dropdown saat buat laporan).

- **Endpoint**: `/my-tasks`
- **Method**: `GET`
- **Response**:
  ```json
  {
    "status": "success",
    "data": [
      {
        "id": 1,
        "judul_tugas": "Mengelola Arsip Surat"
      }
    ]
  }
  ```

---

## 7. User Management (Sekertaris Only)

### CRUD Users
Hanya role `sekertaris` yang bisa mengakses.

- **List Users**: `GET /users`
- **Detail User**: `GET /users/:id`
- **Create User**: `POST /users`
- **Update User**: `PUT /users/:id`
- **Delete User**: `DELETE /users/:id`

**Body Create/Update User**:
```json
{
  "nip": "12345",
  "nama": "User Baru",
  "password": "password123", /* Required only on create */
  "role": "staf",
  "jabatan_id": 3,
  "supervisor_id": 2
}
```
