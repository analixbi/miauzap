package main

import (
	"image"
	"image/png"
	"log"
	"os"

	"github.com/nfnt/resize"
)

func main() {
	// Open the source image
	file, err := os.Open("static/images/miauzap_logo.png")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Decode the image
	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// Create 16x16 favicon
	favicon16 := resize.Resize(16, 16, img, resize.Lanczos3)
	out16, err := os.Create("static/api/favicon-16x16.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out16.Close()
	png.Encode(out16, favicon16)

	// Create 32x32 favicon
	favicon32 := resize.Resize(32, 32, img, resize.Lanczos3)
	out32, err := os.Create("static/api/favicon-32x32.png")
	if err != nil {
		log.Fatal(err)
	}
	defer out32.Close()
	png.Encode(out32, favicon32)

	// Copy to static/images/favicon.png (full size or resized)
	faviconMain := resize.Resize(256, 256, img, resize.Lanczos3)
	outMain, err := os.Create("static/images/favicon.png")
	if err != nil {
		log.Fatal(err)
	}
	defer outMain.Close()
	png.Encode(outMain, faviconMain)

	log.Println("Favicons created successfully!")
}
