package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/go-pdf/fpdf"
)

func main() {
	fPath := "images/splash_illustration.png"
	file, err := os.Open(fPath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	img, format, err := image.DecodeConfig(file)
	if err != nil {
		fmt.Println("Error decoding image config:", err)
		return
	}
	fmt.Printf("Image format: %s, Dimensions: %dx%d\n", format, img.Width, img.Height)

	// Test fpdf
	pdf := fpdf.New("P", "mm", "A4", "")
	opt := fpdf.ImageOptions{ImageType: "PNG"}
	
	pdf.RegisterImageOptions(fPath, opt)
	if pdf.Error() != nil {
		fmt.Println("FPDF Register Error:", pdf.Error())
	} else {
		fmt.Println("FPDF Registration Success.")
	}
}
