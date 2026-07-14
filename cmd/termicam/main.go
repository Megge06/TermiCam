package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"

	"github.com/Megge06/TermiCam/internal/ascii"
)

func main() {
	// Open file
	file, err := os.Open("test.png")
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	// Decode file
	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Printf("Error decoding image: %v\n", err)
		return
	}

	// Convert image to ASCII with color enabled
	asciiArt, err := ascii.ConvertImageToASCII(img, 80, true, ascii.PaletteSimple)
	if err != nil {
		fmt.Printf("Conversion error: %v\n", err)
		return
	}

	// Output the result
	fmt.Print(asciiArt)
}
