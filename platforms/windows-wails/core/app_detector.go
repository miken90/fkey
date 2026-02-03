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
	procCreateToolhelp32Snapshot   = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW            = kernel32.NewProc("Process32FirstW")
	procProcess32NextW             = kernel32.NewProc("Process32NextW")
)

const (
	PROCESS_QUERY_LIMITED_INFORMATION = 0x1000
	TH32CS_SNAPPROCESS                = 0x00000002
	INVALID_HANDLE_VALUE              = ^uintptr(0)
)

// PROCESSENTRY32W structure for process enumeration
type PROCESSENTRY32W struct {
	DwSize              uint32
	CntUsage            uint32
	Th32ProcessID       uint32
	Th32DefaultHeapID   uintptr
	Th32ModuleID        uint32
	CntThreads          uint32
	Th32ParentProcessID uint32
	PcPriClassBase      int32
	DwFlags             uint32
	SzExeFile           [260]uint16
}

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
	// Paste profile: uses clipboard + Ctrl+V for apps that don't render KEYEVENTF_UNICODE
	// Warp terminal doesn't display Unicode input but handles paste correctly
	ProfilePaste = AppProfile{Method: MethodPaste, Coalesce: false}
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

	// Warp terminal - uses paste mode because it doesn't render KEYEVENTF_UNICODE
	// Known issue: https://github.com/warpdotdev/warp/issues/6759
	"warp": ProfilePaste,
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

// terminalProcesses lists known terminal emulators that may host CLI apps
var terminalProcesses = map[string]bool{
	"windowsterminal": true,
	"cmd":             true,
	"powershell":      true,
	"pwsh":            true,
	"wezterm":         true,
	"alacritty":       true,
	"hyper":           true,
	"mintty":          true,
	"wave":            true,
	"waveterm":        true,
	"conhost":         true,
	"warp":            true, // Warp terminal - uses paste mode
}

// cliAppProfiles maps CLI app names to their specific profiles
// These are detected as child processes of terminals
var cliAppProfiles = map[string]AppProfile{
	"claude":  ProfileSlow,                         // Claude Code CLI - slow mode works best
	"auggie":  ProfileAugment,                      // Augment CLI - needs Unicode backspace
	"augment": ProfileAugment,
}

// Cached CLI app detection result
var (
	cachedCLIApp     string // detected CLI app name (empty if none)
	cachedCLIProfile *AppProfile
)

// isTerminalProcess checks if process name is a known terminal
func isTerminalProcess(processName string) bool {
	return terminalProcesses[strings.ToLower(processName)]
}

// getChildProcesses returns direct child process names for a given parent PID
func getChildProcesses(parentPID uint32) []string {
	var children []string

	// Create snapshot of all processes
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == INVALID_HANDLE_VALUE {
		return children
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32W
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	// Get first process
	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return children
	}

	for {
		if entry.Th32ParentProcessID == parentPID {
			// Extract process name from SzExeFile
			name := syscall.UTF16ToString(entry.SzExeFile[:])
			name = ExtractProcessName(name)
			children = append(children, name)
		}

		// Get next process
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return children
}

// processInfo holds PID and parent PID for building process tree
type processInfo struct {
	pid       uint32
	parentPID uint32
	name      string
}

// getAllProcesses returns all processes with their PIDs and parent PIDs
func getAllProcesses() []processInfo {
	var processes []processInfo

	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(TH32CS_SNAPPROCESS, 0)
	if snapshot == INVALID_HANDLE_VALUE {
		return processes
	}
	defer procCloseHandle.Call(snapshot)

	var entry PROCESSENTRY32W
	entry.DwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return processes
	}

	for {
		name := syscall.UTF16ToString(entry.SzExeFile[:])
		name = ExtractProcessName(name)
		processes = append(processes, processInfo{
			pid:       entry.Th32ProcessID,
			parentPID: entry.Th32ParentProcessID,
			name:      name,
		})

		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	return processes
}

// isDescendantOf checks if a process is a descendant of the given ancestor PID
// Uses iterative approach to avoid stack overflow
func isDescendantOf(processes []processInfo, pid uint32, ancestorPID uint32, maxDepth int) bool {
	visited := make(map[uint32]bool)
	currentPID := pid

	for depth := 0; depth < maxDepth; depth++ {
		if currentPID == 0 || visited[currentPID] {
			return false
		}
		visited[currentPID] = true

		// Find parent of current process
		var parentPID uint32
		found := false
		for _, p := range processes {
			if p.pid == currentPID {
				parentPID = p.parentPID
				found = true
				break
			}
		}

		if !found {
			return false
		}

		if parentPID == ancestorPID {
			return true
		}

		currentPID = parentPID
	}

	return false
}

// findCLIAppDescendant searches all processes to find a CLI app that is a descendant of terminalPID
// Prioritizes Augment over Claude because Augment needs BackspaceUnicode while Claude uses terminal default
func findCLIAppDescendant(terminalPID uint32) (string, *AppProfile) {
	processes := getAllProcesses()

	// Collect all CLI apps found as descendants
	var foundCLIApps []struct {
		name    string
		profile AppProfile
		pid     uint32
	}

	for _, p := range processes {
		nameLower := strings.ToLower(p.name)
		if profile, ok := cliAppProfiles[nameLower]; ok {
			// Check if this CLI app is a descendant of the terminal
			if isDescendantOf(processes, p.pid, terminalPID, 10) {
				foundCLIApps = append(foundCLIApps, struct {
					name    string
					profile AppProfile
					pid     uint32
				}{nameLower, profile, p.pid})
			}
		}
	}

	if len(foundCLIApps) == 0 {
		return "", nil
	}

	// Prioritize: Augment (needs BackspaceUnicode) > others
	// Augment is the only CLI that actually needs different treatment
	for _, app := range foundCLIApps {
		if app.name == "auggie" || app.name == "augment" {
			return app.name, &app.profile
		}
	}

	// Return first found (likely Claude which uses terminal default anyway)
	app := foundCLIApps[0]
	return app.name, &app.profile
}

// detectCLIAppInTerminal checks if a known CLI app is running inside terminal
// First tries direct children, then searches descendants
func detectCLIAppInTerminal(terminalPID uint32) (string, *AppProfile) {
	// First try direct children (fast path)
	children := getChildProcesses(terminalPID)
	
	// Collect matching CLI apps from direct children
	var foundCLIApps []struct {
		name    string
		profile AppProfile
	}
	
	for _, child := range children {
		childLower := strings.ToLower(child)
		if profile, ok := cliAppProfiles[childLower]; ok {
			foundCLIApps = append(foundCLIApps, struct {
				name    string
				profile AppProfile
			}{childLower, profile})
		}
	}
	
	// Prioritize Augment if found among direct children
	for _, app := range foundCLIApps {
		if app.name == "auggie" || app.name == "augment" {
			return app.name, &app.profile
		}
	}
	
	// Return first direct child CLI if any
	if len(foundCLIApps) > 0 {
		app := foundCLIApps[0]
		return app.name, &app.profile
	}

	// If no direct CLI child found, search all descendants (slower but thorough)
	return findCLIAppDescendant(terminalPID)
}

// GetAppProfileForTerminal returns the appropriate profile when foreground is a terminal
// It checks for CLI apps running inside the terminal and returns their specific profile
func GetAppProfileForTerminal(terminalName string) AppProfile {
	// Get the terminal's PID
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ProfileSlow // fallback
	}

	var pid uint32
	procGetWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ProfileSlow // fallback
	}

	// Check for CLI apps in terminal
	cliApp, profile := detectCLIAppInTerminal(pid)
	if profile != nil {
		// Cache the result
		cachedCLIApp = cliApp
		cachedCLIProfile = profile
		return *profile
	}

	// No CLI app found, use default slow mode for terminals
	cachedCLIApp = ""
	cachedCLIProfile = nil
	return ProfileSlow
}

// GetSmartAppProfile returns the best profile considering CLI apps in terminals
// This should be called instead of GetAppProfile when you want CLI detection
func GetSmartAppProfile(processName string) AppProfile {
	name := strings.ToLower(processName)

	// First check if terminal has a specific profile (e.g., Warp needs paste mode)
	if profile, ok := appProfiles[name]; ok {
		return profile
	}

	// Check if it's a terminal - if so, look for CLI apps inside
	if isTerminalProcess(name) {
		return GetAppProfileForTerminal(name)
	}

	return ProfileFast
}

// GetCachedCLIApp returns the cached CLI app name (for debugging)
func GetCachedCLIApp() string {
	return cachedCLIApp
}
