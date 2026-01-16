package main

import (
	"fmt"
	"io/fs"
	"log"
	"syscall"
	"time"
	"unsafe"

	"fkey/core"
	"fkey/services"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Windows MessageBox constants
const (
	MB_OK              = 0x00000000
	MB_OKCANCEL        = 0x00000001
	MB_YESNO           = 0x00000004
	MB_ICONINFORMATION = 0x00000040
	MB_ICONWARNING     = 0x00000030
	MB_ICONQUESTION    = 0x00000020
	IDYES              = 6
	IDOK               = 1
)

var (
	user32DLL       = syscall.NewLazyDLL("user32.dll")
	procMessageBoxW = user32DLL.NewProc("MessageBoxW")
)

// showMessageBox shows a Windows message box
func showMessageBox(title, message string, flags uint32) int {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	messagePtr, _ := syscall.UTF16PtrFromString(message)
	ret, _, _ := procMessageBoxW.Call(
		0,
		uintptr(unsafe.Pointer(messagePtr)),
		uintptr(unsafe.Pointer(titlePtr)),
		uintptr(flags),
	)
	return int(ret)
}

// FKey - Vietnamese Input Method
// Wails v3 implementation (target: ~5MB)

// Version is set via -ldflags at build time: -X main.Version=x.x.x
var Version = "dev"

// Icons generated at runtime
var (
	iconOn  []byte
	iconOff []byte
)

// Global references for updates
var (
	globalApp        *application.App
	globalTray       *application.SystemTray
	globalMenu       *application.Menu
	globalImeLoop    *core.ImeLoop
	globalSettingsWin application.Window
	settingsSvc      *services.SettingsService
	updaterSvc       *services.UpdaterService
)

func main() {
	// Extract embedded DLL (single-exe distribution)
	dllPath, err := GetDLLPath()
	if err != nil {
		log.Fatalf("Failed to extract DLL: %v", err)
	}
	core.DLLPath = dllPath
	log.Printf("Using DLL: %s", dllPath)

	// Generate icons
	iconOn = CreateIconOn()
	iconOff = CreateIconOff()

	// Initialize services
	settingsSvc = services.NewSettingsService()
	if err := settingsSvc.Load(); err != nil {
		log.Printf("Failed to load settings: %v", err)
	}
	settings := settingsSvc.Settings()

	// Initialize IME loop
	globalImeLoop, err = core.NewImeLoop()
	if err != nil {
		log.Fatalf("Failed to create IME loop: %v", err)
	}

	// Apply settings to IME
	applySettings(globalImeLoop, settings)

	// Create App bindings
	appBindings := NewAppBindings(globalImeLoop, settingsSvc)

	// Create embedded assets filesystem
	frontendFS, err := fs.Sub(assets, "frontend")
	if err != nil {
		log.Fatalf("Failed to create frontend filesystem: %v", err)
	}

	// Create Wails application with bundled asset server (injects runtime automatically)
	globalApp = application.New(application.Options{
		Name:        "FKey",
		Description: "Vietnamese Input Method",
		Icon:        iconOn, // Application icon for windows
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(frontendFS),
		},
		Services: []application.Service{
			application.NewService(appBindings), // Pass pointer directly
		},
		Windows: application.WindowsOptions{
			// Windows-specific options
		},
	})

	// Create system tray
	globalTray = globalApp.SystemTray.New()
	if settings.Enabled {
		globalTray.SetIcon(iconOn)
		globalTray.SetTooltip("FKey - Tiếng Việt (Bật)")
	} else {
		globalTray.SetIcon(iconOff)
		globalTray.SetTooltip("FKey - Tiếng Việt (Tắt)")
	}

	// Create settings window (hidden by default)
	globalSettingsWin = globalApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:                       "FKey Settings",
		Title:                      "FKey - Cài đặt",
		Width:                      400,
		Height:                     500,
		Hidden:                     true,
		DisableResize:              false,
		URL:                        "/",
		DevToolsEnabled:            false,
		DefaultContextMenuDisabled: true,
		Windows: application.WindowsWindow{
			// Show on taskbar when window is visible
			HiddenOnTaskbar: false,
		},
	})

	// Hide window on close instead of quitting
	globalSettingsWin.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
		globalSettingsWin.Hide()
		e.Cancel()
	})

	// Create tray menu
	globalMenu = createTrayMenu(settings.Enabled)
	globalTray.SetMenu(globalMenu)

	// Left-click toggles IME
	globalTray.OnClick(func() {
		toggleIME()
	})

	// Status callback - called when hotkey toggles IME
	globalImeLoop.OnEnabledChanged = func(enabled bool) {
		updateUI(enabled)
	}

	// Start IME loop BEFORE app.Run() so keyboard hook is active
	if err := globalImeLoop.Start(); err != nil {
		log.Fatalf("Failed to start IME loop: %v", err)
	}

	// Initialize updater service
	updaterSvc = services.NewUpdaterService(Version)

	// Check for updates in background (non-blocking)
	go checkForUpdatesBackground()

	log.Printf("FKey started. IME: %s, Method: %d", 
		map[bool]string{true: "ON", false: "OFF"}[settings.Enabled],
		settings.InputMethod)

	// Run application (blocks until quit)
	if err := globalApp.Run(); err != nil {
		log.Fatal(err)
	}

	// Cleanup
	globalImeLoop.Stop()
}

// checkForUpdatesBackground checks for updates silently at startup
func checkForUpdatesBackground() {
	// Wait a bit for app to fully initialize
	time.Sleep(3 * time.Second)

	info, err := updaterSvc.CheckForUpdates(false)
	if err != nil {
		log.Printf("Update check failed: %v", err)
		return
	}

	if info.Available {
		log.Printf("Update available: %s -> %s", info.CurrentVersion, info.LatestVersion)
		// Show update notification dialog
		result := showMessageBox("Có phiên bản mới!",
			fmt.Sprintf("Phiên bản mới: %s\nPhiên bản hiện tại: %s\n\nBạn có muốn mở trang tải về?",
				info.LatestVersion, info.CurrentVersion),
			MB_YESNO|MB_ICONINFORMATION)
		if result == IDYES {
			updaterSvc.OpenReleasePage(info.ReleaseURL)
		}
	}
}

func toggleIME() {
	enabled := globalImeLoop.Toggle()
	settingsSvc.Settings().Enabled = enabled
	settingsSvc.Save()
	updateUI(enabled)
}

func updateUI(enabled bool) {
	// Update tray icon
	if enabled {
		globalTray.SetIcon(iconOn)
		globalTray.SetTooltip("FKey - Tiếng Việt (Bật)")
	} else {
		globalTray.SetIcon(iconOff)
		globalTray.SetTooltip("FKey - Tiếng Việt (Tắt)")
	}

	// Rebuild menu with new state
	globalMenu = createTrayMenu(enabled)
	globalTray.SetMenu(globalMenu)
}

func applySettings(loop *core.ImeLoop, settings *services.Settings) {
	imeSettings := &core.ImeSettings{
		Enabled:            settings.Enabled,
		InputMethod:        core.InputMethod(settings.InputMethod),
		ModernTone:         settings.ModernTone,
		SkipWShortcut:      settings.SkipWShortcut,
		EscRestore:         settings.EscRestore,
		FreeTone:           settings.FreeTone,
		EnglishAutoRestore: settings.EnglishAutoRestore,
		AutoCapitalize:     settings.AutoCapitalize,
	}
	loop.UpdateSettings(imeSettings)

	// Set hotkey
	keyCode, ctrl, alt, shift := services.ParseHotkey(settings.ToggleHotkey)
	loop.SetHotkey(keyCode, ctrl, alt, shift)

	// Load shortcuts
	shortcuts, err := settingsSvc.LoadShortcuts()
	if err == nil {
		for _, sc := range shortcuts {
			if sc.Enabled {
				loop.AddShortcut(sc.Trigger, sc.Replacement)
			}
		}
	}
}

func createTrayMenu(enabled bool) *application.Menu {
	menu := globalApp.NewMenu()
	settings := settingsSvc.Settings()

	// Status indicator with checkbox
	enabledItem := menu.AddCheckbox("Tiếng Việt", enabled)
	enabledItem.OnClick(func(ctx *application.Context) {
		toggleIME()
	})

	menu.AddSeparator()

	// Input method
	methodMenu := menu.AddSubmenu("Kiểu gõ")
	telexItem := methodMenu.AddRadio("Telex", settings.InputMethod == 0)
	vniItem := methodMenu.AddRadio("VNI", settings.InputMethod == 1)

	telexItem.OnClick(func(ctx *application.Context) {
		globalImeLoop.UpdateSettings(&core.ImeSettings{
			Enabled:     settingsSvc.Settings().Enabled,
			InputMethod: core.Telex,
		})
		settingsSvc.Settings().InputMethod = 0
		settingsSvc.Save()
	})
	vniItem.OnClick(func(ctx *application.Context) {
		globalImeLoop.UpdateSettings(&core.ImeSettings{
			Enabled:     settingsSvc.Settings().Enabled,
			InputMethod: core.VNI,
		})
		settingsSvc.Settings().InputMethod = 1
		settingsSvc.Save()
	})

	menu.AddSeparator()

	// Settings
	menu.Add("Cài đặt...").OnClick(func(ctx *application.Context) {
		globalSettingsWin.Show()
		globalSettingsWin.Focus()
	})

	// Check for updates
	menu.Add("Kiểm tra cập nhật...").OnClick(func(ctx *application.Context) {
		go func() {
			info, err := updaterSvc.CheckForUpdates(true)
			if err != nil {
				log.Printf("Update check failed: %v", err)
				showMessageBox("Kiểm tra cập nhật", 
					"Không thể kiểm tra cập nhật.\n\n"+err.Error(), 
					MB_OK|MB_ICONWARNING)
				return
			}
			if info.Available {
				log.Printf("Update available: %s", info.LatestVersion)
				result := showMessageBox("Có phiên bản mới!", 
					fmt.Sprintf("Phiên bản mới: %s\nPhiên bản hiện tại: %s\n\nBạn có muốn mở trang tải về?", 
						info.LatestVersion, info.CurrentVersion),
					MB_YESNO|MB_ICONQUESTION)
				if result == IDYES {
					updaterSvc.OpenReleasePage(info.ReleaseURL)
				}
			} else {
				log.Printf("Already at latest version: %s", info.CurrentVersion)
				showMessageBox("Kiểm tra cập nhật", 
					fmt.Sprintf("Bạn đang sử dụng phiên bản mới nhất.\n\nPhiên bản: %s", info.CurrentVersion),
					MB_OK|MB_ICONINFORMATION)
			}
		}()
	})

	menu.AddSeparator()

	// Version
	menu.Add(fmt.Sprintf("FKey v%s", Version)).SetEnabled(false)

	menu.AddSeparator()

	// Quit
	menu.Add("Thoát").OnClick(func(ctx *application.Context) {
		globalApp.Quit()
	})

	return menu
}
