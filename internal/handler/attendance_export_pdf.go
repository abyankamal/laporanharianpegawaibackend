package handler

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v3"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/service"
)

// ExportPDF menghasilkan PDF rekap absensi bulanan dalam format grid kehadiran (landscape F4).
// GET /api/mobile/absensi/export/pdf?bulan=1&tahun=2026
func (h *AbsensiHandler) ExportPDF(c fiber.Ctx) error {
	requesterRole, _ := c.Locals("role").(string)
	roleLower := strings.ToLower(requesterRole)

	// Hanya lurah, sekretaris, admin yang boleh export
	if roleLower != "lurah" && roleLower != "sekertaris" && roleLower != "sekretaris" && roleLower != "admin" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"status": "error", "message": "Akses ditolak",
		})
	}

	bulan, _ := strconv.Atoi(c.Query("bulan"))
	tahun, _ := strconv.Atoi(c.Query("tahun"))

	if bulan < 1 || bulan > 12 || tahun < 2020 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status": "error", "message": "Parameter bulan (1-12) dan tahun wajib diisi",
		})
	}

	// Ambil semua user (exclude admin)
	allUsers, err := h.userService.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil data pegawai",
		})
	}

	var users []domain.User
	for _, u := range allUsers {
		if strings.ToLower(u.Role) != "admin" {
			users = append(users, u)
		}
	}

	// Ambil rekap absensi
	recaps, err := h.absensiService.GetAllMonthlyRecap(bulan, tahun, users)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal mengambil data absensi",
		})
	}

	// Cari Lurah dan Sekretaris untuk tanda tangan
	var lurah, sekretaris *domain.User
	for i, u := range users {
		switch strings.ToLower(u.Role) {
		case "lurah":
			lurah = &users[i]
		case "sekertaris", "sekretaris":
			sekretaris = &users[i]
		}
	}

	// Generate PDF
	pdfBytes, err := generateAbsensiPDF(recaps, users, bulan, tahun, lurah, sekretaris)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status": "error", "message": "Gagal membuat PDF: " + err.Error(),
		})
	}

	filename := fmt.Sprintf("daftar_hadir_%s_%d.pdf", getIndonesianMonthName(time.Month(bulan)), tahun)
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/pdf")

	return c.Send(pdfBytes)
}

// generateAbsensiPDF membuat PDF grid kehadiran format landscape F4.
func generateAbsensiPDF(
	recaps []service.UserAbsensiRecap,
	users []domain.User,
	bulan, tahun int,
	lurah, sekretaris *domain.User,
) ([]byte, error) {
	// F4 Landscape: 330.2mm x 215.9mm
	pdf := fpdf.New("L", "mm", "A4", "")
	f4Size := fpdf.SizeType{Wd: 330.2, Ht: 215.9}

	marginL, marginR, marginT := 10.0, 10.0, 10.0
	pdf.SetMargins(marginL, marginT, marginR)
	pageW := f4Size.Wd - marginL - marginR

	// Jumlah hari dalam bulan
	firstDay := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.Local)
	lastDay := firstDay.AddDate(0, 1, -1)
	daysInMonth := lastDay.Day()

	// Lebar kolom
	noW := 8.0
	namaW := 60.0
	sisaW := pageW - noW - namaW
	dayW := sisaW / float64(daysInMonth)
	if dayW > 9.0 {
		dayW = 9.0
	}

	// Init cached logos
	cachedLogosOnce.Do(loadLogosToCache)
	for name, buf := range cachedLogos {
		opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
		pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(buf.Bytes()))
	}

	// Buat map hari libur dalam bulan ini
	holidayDays := make(map[int]bool)
	for d := 1; d <= daysInMonth; d++ {
		date := time.Date(tahun, time.Month(bulan), d, 0, 0, 0, 0, time.Local)
		weekday := date.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			holidayDays[d] = true
		}
	}

	// Buat map absensi per user per tanggal
	absensiMap := make(map[uint]map[int]string) // userID -> day -> status
	for _, recap := range recaps {
		userMap := make(map[int]string)
		if recap.Details != nil {
			for _, a := range recap.Details {
				userMap[a.Tanggal.Day()] = a.Status
			}
		}
		absensiMap[recap.User.ID] = userMap
	}

	// Kelompokkan user berdasarkan kategori
	var pnsUsers, p3kUsers []domain.User
	for _, u := range users {
		kategori := ""
		if u.KategoriPegawai != nil {
			kategori = strings.ToLower(u.KategoriPegawai.KodeKategori)
		}
		if kategori == "p3k_penuh" || kategori == "p3k_paruh" {
			p3kUsers = append(p3kUsers, u)
		} else {
			pnsUsers = append(pnsUsers, u)
		}
	}

	// ===================== PAGE =====================
	pdf.AddPageFormat("L", f4Size)

	// Header: Judul
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(pageW, 7, "DAFTAR HADIR PEGAWAI KELURAHAN SUKANEGLA", "", 1, "C", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(pageW, 7, fmt.Sprintf("TAHUN %d", tahun), "", 1, "C", false, 0, "")
	pdf.Ln(2)

	// Subtitle: Bulan
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(pageW, 6, fmt.Sprintf("Bulan : %s", strings.ToUpper(getIndonesianMonthName(time.Month(bulan)))), "", 1, "L", false, 0, "")
	pdf.Ln(1)

	// ============ DRAW TABLE HEADER ============
	drawGridHeader := func() {
		headerY := pdf.GetY()

		// Baris 1: NO, NAMA, TANGGAL (merged)
		pdf.SetFont("Arial", "B", 7)
		pdf.SetDrawColor(0, 0, 0)
		pdf.SetFillColor(255, 255, 255)
		pdf.SetTextColor(0, 0, 0)

		pdf.SetXY(marginL, headerY)
		pdf.CellFormat(noW, 12, "NO", "1", 0, "C", false, 0, "")
		pdf.CellFormat(namaW, 12, "NAMA", "1", 0, "C", false, 0, "")

		// Header "TANGGAL" (spanning all day columns)
		totalDayW := dayW * float64(daysInMonth)
		pdf.CellFormat(totalDayW, 6, "TANGGAL", "1", 1, "C", false, 0, "")

		// Baris 2: Angka tanggal 1-31
		pdf.SetXY(marginL+noW+namaW, headerY+6)
		for d := 1; d <= daysInMonth; d++ {
			isHoliday := holidayDays[d]
			if isHoliday {
				pdf.SetFillColor(255, 200, 200) // Merah muda untuk weekend/libur
			} else {
				pdf.SetFillColor(255, 255, 255)
			}
			pdf.CellFormat(dayW, 6, strconv.Itoa(d), "1", 0, "C", true, 0, "")
		}
		pdf.Ln(0)
		pdf.SetY(headerY + 12)
	}

	drawGridHeader()

	// ============ DRAW USER ROWS ============
	rowNo := 0

	drawUserRow := func(no int, user domain.User) {
		rowH := 10.0
		startY := pdf.GetY()

		// Cek page break
		if startY+rowH > f4Size.Ht-20 {
			pdf.AddPageFormat("L", f4Size)
			drawGridHeader()
			startY = pdf.GetY()
		}

		pdf.SetFont("Arial", "", 7)

		// NO
		pdf.SetXY(marginL, startY)
		pdf.CellFormat(noW, rowH, strconv.Itoa(no), "1", 0, "C", false, 0, "")

		// NAMA + NIP (2 baris dalam 1 cell)
		pdf.SetXY(marginL+noW, startY)
		pdf.CellFormat(namaW, rowH, "", "1", 0, "", false, 0, "")

		// Tulis nama (baris 1)
		pdf.SetFont("Arial", "B", 6.5)
		pdf.SetXY(marginL+noW+1, startY+1)
		namaUpper := strings.ToUpper(user.Nama)
		if len(namaUpper) > 35 {
			namaUpper = namaUpper[:35] + "..."
		}
		pdf.CellFormat(namaW-2, 4, namaUpper, "", 0, "L", false, 0, "")

		// Tulis NIP (baris 2)
		pdf.SetFont("Arial", "", 6)
		pdf.SetXY(marginL+noW+1, startY+5)
		pdf.CellFormat(namaW-2, 4, "Nip : "+user.NIP, "", 0, "L", false, 0, "")

		// Kolom tanggal
		userAbsensi := absensiMap[user.ID]
		for d := 1; d <= daysInMonth; d++ {
			cellX := marginL + noW + namaW + dayW*float64(d-1)
			isHoliday := holidayDays[d]

			if isHoliday {
				pdf.SetFillColor(255, 200, 200) // Merah muda
			} else {
				pdf.SetFillColor(255, 255, 255)
			}

			pdf.SetXY(cellX, startY)
			pdf.CellFormat(dayW, rowH, "", "1", 0, "C", true, 0, "")

			// Tandai kehadiran dengan centang jika hari kerja dan status hadir/terlambat/pulang_cepat/dinas_luar
			if !isHoliday && userAbsensi != nil {
				status, exists := userAbsensi[d]
				if exists && (status == "hadir" || status == "terlambat" || status == "pulang_cepat" || status == "dinas_luar") {
					pdf.SetFont("Arial", "B", 8)
					pdf.SetXY(cellX, startY+2)
					// Gunakan karakter centang ASCII
					pdf.CellFormat(dayW, 6, "v", "", 0, "C", false, 0, "")
					pdf.SetFont("Arial", "", 7)
				}
			}
		}

		pdf.SetY(startY + rowH)
	}

	// Draw PNS users
	for _, user := range pnsUsers {
		rowNo++
		drawUserRow(rowNo, user)
	}

	// Draw P3K separator jika ada
	if len(p3kUsers) > 0 {
		sepY := pdf.GetY()
		if sepY+8 > f4Size.Ht-20 {
			pdf.AddPageFormat("L", f4Size)
			drawGridHeader()
			sepY = pdf.GetY()
		}

		totalW := noW + namaW + dayW*float64(daysInMonth)
		pdf.SetXY(marginL, sepY)
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(230, 230, 230)
		pdf.CellFormat(totalW, 7, "P3K PARUH WAKTU", "1", 1, "C", true, 0, "")
		pdf.SetFillColor(255, 255, 255)

		for _, user := range p3kUsers {
			rowNo++
			drawUserRow(rowNo, user)
		}
	}

	// ============ TANDA TANGAN ============
	pdf.Ln(15)

	sigY := pdf.GetY()
	if sigY+45 > f4Size.Ht-10 {
		pdf.AddPageFormat("L", f4Size)
		sigY = pdf.GetY() + 10
	}

	sigColW := 80.0

	// Sekretaris Kelurahan (kiri)
	pdf.SetFont("Arial", "B", 10)
	pdf.SetXY(marginL+20, sigY)
	pdf.CellFormat(sigColW, 5, "Sekretaris Kelurahan", "", 2, "C", false, 0, "")
	pdf.CellFormat(sigColW, 5, "Sukanegla", "", 2, "C", false, 0, "")
	pdf.Ln(20)

	pdf.SetX(marginL + 20)
	if sekretaris != nil {
		pdf.SetFont("Arial", "BU", 10)
		pdf.CellFormat(sigColW, 5, strings.ToUpper(sekretaris.Nama), "", 2, "C", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(sigColW, 5, "Nip : "+sekretaris.NIP, "", 1, "C", false, 0, "")
	} else {
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(sigColW, 5, "( ........................................ )", "", 2, "C", false, 0, "")
		pdf.CellFormat(sigColW, 5, "Nip : ................................", "", 1, "C", false, 0, "")
	}

	// Kepala Kelurahan / Lurah (kanan)
	pdf.SetFont("Arial", "B", 10)
	rightX := pageW - sigColW + marginL - 20
	pdf.SetXY(rightX, sigY)
	pdf.CellFormat(sigColW, 5, "Kepala Kelurahan", "", 2, "C", false, 0, "")
	pdf.CellFormat(sigColW, 5, "Sukanegla", "", 2, "C", false, 0, "")
	pdf.Ln(20)

	pdf.SetX(rightX)
	if lurah != nil {
		pdf.SetFont("Arial", "BU", 10)
		pdf.CellFormat(sigColW, 5, strings.ToUpper(lurah.Nama), "", 2, "C", false, 0, "")
		pdf.SetFont("Arial", "", 9)
		pdf.CellFormat(sigColW, 5, "Nip : "+lurah.NIP, "", 1, "C", false, 0, "")
	} else {
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(sigColW, 5, "( ........................................ )", "", 2, "C", false, 0, "")
		pdf.CellFormat(sigColW, 5, "Nip : ................................", "", 1, "C", false, 0, "")
	}

	// Output ke bytes
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// getIndonesianMonthName mengembalikan nama bulan dalam Bahasa Indonesia.
func getIndonesianMonthName(m time.Month) string {
	months := map[time.Month]string{
		time.January:   "Januari",
		time.February:  "Februari",
		time.March:     "Maret",
		time.April:     "April",
		time.May:       "Mei",
		time.June:      "Juni",
		time.July:      "Juli",
		time.August:    "Agustus",
		time.September: "September",
		time.October:   "Oktober",
		time.November:  "November",
		time.December:  "Desember",
	}
	return months[m]
}
