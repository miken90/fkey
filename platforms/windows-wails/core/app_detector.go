package core

// AppDetector detects foreground app and determines best text injection method.
// Port of AppDetector.cs from .NET implementation.
// Caches process lookups by window handle for performance (<1ms overhead).

import (
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// Win32 API procs (user32 and kernel32 declared in keyboard_hook.go)
var (
	procGetForegroundWindow        = user32.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessId   = user32.NewProc("GetWindowThreadProcessId")
	procOpenProcess                = kernel32.NewProc("OpenProcess")
	procCloseHandle                = kernel32.NewProc("CloseHandle")
	procQueryFullProcessImageNameW = kernel32.NewProc("QueryFullProcessImageNameW")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
)

// BackspaceMode determines how backspace is sent
type BackspaceMode int

const (
	BackspaceVK      BackspaceMode = iota // VK_BACK virtual key (default, works for most apps)
	BackspaceUnicode                      // Unicode BS (0x08) via KEYEVENTF_UNICODE (for CLI apps that don't handle DEL)
)

// AppProfile defines injection behavior for an app
type AppProfile struct {
	Method        InjectionMethod
	Coalesce      bool          // Whether to use coalescing
	CoalesceMs    int           // Coalescing timer (0 = use default 25ms)
	BackspaceMode BackspaceMode // How to send backspace (default: BackspaceVK)
}

// Default profiles
var (
	ProfileFast   = AppProfile{Method: MethodFast, Coalesce: false}
	ProfileSlow   = AppProfile{Method: MethodSlow, Coalesce: false}
	ProfileAtomic = AppProfile{Method: MethodAtomic, Coalesce: false}
	// Discord profile: use slow mode like other Electron apps (atomic+coalesce caused lag)
	ProfileDiscord = AppProfile{Method: MethodSlow, Coalesce: false}
	// Terminal profile: atomic mode with VK_BACK (default)
	// Note: BackspaceUnicode caused issues with Wave, Claude Code
	ProfileTerminal = AppProfile{Method: MethodAtomic, Coalesce: false, BackspaceMode: BackspaceVK}
	// Augment CLI profile: uses Unicode BS - only for explicit auggie/augment process
	// This is a workaround; user running auggie in terminal may still need script patch
	ProfileAugment = AppProfile{Method: MethodAtomic, Coalesce: false, BackspaceMode: BackspaceUnicode}
)

// appProfiles maps process names to their injection profiles
// Add new apps here with custom settings
var appProfiles = map[string]AppProfile{
	// Discord - atomic mode with short coalescing for smooth diacritics
	"discord":       ProfileDiscord,
	"discordcanary": ProfileDiscord,
	"discordptb":    ProfileDiscord,

	// Electron apps - slow mode with delays
	"notion":   ProfileSlow,
	"slack":    ProfileSlow,
	"teams":    ProfileSlow,
	"code":     ProfileSlow,
	"vscode":   ProfileSlow,
	"cursor":   ProfileSlow,
	"obsidian": ProfileSlow,
	"figma":    ProfileSlow,

	// Claude Code CLI - keep slow mode (was working in v2.2.4)
	// Process name can be "claude" or "claude code" depending on how it's launched
	"claude":      ProfileSlow,
	"claude code": ProfileSlow,

	// Terminals - slow mode (same as v2.2.4 which was stable)
	// Note: ProfileTerminal (atomic) caused missing chars in Claude Code
	"windowsterminal": ProfileSlow,
	"cmd":             ProfileSlow,
	"powershell":      ProfileSlow,
	"pwsh":            ProfileSlow,
	"wezterm":         ProfileSlow,
	"alacritty":       ProfileSlow,
	"hyper":           ProfileSlow,
	"mintty":          ProfileSlow,
	"wave":            ProfileSlow,
	"waveterm":        ProfileSlow,

	// Augment CLI (auggie) - uses Unicode BS to fix duplicate chars issue
	// npm package: @augmentcode/auggie, command: auggie
	// Process name may vary: auggie, augment, or node (when running via npx)
	// Adding common variants
	"auggie":  ProfileAugment,
	"augment": ProfileAugment,

	// Browsers - slow mode as safe default
	"chrome":  ProfileSlow,
	"msedge":  ProfileSlow,
	"firefox": ProfileSlow,
	"brave":   ProfileSlow,
	"opera":   ProfileSlow,
	"vivaldi": ProfileSlow,
	"arc":     ProfileSlow,
}

// GetAppProfile returns the injection profile for a process name
func GetAppProfile(processName string) AppProfile {
	name := strings.ToLower(processName)
	if profile, ok := appProfiles[name]; ok {
		return profile
	}
	return ProfileFast // Default
}

// Cache to avoid repeated process lookups
var (
	cachedProcessName string
	cachedWindow      uintptr
	cacheMu           sync.RWMutex
)

// GetInjectionMethod returns the best injection method for current foreground app.
// Uses cached result if same window handle.
func GetInjectionMethod() InjectionMethod {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return MethodFast
	}

	// Check cache first
	cacheMu.RLock()
	if hwnd == cachedWindow && cachedProcessName != "" {
		method := DetermineMethod(cachedProcessName)
		cacheMu.RUnlock()
		return method
	}
	cacheMu.RUnlock()

	// Get process name
	processName := getProcessName(hwnd)
	if processName == "" {
		return MethodFast
	}

	// Update cache
	cacheMu.Lock()
	cachedProcessName = processName
	cachedWindow = hwnd
	cacheMu.Unlock()

	return DetermineMethod(processName)
}

// getProcessName gets the process name from window handle
func getProcessName(hwnd uintptr) string {
	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ""
	}

	// Open process with limited query rights
	hProcess, _, _ := procOpenProcess.Call(
		PROCESS_QUERY_LIMITED_INFORMATION,
		0,
		uintptr(pid),
	)
	if hProcess == 0 {
		return ""
	}
	defer procCloseHandle.Call(hProcess)

	// Get process image name
	var buf [260]uint16
	size := uint32(len(buf))
	ret, _, _ := procQueryFullProcessImageNameW.Call(
		hProcess,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return ""
	}

	// Convert to string and extract filename
	fullPath := syscall.UTF16ToString(buf[:size])
	return ExtractProcessName(fullPath)
}

// ExtractProcessName extracts process name from full path (without .exe) - exported for testing
func ExtractProcessName(fullPath string) string {
	// Find last backslash
	lastSlash := strings.LastIndex(fullPath, "\\")
	if lastSlash >= 0 {
		fullPath = fullPath[lastSlash+1:]
	}

	// Remove .exe extension
	if strings.HasSuffix(strings.ToLower(fullPath), ".exe") {
		fullPath = fullPath[:len(fullPath)-4]
	}

	return strings.ToLower(fullPath)
}

// DetermineMethod checks if process name needs slow mode - exported for testing
func DetermineMethod(processName string) InjectionMethod {
	return GetAppProfile(processName).Method
}

// InvalidateCache forces refresh of cached process info
func InvalidateCache() {
	cacheMu.Lock()
	cachedProcessName = ""
	cachedWindow = 0
	cacheMu.Unlock()
}

// AppChanged checks if foreground app changed since last call
// Returns true if app changed, and updates cache
func AppChanged() bool {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return false
	}

	cacheMu.RLock()
	changed := hwnd != cachedWindow
	cacheMu.RUnlock()

	if changed {
		// Update cache with new window/process
		processName := getProcessName(hwnd)
		cacheMu.Lock()
		cachedProcessName = processName
		cachedWindow = hwnd
		cacheMu.Unlock()
	}

	return changed
}

// GetCurrentProcessName returns cached process name (for debugging)
func GetCurrentProcessName() string {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return cachedProcessName
}

// DetectForegroundApp forces a fresh detection of the currently focused app
// Used by the UI detect button
func DetectForegroundApp() string {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}
	return getProcessName(hwnd)
}
