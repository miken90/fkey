package tests

import (
	"testing"

	"fkey/core"
	"fkey/services"
)

// ==================== Settings Tests ====================

func TestParseHotkey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		keyCode  uint16
		ctrl     bool
		alt      bool
		shift    bool
	}{
		{"Ctrl+Space", "32,1", 32, true, false, false},
		{"Alt+A", "65,2", 65, false, true, false},
		{"Ctrl+Alt+Shift+F1", "112,7", 112, true, true, true},
		{"Shift only", "65,4", 65, false, false, true},
		{"No modifiers", "65,0", 65, false, false, false},
		{"Invalid format", "invalid", 0, false, false, false},
		{"Empty string", "", 0, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keyCode, ctrl, alt, shift := services.ParseHotkey(tt.input)
			if keyCode != tt.keyCode {
				t.Errorf("keyCode = %d, want %d", keyCode, tt.keyCode)
			}
			if ctrl != tt.ctrl {
				t.Errorf("ctrl = %v, want %v", ctrl, tt.ctrl)
			}
			if alt != tt.alt {
				t.Errorf("alt = %v, want %v", alt, tt.alt)
			}
			if shift != tt.shift {
				t.Errorf("shift = %v, want %v", shift, tt.shift)
			}
		})
	}
}

func TestFormatHotkey(t *testing.T) {
	tests := []struct {
		name     string
		keyCode  uint16
		ctrl     bool
		alt      bool
		shift    bool
		expected string
	}{
		{"Ctrl+Space", 32, true, false, false, "32,1"},
		{"Alt+A", 65, false, true, false, "65,2"},
		{"Ctrl+Alt+Shift+F1", 112, true, true, true, "112,7"},
		{"No modifiers", 65, false, false, false, "65,0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := services.FormatHotkey(tt.keyCode, tt.ctrl, tt.alt, tt.shift)
			if result != tt.expected {
				t.Errorf("FormatHotkey() = %s, want %s", result, tt.expected)
			}
		})
	}
}

func TestHotkeyRoundTrip(t *testing.T) {
	testCases := []struct {
		keyCode uint16
		ctrl    bool
		alt     bool
		shift   bool
	}{
		{32, true, false, false},
		{65, false, true, false},
		{112, true, true, true},
		{27, false, false, true},
	}

	for _, tc := range testCases {
		formatted := services.FormatHotkey(tc.keyCode, tc.ctrl, tc.alt, tc.shift)
		keyCode, ctrl, alt, shift := services.ParseHotkey(formatted)

		if keyCode != tc.keyCode || ctrl != tc.ctrl || alt != tc.alt || shift != tc.shift {
			t.Errorf("Round trip failed for %+v: got keyCode=%d, ctrl=%v, alt=%v, shift=%v",
				tc, keyCode, ctrl, alt, shift)
		}
	}
}

func TestDefaultSettings(t *testing.T) {
	s := services.DefaultSettings()

	if s.InputMethod != 0 {
		t.Errorf("InputMethod = %d, want 0 (Telex)", s.InputMethod)
	}
	if !s.ModernTone {
		t.Error("ModernTone should be true by default")
	}
	if !s.Enabled {
		t.Error("Enabled should be true by default")
	}
	if !s.FirstRun {
		t.Error("FirstRun should be true by default")
	}
	if s.AutoStart {
		t.Error("AutoStart should be false by default")
	}
	if s.SkipWShortcut {
		t.Error("SkipWShortcut should be false by default")
	}
	if !s.EscRestore {
		t.Error("EscRestore should be true by default")
	}
	if s.FreeTone {
		t.Error("FreeTone should be false by default")
	}
	if s.EnglishAutoRestore {
		t.Error("EnglishAutoRestore should be false by default")
	}
	if !s.AutoCapitalize {
		t.Error("AutoCapitalize should be true by default")
	}
	if s.ToggleHotkey != "32,1" {
		t.Errorf("ToggleHotkey = %s, want 32,1", s.ToggleHotkey)
	}
}

func TestNewSettingsService(t *testing.T) {
	svc := services.NewSettingsService()
	if svc == nil {
		t.Fatal("NewSettingsService() returned nil")
	}

	settings := svc.Settings()
	if settings == nil {
		t.Fatal("Settings() returned nil")
	}

	if settings.InputMethod != 0 {
		t.Errorf("Expected default InputMethod 0, got %d", settings.InputMethod)
	}
}

// ==================== Updater Tests ====================

func TestNewUpdaterService(t *testing.T) {
	svc := services.NewUpdaterService("2.0.0")
	if svc == nil {
		t.Fatal("NewUpdaterService() returned nil")
	}

	if svc.GetCurrentVersion() != "2.0.0" {
		t.Errorf("GetCurrentVersion() = %s, want 2.0.0", svc.GetCurrentVersion())
	}
}

func TestIsNewerVersion(t *testing.T) {
	tests := []struct {
		name    string
		current string
		latest  string
		want    bool
	}{
		{"major upgrade", "1.0.0", "2.0.0", true},
		{"minor upgrade", "1.0.0", "1.1.0", true},
		{"patch upgrade", "1.0.0", "1.0.1", true},
		{"same version", "1.0.0", "1.0.0", false},
		{"current newer major", "2.0.0", "1.0.0", false},
		{"current newer minor", "1.1.0", "1.0.0", false},
		{"current newer patch", "1.0.1", "1.0.0", false},
		{"with v prefix", "v1.0.0", "v2.0.0", true},
		{"mixed v prefix", "1.0.0", "v2.0.0", true},
		{"with suffix -wails", "1.0.0-wails", "2.0.0", true},
		{"with suffix -beta", "2.0.0-beta", "2.0.0", false},
		{"both with suffix", "1.0.0-wails", "2.0.0-wails", true},
	}

	svc := services.NewUpdaterService("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.IsNewerVersion(tt.current, tt.latest)
			if got != tt.want {
				t.Errorf("IsNewerVersion(%q, %q) = %v, want %v",
					tt.current, tt.latest, got, tt.want)
			}
		})
	}
}

func TestIsWindowsAsset(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"fkey-windows-amd64.zip", true},
		{"fkey-win64.zip", true},
		{"fkey-win32.zip", true},
		{"FKey.exe", true},
		{"fkey-portable.zip", true},
		{"fkey-darwin-arm64.zip", false},
		{"fkey-macos.zip", false},
		{"fkey-mac.zip", false},
		{"fkey-linux-amd64.tar.gz", false},
		{"fkey-linux.deb", false},
	}

	svc := services.NewUpdaterService("test")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.IsWindowsAsset(tt.name)
			if got != tt.want {
				t.Errorf("IsWindowsAsset(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// ==================== Bridge/Keycode Tests ====================

func TestTranslateToMacKeycode_Letters(t *testing.T) {
	tests := []struct {
		windowsVK uint16
		macKey    uint16
		name      string
	}{
		{0x41, core.MAC_A, "A"},
		{0x42, core.MAC_B, "B"},
		{0x43, core.MAC_C, "C"},
		{0x44, core.MAC_D, "D"},
		{0x45, core.MAC_E, "E"},
		{0x46, core.MAC_F, "F"},
		{0x47, core.MAC_G, "G"},
		{0x48, core.MAC_H, "H"},
		{0x49, core.MAC_I, "I"},
		{0x4A, core.MAC_J, "J"},
		{0x4B, core.MAC_K, "K"},
		{0x4C, core.MAC_L, "L"},
		{0x4D, core.MAC_M, "M"},
		{0x4E, core.MAC_N, "N"},
		{0x4F, core.MAC_O, "O"},
		{0x50, core.MAC_P, "P"},
		{0x51, core.MAC_Q, "Q"},
		{0x52, core.MAC_R, "R"},
		{0x53, core.MAC_S, "S"},
		{0x54, core.MAC_T, "T"},
		{0x55, core.MAC_U, "U"},
		{0x56, core.MAC_V, "V"},
		{0x57, core.MAC_W, "W"},
		{0x58, core.MAC_X, "X"},
		{0x59, core.MAC_Y, "Y"},
		{0x5A, core.MAC_Z, "Z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.TranslateToMacKeycode(tt.windowsVK)
			if result != tt.macKey {
				t.Errorf("TranslateToMacKeycode(0x%02X) = 0x%02X, want 0x%02X",
					tt.windowsVK, result, tt.macKey)
			}
		})
	}
}

func TestTranslateToMacKeycode_Numbers(t *testing.T) {
	tests := []struct {
		windowsVK uint16
		macKey    uint16
		name      string
	}{
		{0x30, core.MAC_N0, "0"},
		{0x31, core.MAC_N1, "1"},
		{0x32, core.MAC_N2, "2"},
		{0x33, core.MAC_N3, "3"},
		{0x34, core.MAC_N4, "4"},
		{0x35, core.MAC_N5, "5"},
		{0x36, core.MAC_N6, "6"},
		{0x37, core.MAC_N7, "7"},
		{0x38, core.MAC_N8, "8"},
		{0x39, core.MAC_N9, "9"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.TranslateToMacKeycode(tt.windowsVK)
			if result != tt.macKey {
				t.Errorf("TranslateToMacKeycode(0x%02X) = 0x%02X, want 0x%02X",
					tt.windowsVK, result, tt.macKey)
			}
		})
	}
}

func TestTranslateToMacKeycode_SpecialKeys(t *testing.T) {
	tests := []struct {
		windowsVK uint16
		macKey    uint16
		name      string
	}{
		{0x08, core.MAC_DELETE, "Backspace"},
		{0x09, core.MAC_TAB, "Tab"},
		{0x0D, core.MAC_RETURN, "Enter"},
		{0x1B, core.MAC_ESC, "Escape"},
		{0x20, core.MAC_SPACE, "Space"},
		{0xDB, core.MAC_LBRACKET, "LeftBracket"},
		{0xDD, core.MAC_RBRACKET, "RightBracket"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.TranslateToMacKeycode(tt.windowsVK)
			if result != tt.macKey {
				t.Errorf("TranslateToMacKeycode(0x%02X) = 0x%02X, want 0x%02X",
					tt.windowsVK, result, tt.macKey)
			}
		})
	}
}

func TestTranslateToMacKeycode_Punctuation(t *testing.T) {
	tests := []struct {
		windowsVK uint16
		macKey    uint16
		name      string
	}{
		{0xBE, core.MAC_DOT, "Period"},
		{0xBC, core.MAC_COMMA, "Comma"},
		{0xBF, core.MAC_SLASH, "Slash"},
		{0xBA, core.MAC_SEMICOLON, "Semicolon"},
		{0xDE, core.MAC_QUOTE, "Quote"},
		{0xBD, core.MAC_MINUS, "Minus"},
		{0xBB, core.MAC_EQUAL, "Equal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.TranslateToMacKeycode(tt.windowsVK)
			if result != tt.macKey {
				t.Errorf("TranslateToMacKeycode(0x%02X) = 0x%02X, want 0x%02X",
					tt.windowsVK, result, tt.macKey)
			}
		})
	}
}

func TestTranslateToMacKeycode_UnmappedKeys(t *testing.T) {
	unmappedKeys := []uint16{
		0x00, // Null
		0x10, // Shift
		0x11, // Ctrl
		0x12, // Alt
		0x70, // F1
		0x7F, // DEL
		0xFF, // Invalid
	}

	for _, vk := range unmappedKeys {
		result := core.TranslateToMacKeycode(vk)
		if result != 0xFFFF {
			t.Errorf("TranslateToMacKeycode(0x%02X) = 0x%02X, want 0xFFFF (unmapped)",
				vk, result)
		}
	}
}

func TestImeResult_GetText(t *testing.T) {
	tests := []struct {
		name     string
		chars    []rune
		expected string
	}{
		{"empty", []rune{}, ""},
		{"single char", []rune{'a'}, "a"},
		{"vietnamese", []rune{'v', 'i', 'ệ', 't'}, "việt"},
		{"unicode", []rune{'ă', 'â', 'ê', 'ô', 'ơ', 'ư'}, "ăâêôơư"},
		{"tones", []rune{'á', 'à', 'ả', 'ã', 'ạ'}, "áàảãạ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := core.ImeResult{Chars: tt.chars}
			if result.GetText() != tt.expected {
				t.Errorf("GetText() = %s, want %s", result.GetText(), tt.expected)
			}
		})
	}
}

func TestInputMethodConstants(t *testing.T) {
	if core.Telex != 0 {
		t.Errorf("Telex = %d, want 0", core.Telex)
	}
	if core.VNI != 1 {
		t.Errorf("VNI = %d, want 1", core.VNI)
	}
}

func TestImeActionConstants(t *testing.T) {
	if core.ActionNone != 0 {
		t.Errorf("ActionNone = %d, want 0", core.ActionNone)
	}
	if core.ActionSend != 1 {
		t.Errorf("ActionSend = %d, want 1", core.ActionSend)
	}
	if core.ActionRestore != 2 {
		t.Errorf("ActionRestore = %d, want 2", core.ActionRestore)
	}
}

// ==================== IME Loop Tests ====================

func TestDefaultImeSettings(t *testing.T) {
	settings := core.DefaultImeSettings()

	if settings == nil {
		t.Fatal("DefaultImeSettings() returned nil")
	}

	if !settings.Enabled {
		t.Error("Enabled should be true by default")
	}
	if settings.InputMethod != core.Telex {
		t.Errorf("InputMethod = %v, want Telex", settings.InputMethod)
	}
	if !settings.ModernTone {
		t.Error("ModernTone should be true by default")
	}
	if settings.SkipWShortcut {
		t.Error("SkipWShortcut should be false by default")
	}
	if settings.BracketShortcut {
		t.Error("BracketShortcut should be false by default")
	}
	if !settings.EscRestore {
		t.Error("EscRestore should be true by default")
	}
	if settings.FreeTone {
		t.Error("FreeTone should be false by default")
	}
	if settings.EnglishAutoRestore {
		t.Error("EnglishAutoRestore should be false by default")
	}
	if !settings.AutoCapitalize {
		t.Error("AutoCapitalize should be true by default")
	}
}

// ==================== Keyboard Hook Helper Tests ====================

func TestIsLetterKey(t *testing.T) {
	tests := []struct {
		vk   uint16
		want bool
	}{
		{0x41, true},  // A
		{0x5A, true},  // Z
		{0x4D, true},  // M
		{0x30, false}, // 0
		{0x39, false}, // 9
		{0x20, false}, // Space
		{0x08, false}, // Backspace
	}

	for _, tt := range tests {
		got := core.IsLetterKey(tt.vk)
		if got != tt.want {
			t.Errorf("IsLetterKey(0x%02X) = %v, want %v", tt.vk, got, tt.want)
		}
	}
}

func TestIsNumberKey(t *testing.T) {
	tests := []struct {
		vk   uint16
		want bool
	}{
		{0x30, true},  // 0
		{0x39, true},  // 9
		{0x35, true},  // 5
		{0x41, false}, // A
		{0x5A, false}, // Z
		{0x20, false}, // Space
	}

	for _, tt := range tests {
		got := core.IsNumberKey(tt.vk)
		if got != tt.want {
			t.Errorf("IsNumberKey(0x%02X) = %v, want %v", tt.vk, got, tt.want)
		}
	}
}

func TestIsRelevantKey(t *testing.T) {
	tests := []struct {
		vk   uint16
		want bool
		name string
	}{
		{0x41, true, "A"},
		{0x5A, true, "Z"},
		{0x30, true, "0"},
		{0x39, true, "9"},
		{0x08, true, "Backspace"},
		{0x20, true, "Space"},
		{0x0D, true, "Enter"},
		{0x09, true, "Tab"},
		{0x1B, true, "Escape"},
		{0xDB, true, "LeftBracket"},
		{0xDD, true, "RightBracket"},
		{0xBE, true, "Period"},
		{0xBC, true, "Comma"},
		{0x10, false, "Shift"},
		{0x11, false, "Ctrl"},
		{0x12, false, "Alt"},
		{0x70, false, "F1"},
		{0x7B, false, "F12"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := core.IsRelevantKey(tt.vk)
			if got != tt.want {
				t.Errorf("IsRelevantKey(0x%02X) = %v, want %v", tt.vk, got, tt.want)
			}
		})
	}
}

// ==================== Keyboard Shortcut Tests ====================

func TestKeyboardShortcutMatches(t *testing.T) {
	tests := []struct {
		name    string
		ks      core.KeyboardShortcut
		keyCode uint16
		ctrl    bool
		alt     bool
		shift   bool
		want    bool
	}{
		{
			name:    "Ctrl+Space matches",
			ks:      core.KeyboardShortcut{KeyCode: 0x20, Ctrl: true, Alt: false, Shift: false},
			keyCode: 0x20, ctrl: true, alt: false, shift: false,
			want: true,
		},
		{
			name:    "Ctrl+Space not matched by Alt+Space",
			ks:      core.KeyboardShortcut{KeyCode: 0x20, Ctrl: true, Alt: false, Shift: false},
			keyCode: 0x20, ctrl: false, alt: true, shift: false,
			want: false,
		},
		{
			name:    "Ctrl+Alt+Delete matches",
			ks:      core.KeyboardShortcut{KeyCode: 0x2E, Ctrl: true, Alt: true, Shift: false},
			keyCode: 0x2E, ctrl: true, alt: true, shift: false,
			want: true,
		},
		{
			name:    "Wrong keycode fails",
			ks:      core.KeyboardShortcut{KeyCode: 0x41, Ctrl: true, Alt: false, Shift: false},
			keyCode: 0x42, ctrl: true, alt: false, shift: false,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.ks.Matches(tt.keyCode, tt.ctrl, tt.alt, tt.shift)
			if got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ==================== App Detector Tests ====================

func TestExtractProcessName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`C:\Windows\System32\notepad.exe`, "notepad"},
		{`C:\Program Files\Code\Code.exe`, "code"},
		{`C:\Users\Test\AppData\Local\Slack\slack.exe`, "slack"},
		{`cmd.exe`, "cmd"},
		{`notepad`, "notepad"},
		{`C:\Apps\MyApp.EXE`, "myapp"},
	}

	for _, tt := range tests {
		result := core.ExtractProcessName(tt.input)
		if result != tt.expected {
			t.Errorf("ExtractProcessName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDetermineMethod(t *testing.T) {
	tests := []struct {
		processName string
		want        core.InjectionMethod
	}{
		// Slow apps (Electron, terminals, browsers)
		{"code", core.MethodSlow},
		{"vscode", core.MethodSlow},
		{"cursor", core.MethodSlow},
		{"slack", core.MethodSlow},
		{"discord", core.MethodSlow},
		{"notion", core.MethodSlow},
		{"chrome", core.MethodSlow},
		{"msedge", core.MethodSlow},
		{"firefox", core.MethodSlow},
		{"windowsterminal", core.MethodSlow},
		{"powershell", core.MethodSlow},
		{"wave", core.MethodSlow},
		{"waveterm", core.MethodSlow},
		{"claude", core.MethodSlow},
		// Fast apps
		{"notepad", core.MethodFast},
		{"winword", core.MethodFast},
		{"excel", core.MethodFast},
		{"explorer", core.MethodFast},
		{"unknown", core.MethodFast},
	}

	for _, tt := range tests {
		t.Run(tt.processName, func(t *testing.T) {
			got := core.DetermineMethod(tt.processName)
			if got != tt.want {
				t.Errorf("DetermineMethod(%q) = %v, want %v", tt.processName, got, tt.want)
			}
		})
	}
}

// ==================== Text Sender Constants Tests ====================

func TestInjectionMethodConstants(t *testing.T) {
	if core.MethodFast != 0 {
		t.Errorf("MethodFast = %d, want 0", core.MethodFast)
	}
	if core.MethodSlow != 1 {
		t.Errorf("MethodSlow = %d, want 1", core.MethodSlow)
	}
}

func TestTextSenderDelays(t *testing.T) {
	if core.SlowModeKeyDelay != 5 {
		t.Errorf("SlowModeKeyDelay = %d, want 5", core.SlowModeKeyDelay)
	}
	if core.SlowModePreDelay != 20 {
		t.Errorf("SlowModePreDelay = %d, want 20", core.SlowModePreDelay)
	}
	if core.SlowModePostDelay != 15 {
		t.Errorf("SlowModePostDelay = %d, want 15", core.SlowModePostDelay)
	}
	if core.FastModeDelay != 5 {
		t.Errorf("FastModeDelay = %d, want 5", core.FastModeDelay)
	}
}
