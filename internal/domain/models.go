package domain

import "time"

// RefJabatan adalah tabel master untuk data jabatan pegawai.
type RefJabatan struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NamaJabatan string    `gorm:"column:nama_jabatan;type:varchar(255);not null;unique" json:"nama_jabatan"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (RefJabatan) TableName() string {
	return "ref_jabatan"
}

// User adalah tabel untuk data pengguna sistem.
type User struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	NIP          string    `gorm:"column:nip;type:varchar(20);unique;not null" json:"nip"`
	Nama         string    `gorm:"column:nama;type:varchar(255);not null" json:"nama"`
	Password     string    `gorm:"column:password;type:varchar(255);not null" json:"-"`
	Role         string    `gorm:"column:role;type:varchar(50);not null" json:"role"` // 'lurah', 'sekertaris', 'kasi', 'staf'
	JabatanID    *uint     `gorm:"column:jabatan_id" json:"jabatan_id"`
	SupervisorID *uint     `gorm:"column:supervisor_id" json:"supervisor_id"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`

	// Relasi
	Jabatan    *RefJabatan `gorm:"foreignKey:JabatanID" json:"jabatan,omitempty"`
	Supervisor *User       `gorm:"foreignKey:SupervisorID" json:"supervisor,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// RefSkorPenilaian adalah tabel referensi untuk skor penilaian kinerja.
type RefSkorPenilaian struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Keterangan string `gorm:"column:keterangan;type:varchar(255)" json:"keterangan"`
	BobotNilai int    `gorm:"column:bobot_nilai" json:"bobot_nilai"`
}

func (RefSkorPenilaian) TableName() string {
	return "ref_skor_penilaian"
}

// TugasPokok adalah tabel untuk menyimpan tugas pokok pegawai.
// dikomen dulu, nanti diaktifkan lagi kalau sudah ada fitur tugas pokok
// type TugasPokok struct {
// 	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
// 	UserID     *uint     `gorm:"column:user_id" json:"user_id"`
// 	JudulTugas string    `gorm:"column:judul_tugas;type:varchar(255)" json:"judul_tugas"`
// 	Deskripsi  string    `gorm:"column:deskripsi;type:text" json:"deskripsi"`
// 	CreatedBy  *uint     `gorm:"column:created_by" json:"created_by"`
// 	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`

// 	// Relasi
// 	User    *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
// 	Creator *User `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
// }

// func (TugasPokok) TableName() string {
// 	return "tugas_pokok"
// }

// Laporan adalah tabel untuk menyimpan laporan kinerja harian.
type Laporan struct {
	ID          uint  `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      *uint `gorm:"column:user_id" json:"user_id"`
	TipeLaporan bool  `gorm:"column:tipe_laporan;type:boolean" json:"tipe_laporan"` // true = pokok, false = tambahan
	// TugasPokokID   *uint     `gorm:"column:tugas_pokok_id" json:"tugas_pokok_id"`              // nullable
	JudulKegiatan  string    `gorm:"column:judul_kegiatan;type:varchar(255)" json:"judul_kegiatan"`
	DeskripsiHasil string    `gorm:"column:deskripsi_hasil;type:text" json:"deskripsi_hasil"`
	WaktuMulai     time.Time `gorm:"column:waktu_mulai" json:"waktu_mulai"`
	WaktuSelesai   time.Time `gorm:"column:waktu_selesai" json:"waktu_selesai"`
	IsOvertime     bool      `gorm:"column:is_overtime;default:false" json:"is_overtime"`
	LokasiLat      string    `gorm:"column:lokasi_lat;type:varchar(50)" json:"lokasi_lat"`
	LokasiLong     string    `gorm:"column:lokasi_long;type:varchar(50)" json:"lokasi_long"`
	AlamatLokasi   string    `gorm:"column:alamat_lokasi;type:text" json:"alamat_lokasi"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`

	// Relasi
	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
	// TugasPokok *TugasPokok `gorm:"foreignKey:TugasPokokID" json:"tugas_pokok,omitempty"`
}

func (Laporan) TableName() string {
	return "laporan"
}

// FileLaporan adalah tabel untuk menyimpan file bukti laporan.
type FileLaporan struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	LaporanID    *uint     `gorm:"column:laporan_id" json:"laporan_id"`
	TipeFile     string    `gorm:"column:tipe_file;type:varchar(50)" json:"tipe_file"`
	FilePath     string    `gorm:"column:file_path;type:varchar(255)" json:"file_path"`
	MetadataExif string    `gorm:"column:metadata_exif;type:text" json:"metadata_exif"`
	UploadedAt   time.Time `gorm:"column:uploaded_at" json:"uploaded_at"`

	// Relasi
	Laporan *Laporan `gorm:"foreignKey:LaporanID" json:"laporan,omitempty"`
}

func (FileLaporan) TableName() string {
	return "file_laporan"
}

// Penilaian adalah tabel untuk menyimpan hasil penilaian kinerja pegawai.
type Penilaian struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          *uint     `gorm:"column:user_id" json:"user_id"`
	PenilaiID       *uint     `gorm:"column:penilai_id" json:"penilai_id"`
	Periode         string    `gorm:"column:periode;type:varchar(50)" json:"periode"`
	TanggalPeriode  time.Time `gorm:"column:tanggal_periode;type:date" json:"tanggal_periode"`
	SkorID          *uint     `gorm:"column:skor_id" json:"skor_id"`
	CatatanEvaluasi string    `gorm:"column:catatan_evaluasi;type:text" json:"catatan_evaluasi"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`

	// Relasi
	User    *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Penilai *User             `gorm:"foreignKey:PenilaiID" json:"penilai,omitempty"`
	Skor    *RefSkorPenilaian `gorm:"foreignKey:SkorID" json:"skor,omitempty"`
}

func (Penilaian) TableName() string {
	return "penilaian"
}

// HariLibur adalah tabel untuk menyimpan data hari libur.
// dikomen dulu, nanti diaktifkan lagi kalau sudah ada fitur hari libur
// type HariLibur struct {
// 	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
// 	Tanggal    time.Time `gorm:"column:tanggal;type:date;unique" json:"tanggal"`
// 	Keterangan string    `gorm:"column:keterangan;type:varchar(255)" json:"keterangan"`
// }

// func (HariLibur) TableName() string {
// 	return "hari_libur"
// }
