package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// Palette presets
const (
	paletteSimple   = " .:-=+*#%@"
	paletteDetailed = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
)

func main() {
	//Open file
	file, err := os.Open("test.jpeg")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	//Decode file
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		return
	}

	// Convert image to ASCII with color enabled
	asciiArt, err := ConvertImageToASCII(img, 80, true, paletteSimple)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	// Output the result
	fmt.Print(asciiArt)
}

// ConvertImageToASCII handles the core pixel-binning and character translation
func ConvertImageToASCII(img image.Image, targetWidth int, useColor bool, palette string) (string, error) {
	// Define image dimensions
	bounds := img.Bounds()
	imgWidth := bounds.Dx()
	imgHeight := bounds.Dy()

	if imgWidth <= targetWidth || imgHeight <= targetWidth {
		return "", fmt.Errorf("image dimensions too small: must be wider/taller than %dpx", targetWidth)
	}

	rawHeight := targetWidth * imgHeight / imgWidth
	targetHeight := int(float64(rawHeight) * 0.48)

	if targetHeight == 0 {
		targetHeight = 1
	}

	// Calculate the size of each pixel-bin block
	blockWidth := imgWidth / targetWidth
	blockHeight := imgHeight / targetHeight

	var buf bytes.Buffer

	// Go through each character block
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			// Find the bounding box of pixels for this specific character bin
			startX := bounds.Min.X + (x * blockWidth)
			startY := bounds.Min.Y + (y * blockHeight)
			endX := startX + blockWidth
			endY := startY + blockHeight

			// Safety boundary checks
			if endX > bounds.Max.X {
				endX = bounds.Max.X
			}
			if endY > bounds.Max.Y {
				endY = bounds.Max.Y
			}

			// Calculate average 8-bit color channels for this block
			r, g, b := getBlockAverageRGB(img, startX, startY, endX, endY)

			// Standard formula luminance (results in 0-255 scale)
			luminance := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)

			// Map the 8-bit luminance (0-255) to our palette index
			paletteIndex := int((luminance / 255.0) * float64(len(palette)-1))
			char := palette[paletteIndex]

			if useColor {
				buf.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm%c\x1b[0m", r, g, b, char))
			} else {
				buf.WriteRune(rune(char))
			}
		}
		buf.WriteByte('\n')
	}

	return buf.String(), nil
}

// getBlockAverageRGB calculates the averaged 8-bit RGB color channels for a pixel neighborhood
func getBlockAverageRGB(img image.Image, startX, startY, endX, endY int) (uint8, uint8, uint8) {
	var totalR, totalG, totalB uint64
	var count uint64

	for y := startY; y < endY; y++ {
		for x := startX; x < endX; x++ {
			r, g, b, _ := img.At(x, y).RGBA()

			totalR += uint64(r / 257)
			totalG += uint64(g / 257)
			totalB += uint64(b / 257)
			count++
		}
	}

	if count == 0 {
		return 0, 0, 0
	}

	return uint8(totalR / count), uint8(totalG / count), uint8(totalB / count)
}
