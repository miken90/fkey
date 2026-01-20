package core

import (
	"log"
	"sync"
	"time"
	"unicode"
	"unsafe"

	"golang.org/x/text/encoding/charmap"
)

// SmartPaste global settings accessor
var (
	smartPasteEnabled     bool = true
	smartPasteEnabledMu   sync.RWMutex
)

// SetSmartPasteEnabled sets the SmartPaste enabled state (called from main.go)
func SetSmartPasteEnabled(enabled bool) {
	smartPasteEnabledMu.Lock()
	defer smartPasteEnabledMu.Unlock()
	smartPasteEnabled = enabled
}

// IsSmartPasteEnabled returns whether SmartPaste is enabled
func IsSmartPasteEnabled() bool {
	smartPasteEnabledMu.RLock()
	defer smartPasteEnabledMu.RUnlock()
	return smartPasteEnabled
}

// IsMojibake detects if a string contains mojibake patterns from UTF-8
// being misinterpreted as a legacy encoding.
func IsMojibake(s string) bool {
	signatures := []string{
		"ß╗", // Common in Vietnamese mojibake (ộ, ố, etc.) - from UTF-8 prefix 0xe1 0xbb
		"├",  // Latin letter with combining mark - from UTF-8 prefix 0xc3
		"║",  // Box drawing character - from 0xba
		"ƒÄ", // Part of emoji mojibake pattern in CP850
	}

	for _, sig := range signatures {
		for i := 0; i <= len(s)-len(sig); i++ {
			if s[i:i+len(sig)] == sig {
				return true
			}
		}
	}
	return false
}

// FixMojibake attempts to fix mojibake text by re-encoding through legacy charsets.
// Returns the fixed string and whether any changes were made.
func FixMojibake(s string) (string, bool) {
	if !IsMojibake(s) {
		return s, false
	}

	// Try CP850 first (matches the user's observed pattern)
	if fixed, ok := tryFixWithCharmap(s, charmap.CodePage850); ok {
		return fixed, true
	}

	// Fallback to Windows-1252
	if fixed, ok := tryFixWithCharmap(s, charmap.Windows1252); ok {
		return fixed, true
	}

	return s, false
}

// tryFixWithCharmap attempts to fix mojibake by reversing the encoding process:
// When UTF-8 bytes are wrongly decoded as CP850, each UTF-8 byte becomes a CP850 character.
// To reverse: encode each character back to CP850 to recover the original UTF-8 bytes.
func tryFixWithCharmap(s string, cm *charmap.Charmap) (string, bool) {
	encoder := cm.NewEncoder()

	// Build reverse lookup: for each character in the mojibake string,
	// find what byte value it represents in the legacy encoding
	var rawBytes []byte
	for _, r := range s {
		if r < 128 {
			// ASCII characters map directly
			rawBytes = append(rawBytes, byte(r))
		} else {
			// Try to encode this character back to get the original byte
			charBytes := []byte(string(r))
			encoded, err := encoder.Bytes(charBytes)
			if err == nil && len(encoded) > 0 {
				rawBytes = append(rawBytes, encoded...)
			} else {
				// Character doesn't exist in this encoding, skip or use replacement
				return "", false
			}
		}
	}

	// Interpret the recovered bytes as UTF-8
	result := string(rawBytes)

	// Validate: check if result contains Vietnamese characters
	if containsVietnamese(result) {
		return result, true
	}

	return "", false
}

// containsVietnamese checks if a string contains Vietnamese-specific Unicode characters
// in the range U+1EA0–U+1EFF (Vietnamese Extended block).
func containsVietnamese(s string) bool {
	for _, r := range s {
		if r >= 0x1EA0 && r <= 0x1EFF {
			return true
		}
		// Also check for common Vietnamese tone marks on base letters
		if isVietnameseTonedLetter(r) {
			return true
		}
	}
	return false
}

// isVietnameseTonedLetter checks for Vietnamese letters with diacritics
// that fall outside the main Vietnamese Extended block.
func isVietnameseTonedLetter(r rune) bool {
	switch r {
	case 'À', 'Á', 'Â', 'Ã', 'È', 'É', 'Ê', 'Ì', 'Í', 'Ò', 'Ó', 'Ô', 'Õ', 'Ù', 'Ú', 'Ý',
		'à', 'á', 'â', 'ã', 'è', 'é', 'ê', 'ì', 'í', 'ò', 'ó', 'ô', 'õ', 'ù', 'ú', 'ý',
		'Đ', 'đ', 'Ă', 'ă', 'Ơ', 'ơ', 'Ư', 'ư':
		return true
	}
	// Check for combining marks (used in NFD normalization)
	return unicode.Is(unicode.Mn, r)
}

// HandleSmartPaste reads clipboard, fixes mojibake if detected, and pastes
func HandleSmartPaste() {
	text, err := GetClipboardTextRetry()
	if err != nil || text == "" {
		simulatePaste()
		return
	}

	fixed, changed := FixMojibake(text)
	if !changed {
		simulatePaste()
		return
	}

	log.Printf("[SmartPaste] Fixed mojibake: %d chars -> %d chars", len(text), len(fixed))

	if err := SetClipboardTextRetry(fixed); err != nil {
		log.Printf("[SmartPaste] Failed to set clipboard: %v", err)
		simulatePaste()
		return
	}

	time.Sleep(30 * time.Millisecond)
	simulatePaste()
}

// simulatePaste releases Shift (since user is holding Ctrl+Shift+V) and sends Ctrl+V
func simulatePaste() {
	inputs := [6]INPUT{
		// Release Shift first (user is holding it)
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_SHIFT,
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// Ctrl down
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_CONTROL,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// V down
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         0x56, // VK_V
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// V up
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         0x56, // VK_V
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// Ctrl up
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_CONTROL,
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// Re-press Shift (restore user's held state)
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_SHIFT,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
	}

	procSendInput.Call(
		6,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}
