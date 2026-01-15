package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

// generateIcon creates a simple 32x32 icon with a letter
func generateIcon(letter string, bgColor, textColor color.Color) []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Fill background
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Draw simple letter pattern (V or E)
	switch letter {
	case "V":
		drawV(img, textColor)
	case "E":
		drawE(img, textColor)
	}

	// Encode to PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func drawV(img *image.RGBA, c color.Color) {
	// Simple V shape (32x32)
	// Left diagonal
	for i := 0; i < 16; i++ {
		for w := 0; w < 4; w++ {
			x := 6 + i/2 + w
			y := 6 + i
			if x < 32 && y < 26 {
				img.Set(x, y, c)
			}
		}
	}
	// Right diagonal
	for i := 0; i < 16; i++ {
		for w := 0; w < 4; w++ {
			x := 26 - i/2 - w
			y := 6 + i
			if x >= 0 && y < 26 {
				img.Set(x, y, c)
			}
		}
	}
}

func drawE(img *image.RGBA, c color.Color) {
	// Simple E shape (32x32)
	// Vertical bar
	for y := 6; y < 26; y++ {
		for x := 8; x < 12; x++ {
			img.Set(x, y, c)
		}
	}
	// Top bar
	for x := 8; x < 24; x++ {
		for y := 6; y < 10; y++ {
			img.Set(x, y, c)
		}
	}
	// Middle bar
	for x := 8; x < 22; x++ {
		for y := 14; y < 18; y++ {
			img.Set(x, y, c)
		}
	}
	// Bottom bar
	for x := 8; x < 24; x++ {
		for y := 22; y < 26; y++ {
			img.Set(x, y, c)
		}
	}
}

// CreateIconOn creates the Vietnamese "V" icon (green background)
func CreateIconOn() []byte {
	bg := color.RGBA{0x22, 0xC5, 0x5E, 0xFF}   // Green
	text := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} // White
	return generateIcon("V", bg, text)
}

// CreateIconOff creates the English "E" icon (gray background)
func CreateIconOff() []byte {
	bg := color.RGBA{0x6B, 0x72, 0x80, 0xFF}   // Gray
	text := color.RGBA{0xFF, 0xFF, 0xFF, 0xFF} // White
	return generateIcon("E", bg, text)
}
