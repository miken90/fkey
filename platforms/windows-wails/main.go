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
	MB_YESNOCANCEL     = 0x00000003
	MB_ICONINFORMATION = 0x00000040
	MB_ICONWARNING     = 0x00000030
	MB_ICONQUESTION    = 0x00000020
	IDYES              = 6
	IDNO               = 7
	IDOK               = 1
	IDCANCEL           = 2
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
	formattingSvc    *services.FormattingService
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

	// Initialize formatting service
	formattingSvc = services.NewFormattingService()
	if err := formattingSvc.Load(); err != nil {
		log.Printf("Failed to load formatting config: %v", err)
	}

	// Initialize IME loop
	globalImeLoop, err = core.NewImeLoop()
	if err != nil {
		log.Fatalf("Failed to create IME loop: %v", err)
	}

	// Apply settings to IME
	applySettings(globalImeLoop, settings)

	// Initialize format handler for text formatting feature
	core.InitFormatHandler(formattingSvc)

	// Create App bindings
	appBindings := NewAppBindings(globalImeLoop, settingsSvc, formattingSvc)

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
		Width:                      520,
		Height:                     560,
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
		// Play beep sound when toggled via hotkey
		core.PlayBeep(enabled)
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

	if info.Available && info.DownloadURL != "" {
		log.Printf("Update available: %s -> %s", info.CurrentVersion, info.LatestVersion)
		// Show update notification dialog with auto-update option
		result := showMessageBox("Có phiên bản mới!",
			fmt.Sprintf("Phiên bản mới: %s\nPhiên bản hiện tại: %s\n\nBạn có muốn tự động cập nhật?\n\n(Chọn No để mở trang tải về)",
				info.LatestVersion, info.CurrentVersion),
			MB_YESNOCANCEL|MB_ICONINFORMATION)
		if result == IDYES {
			performAutoUpdate(info.DownloadURL)
		} else if result == IDNO {
			updaterSvc.OpenReleasePage(info.ReleaseURL)
		}
	}
}

// performAutoUpdate downloads and installs update automatically
func performAutoUpdate(downloadURL string) {
	log.Printf("Starting auto-update from: %s", downloadURL)
	
	// Show downloading message
	showMessageBox("Đang tải về...", 
		"FKey đang tải bản cập nhật.\nVui lòng đợi...", 
		MB_OK|MB_ICONINFORMATION)
	
	// Download update
	zipPath, err := updaterSvc.DownloadUpdate(downloadURL, nil)
	if err != nil {
		log.Printf("Download failed: %v", err)
		showMessageBox("Lỗi cập nhật", 
			"Không thể tải bản cập nhật.\n\n"+err.Error(), 
			MB_OK|MB_ICONWARNING)
		return
	}
	log.Printf("Downloaded to: %s", zipPath)
	
	// Install update (creates batch script)
	batchPath, err := updaterSvc.InstallUpdate(zipPath)
	if err != nil {
		log.Printf("Install failed: %v", err)
		showMessageBox("Lỗi cập nhật", 
			"Không thể cài đặt bản cập nhật.\n\n"+err.Error(), 
			MB_OK|MB_ICONWARNING)
		return
	}
	log.Printf("Update script created: %s", batchPath)
	
	// Run update script and quit app
	if err := updaterSvc.RunUpdateScript(batchPath); err != nil {
		log.Printf("Failed to run update script: %v", err)
		showMessageBox("Lỗi cập nhật", 
			"Không thể chạy script cập nhật.\n\n"+err.Error(), 
			MB_OK|MB_ICONWARNING)
		return
	}
	
	// Quit app to allow update
	log.Printf("Quitting for update...")
	globalApp.Quit()
}

func toggleIME() {
	enabled := globalImeLoop.Toggle()
	settingsSvc.Settings().Enabled = enabled
	settingsSvc.Save()
	updateUI(enabled)
	// Play beep sound to indicate toggle
	core.PlayBeep(enabled)
}

// showOSDPopup displays a brief on-screen notification when switching language
func showOSDPopup(isVietnamese bool) {
	var title, message string
	if isVietnamese {
		title = "FKey"
		message = "🇻🇳 Tiếng Việt"
	} else {
		title = "FKey"
		message = "🇺🇸 English"
	}
	
	// Use Windows MessageBox with auto-close via timer
	// For non-blocking: spawn a goroutine that shows a quick tooltip-style message
	time.Sleep(100 * time.Millisecond) // Brief delay to avoid UI race
	
	// Create a simple tooltip-style window using MessageBox with timeout
	// Note: This is a temporary solution. Proper OSD would use layered windows.
	showTooltipNotification(title, message)
}

// showTooltipNotification shows a brief tooltip notification
func showTooltipNotification(title, message string) {
	// Update the tooltip temporarily to show language change
	// The tooltip will be shown when user hovers over the tray icon
	globalTray.SetTooltip(message)
	
	// Restore normal tooltip after a delay
	go func() {
		time.Sleep(2 * time.Second)
		if settingsSvc.Settings().Enabled {
			globalTray.SetTooltip("FKey - Tiếng Việt (Bật)")
		} else {
			globalTray.SetTooltip("FKey - Tiếng Việt (Tắt)")
		}
	}()
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

	// Emit event to frontend so Settings UI can update status indicator
	globalApp.Event.Emit("ime:status-changed", enabled)

	// Show OSD popup if enabled
	if settingsSvc.Settings().ShowOSD {
		go showOSDPopup(enabled)
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
			if info.Available && info.DownloadURL != "" {
				log.Printf("Update available: %s", info.LatestVersion)
				result := showMessageBox("Có phiên bản mới!", 
					fmt.Sprintf("Phiên bản mới: %s\nPhiên bản hiện tại: %s\n\nBạn có muốn tự động cập nhật?\n\n(Chọn No để mở trang tải về)", 
						info.LatestVersion, info.CurrentVersion),
					MB_YESNOCANCEL|MB_ICONQUESTION)
				if result == IDYES {
					performAutoUpdate(info.DownloadURL)
				} else if result == IDNO {
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
