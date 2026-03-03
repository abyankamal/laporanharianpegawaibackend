package utils

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"os"

	"github.com/nfnt/resize"
)

// CompressImage menerima input file multipart, mengompresinya jika melebihi target size, lalu menyimpan ke destinasi.
// Hanya mendukung JPG/JPEG dan PNG.
func CompressImage(fileHeader *multipart.FileHeader, destPath string, maxSizeMB int) error {
	// Buka file sumber
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("gagal membuka file sumber: %w", err)
	}
	defer src.Close()

	// Konversi max size ke bytes
	maxBytes := int64(maxSizeMB * 1024 * 1024)

	// Jika ukuran file SUDAH DI BAWAH batas maksimal, kita TIDAK PERLU kompres.
	// Langsung copy saja (lebih cepat & tidak merusak kualitas).
	if fileHeader.Size <= maxBytes {
		return copyFileOriginal(src, destPath)
	}

	// --- Jika ukuran > maxSizeMB, lakukan kompresi ---

	// Decode gambar (hanya support standard jpeg & png di sini, format ain butuh registrasi)
	img, format, err := image.Decode(src)
	if err != nil {
		// Jika gagal decode (mungkin format tidak disupport atau korup), fallback simpan original
		if seeker, ok := src.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		return copyFileOriginal(src, destPath)
	}

	// Resize gambar.
	// Kita tetapkan max width = 1920 (Full HD). Tingginya menyesuaikan aspek rasio secara otomatis (0).
	// Bilinear adalah algoritma resizing yang lebih cepat dan menghemat banyak memory dibanding Lanczos3
	resizedImg := resize.Resize(1920, 0, img, resize.Bilinear)

	// Buka file tujuan
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("gagal membuat file tujuan: %w", err)
	}
	defer dst.Close()

	// Encode dan simpan. Selalu simpan sebagai JPEG untuk mendapatkan kompresi lossy terbaik,
	// meskipun sumber aslinya PNG (karena PNG di resize sulit turun sizenya secara signifikan).
	// Namun kita harus sesuaikan ekstensi di nama file agar konsisten di level caller.
	// Opsi amannya: Kita hormati format asli. Jika PNG encode PNG, jika JPG encode JPG.
	// Jika PNG terlalu besar, encode ke PNG lambat. Kita ubah saja semua jadi compress JPEG 75%.

	options := &jpeg.Options{Quality: 75}
	if format == "png" {
		// KARENA objektif kita adalah kompresi radikal (turun dari misal 10MB ke >5MB),
		// memaksa encode ke JPEG walau ekstensi filenya ".png" kadang bikin error di beberapa OS reader tua.
		// Namun untuk web browser/Flutter biasanya tetap bisa render. Untuk amannya,
		// kita simpan sebagai PNG jika aslinya PNG (harapannya proses resize dimension
		// dari 4000x4000 menjadi 1920xH sudah cukup memangkas byte).
		var enc png.Encoder
		enc.CompressionLevel = png.BestCompression
		err = enc.Encode(dst, resizedImg)
	} else {
		// Asumsi aslinya jpeg/jpg
		err = jpeg.Encode(dst, resizedImg, options)
	}

	if err != nil {
		return fmt.Errorf("gagal mengenkode gambar hasil kompresi: %w", err)
	}

	return nil
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
