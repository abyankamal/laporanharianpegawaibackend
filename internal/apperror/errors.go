package apperror

import "errors"

var (
	// Domain: Report
	ErrReportNotFound = errors.New("laporan tidak ditemukan")
	
	// Domain: Auth & Permission
	ErrForbidden    = errors.New("akses ditolak")
	ErrUnauthorized = errors.New("unauthorized")
	
	// Domain: Common
	ErrBadRequest   = errors.New("request tidak valid")
	ErrInternal     = errors.New("terjadi kesalahan internal server")
)
