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

// Apps requiring slow injection (Electron, terminals, browsers).
// Slow mode adds small delays between backspaces and text.
var slowApps = map[string]bool{
	// Electron apps
	"claude":   true,
	"notion":   true,
	"slack":    true,
	"teams":    true,
	"code":     true,
	"vscode":   true,
	"cursor":   true,
	"obsidian": true,
	"figma":    true,
	// Terminals
	"windowsterminal": true,
	"cmd":             true,
	"powershell":      true,
	"pwsh":            true,
	"wezterm":         true,
	"alacritty":       true,
	"hyper":           true,
	"mintty":          true,
	// Browsers (use slow mode as safe default)
	"chrome":  true,
	"msedge":  true,
	"firefox": true,
	"brave":   true,
	"opera":   true,
	"vivaldi": true,
	"arc":     true,
}

// Apps requiring extra slow injection (problematic apps that drop chars with normal slow mode)
var extraSlowApps = map[string]bool{
	"discord":       true,
	"discordcanary": true,
	"discordptb":    true,
	"wave":          true,
	"waveterm":      true,
}

// DefaultCoalescingApps - apps that benefit from coalescing (heavy rich-text editors)
var DefaultCoalescingApps = []string{
	"discord",
	"discordcanary",
	"discordptb",
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
	name := strings.ToLower(processName)
	if extraSlowApps[name] {
		return MethodExtraSlow
	}
	if slowApps[name] {
		return MethodSlow
	}
	return MethodFast
}

// InvalidateCache forces refresh of cached process info
func InvalidateCache() {
	cacheMu.Lock()
	cachedProcessName = ""
	cachedWindow = 0
	cacheMu.Unlock()
}

// GetCurrentProcessName returns cached process name (for debugging)
func GetCurrentProcessName() string {
	cacheMu.RLock()
	defer cacheMu.RUnlock()
	return cachedProcessName
}
