package apperror

import "errors"

var (
	// Domain: Report
	ErrReportNotFound           = errors.New("laporan tidak ditemukan")
	ErrReportAlreadyReviewed     = errors.New("laporan ini sudah disetujui dan tidak dapat dievaluasi ulang")
	ErrReportAlreadyApproved     = errors.New("laporan yang sudah disetujui tidak dapat diubah")
	ErrInvalidEvaluationStatus   = errors.New("status evaluasi tidak valid (harus 'disetujui' atau 'ditolak')")
	ErrReasonRequired            = errors.New("alasan (komentar) wajib diisi jika laporan ditolak")
	ErrSecretaryStaffOnly        = errors.New("Sekertaris hanya memiliki hak untuk mengevaluasi laporan Staf")
	ErrOnlyLurahCanDeleteReport  = errors.New("akses ditolak: hanya role Lurah yang diperbolehkan menghapus laporan")
	ErrOnlyOwnReportAllowed      = errors.New("akses ditolak: hanya dapat melihat laporan milik sendiri")
	ErrOnlyStaffOrOwnAllowed     = errors.New("akses ditolak: hanya dapat melihat laporan staf atau milik sendiri")
	ErrOnlyOwnReportModifiable   = errors.New("akses ditolak: hanya dapat mengubah laporan milik sendiri")

	// Domain: User
	ErrUserNotFound             = errors.New("user tidak ditemukan")
	ErrNIPAlreadyExists         = errors.New("NIP sudah terdaftar")
	ErrOldPasswordMismatch      = errors.New("password lama tidak sesuai")
	ErrSamePassword             = errors.New("password baru tidak boleh sama dengan password lama")

	// Domain: Auth & Permission
	ErrForbidden                = errors.New("akses ditolak")
	ErrUnauthorized             = errors.New("unauthorized")
	ErrInvalidToken             = errors.New("token tidak valid atau sudah kadaluarsa")

	// Domain: Common
	ErrBadRequest               = errors.New("request tidak valid")
	ErrInternal                 = errors.New("terjadi kesalahan internal server")
)
