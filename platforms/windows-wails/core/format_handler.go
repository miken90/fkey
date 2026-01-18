package core

import (
	"fkey/services"
	"sync"
	"time"
	"unsafe"
)

// Format handler delays (milliseconds)
const (
	ClipboardCopyDelay   = 50  // Fallback delay if polling fails
	ClipboardPasteDelay  = 30  // Delay after paste for restoration
	ClipboardRestoreWait = 150 // Wait before restoring original clipboard

	// L/R modifier virtual key codes
	VK_LCONTROL = 0xA2
	VK_RCONTROL = 0xA3
	VK_LSHIFT   = 0xA0
	VK_RSHIFT   = 0xA1
	VK_LMENU    = 0xA4
	VK_RMENU    = 0xA5
)

var (
	formatHandler     *FormatHandler
	formatHandlerOnce sync.Once
	formatHandlerMu   sync.RWMutex
	formatOpMu        sync.Mutex // Serializes all formatting operations
)

// FormatHandler handles text formatting via clipboard operations
type FormatHandler struct {
	service *services.FormattingService
	enabled bool
	mu      sync.RWMutex
}

// InitFormatHandler initializes the global format handler with a formatting service
func InitFormatHandler(svc *services.FormattingService) *FormatHandler {
	formatHandlerOnce.Do(func() {
		formatHandler = &FormatHandler{
			service: svc,
			enabled: true,
		}
	})
	return formatHandler
}

// GetFormatHandler returns the global format handler
func GetFormatHandler() *FormatHandler {
	formatHandlerMu.RLock()
	defer formatHandlerMu.RUnlock()
	return formatHandler
}

// SetFormatHandler sets the global format handler (used by services)
func SetFormatHandler(h *FormatHandler) {
	formatHandlerMu.Lock()
	defer formatHandlerMu.Unlock()
	formatHandler = h
}

// IsEnabled returns whether formatting is enabled
func (h *FormatHandler) IsEnabled() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.enabled && h.service != nil && h.service.IsEnabled()
}

// SetEnabled enables or disables the format handler
func (h *FormatHandler) SetEnabled(enabled bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.enabled = enabled
}

// Service returns the formatting service
func (h *FormatHandler) Service() *services.FormattingService {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.service
}

// GetProfileForApp returns the format profile for a given process name
func (h *FormatHandler) GetProfileForApp(processName string) string {
	h.mu.RLock()
	svc := h.service
	h.mu.RUnlock()

	if svc != nil {
		return svc.GetProfileForApp(processName)
	}
	return "disabled"
}

// IsHotkeyExcluded checks if a hotkey should be excluded for a given process
func (h *FormatHandler) IsHotkeyExcluded(processName, formatType string) bool {
	h.mu.RLock()
	svc := h.service
	h.mu.RUnlock()

	if svc != nil {
		return svc.IsHotkeyExcluded(processName, formatType)
	}
	return false
}

// GetCustomHotkey returns the custom hotkey string for a format type in a specific app
func (h *FormatHandler) GetCustomHotkey(processName, formatType string) string {
	h.mu.RLock()
	svc := h.service
	h.mu.RUnlock()

	if svc != nil {
		return svc.GetCustomHotkey(processName, formatType)
	}
	return ""
}

// MatchesCustomHotkey checks if the key combination matches any custom hotkey for an app
// Returns formatType if matched, empty string otherwise
func (h *FormatHandler) MatchesCustomHotkey(processName string, keyCode uint16, ctrl, alt, shift bool) string {
	h.mu.RLock()
	svc := h.service
	h.mu.RUnlock()

	if svc == nil {
		return ""
	}

	appConfig := svc.GetAppConfig(processName)
	if appConfig == nil || len(appConfig.CustomHotkeys) == 0 {
		return ""
	}

	// Build current key string
	currentKey := BuildHotkeyString(keyCode, ctrl, alt, shift)

	for formatType, customHotkey := range appConfig.CustomHotkeys {
		if customHotkey == currentKey {
			return formatType
		}
	}
	return ""
}

// MatchesGlobalHotkey checks if the key combination matches any global custom hotkey
// Returns formatType if matched, empty string otherwise
func (h *FormatHandler) MatchesGlobalHotkey(keyCode uint16, ctrl, alt, shift bool) string {
	h.mu.RLock()
	svc := h.service
	h.mu.RUnlock()

	if svc == nil {
		return ""
	}

	config := svc.Config()
	if config == nil || len(config.Hotkeys) == 0 {
		return ""
	}

	// Build current key string
	currentKey := BuildHotkeyString(keyCode, ctrl, alt, shift)

	for formatType, hotkeyStr := range config.Hotkeys {
		if hotkeyStr == currentKey {
			return formatType
		}
	}
	return ""
}

// BuildHotkeyString builds a hotkey string from key code and modifiers (e.g. "Ctrl+Alt+S")
func BuildHotkeyString(keyCode uint16, ctrl, alt, shift bool) string {
	var parts []string
	if ctrl {
		parts = append(parts, "Ctrl")
	}
	if alt {
		parts = append(parts, "Alt")
	}
	if shift {
		parts = append(parts, "Shift")
	}
	parts = append(parts, KeyCodeToString(keyCode))
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "+"
		}
		result += p
	}
	return result
}

// KeyCodeToString converts a virtual key code to a string representation
func KeyCodeToString(keyCode uint16) string {
	switch keyCode {
	case VK_SPACE:
		return "Space"
	case VK_RETURN:
		return "Enter"
	case VK_TAB:
		return "Tab"
	case VK_ESCAPE:
		return "Esc"
	case VK_BACK:
		return "Backspace"
	case VK_OEM_3:
		return "`"
	case VK_OEM_1:
		return ";"
	case VK_OEM_2:
		return "/"
	case VK_OEM_4:
		return "["
	case VK_OEM_5:
		return "\\"
	case VK_OEM_6:
		return "]"
	case VK_OEM_7:
		return "'"
	case VK_OEM_PLUS:
		return "="
	case VK_OEM_COMMA:
		return ","
	case VK_OEM_MINUS:
		return "-"
	case VK_OEM_PERIOD:
		return "."
	default:
		// Letters A-Z
		if keyCode >= VK_A && keyCode <= VK_Z {
			return string(rune('A' + keyCode - VK_A))
		}
		// Numbers 0-9
		if keyCode >= VK_0 && keyCode <= VK_9 {
			return string(rune('0' + keyCode - VK_0))
		}
		return ""
	}
}

// HandleFormatHotkey processes a format hotkey with clipboard flow
// Flow: save clipboard → release modifiers → Ctrl+C → poll clipboard → format → set clipboard → Ctrl+V → restore
func (h *FormatHandler) HandleFormatHotkey(formatType, profile string) {
	// Serialize all formatting operations to prevent race conditions
	formatOpMu.Lock()
	defer formatOpMu.Unlock()

	h.mu.RLock()
	if !h.enabled || h.service == nil {
		h.mu.RUnlock()
		return
	}
	svc := h.service
	h.mu.RUnlock()

	// Step 1: Get clipboard sequence number before copy
	seqBefore := GetClipboardSequenceNumber()

	// Step 2: Save original clipboard content (with retry)
	originalClipboard, _ := GetClipboardTextRetry()
	wasEmpty := originalClipboard == ""

	// Step 3: Release all modifier keys (including L/R variants)
	ReleaseAllModifiers()

	// Step 4: Simulate Ctrl+C to copy selected text
	SimulateCtrlC()

	// Step 5: Wait for clipboard to update using sequence polling (more reliable than fixed delay)
	clipboardChanged := WaitClipboardChange(seqBefore)
	if !clipboardChanged {
		// Fallback: use fixed delay
		time.Sleep(ClipboardCopyDelay * time.Millisecond)
	}

	// Step 6: Get selected text from clipboard (with retry)
	selectedText, err := GetClipboardTextRetry()
	if err != nil {
		restoreClipboard(originalClipboard, wasEmpty)
		return
	}

	// Step 7: Check if selection is valid (not empty and different from original)
	if selectedText == "" || selectedText == originalClipboard {
		restoreClipboard(originalClipboard, wasEmpty)
		return
	}

	// Step 8: Apply formatting
	formattedText := svc.Format(formatType, selectedText, profile)

	// Step 9: Set formatted text to clipboard (with retry)
	if err := SetClipboardTextRetry(formattedText); err != nil {
		restoreClipboard(originalClipboard, wasEmpty)
		return
	}

	// Step 10: Simulate Ctrl+V to paste
	SimulateCtrlV()

	// Step 11: Wait and restore original clipboard synchronously
	time.Sleep(ClipboardRestoreWait * time.Millisecond)
	restoreClipboard(originalClipboard, wasEmpty)
}

// restoreClipboard restores the original clipboard content or clears it
func restoreClipboard(original string, wasEmpty bool) {
	if wasEmpty {
		ClearClipboard()
	} else {
		SetClipboardTextRetry(original)
	}
}

// ReleaseAllModifiers releases Ctrl, Alt, Shift keys (including L/R variants) if they are down
func ReleaseAllModifiers() {
	inputs := []INPUT{}
	
	// Release all modifier variants
	keysToCheck := []uint16{
		VK_CONTROL, VK_LCONTROL, VK_RCONTROL,
		VK_SHIFT, VK_LSHIFT, VK_RSHIFT,
		VK_MENU, VK_LMENU, VK_RMENU,
	}
	
	for _, vk := range keysToCheck {
		if isKeyDown(int(vk)) {
			inputs = append(inputs, INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         vk,
					DwFlags:     KEYEVENTF_KEYUP,
					DwExtraInfo: InjectedKeyMarker,
				},
			})
		}
	}
	
	if len(inputs) > 0 {
		procSendInput.Call(
			uintptr(len(inputs)),
			uintptr(unsafe.Pointer(&inputs[0])),
			uintptr(inputSize),
		)
		time.Sleep(10 * time.Millisecond)
	}
}

// SimulateCtrlC sends Ctrl+C using SendInput
// Uses 4 events: Ctrl↓, C↓, C↑, Ctrl↑
func SimulateCtrlC() {
	const VK_C = 0x43

	inputs := [4]INPUT{
		// Ctrl down
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_CONTROL,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// C down
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_C,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// C up
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_C,
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
	}

	procSendInput.Call(
		4,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}

// SimulateCtrlV sends Ctrl+V using SendInput
// Uses 4 events: Ctrl↓, V↓, V↑, Ctrl↑
func SimulateCtrlV() {
	const VK_V = 0x56

	inputs := [4]INPUT{
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
				WVk:         VK_V,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// V up
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_V,
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
	}

	procSendInput.Call(
		4,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}
