package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"sync/atomic"
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
	globalApp         *application.App
	globalTray        *application.SystemTray
	globalMenu        *application.Menu
	globalImeLoop     *core.ImeLoop
	globalSettingsWin application.Window
	settingsSvc       *services.SettingsService
	updaterSvc        *services.UpdaterService
	formattingSvc     *services.FormattingService
	wantQuit          bool // Flag to allow quit via tray menu
	hubChannel        *core.HubChannel
	appStarted        atomic.Bool // true once Wails Run() has set up the main thread
)

func main() {
	core.StartupTraceBegin()

	launch := core.ParseLaunchArgs(os.Args[1:])

	if err := core.AcquireMutex(launch.Relaunch); err != nil {
		showMessageBox("FKey", err.Error(), MB_OK|MB_ICONWARNING)
		os.Exit(1)
	}
	defer core.ReleaseMutex()
	core.StartupTraceStage("mutex")

	core.QuitApp = func() {
		wantQuit = true
		globalApp.Quit()
	}

	core.RevertRunAsAdmin = func() {
		if settingsSvc != nil {
			settingsSvc.Settings().RunAsAdmin = false
			settingsSvc.Save()
		}
	}

	// Extract embedded DLL (single-exe distribution)
	dllPath, err := GetDLLPath()
	if err != nil {
		log.Fatalf("Failed to extract DLL: %v", err)
	}
	core.DLLPath = dllPath
	log.Printf("Using DLL: %s", dllPath)
	core.StartupTraceStage("dll-extract")

	// Generate icons
	iconOn = CreateIconOn()
	iconOff = CreateIconOff()
	core.StartupTraceStage("icons")

	// Initialize services
	settingsSvc = services.NewSettingsService()
	if err := settingsSvc.Load(); err != nil {
		log.Printf("Failed to load settings: %v", err)
	}
	settings := settingsSvc.Settings()
	core.StartupTraceStage("settings-load")

	// An elevated FKey is unsupported in hub mode for now; it keeps its tray.
	hubMode := core.HubModeEnabled(launch.Hub, settings.RunAsAdmin, services.IsElevated())
	settingsSvc.SetHubManaged(hubMode)

	if settings.RunAsAdmin && !services.IsElevated() {
		log.Printf("RunAsAdmin enabled but not elevated, re-launching...")
		core.ElevateAndRelaunch()
		return
	}

	if !hubMode && settings.AutoStart && settings.RunAsAdmin && services.IsElevated() {
		// Off the boot path: this shells out to schtasks.exe, which has no
		// business blocking typing-readiness. Only affects RunAsAdmin users.
		go func() {
			time.Sleep(3 * time.Second)
			settingsSvc.ReconcileScheduledTaskPath()
		}()
	}

	// Initialize formatting service
	formattingSvc = services.NewFormattingService()
	if err := formattingSvc.Load(); err != nil {
		log.Printf("Failed to load formatting config: %v", err)
	}
	core.StartupTraceStage("formatting-load")

	// Initialize IME loop
	globalImeLoop, err = core.NewImeLoop()
	if err != nil {
		log.Fatalf("Failed to create IME loop: %v", err)
	}
	core.StartupTraceStage("imeloop-create")

	// Apply settings to IME
	applySettings(globalImeLoop, settings)

	// Initialize format handler for text formatting feature
	core.InitFormatHandler(formattingSvc)

	// Status callback - called when hotkey toggles IME. Wired before Start()
	// since the hook thread (started below) can invoke this the instant a
	// user presses the toggle hotkey - before Wails (globalApp/globalTray/
	// globalMenu) exists yet. updateUI is marshaled onto the Wails main
	// thread and nil-guarded for exactly that reason; PlayBeep is a plain
	// Win32 call with no GUI-thread affinity, safe to run directly here.
	// Before Run() has set up the main thread there is nothing to marshal
	// onto (InvokeAsync would panic), so only beep; updateUI runs once at start
	// through the startup publish.
	globalImeLoop.OnEnabledChanged = core.NewEnabledChangedHandler(
		core.PlayBeep,
		func(f func()) {
			if appStarted.Load() {
				application.InvokeAsync(f)
			}
		},
		updateUI,
	)

	// Start the keyboard/mouse hooks on their own dedicated, pumping OS
	// thread NOW - before Wails initializes - so typing works while the
	// webview/tray/window are still being set up. This must run
	// independently of app.Run()'s message loop; see hook_pump.go.
	if err := globalImeLoop.Start(); err != nil {
		log.Fatalf("Failed to start IME loop: %v", err)
	}
	core.StartupTraceStage("hook-start")

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
		// Prevent app from quitting when settings window closes
		// App should only quit via tray menu "Thoát"
		ShouldQuit: func() bool {
			return wantQuit
		},
		Windows: application.WindowsOptions{
			// Windows-specific options
		},
	})
	core.StartupTraceStage("wails-new")

	// Hub mode: publish state through the channel and hide the tray. If the
	// channel cannot be created, keep the tray (fail-safe).
	if hubMode {
		ch, err := core.CreateHubChannel(core.HubChannelPrefix)
		if err != nil {
			log.Printf("hub channel unavailable, keeping tray: %v", err)
		} else {
			hubChannel = ch
			defer ch.Close()
		}
	}

	if hubChannel == nil {
		// Create system tray
		globalTray = globalApp.SystemTray.New()
		if settings.Enabled {
			globalTray.SetIcon(iconOn)
			globalTray.SetTooltip("FKey - Tiếng Việt (Bật)")
		} else {
			globalTray.SetIcon(iconOff)
			globalTray.SetTooltip("FKey - Tiếng Việt (Tắt)")
		}

		// Settings window created on-demand (lazy-load for RAM optimization)
		// globalSettingsWin starts as nil - created when user opens settings

		// Create tray menu
		globalMenu = createTrayMenu(settings.Enabled)
		globalTray.SetMenu(globalMenu)

		// Left-click toggles IME
		globalTray.OnClick(func() {
			toggleIME()
		})
		core.StartupTraceStage("tray-ready")
	}
	if hubChannel != nil {
		hubChannel.Publish(settings.Enabled)
	}
	// The request loop starts only after Run() has set up the main thread,
	// because InvokeAsync panics before that. Toggle runs on the same thread
	// as a tray click and fires OnEnabledChanged (beep + updateUI). A request
	// signalled earlier stays pending in the auto-reset event.
	globalApp.Event.OnApplicationEvent(events.Common.ApplicationStarted, func(*application.ApplicationEvent) {
		appStarted.Store(true)
		// A hotkey toggle before Run() only beeped; bring the tray, hub state and stored Enabled up to date once.
		application.InvokeAsync(func() {
			if e := globalImeLoop.IsEnabled(); e != settingsSvc.Settings().Enabled {
				updateUI(e)
			}
		})
		if hubChannel != nil {
			go hubChannel.RunRequestLoop(
				func() { application.InvokeAsync(func() { globalImeLoop.Toggle() }) },
				func() { application.InvokeAsync(reloadSettings) },
			)
		}
	})

	// Initialize updater service
	updaterSvc = services.NewUpdaterService(Version)

	// Check for updates in background (non-blocking). Off in hub mode: the
	// updater's relaunch would drop --hub; the hub's catalog owns updates.
	if core.ShouldCheckUpdates(hubMode) {
		go checkForUpdatesBackground()
	}

	log.Printf("FKey started. IME: %s, Method: %d",
		map[bool]string{true: "ON", false: "OFF"}[settings.Enabled],
		settings.InputMethod)

	core.StartupTraceStage("run")
	core.StartupTraceFinish()

	// Run application (blocks until quit)
	if err := globalApp.Run(); err != nil {
		log.Fatal(err)
	}

	// Cleanup
	globalImeLoop.Stop()
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

// showSettingsWindow creates settings window on-demand (lazy-load) and hides on close
// WebView2 stays in RAM after first open, but this prevents app crash
func showSettingsWindow() {
	if globalSettingsWin == nil {
		globalSettingsWin = globalApp.Window.NewWithOptions(application.WebviewWindowOptions{
			Name:                       "FKey Settings",
			Title:                      "FKey - Cài đặt",
			Width:                      520,
			Height:                     560,
			Hidden:                     false, // Show immediately
			DisableResize:              false,
			URL:                        "/",
			DevToolsEnabled:            false,
			DefaultContextMenuDisabled: true,
			Windows: application.WindowsWindow{
				HiddenOnTaskbar: false,
			},
		})

		// Hide window on close instead of destroying
		// Wails v3 quits app when last window closes, even with ShouldQuit
		globalSettingsWin.RegisterHook(events.Common.WindowClosing, func(e *application.WindowEvent) {
			globalSettingsWin.Hide()
			e.Cancel()
		})
	}
	globalSettingsWin.Show()
	globalSettingsWin.Focus()
}

// showTooltipNotification shows a brief tooltip notification
func showTooltipNotification(title, message string) {
	// No tray in hub mode: the OSD is a tray tooltip, so it shows nothing there.
	if globalTray == nil {
		return
	}
	// Update the tooltip temporarily to show language change
	// The tooltip will be shown when user hovers over the tray icon
	globalTray.SetTooltip(message)

	// Restore normal tooltip after a delay
	go func() {
		time.Sleep(2 * time.Second)
		if globalTray == nil {
			return
		}
		if settingsSvc.Settings().Enabled {
			globalTray.SetTooltip("FKey - Tiếng Việt (Bật)")
		} else {
			globalTray.SetTooltip("FKey - Tiếng Việt (Tắt)")
		}
	}()
}

// updateUI runs on the Wails main thread (posted by the enabled-changed
// handler, or called at startup). Order: hub state, persisted Enabled, frontend
// event, OSD, then the tray part (absent in hub mode).
func updateUI(enabled bool) {
	if hubChannel != nil {
		hubChannel.Publish(enabled)
	}

	// Hotkey toggles bypass toggleIME, so persist Enabled here; SaveEnabled
	// writes one DWORD and never touches autostart.
	if settingsSvc != nil && settingsSvc.Settings().Enabled != enabled {
		settingsSvc.Settings().Enabled = enabled
		if err := settingsSvc.SaveEnabled(); err != nil {
			log.Printf("persist Enabled: %v", err)
		}
	}

	if globalApp == nil {
		return
	}

	// Emit event to frontend so Settings UI can update status indicator
	globalApp.Event.Emit("ime:status-changed", enabled)

	// Show OSD popup if enabled
	if settingsSvc != nil && settingsSvc.Settings().ShowOSD {
		go showOSDPopup(enabled)
	}

	if globalTray == nil || settingsSvc == nil {
		return
	}

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

// reloadSettings applies settings the hub wrote. Runs on the main thread (InvokeAsync).
// It never calls Save(), so the hub-owned Run value is never touched. Known
// limit: a hotkey toggle in the microseconds between IsEnabled() and
// UpdateSettings can be undone; the tray Telex/VNI handler has the same
// unsynchronised pattern. It never runs on the hook thread.
func reloadSettings() {
	if err := settingsSvc.Reload(); err != nil {
		log.Printf("hub reload: %v", err)
		return
	}
	settingsSvc.Settings().Enabled = globalImeLoop.IsEnabled()
	if formattingSvc != nil {
		_ = formattingSvc.Load()
	}
	globalImeLoop.ClearShortcuts()
	applySettings(globalImeLoop, settingsSvc.Settings())
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

	// Set SmartPaste enabled state
	core.SetSmartPasteEnabled(settings.SmartPaste)

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
		showSettingsWindow()
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
		wantQuit = true
		globalApp.Quit()
	})

	return menu
}
