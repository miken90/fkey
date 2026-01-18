package core

// Clipboard operations using Windows API
// Supports Unicode text (UTF-16) for Vietnamese characters

import (
	"errors"
	"time"
	"unicode/utf16"
	"unsafe"
)

const (
	CF_UNICODETEXT = 13
	GMEM_MOVEABLE  = 0x0002

	// Clipboard retry settings for Electron apps (Discord, Slack, etc.)
	ClipboardMaxRetries  = 5
	ClipboardRetryDelay  = 20 * time.Millisecond
	ClipboardPollTimeout = 600 * time.Millisecond
	ClipboardPollInterval = 10 * time.Millisecond
)

var (
	procOpenClipboard    = user32.NewProc("OpenClipboard")
	procCloseClipboard   = user32.NewProc("CloseClipboard")
	procGetClipboardData = user32.NewProc("GetClipboardData")
	procSetClipboardData = user32.NewProc("SetClipboardData")
	procEmptyClipboard   = user32.NewProc("EmptyClipboard")

	procGlobalAlloc  = kernel32.NewProc("GlobalAlloc")
	procGlobalLock   = kernel32.NewProc("GlobalLock")
	procGlobalUnlock = kernel32.NewProc("GlobalUnlock")
	procGlobalFree   = kernel32.NewProc("GlobalFree")
	procGlobalSize   = kernel32.NewProc("GlobalSize")

	procLstrlenW = kernel32.NewProc("lstrlenW")

	procGetClipboardSequenceNumber = user32.NewProc("GetClipboardSequenceNumber")
)

var (
	ErrOpenClipboard  = errors.New("failed to open clipboard")
	ErrGetClipboard   = errors.New("failed to get clipboard data")
	ErrSetClipboard   = errors.New("failed to set clipboard data")
	ErrEmptyClipboard = errors.New("failed to empty clipboard")
	ErrGlobalAlloc    = errors.New("failed to allocate global memory")
	ErrGlobalLock     = errors.New("failed to lock global memory")
)

// GetClipboardText retrieves Unicode text from the clipboard
func GetClipboardText() (string, error) {
	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return "", ErrOpenClipboard
	}
	defer procCloseClipboard.Call()

	hData, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if hData == 0 {
		return "", nil // No text data, return empty string
	}

	ptr, _, _ := procGlobalLock.Call(hData)
	if ptr == 0 {
		return "", ErrGlobalLock
	}
	defer procGlobalUnlock.Call(hData)

	// Get string length
	length, _, _ := procLstrlenW.Call(ptr)
	if length == 0 {
		return "", nil
	}

	// Read UTF-16 data
	utf16Slice := make([]uint16, length)
	for i := uintptr(0); i < length; i++ {
		utf16Slice[i] = *(*uint16)(unsafe.Pointer(ptr + i*2))
	}

	return string(utf16.Decode(utf16Slice)), nil
}

// SetClipboardText sets Unicode text to the clipboard
func SetClipboardText(text string) error {
	if text == "" {
		return nil
	}

	ret, _, _ := procOpenClipboard.Call(0)
	if ret == 0 {
		return ErrOpenClipboard
	}
	defer procCloseClipboard.Call()

	ret, _, _ = procEmptyClipboard.Call()
	if ret == 0 {
		return ErrEmptyClipboard
	}

	// Convert to UTF-16 with null terminator
	utf16Text := utf16.Encode([]rune(text))
	utf16Text = append(utf16Text, 0) // Null terminator

	// Allocate global memory
	size := uintptr(len(utf16Text) * 2)
	hMem, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if hMem == 0 {
		return ErrGlobalAlloc
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		procGlobalFree.Call(hMem)
		return ErrGlobalLock
	}

	// Copy UTF-16 data
	for i, ch := range utf16Text {
		*(*uint16)(unsafe.Pointer(ptr + uintptr(i*2))) = ch
	}

	procGlobalUnlock.Call(hMem)

	// Set clipboard data (clipboard takes ownership of memory)
	ret, _, _ = procSetClipboardData.Call(CF_UNICODETEXT, hMem)
	if ret == 0 {
		procGlobalFree.Call(hMem)
		return ErrSetClipboard
	}

	return nil
}

// GetClipboardSequenceNumber returns the current clipboard sequence number
func GetClipboardSequenceNumber() uint32 {
	ret, _, _ := procGetClipboardSequenceNumber.Call()
	return uint32(ret)
}

// OpenClipboardRetry attempts to open clipboard with retry logic for Electron apps
func OpenClipboardRetry() error {
	for i := 0; i < ClipboardMaxRetries; i++ {
		ret, _, _ := procOpenClipboard.Call(0)
		if ret != 0 {
			return nil
		}
		if i < ClipboardMaxRetries-1 {
			time.Sleep(ClipboardRetryDelay)
		}
	}
	return ErrOpenClipboard
}

// WaitClipboardChange waits for clipboard sequence number to change
func WaitClipboardChange(oldSeq uint32) bool {
	deadline := time.Now().Add(ClipboardPollTimeout)
	for time.Now().Before(deadline) {
		newSeq := GetClipboardSequenceNumber()
		if newSeq != oldSeq {
			return true
		}
		time.Sleep(ClipboardPollInterval)
	}
	return false
}

// GetClipboardTextRetry retrieves clipboard text with retry logic
func GetClipboardTextRetry() (string, error) {
	if err := OpenClipboardRetry(); err != nil {
		return "", err
	}
	defer procCloseClipboard.Call()

	hData, _, _ := procGetClipboardData.Call(CF_UNICODETEXT)
	if hData == 0 {
		return "", nil
	}

	ptr, _, _ := procGlobalLock.Call(hData)
	if ptr == 0 {
		return "", ErrGlobalLock
	}
	defer procGlobalUnlock.Call(hData)

	length, _, _ := procLstrlenW.Call(ptr)
	if length == 0 {
		return "", nil
	}

	utf16Slice := make([]uint16, length)
	for i := uintptr(0); i < length; i++ {
		utf16Slice[i] = *(*uint16)(unsafe.Pointer(ptr + i*2))
	}

	return string(utf16.Decode(utf16Slice)), nil
}

// SetClipboardTextRetry sets clipboard text with retry logic
func SetClipboardTextRetry(text string) error {
	if text == "" {
		return nil
	}

	if err := OpenClipboardRetry(); err != nil {
		return err
	}
	defer procCloseClipboard.Call()

	ret, _, _ := procEmptyClipboard.Call()
	if ret == 0 {
		return ErrEmptyClipboard
	}

	utf16Text := utf16.Encode([]rune(text))
	utf16Text = append(utf16Text, 0)

	size := uintptr(len(utf16Text) * 2)
	hMem, _, _ := procGlobalAlloc.Call(GMEM_MOVEABLE, size)
	if hMem == 0 {
		return ErrGlobalAlloc
	}

	ptr, _, _ := procGlobalLock.Call(hMem)
	if ptr == 0 {
		procGlobalFree.Call(hMem)
		return ErrGlobalLock
	}

	for i, ch := range utf16Text {
		*(*uint16)(unsafe.Pointer(ptr + uintptr(i*2))) = ch
	}

	procGlobalUnlock.Call(hMem)

	ret, _, _ = procSetClipboardData.Call(CF_UNICODETEXT, hMem)
	if ret == 0 {
		procGlobalFree.Call(hMem)
		return ErrSetClipboard
	}

	return nil
}

// ClearClipboard clears the clipboard contents
func ClearClipboard() error {
	if err := OpenClipboardRetry(); err != nil {
		return err
	}
	defer procCloseClipboard.Call()

	ret, _, _ := procEmptyClipboard.Call()
	if ret == 0 {
		return ErrEmptyClipboard
	}
	return nil
}
