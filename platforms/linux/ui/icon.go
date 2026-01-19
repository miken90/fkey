package ui

import (
	_ "embed"
)

//go:embed icon.png
var iconBytes []byte

// GetIconBytes returns the embedded tray icon
func GetIconBytes() []byte {
	return iconBytes
}
