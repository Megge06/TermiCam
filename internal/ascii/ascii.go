package ascii

import (
	"bytes"
	"fmt"
)

// Palette presets
const (
	PaletteSimple   = " .:-=+*#%@"
	PaletteDetailed = " .'`^\",:;Il!i><~+_-?][}{1)(|\\/tfjrxnuvczXYUJCLQ0OZmwqpdbkhao*#MW&8%B@$"
)

// ConvertRGB24ToASCII handles the conversion of an RGB24 image to ASCII art
func ConvertRGB24ToASCII(pix []byte, imgWidth, imgHeight, targetWidth int, useColor bool, palette string) (string, error) {
	if imgWidth <= 0 || imgHeight <= 0 || targetWidth <= 0 {
		return "", fmt.Errorf("invalid dimensions: imgWidth=%d, imgHeight=%d, targetWidth=%d", imgWidth, imgHeight, targetWidth)
	}

	rawHeight := targetWidth * imgHeight / imgWidth
	targetHeight := int(float64(rawHeight) * 0.45)

	if targetHeight <= 0 {
		targetHeight = 1
	}

	var buf bytes.Buffer

	// Go through each character block
	for y := 0; y < targetHeight; y++ {
		for x := 0; x < targetWidth; x++ {
			startX := int(float64(x) * float64(imgWidth) / float64(targetWidth))
			startY := int(float64(y) * float64(imgHeight) / float64(targetHeight))
			endX := int(float64(x+1) * float64(imgWidth) / float64(targetWidth))
			endY := int(float64(y+1) * float64(imgHeight) / float64(targetHeight))

			// Safety boundary checks
			if endX > imgWidth {
				endX = imgWidth
			}
			if endY > imgHeight {
				endY = imgHeight
			}
			if startX >= endX {
				endX = startX + 1
			}
			if startY >= endY {
				endY = startY + 1
			}

			// Calculate average 8-bit color channels for this block
			r, g, b := getBlockAverageRGB24(pix, imgWidth, startX, startY, endX, endY)

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
		if y < targetHeight-1 {
			buf.WriteByte('\n')
		}
	}

	return buf.String(), nil
}

// getBlockAverageRGB calculates the averaged 8-bit RGB color channels for a pixel neighborhood
func getBlockAverageRGB24(pix []byte, imgWidth int, startX, startY, endX, endY int) (uint8, uint8, uint8) {
	var totalR, totalG, totalB uint64
	var count uint64

	for y := startY; y < endY; y++ {
		rowOffset := y * imgWidth * 3 // Jump directly to the start of the row
		for x := startX; x < endX; x++ {
			idx := rowOffset + (x * 3)

			if idx+2 < len(pix) {
				totalR += uint64(pix[idx])
				totalG += uint64(pix[idx+1])
				totalB += uint64(pix[idx+2])
				count++
			}
		}
	}

	if count == 0 {
		return 0, 0, 0
	}

	return uint8(totalR / count), uint8(totalG / count), uint8(totalB / count)
}
