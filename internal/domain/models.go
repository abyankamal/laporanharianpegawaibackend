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
	FotoPath     *string   `gorm:"column:foto_path;type:varchar(255)" json:"foto_path"`
	FCMToken     *string   `gorm:"column:fcm_token;type:varchar(255)" json:"fcm_token"`
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

// TugasOrganisasi adalah tabel untuk menyimpan tugas organisasi yang ditetapkan oleh Lurah.
type TugasOrganisasi struct {
	ID         uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	JudulTugas string     `gorm:"column:judul_tugas;type:varchar(255)" json:"judul_tugas"`
	Deskripsi  string     `gorm:"column:deskripsi;type:text" json:"deskripsi"`
	FileBukti  *string    `gorm:"column:file_bukti;type:varchar(255)" json:"file_bukti"` // URL dokumen pendukung
	Deadline   *time.Time `gorm:"column:deadline;type:timestamp" json:"deadline"`
	CreatedBy  *uint      `gorm:"column:created_by" json:"created_by"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`

	// Relasi
	Creator   *User  `gorm:"foreignKey:CreatedBy" json:"creator,omitempty"`
	Assignees []User `gorm:"many2many:tugas_assignees;foreignKey:ID;joinForeignKey:TugasOrganisasiID;References:ID;joinReferences:UserID" json:"assignees,omitempty"`
}

func (TugasOrganisasi) TableName() string {
	return "tugas_organisasi"
}

// Laporan adalah tabel untuk menyimpan laporan kinerja harian.
type Laporan struct {
	ID                uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            *uint     `gorm:"column:user_id" json:"user_id"`
	TipeLaporan       bool      `gorm:"column:tipe_laporan;type:boolean" json:"tipe_laporan"`  // true = Pokok (linked or manual), false = Tambahan
	TugasOrganisasiID *uint     `gorm:"column:tugas_organisasi_id" json:"tugas_organisasi_id"` // nullable, for linking to organizational tasks
	JudulKegiatan     string    `gorm:"column:judul_kegiatan;type:varchar(255)" json:"judul_kegiatan"`
	DeskripsiHasil    string    `gorm:"column:deskripsi_hasil;type:text" json:"deskripsi_hasil"`
	WaktuPelaporan    time.Time `gorm:"column:waktu_pelaporan" json:"waktu_pelaporan"`
	IsOvertime        bool      `gorm:"column:is_overtime;default:false" json:"is_overtime"`
	LokasiLat         *string   `gorm:"column:lokasi_lat;type:varchar(50)" json:"lokasi_lat"`
	LokasiLong        *string   `gorm:"column:lokasi_long;type:varchar(50)" json:"lokasi_long"`
	AlamatLokasi      *string   `gorm:"column:alamat_lokasi;type:text" json:"alamat_lokasi"`
	FotoURL           *string   `gorm:"column:foto_url;type:varchar(255)" json:"foto_url"`       // URL file foto lampiran (opsional)
	DokumenURL        *string   `gorm:"column:dokumen_url;type:varchar(255)" json:"dokumen_url"` // URL file dokumen lampiran (opsional)
	Status            string    `gorm:"column:status;type:varchar(50);default:'menunggu_review'" json:"status"`
	JamKerja          int       `gorm:"column:jam_kerja;default:0" json:"jam_kerja"`
	KomentarAtasan    *string   `gorm:"column:komentar_atasan;type:text" json:"komentar_atasan"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`

	// Relasi
	User            *User            `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TugasOrganisasi *TugasOrganisasi `gorm:"foreignKey:TugasOrganisasiID" json:"tugas_organisasi,omitempty"`
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
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         *uint     `gorm:"column:user_id;uniqueIndex:idx_user_bulan_tahun" json:"user_id"`
	PenilaiID      *uint     `gorm:"column:penilai_id" json:"penilai_id"`
	SkorID         *uint     `gorm:"column:skor_id" json:"skor_id"`
	JenisPeriode   string    `gorm:"column:jenis_periode;type:varchar(50)" json:"jenis_periode"` // 'Harian', 'Mingguan', 'Bulanan', 'Custom'
	Bulan          int       `gorm:"column:bulan;type:int;uniqueIndex:idx_user_bulan_tahun" json:"bulan"`
	Tahun          int       `gorm:"column:tahun;type:int;uniqueIndex:idx_user_bulan_tahun" json:"tahun"`
	TanggalMulai   time.Time `gorm:"column:tanggal_mulai;type:date" json:"tanggal_mulai"`
	TanggalSelesai time.Time `gorm:"column:tanggal_selesai;type:date" json:"tanggal_selesai"`
	Catatan        string    `gorm:"column:catatan;type:text" json:"catatan"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Relasi
	User    *User             `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Penilai *User             `gorm:"foreignKey:PenilaiID" json:"penilai,omitempty"`
	Skor    *RefSkorPenilaian `gorm:"foreignKey:SkorID" json:"skor,omitempty"`
}

func (Penilaian) TableName() string {
	return "penilaian"
}

// Notification adalah tabel untuk menyimpan notifikasi pengguna.
type Notification struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int       `gorm:"column:user_id;not null" json:"user_id"`
	Kategori  string    `gorm:"column:kategori;type:varchar(50);not null" json:"kategori"` // 'Tugas', 'Penilaian', 'Sistem'
	Judul     string    `gorm:"column:judul;type:varchar(255);not null" json:"judul"`
	Pesan     string    `gorm:"column:pesan;type:text;not null" json:"pesan"`
	IsRead    bool      `gorm:"column:is_read;default:false" json:"is_read"`
	TerkaitID int       `gorm:"column:terkait_id" json:"terkait_id"` // ID referensi (Tugas atau Penilaian)
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}

// Holiday adalah tabel untuk menyimpan data hari libur.
type Holiday struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TanggalMulai   time.Time `gorm:"column:tanggal_mulai;type:date;not null" json:"tanggal_mulai"`
	TanggalSelesai time.Time `gorm:"column:tanggal_selesai;type:date;not null" json:"tanggal_selesai"`
	Keterangan     string    `gorm:"column:keterangan;type:varchar(255)" json:"keterangan"`
}

func (Holiday) TableName() string {
	return "holiday"
}

// WorkHour adalah tabel untuk menyimpan pengaturan sistem seperti jam kerja (hanya 1 record/baris).
type WorkHour struct {
	ID             uint   `gorm:"primaryKey;autoIncrement" json:"id"`
	JamMasuk       string `gorm:"column:jam_masuk;type:varchar(5)" json:"jam_masuk"`   // Senin-Kamis
	JamPulang      string `gorm:"column:jam_pulang;type:varchar(5)" json:"jam_pulang"` // Senin-Kamis
	JamMasukJumat  string `gorm:"column:jam_masuk_jumat;type:varchar(5)" json:"jam_masuk_jumat"`
	JamPulangJumat string `gorm:"column:jam_pulang_jumat;type:varchar(5)" json:"jam_pulang_jumat"`
}

func (WorkHour) TableName() string {
	return "work_hour"
}
