package utils

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

// CompressImage menerima input file multipart, mengompresinya jika melebihi target size, lalu menyimpan ke destinasi.
// NOTE: Kompresi gambar sementara dimatikan dan menggunakan logika copy file langsung
// untuk menghindari panic memory (disconnect) saat memproses file dari kamera HP resolusi tinggi.
func CompressImage(fileHeader *multipart.FileHeader, destPath string, maxSizeMB int) error {
	// Buka file sumber
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("gagal membuka file sumber: %w", err)
	}
	defer src.Close()

	// Langsung gunakan logika copy original agar stabil seperti upload dokumen
	return copyFileOriginal(src, destPath)
}

// copyFileOriginal mengkopi file langsung tanpa decode-encode gambar
func copyFileOriginal(src io.Reader, destPath string) error {
	// Pastikan reader mulai dari awal
	if seeker, ok := src.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file original: %w", err)
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("gagal mengkopi file original: %w", err)
	}

	return nil
}
