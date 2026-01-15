package main

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

//go:embed gonhanh_core.dll
var embeddedDLL embed.FS

// ExtractDLL extracts the embedded DLL to a temp location and returns its path.
// Uses content hash for cache invalidation - only extracts if DLL changed.
func ExtractDLL() (string, error) {
	dllData, err := embeddedDLL.ReadFile("gonhanh_core.dll")
	if err != nil {
		return "", err
	}

	// Create hash-based filename for cache invalidation
	hash := sha256.Sum256(dllData)
	hashStr := hex.EncodeToString(hash[:8]) // First 8 bytes = 16 hex chars

	// Use %LOCALAPPDATA%\FKey for DLL storage (persists across runs)
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		localAppData = os.TempDir()
	}
	dllDir := filepath.Join(localAppData, "FKey")
	if err := os.MkdirAll(dllDir, 0755); err != nil {
		return "", err
	}

	dllPath := filepath.Join(dllDir, "fkey_core_"+hashStr+".dll")

	// Check if already extracted (skip if exists and matches)
	if info, err := os.Stat(dllPath); err == nil && info.Size() == int64(len(dllData)) {
		return dllPath, nil
	}

	// Clean old versions
	cleanOldDLLs(dllDir, "fkey_core_"+hashStr+".dll")

	// Extract DLL
	if err := os.WriteFile(dllPath, dllData, 0755); err != nil {
		return "", err
	}

	return dllPath, nil
}

// cleanOldDLLs removes old DLL versions, keeping only the current one
func cleanOldDLLs(dir, keepName string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		name := entry.Name()
		if name != keepName && len(name) > 10 && name[:10] == "fkey_core_" {
			os.Remove(filepath.Join(dir, name))
		}
	}
}

// GetDLLPath returns the path to the DLL (embedded or external)
// Priority: 1. Same directory as exe (for development), 2. Embedded
func GetDLLPath() (string, error) {
	// Check for external DLL first (development mode)
	exePath, err := os.Executable()
	if err == nil {
		exeDir := filepath.Dir(exePath)
		for _, name := range []string{"fkey_core.dll", "gonhanh_core.dll"} {
			localPath := filepath.Join(exeDir, name)
			if _, err := os.Stat(localPath); err == nil {
				return localPath, nil
			}
		}
	}

	// Extract embedded DLL
	return ExtractDLL()
}

// CopyDLLToExeDir copies embedded DLL next to executable (for portable mode)
func CopyDLLToExeDir() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	dllData, err := embeddedDLL.ReadFile("gonhanh_core.dll")
	if err != nil {
		return err
	}

	destPath := filepath.Join(filepath.Dir(exePath), "fkey_core.dll")

	// Check if already exists and up-to-date
	if existing, err := os.ReadFile(destPath); err == nil {
		existingHash := sha256.Sum256(existing)
		newHash := sha256.Sum256(dllData)
		if existingHash == newHash {
			return nil
		}
	}

	dest, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, io.NewSectionReader(
		&byteReader{data: dllData}, 0, int64(len(dllData))))
	return err
}

type byteReader struct {
	data []byte
}

func (r *byteReader) ReadAt(p []byte, off int64) (n int, err error) {
	if off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	n = copy(p, r.data[off:])
	if off+int64(n) >= int64(len(r.data)) {
		err = io.EOF
	}
	return
}
