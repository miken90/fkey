package core

// Keyboard hook implementation using Windows low-level keyboard hook
// Port of KeyboardHook.cs from .NET implementation

import (
	"log"
	"sync"
	"syscall"
	"unsafe"
)

const (
	WH_KEYBOARD_LL = 13
	WM_KEYDOWN     = 0x0100
	WM_KEYUP       = 0x0101
	WM_SYSKEYDOWN  = 0x0104
	WM_SYSKEYUP    = 0x0105
	LLKHF_INJECTED = 0x10
)

// Virtual key codes (Windows)
const (
	VK_BACK    = 0x08
	VK_TAB     = 0x09
	VK_RETURN  = 0x0D
	VK_SHIFT   = 0x10
	VK_CONTROL = 0x11
	VK_MENU    = 0x12 // Alt
	VK_CAPITAL = 0x14 // CapsLock
	VK_ESCAPE  = 0x1B
	VK_SPACE   = 0x20
	VK_A       = 0x41
	VK_Z       = 0x5A
	VK_0       = 0x30
	VK_9       = 0x39

	// OEM keys
	VK_OEM_1      = 0xBA // ;:
	VK_OEM_2      = 0xBF // /?
	VK_OEM_3      = 0xC0 // `~
	VK_OEM_4      = 0xDB // [{
	VK_OEM_5      = 0xDC // \|
	VK_OEM_6      = 0xDD // ]}
	VK_OEM_7      = 0xDE // '"
	VK_OEM_PLUS   = 0xBB // =+
	VK_OEM_COMMA  = 0xBC // ,<
	VK_OEM_MINUS  = 0xBD // -_
	VK_OEM_PERIOD = 0xBE // .>
)

// KBDLLHOOKSTRUCT matches Windows structure
type KBDLLHOOKSTRUCT struct {
	VkCode      uint32
	ScanCode    uint32
	Flags       uint32
	Time        uint32
	DwExtraInfo uintptr
}

// Win32 API
var (
	user32                  = syscall.NewLazyDLL("user32.dll")
	kernel32                = syscall.NewLazyDLL("kernel32.dll")
	procSetWindowsHookEx    = user32.NewProc("SetWindowsHookExW")
	procUnhookWindowsHookEx = user32.NewProc("UnhookWindowsHookEx")
	procCallNextHookEx      = user32.NewProc("CallNextHookEx")
	procGetModuleHandle     = kernel32.NewProc("GetModuleHandleW")
	procGetKeyState         = user32.NewProc("GetKeyState")
	procGetAsyncKeyState    = user32.NewProc("GetAsyncKeyState")
	procMessageBeep         = user32.NewProc("MessageBeep")
)

// InjectedKeyMarker identifies keys we injected (to skip processing)
// "FKEY" in hex: 0x464B4559
var InjectedKeyMarker = uintptr(0x464B4559)

// KeyboardHook manages low-level keyboard interception
type KeyboardHook struct {
	hookID       uintptr
	hookProc     uintptr // prevent GC
	isProcessing bool
	mu           sync.Mutex

	// Callbacks
	OnKeyPressed func(keyCode uint16, shift, capsLock bool) bool // returns true if handled
	OnHotkey     func()

	// Hotkey configuration
	Hotkey        *KeyboardShortcut
	HotkeyEnabled bool
}

// KeyboardShortcut represents a keyboard shortcut
type KeyboardShortcut struct {
	KeyCode      uint16
	Ctrl         bool
	Alt          bool
	Shift        bool
	ModifierOnly bool // If true, trigger on modifier release (e.g., Ctrl+Shift)
}

// Matches checks if the shortcut matches current key state
func (ks *KeyboardShortcut) Matches(keyCode uint16, ctrl, alt, shift bool) bool {
	// For modifier-only shortcuts (e.g., Ctrl+Shift), we match when:
	// - KeyCode is 0 (or a modifier VK like VK_SHIFT)
	// - The required modifiers are pressed
	if ks.ModifierOnly {
		// Modifier-only: check modifiers match, ignore keyCode
		return ks.Ctrl == ctrl && ks.Alt == alt && ks.Shift == shift
	}
	
	return ks.KeyCode == keyCode &&
		ks.Ctrl == ctrl &&
		ks.Alt == alt &&
		ks.Shift == shift
}

// NewKeyboardHook creates a new keyboard hook
func NewKeyboardHook() *KeyboardHook {
	return &KeyboardHook{
		HotkeyEnabled: true,
	}
}

// Start begins keyboard interception
func (h *KeyboardHook) Start() error {
	if h.hookID != 0 {
		return nil // Already started
	}

	// Create callback
	h.hookProc = syscall.NewCallback(h.hookCallback)

	// Get module handle
	hMod, _, _ := procGetModuleHandle.Call(0)

	// Install hook
	hookID, _, err := procSetWindowsHookEx.Call(
		WH_KEYBOARD_LL,
		h.hookProc,
		hMod,
		0,
	)

	if hookID == 0 {
		return err
	}

	h.hookID = hookID
	return nil
}

// Stop ends keyboard interception
func (h *KeyboardHook) Stop() {
	if h.hookID != 0 {
		procUnhookWindowsHookEx.Call(h.hookID)
		h.hookID = 0
	}
}

// hookCallback is the low-level keyboard procedure
func (h *KeyboardHook) hookCallback(nCode int, wParam uintptr, lParam uintptr) uintptr {
	// Don't process if already processing (prevents recursion)
	if h.isProcessing {
		ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
		return ret
	}

	// Only process key down events
	if nCode >= 0 && (wParam == WM_KEYDOWN || wParam == WM_SYSKEYDOWN) {
		hookStruct := (*KBDLLHOOKSTRUCT)(unsafe.Pointer(lParam))

		// Skip our own injected keys
		if hookStruct.DwExtraInfo == InjectedKeyMarker {
			ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
			return ret
		}

		// Skip injected keys from other sources
		if (hookStruct.Flags & LLKHF_INJECTED) != 0 {
			ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
			return ret
		}

		keyCode := uint16(hookStruct.VkCode)

		// Get modifier states
		// Note: The key currently being pressed may not be reflected in GetAsyncKeyState yet
		// So we need to account for it based on keyCode
		shift := isKeyDown(VK_SHIFT)
		ctrl := isKeyDown(VK_CONTROL)
		alt := isKeyDown(VK_MENU)
		capsLock := isCapsLockOn()
		
		// If the current keyCode IS a modifier, ensure it's counted as pressed
		// GetAsyncKeyState may not have updated yet for the key being pressed
		isShiftKey := keyCode == VK_SHIFT || keyCode == VK_LSHIFT || keyCode == VK_RSHIFT
		isCtrlKey := keyCode == VK_CONTROL || keyCode == VK_LCONTROL || keyCode == VK_RCONTROL
		isAltKey := keyCode == VK_MENU || keyCode == VK_LMENU || keyCode == VK_RMENU
		if isShiftKey {
			shift = true
		}
		if isCtrlKey {
			ctrl = true
		}
		if isAltKey {
			alt = true
		}

		// Check format hotkeys BEFORE toggle hotkey
		// Need at least Ctrl or Alt modifier for format hotkeys
		if ctrl || alt {
			handler := GetFormatHandler()
			if handler != nil && handler.IsEnabled() {
				// Force fresh detection instead of using cached value
				processName := DetectForegroundApp()
				if processName == "" {
					processName = GetCurrentProcessName()
				}

				// Step 1: Check CUSTOM hotkeys first (app-specific overrides)
				customFormatType := handler.MatchesCustomHotkey(processName, keyCode, ctrl, alt, shift)
				if customFormatType != "" {
					profile := handler.GetProfileForApp(processName)
					log.Printf("[FormatHotkey] CUSTOM key=0x%X formatType=%s process=%s profile=%s",
						keyCode, customFormatType, processName, profile)
					if profile != "disabled" {
						go handler.HandleFormatHotkey(customFormatType, profile)
						return 1 // Block key
					}
				}

				// Step 2: Check GLOBAL custom hotkeys (user-defined replacements for defaults)
				globalFormatType := handler.MatchesGlobalHotkey(keyCode, ctrl, alt, shift)
				if globalFormatType != "" {
					// Check if this hotkey is excluded for this app
					if handler.IsHotkeyExcluded(processName, globalFormatType) {
						log.Printf("[FormatHotkey] EXCLUDED key=0x%X formatType=%s process=%s", keyCode, globalFormatType, processName)
						ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
						return ret
					}

					profile := handler.GetProfileForApp(processName)
					log.Printf("[FormatHotkey] GLOBAL key=0x%X formatType=%s process=%s profile=%s",
						keyCode, globalFormatType, processName, profile)
					if profile != "disabled" {
						go handler.HandleFormatHotkey(globalFormatType, profile)
						return 1 // Block key
					}
				}

				// Step 3: Check DEFAULT hotkeys (Ctrl+B, Ctrl+I, Ctrl+Alt+S, etc.)
				if ctrl {
					if formatType, matched := IsFormatHotkey(keyCode, ctrl, alt, shift); matched {
						// Check if this formatType has a global custom hotkey override
						globalHotkey := handler.Service().GetGlobalHotkey(formatType)
						if globalHotkey != "" {
							// This formatType uses a custom global hotkey, skip default
							ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
							return ret
						}

						// Check if default hotkey has been overridden by a per-app custom one
						customHotkey := handler.GetCustomHotkey(processName, formatType)
						if customHotkey != "" {
							// This formatType has a custom hotkey, skip default handling
							// Let the key pass through
							ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
							return ret
						}

						// Check if this hotkey is excluded for this app
						if handler.IsHotkeyExcluded(processName, formatType) {
							log.Printf("[FormatHotkey] EXCLUDED key=0x%X formatType=%s process=%s", keyCode, formatType, processName)
							// Don't return 1, let the key pass through to native app
							ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
							return ret
						}

						profile := handler.GetProfileForApp(processName)
						log.Printf("[FormatHotkey] key=0x%X shift=%v formatType=%s process=%s profile=%s",
							keyCode, shift, formatType, processName, profile)
						if profile != "disabled" {
							go handler.HandleFormatHotkey(formatType, profile)
							return 1 // Block key
						}
					}
				}
				}
				}

		// Check for toggle hotkey
		// For modifier-only shortcuts (like Ctrl+Shift), trigger when the last modifier is pressed
		if h.HotkeyEnabled && h.Hotkey != nil {
			if h.Hotkey.ModifierOnly {
				// Modifier-only: trigger when pressing a modifier key while other required modifiers are held
				// E.g., Ctrl+Shift triggers when pressing Shift while Ctrl is held (or vice versa)
				if isShiftKey || isCtrlKey || isAltKey {
					if h.Hotkey.Matches(keyCode, ctrl, alt, shift) {
						if h.OnHotkey != nil {
							h.OnHotkey()
						}
						return 1 // Consume the key
					}
				}
			} else if h.Hotkey.Matches(keyCode, ctrl, alt, shift) {
				if h.OnHotkey != nil {
					h.OnHotkey()
				}
				return 1 // Consume the key
			}
		}

		// Only process relevant keys for Vietnamese input
		if IsRelevantKey(keyCode) {
			// Skip if Ctrl or Alt is pressed (shortcuts)
			if ctrl || alt {
				// Clear buffer on Ctrl+key combinations
				if ctrl {
					bridge, _ := GetBridge()
					if bridge != nil {
						bridge.Clear()
					}
				}
				ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
				return ret
			}

			// Handle buffer-clearing keys (TAB only - Space/Enter go through IME)
			if keyCode == VK_TAB {
				bridge, _ := GetBridge()
				if bridge != nil {
					bridge.Clear()
				}
				ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
				return ret
			}

			// Process the key through IME callback
			if h.OnKeyPressed != nil {
				h.mu.Lock()
				h.isProcessing = true
				handled := h.OnKeyPressed(keyCode, shift, capsLock)
				h.isProcessing = false
				h.mu.Unlock()

				if handled {
					return 1 // Block the original key
				}
			}
		}
	}

	ret, _, _ := procCallNextHookEx.Call(h.hookID, uintptr(nCode), wParam, lParam)
	return ret
}

// isKeyDown checks if a key is currently pressed
func isKeyDown(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return (ret & 0x8000) != 0
}

// isCapsLockOn checks if CapsLock is toggled on
func isCapsLockOn() bool {
	ret, _, _ := procGetKeyState.Call(uintptr(VK_CAPITAL))
	return (ret & 0x0001) != 0
}

// IsLetterKey checks if virtual key is a letter (A-Z)
func IsLetterKey(vk uint16) bool {
	return vk >= VK_A && vk <= VK_Z
}

// IsNumberKey checks if virtual key is a number (0-9)
func IsNumberKey(vk uint16) bool {
	return vk >= VK_0 && vk <= VK_9
}

// IsRelevantKey checks if key should be processed by IME
func IsRelevantKey(vk uint16) bool {
	// Letters
	if IsLetterKey(vk) {
		return true
	}
	// Numbers
	if IsNumberKey(vk) {
		return true
	}
	// Special keys
	switch vk {
	case VK_BACK, VK_SPACE, VK_RETURN, VK_TAB, VK_ESCAPE,
		VK_OEM_4, VK_OEM_6, VK_OEM_PERIOD, VK_OEM_COMMA, VK_OEM_2,
		VK_OEM_1, VK_OEM_7, VK_OEM_MINUS, VK_OEM_PLUS:
		return true
	}
	return false
}

// MessageBeep sound types
const (
	MB_OK          = 0x00000000 // Default beep
	MB_ICONHAND    = 0x00000010 // Critical stop
	MB_ICONQUESTION = 0x00000020 // Question
	MB_ICONEXCLAMATION = 0x00000030 // Exclamation
	MB_ICONASTERISK = 0x00000040 // Asterisk (info)
)

// PlayBeep plays a Windows system beep sound
// soundType: true = Vietnamese on (higher pitch), false = English (lower pitch)
func PlayBeep(isVietnamese bool) {
	if isVietnamese {
		// Higher pitch beep for Vietnamese
		procMessageBeep.Call(uintptr(MB_ICONASTERISK))
	} else {
		// Lower pitch beep for English
		procMessageBeep.Call(uintptr(MB_OK))
	}
}
