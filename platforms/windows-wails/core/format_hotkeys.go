package core

// Format hotkey definitions for text formatting feature
// Maps keyboard shortcuts to formatting types

const (
	VK_B = 0x42 // Ctrl+B → bold
	VK_I = 0x49 // Ctrl+I → italic
	VK_U = 0x55 // Ctrl+U → underline
	VK_K = 0x4B // Ctrl+K → link
	VK_S = 0x53 // Ctrl+Alt+S → strikethrough
	// VK_OEM_3 = 0xC0 already defined in keyboard_hook.go → Ctrl+` → code
)

// IsFormatHotkey checks if the key combination is a format hotkey
// Returns the format type and whether it matched
func IsFormatHotkey(keyCode uint16, ctrl, alt, shift bool) (formatType string, matched bool) {
	// Format hotkeys require Ctrl
	if !ctrl {
		return "", false
	}

	switch keyCode {
	case VK_B:
		if !alt && !shift {
			return "bold", true
		}
	case VK_I:
		if !alt && !shift {
			return "italic", true
		}
	case VK_U:
		if !alt && !shift {
			return "underline", true
		}
	case VK_K:
		if !alt && !shift {
			return "link", true
		}
	case VK_S:
		// Ctrl+Alt+S for strikethrough (avoids browser conflicts)
		if alt && !shift {
			return "strikethrough", true
		}
	case VK_OEM_3:
		if !alt && !shift {
			return "code", true
		}
	}

	return "", false
}
