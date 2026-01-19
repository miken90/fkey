package ui

import (
	"log"

	"fkey-linux/config"

	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// Tray manages the system tray icon and menu
type Tray struct {
	app          *gtk.Application
	statusIcon   *gtk.StatusIcon
	config       *config.Config
	onToggle     func(bool)
	settingsWin  *gtk.Window
}

// NewTray creates a system tray icon
func NewTray(cfg *config.Config, onToggle func(bool)) (*Tray, error) {
	gtk.Init(nil)

	app, err := gtk.ApplicationNew("org.gonhanh.fkey", glib.APPLICATION_FLAGS_NONE)
	if err != nil {
		return nil, err
	}

	t := &Tray{
		app:      app,
		config:   cfg,
		onToggle: onToggle,
	}

	app.Connect("activate", func() {
		t.setupTray()
	})

	return t, nil
}

func (t *Tray) setupTray() {
	// Create status icon
	icon, err := gtk.StatusIconNewFromIconName(t.getIconName())
	if err != nil {
		log.Printf("Failed to create status icon: %v", err)
		return
	}
	t.statusIcon = icon

	t.statusIcon.SetTooltipText(t.getTooltip())
	t.statusIcon.SetVisible(true)

	// Left click - toggle
	t.statusIcon.Connect("activate", func() {
		t.toggle()
	})

	// Right click - menu
	t.statusIcon.Connect("popup-menu", func(icon *gtk.StatusIcon, button, activateTime uint) {
		t.showMenu(button, activateTime)
	})
}

func (t *Tray) getIconName() string {
	if t.config.Enabled {
		return "input-keyboard" // Standard icon, replace with custom
	}
	return "input-keyboard-symbolic"
}

func (t *Tray) getTooltip() string {
	status := "Tắt"
	if t.config.Enabled {
		status = "Bật"
	}

	method := "Telex"
	if t.config.InputMethod == 1 {
		method = "VNI"
	}

	return "FKey - Tiếng Việt (" + status + ") - " + method
}

func (t *Tray) toggle() {
	t.config.Enabled = !t.config.Enabled
	if t.onToggle != nil {
		t.onToggle(t.config.Enabled)
	}
	t.updateIcon()
}

func (t *Tray) updateIcon() {
	if t.statusIcon == nil {
		return
	}
	t.statusIcon.SetFromIconName(t.getIconName())
	t.statusIcon.SetTooltipText(t.getTooltip())
}

func (t *Tray) showMenu(button, activateTime uint) {
	menu, _ := gtk.MenuNew()

	// Toggle item
	toggleText := "Bật Tiếng Việt"
	if t.config.Enabled {
		toggleText = "✓ Tiếng Việt"
	}
	toggleItem, _ := gtk.MenuItemNewWithLabel(toggleText)
	toggleItem.Connect("activate", func() {
		t.toggle()
	})
	menu.Append(toggleItem)

	menu.Append(t.createSeparator())

	// Input method submenu
	methodMenu, _ := gtk.MenuNew()
	methodItem, _ := gtk.MenuItemNewWithLabel("Kiểu gõ")
	methodItem.SetSubmenu(methodMenu)

	telexItem, _ := gtk.CheckMenuItemNewWithLabel("Telex")
	telexItem.SetActive(t.config.InputMethod == 0)
	telexItem.Connect("activate", func() {
		t.config.InputMethod = 0
		config.Save(t.config)
	})
	methodMenu.Append(telexItem)

	vniItem, _ := gtk.CheckMenuItemNewWithLabel("VNI")
	vniItem.SetActive(t.config.InputMethod == 1)
	vniItem.Connect("activate", func() {
		t.config.InputMethod = 1
		config.Save(t.config)
	})
	methodMenu.Append(vniItem)

	menu.Append(methodItem)

	menu.Append(t.createSeparator())

	// Settings
	settingsItem, _ := gtk.MenuItemNewWithLabel("Cài đặt...")
	settingsItem.Connect("activate", func() {
		t.showSettings()
	})
	menu.Append(settingsItem)

	menu.Append(t.createSeparator())

	// About
	aboutItem, _ := gtk.MenuItemNewWithLabel("Về FKey")
	aboutItem.Connect("activate", func() {
		t.showAbout()
	})
	menu.Append(aboutItem)

	menu.Append(t.createSeparator())

	// Quit
	quitItem, _ := gtk.MenuItemNewWithLabel("Thoát")
	quitItem.Connect("activate", func() {
		t.Quit()
	})
	menu.Append(quitItem)

	menu.ShowAll()
	menu.PopupAtStatusIcon(t.statusIcon, button, activateTime)
}

func (t *Tray) createSeparator() *gtk.SeparatorMenuItem {
	sep, _ := gtk.SeparatorMenuItemNew()
	return sep
}

func (t *Tray) showSettings() {
	if t.settingsWin != nil {
		t.settingsWin.Present()
		return
	}

	win, _ := gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	win.SetTitle("FKey - Cài đặt")
	win.SetDefaultSize(350, 300)
	win.SetPosition(gtk.WIN_POS_CENTER)

	t.settingsWin = win

	box, _ := gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 10)
	box.SetMarginTop(20)
	box.SetMarginBottom(20)
	box.SetMarginStart(20)
	box.SetMarginEnd(20)

	// Modern tone checkbox
	modernCheck, _ := gtk.CheckButtonNewWithLabel("Bỏ dấu kiểu mới (hoà thay vì hòa)")
	modernCheck.SetActive(t.config.ModernTone)
	modernCheck.Connect("toggled", func() {
		t.config.ModernTone = modernCheck.GetActive()
		config.Save(t.config)
	})
	box.Add(modernCheck)

	// ESC restore checkbox
	escCheck, _ := gtk.CheckButtonNewWithLabel("ESC khôi phục ký tự gốc")
	escCheck.SetActive(t.config.EscRestore)
	escCheck.Connect("toggled", func() {
		t.config.EscRestore = escCheck.GetActive()
		config.Save(t.config)
	})
	box.Add(escCheck)

	// Autostart checkbox
	autoCheck, _ := gtk.CheckButtonNewWithLabel("Khởi động cùng hệ thống")
	autoCheck.SetActive(t.config.AutoStart)
	autoCheck.Connect("toggled", func() {
		t.config.AutoStart = autoCheck.GetActive()
		config.Save(t.config)
		// TODO: Create/remove autostart desktop file
	})
	box.Add(autoCheck)

	win.Add(box)
	win.Connect("destroy", func() {
		t.settingsWin = nil
	})
	win.ShowAll()
}

func (t *Tray) showAbout() {
	dialog, _ := gtk.AboutDialogNew()
	dialog.SetProgramName("FKey")
	dialog.SetVersion("0.1.0")
	dialog.SetComments("Bộ gõ Tiếng Việt cho Linux")
	dialog.SetWebsite("https://github.com/miken90/fkey")
	dialog.SetCopyright("© 2025-2026 GoNhanh.org")
	dialog.SetLicense("MIT License")
	dialog.Run()
	dialog.Destroy()
}

// Run starts the GTK main loop
func (t *Tray) Run() {
	t.app.Run(nil)
}

// Quit exits the application
func (t *Tray) Quit() {
	t.app.Quit()
}

// SetEnabled updates enabled state and refreshes UI
func (t *Tray) SetEnabled(enabled bool) {
	t.config.Enabled = enabled
	t.updateIcon()
}
