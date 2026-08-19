package handler

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/go-pdf/fpdf"
	"github.com/gofiber/fiber/v3"

	embedImages "laporanharianapi/images"
	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

var (
	cachedLogos     map[string]*bytes.Buffer
	cachedLogosOnce sync.Once
)

func loadLogosToCache() {
	cachedLogos = make(map[string]*bytes.Buffer)
	allImages := []string{
		"images/logo.png",
		"images/logo_berakhlak.png",
		"images/Logo_EVP.png",
		"images/splash_illustration.png",
	}

	for _, f := range allImages {
		name := filepath.Base(f)
		file, errFS := embedImages.FS.Open(name)
		if errFS != nil {
			continue
		}
		img, errOpen := imaging.Decode(file, imaging.AutoOrientation(true))
		file.Close()
		if errOpen == nil {
			whiteBg := image.NewRGBA(img.Bounds())
			draw.Draw(whiteBg, whiteBg.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
			draw.Draw(whiteBg, whiteBg.Bounds(), img, img.Bounds().Min, draw.Over)

			var buf bytes.Buffer
			if jpeg.Encode(&buf, whiteBg, &jpeg.Options{Quality: 90}) == nil {
				cachedLogos[name] = &buf
			}
		}
	}
}

func (h *ReportHandler) ExportReportPDFHandler(c fiber.Ctx) error {
	requesterIDFloat, ok := c.Locals("user_id").(float64)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "User tidak terautentikasi"})
	}
	requesterID := uint(requesterIDFloat)

	requesterRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "error", "message": "Role tidak ditemukan"})
	}
	roleBase := strings.ToLower(requesterRole)

	// 1. Tentukan target users berdasarkan RBAC
	targetUserIDsStr := c.Query("user_ids")
	if targetUserIDsStr == "" {
		targetUserIDsStr = c.Query("user_id")
	}

	var targetUserIDs []uint
	if targetUserIDsStr != "" {
		parts := strings.Split(targetUserIDsStr, ",")
		for _, part := range parts {
			part = strings.TrimSpace(part)
			if id, err := strconv.Atoi(part); err == nil && id > 0 {
				targetUserIDs = append(targetUserIDs, uint(id))
			}
		}
	}

	var targetUsers []domain.User

	switch roleBase {
	case "staf", "kasi":
		user, err := h.userService.GetUserByID(requesterID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
		}
		targetUsers = []domain.User{*user}

	case "sekertaris", "sekretaris":
		if len(targetUserIDs) > 0 {
			for _, id := range targetUserIDs {
				user, err := h.userService.GetUserByID(id)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("User dengan ID %d tidak ditemukan", id)})
				}
				roleLower := strings.ToLower(user.Role)
				if roleLower != "staf" && roleLower != "kasi" && roleLower != "sekertaris" && roleLower != "sekretaris" {
					return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("Akses ditolak: Tidak dapat mengekspor laporan user dengan ID %d", id)})
				}
				targetUsers = append(targetUsers, *user)
			}
		} else {
			users, err := h.userService.GetUsersByRoles([]string{"staf", "Staf", "kasi", "Kasi", "sekertaris", "Sekertaris"})
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
			}
			targetUsers = users
		}

	default: // Lurah
		if len(targetUserIDs) > 0 {
			for _, id := range targetUserIDs {
				user, err := h.userService.GetUserByID(id)
				if err != nil {
					return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": fmt.Sprintf("User dengan ID %d tidak ditemukan", id)})
				}
				targetUsers = append(targetUsers, *user)
			}
		} else {
			users, err := h.userService.GetAllUsers()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Gagal mengambil data user"})
			}
			targetUsers = users
		}
	}

	// 2. Parse tanggal
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	var startDate, endDate time.Time
	var err error
	if startDateStr != "" && endDateStr != "" {
		startDate, err = time.ParseInLocation("2006-01-02", startDateStr, time.Local)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Format start_date tidak valid"})
		}
		endDate, err = time.ParseInLocation("2006-01-02", endDateStr, time.Local)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "error", "message": "Format end_date tidak valid"})
		}
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)
	} else {
		now := time.Now()
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
		endDate = startDate.AddDate(0, 1, -1)
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, time.Local)
	}

	// 3. Fetch laporan in a single bulk query (Solusi N+1)
	filter := repository.ReportFilter{
		Limit:     1000000, 
		Offset:    0,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		SortOrder: "asc", 
	}
	
	if len(targetUserIDs) > 0 {
		var ids []int
		for _, u := range targetUsers {
			ids = append(ids, int(u.ID))
		}
		filter.UserIDs = ids
	}

	// Menggunakan requesterRole asli dari token untuk RBAC
	allReports, _, err := h.reportService.GetAllReports(filter, requesterRole, requesterID)
	if err != nil {
		allReports = []domain.Laporan{} // Fallback to empty slice
	}

	// Group reports by user ID
	reportsByUser := make(map[uint][]domain.Laporan)
	for _, rp := range allReports {
		if rp.UserID != nil {
			reportsByUser[*rp.UserID] = append(reportsByUser[*rp.UserID], rp)
		}
	}

	// 4. Parallel Image Processing using Goroutine Pool
	exePath, exeErr := os.Executable()
	var baseDir string
	if exeErr == nil {
		baseDir = filepath.Dir(exePath)
	} else {
		baseDir, _ = os.Getwd()
	}
	if strings.Contains(baseDir, "go-build") || strings.Contains(baseDir, "/tmp") {
		baseDir, _ = os.Getwd()
	}

	colW := []float64{8, 22, 25, 30, 50, 50.9}

	type ProcessedImage struct {
		ReportID  uint
		JPEGBytes []byte
		Width     float64
		Height    float64
		Path      string
	}

	processedImages := make(map[uint]ProcessedImage)
	var mu sync.Mutex
	var wg sync.WaitGroup

	type Job struct {
		Report domain.Laporan
	}
	jobs := make(chan Job, len(allReports))

	numWorkers := runtime.NumCPU()
	if numWorkers < 4 {
		numWorkers = 4
	}
	if numWorkers > 16 {
		numWorkers = 16
	}
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				if job.Report.FotoURL == nil || *job.Report.FotoURL == "" {
					continue
				}

				p := strings.ReplaceAll(*job.Report.FotoURL, "\\", "/")
				p = strings.TrimPrefix(p, "/")
				localPath := filepath.Join(baseDir, filepath.FromSlash(p))
				if _, statErr := os.Stat(localPath); statErr != nil {
					cwd, _ := os.Getwd()
					localPath = filepath.Join(cwd, filepath.FromSlash(p))
				}

				if _, statErr := os.Stat(localPath); statErr == nil {
					img, openErr := imaging.Open(localPath, imaging.AutoOrientation(true))
					if openErr == nil {
						maxW := colW[5] - 2
						maxH := 70.0

						imgW := float64(img.Bounds().Dx())
						imgH := float64(img.Bounds().Dy())
						imgRatio := imgH / imgW

						scaledW := maxW
						scaledH := scaledW * imgRatio
						if scaledH > maxH {
							scaledH = maxH
							scaledW = scaledH / imgRatio
						}

						targetPxW := int(scaledW * 3.78)
						targetPxH := int(scaledH * 3.78)
						if targetPxW > img.Bounds().Dx() {
							targetPxW = img.Bounds().Dx()
							targetPxH = img.Bounds().Dy()
						}

						// Use Box filter instead of Lanczos for massive speedup
						resized := imaging.Resize(img, targetPxW, targetPxH, imaging.Box)

						var buf bytes.Buffer
						if jpeg.Encode(&buf, resized, &jpeg.Options{Quality: 70}) == nil {
							mu.Lock()
							processedImages[job.Report.ID] = ProcessedImage{
								ReportID:  job.Report.ID,
								JPEGBytes: buf.Bytes(),
								Width:     scaledW,
								Height:    scaledH,
								Path:      localPath,
							}
							mu.Unlock()
						}
					}
				}
			}
		}()
	}

	for _, rp := range allReports {
		jobs <- Job{Report: rp}
	}
	close(jobs)
	wg.Wait() // Tunggu semua foto diproses

	// 5. Setup PDF (F4 = 215.9mm x 330.2mm)
	pdf := fpdf.New("P", "mm", "A4", "") 
	f4Size := fpdf.SizeType{Wd: 215.9, Ht: 330.2}

	marginL, marginR := 15.0, 15.0
	pdf.SetMargins(marginL, 20, marginR)
	pageW := 215.9 - marginL - marginR 

	// Init cached logos
	cachedLogosOnce.Do(loadLogosToCache)

	for name, buf := range cachedLogos {
		opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
		pdf.RegisterImageOptionsReader(name, opt, bytes.NewReader(buf.Bytes()))
	}

	type logoTarget struct {
		file   string
		height float64
	}

	logoFiles := []logoTarget{
		{"logo.png", 12.0},             
		{"logo_berakhlak.png", 8.0},    
		{"Logo_EVP.png", 8.0},          
	}

	splashName := "splash_illustration.png"
	logoGap := 2.0
	pdf.SetHeaderFuncMode(func() {
		// Watermark
		if infoSplash := pdf.GetImageInfo(splashName); infoSplash != nil {
			splashW := 90.0 
			splashH := splashW * float64(infoSplash.Height()) / float64(infoSplash.Width())
			
			splashX := (215.9 - splashW) / 2
			splashY := (330.2 - splashH) / 2

			pdf.SetAlpha(0.15, "Normal") 
			opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
			pdf.ImageOptions(splashName, splashX, splashY, splashW, splashH, false, opt, 0, "")
			pdf.SetAlpha(1.0, "Normal") 
		}

		// Logo Header
		logoX := marginL
		baseY := 3.0
		baseH := 12.0 

		for _, logo := range logoFiles {
			if info := pdf.GetImageInfo(logo.file); info != nil {
				opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
				offsetY := (baseH - logo.height) / 2
				currentY := baseY + offsetY

				pdf.ImageOptions(logo.file, logoX, currentY, 0, logo.height, false, opt, 0, "")
				logoW := logo.height * float64(info.Width()) / float64(info.Height())
				logoX += logoW + logoGap
			}
		}
	}, true)

	getIndonesianMonth := func(m time.Month) string {
		months := map[time.Month]string{
			time.January:   "Januari", time.February:  "Februari", time.March:     "Maret",
			time.April:     "April", time.May:       "Mei", time.June:      "Juni",
			time.July:      "Juli", time.August:    "Agustus", time.September: "September",
			time.October:   "Oktober", time.November:  "November", time.December:  "Desember",
		}
		return months[m]
	}

	formatDateIndo := func(t time.Time) string {
		return fmt.Sprintf("%02d %s %d", t.Day(), getIndonesianMonth(t.Month()), t.Year())
	}

	colHeaders := []string{"No", "Waktu\nPelaksanaan", "Jenis\nLaporan", "Judul\nLaporan", "Deskripsi", "Foto"}

	headerBgR, headerBgG, headerBgB := 255, 255, 255
	headerFgR, headerFgG, headerFgB := 0, 0, 0

	drawTableHeader := func() {
		pdf.SetFont("Arial", "B", 8)
		pdf.SetFillColor(headerBgR, headerBgG, headerBgB)
		pdf.SetTextColor(headerFgR, headerFgG, headerFgB)
		pdf.SetDrawColor(200, 200, 200)

		headerH := 10.0
		startY := pdf.GetY()
		startX := marginL

		for i, w := range colW {
			pdf.SetXY(startX, startY)
			pdf.CellFormat(w, headerH, "", "1", 0, "C", false, 0, "")
			lines := strings.Split(colHeaders[i], "\n")
			lineH := 4.0
			totalTextH := float64(len(lines)) * lineH
			offsetY := (headerH - totalTextH) / 2

			pdf.SetXY(startX, startY+offsetY)
			pdf.MultiCell(w, lineH, colHeaders[i], "0", "C", false)
			startX += w
		}
		pdf.SetXY(marginL, startY+headerH)
		pdf.SetTextColor(0, 0, 0)
		pdf.SetFillColor(255, 255, 255)
	}

	calcTextRows := func(text string, colWidth float64, fontSize float64) int {
		if text == "" {
			return 1
		}
		charsPerLine := int((colWidth / fontSize) * 1.9)
		if charsPerLine < 1 {
			charsPerLine = 1
		}
		lineCount := 0
		words := strings.Fields(text)
		currentLen := 0
		for _, w := range words {
			if currentLen+len(w)+1 > charsPerLine {
				lineCount++
				currentLen = len(w)
			} else {
				currentLen += len(w) + 1
			}
		}
		lineCount++ 
		return lineCount
	}

	rowFillAlt := false 

	addReportRow := func(no int, laporan domain.Laporan) {
		lineH := 4.5 

		jenis := "Individu"
		if laporan.TipeLaporan {
			jenis = "Organisasi"
			if laporan.TugasOrganisasi != nil {
				jenis = "Organisasi\n" + laporan.TugasOrganisasi.JudulTugas
			}
		}

		judul := laporan.JudulKegiatan
		desc := laporan.DeskripsiHasil
		tanggal := fmt.Sprintf("%02d/%02d/%d\n%02d:%02d", laporan.WaktuPelaporan.Day(), int(laporan.WaktuPelaporan.Month()), laporan.WaktuPelaporan.Year(), laporan.WaktuPelaporan.Hour(), laporan.WaktuPelaporan.Minute())

		rowsJenis := calcTextRows(jenis, colW[2], 7)
		rowsJudul := calcTextRows(judul, colW[3], 7)
		rowsDesc := calcTextRows(desc, colW[4], 7)
		rowsTanggal := 2

		maxTextRows := rowsJenis
		if rowsJudul > maxTextRows { maxTextRows = rowsJudul }
		if rowsDesc > maxTextRows { maxTextRows = rowsDesc }
		if rowsTanggal > maxTextRows { maxTextRows = rowsTanggal }
		rowH := float64(maxTextRows)*lineH + 4

		photoH := 0.0
		photoW := 0.0
		photoPath := ""
		var photoBytes []byte 

		if pimg, ok := processedImages[laporan.ID]; ok {
			photoW = pimg.Width
			photoH = pimg.Height
			photoPath = pimg.Path
			photoBytes = pimg.JPEGBytes
		}

		if photoH+4 > rowH {
			rowH = photoH + 4
		}
		if rowH < 12 {
			rowH = 12
		}

		_, pageH := pdf.GetPageSize()
		bottomMargin := 20.0
		if pdf.GetY()+rowH > pageH-bottomMargin {
			pdf.AddPageFormat("P", f4Size)
			drawTableHeader()
		}

		startX := pdf.GetX()
		startY := pdf.GetY()

		pdf.SetFillColor(255, 255, 255)
		rowFillAlt = !rowFillAlt

		pdf.SetFont("Arial", "", 7.5)
		pdf.SetDrawColor(200, 200, 200)

		drawCell := func(x, y, w, h float64, txt string, align string) {
			pdf.SetXY(x, y)
			pdf.CellFormat(w, h, "", "1", 0, "", false, 0, "")
			if txt != "" {
				rows := calcTextRows(txt, w, 7)
				realTextH := float64(rows) * lineH
				offsetY := (h - realTextH) / 2
				if offsetY < 0 { offsetY = 0 }
				pdf.SetXY(x, y+offsetY)
				pdf.MultiCell(w, lineH, txt, "0", align, false)
			}
		}

		drawCell(startX, startY, colW[0], rowH, strconv.Itoa(no), "C")
		drawCell(startX+colW[0], startY, colW[1], rowH, tanggal, "C")
		drawCell(startX+colW[0]+colW[1], startY, colW[2], rowH, jenis, "C")
		drawCell(startX+colW[0]+colW[1]+colW[2], startY, colW[3], rowH, judul, "L")
		drawCell(startX+colW[0]+colW[1]+colW[2]+colW[3], startY, colW[4], rowH, desc, "L")

		fotoX := startX + colW[0] + colW[1] + colW[2] + colW[3] + colW[4]
		pdf.SetXY(fotoX, startY)
		pdf.CellFormat(colW[5], rowH, "", "1", 0, "C", false, 0, "")

		if photoPath != "" && len(photoBytes) > 0 {
			imgX := fotoX + (colW[5]-photoW)/2
			imgY := startY + (rowH-photoH)/2

			imgName := fmt.Sprintf("opt_%d", laporan.ID)
			opt := fpdf.ImageOptions{ImageType: "JPEG", ReadDpi: false}
			pdf.RegisterImageOptionsReader(imgName, opt, bytes.NewReader(photoBytes))
			pdf.ImageOptions(imgName, imgX, imgY, photoW, photoH, false, opt, 0, "")
		} else if laporan.FotoURL != nil && *laporan.FotoURL != "" {
			pdf.SetXY(fotoX, startY+(rowH/2)-2)
			pdf.SetFont("Arial", "I", 6)
			pdf.CellFormat(colW[5], 4, "File tdk ditemukan", "", 0, "C", false, 0, "")
		} else {
			pdf.SetXY(fotoX, startY+(rowH/2)-2)
			pdf.SetFont("Arial", "I", 6)
			pdf.CellFormat(colW[5], 4, "- Tanpa Foto -", "", 0, "C", false, 0, "")
		}

		pdf.SetXY(startX, startY+rowH)
	}

	// 6. Generate PDF per user using bulk-fetched reports
	for _, user := range targetUsers {
		reports := reportsByUser[user.ID]
		if len(reports) == 0 {
			continue
		}

		pdf.AddPageFormat("P", f4Size)

		// Kop halaman
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(pageW, 8, "Laporan Harian Pegawai", "", 1, "C", false, 0, "")
		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(pageW, 6, user.Nama, "", 1, "C", false, 0, "")

		pdf.SetFont("Arial", "", 9)
		if user.Jabatan != nil {
			pdf.CellFormat(pageW, 6, user.Jabatan.NamaJabatan, "", 1, "C", false, 0, "")
		}
		pdf.CellFormat(pageW, 6,
			fmt.Sprintf("Periode: %s s/d %s", formatDateIndo(startDate), formatDateIndo(endDate)),
			"", 1, "C", false, 0, "")
		pdf.Ln(2)

		drawTableHeader()

		rowFillAlt = false
		for i, laporan := range reports {
			addReportRow(i+1, laporan)
		}

		pdf.Ln(10)

		_, pageH := pdf.GetPageSize()
		bottomMargin := 20.0
		if pdf.GetY()+40 > pageH-bottomMargin {
			pdf.AddPageFormat("P", f4Size)
		}

		sigY := pdf.GetY()
		pdf.SetFont("Arial", "", 10)

		leftX := marginL
		pdf.SetXY(leftX, sigY)
		pdf.CellFormat(60, 5, "Mengetahui,", "", 2, "C", false, 0, "")
		pdf.CellFormat(60, 5, "Pejabat Penilai Kinerja,", "", 2, "C", false, 0, "")

		pdf.Ln(20)

		pdf.SetX(leftX)
		supervisorName := ""
		supervisorNIP := ""
		if strings.ToLower(user.Role) == "lurah" {
			supervisorName = "Rena Sudrajat, S.Sos., M.Si"
			supervisorNIP = "197208241992031003"
		} else if user.Supervisor != nil {
			supervisorName = user.Supervisor.Nama
			supervisorNIP = user.Supervisor.NIP
		}

		if supervisorName != "" {
			pdf.SetFont("Arial", "BU", 10)
			pdf.CellFormat(60, 5, supervisorName, "", 2, "C", false, 0, "")
			pdf.SetFont("Arial", "", 10)
			pdf.CellFormat(60, 5, "NIP. "+supervisorNIP, "", 1, "C", false, 0, "")
		} else {
			pdf.SetFont("Arial", "", 10)
			pdf.CellFormat(60, 5, "( ........................................ )", "", 2, "C", false, 0, "")
			pdf.CellFormat(60, 5, "NIP. ........................................ ", "", 1, "C", false, 0, "")
		}

		pdf.SetFont("Arial", "", 10)
		rightX := pageW - 60 + marginL 
		pdf.SetXY(rightX, sigY)
		pdf.CellFormat(60, 5, "", "", 2, "C", false, 0, "") 
		pdf.CellFormat(60, 5, "Yang Dinilai,", "", 2, "C", false, 0, "")

		pdf.Ln(20)

		pdf.SetX(rightX)
		pdf.SetFont("Arial", "BU", 10)
		pdf.CellFormat(60, 5, user.Nama, "", 2, "C", false, 0, "")
		pdf.SetFont("Arial", "", 10)
		pdf.CellFormat(60, 5, "NIP. "+user.NIP, "", 1, "C", false, 0, "")
	}

	if pdf.PageCount() == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "error", "message": "Tidak ada laporan dalam periode tersebut"})
	}

	filename := fmt.Sprintf("laporan_harian_%s_sd_%s.pdf", startDate.Format("20060102"), endDate.Format("20060102"))
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/pdf")

	// 7. Tulis ke pipe dan streaming ke response
	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()
		err := pdf.Output(pw)
		if err != nil {
			fmt.Printf("Gagal generate PDF stream: %v\n", err)
		}
	}()

	return c.SendStream(pr)
}
