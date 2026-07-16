package tui

import (
	"image"
	_ "image/jpeg" // Side-effect import to register JPEG decoder
	_ "image/png"  // Side-effect import to register PNG decoder
	"os"
)

// loadImage opens a file path and decodes it into an image.Image
func loadImage(filePath string) (image.Image, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// image.Decode automatically detects the format (PNG, JPEG, etc.)
	// as long as the blank imports above are included
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}
