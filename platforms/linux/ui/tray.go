package ui

import (
	"log"

	"fkey-linux/config"

	"github.com/getlantern/systray"
)

// Tray manages the system tray icon and menu
type Tray struct {
	config     *config.Config
	onToggle   func(bool)
	toggleItem *systray.MenuItem
	telexItem  *systray.MenuItem
	vniItem    *systray.MenuItem
	quitChan   chan struct{}
}

// NewTray creates a system tray manager
func NewTray(cfg *config.Config, onToggle func(bool)) (*Tray, error) {
	return &Tray{
		config:   cfg,
		onToggle: onToggle,
		quitChan: make(chan struct{}),
	}, nil
}

// Run starts the systray (blocks until quit)
func (t *Tray) Run() {
	systray.Run(t.onReady, t.onExit)
}

// Quit exits the tray application
func (t *Tray) Quit() {
	systray.Quit()
}

// SetEnabled updates enabled state and refreshes UI
func (t *Tray) SetEnabled(enabled bool) {
	t.config.Enabled = enabled
	t.updateToggleItem()
}

func (t *Tray) onReady() {
	systray.SetTitle("FKey")
	systray.SetTooltip(t.getTooltip())

	// Toggle item
	t.toggleItem = systray.AddMenuItem(t.getToggleText(), "Bật/Tắt gõ tiếng Việt")

	systray.AddSeparator()

	// Input method submenu
	methodMenu := systray.AddMenuItem("Kiểu gõ", "Chọn kiểu gõ")
	t.telexItem = methodMenu.AddSubMenuItem("Telex", "Kiểu gõ Telex")
	t.vniItem = methodMenu.AddSubMenuItem("VNI", "Kiểu gõ VNI")
	t.updateMethodItems()

	systray.AddSeparator()

	// Settings items
	modernItem := systray.AddMenuItemCheckbox("Bỏ dấu kiểu mới", "hoà thay vì hòa", t.config.ModernTone)
	escItem := systray.AddMenuItemCheckbox("ESC khôi phục ký tự gốc", "", t.config.EscRestore)

	systray.AddSeparator()

	// About
	aboutItem := systray.AddMenuItem("Về FKey", "Thông tin phần mềm")

	systray.AddSeparator()

	// Quit
	quitItem := systray.AddMenuItem("Thoát", "Đóng FKey")

	// Handle menu clicks in goroutine
	go func() {
		for {
			select {
			case <-t.toggleItem.ClickedCh:
				t.toggle()

			case <-t.telexItem.ClickedCh:
				t.config.InputMethod = 0
				t.updateMethodItems()
				config.Save(t.config)

			case <-t.vniItem.ClickedCh:
				t.config.InputMethod = 1
				t.updateMethodItems()
				config.Save(t.config)

			case <-modernItem.ClickedCh:
				t.config.ModernTone = !t.config.ModernTone
				if t.config.ModernTone {
					modernItem.Check()
				} else {
					modernItem.Uncheck()
				}
				config.Save(t.config)

			case <-escItem.ClickedCh:
				t.config.EscRestore = !t.config.EscRestore
				if t.config.EscRestore {
					escItem.Check()
				} else {
					escItem.Uncheck()
				}
				config.Save(t.config)

			case <-aboutItem.ClickedCh:
				log.Println("FKey - Bộ gõ Tiếng Việt cho Linux")
				log.Println("https://github.com/miken90/fkey")

			case <-quitItem.ClickedCh:
				systray.Quit()
				return

			case <-t.quitChan:
				return
			}
		}
	}()

	log.Println("System tray ready")
}

func (t *Tray) onExit() {
	close(t.quitChan)
	log.Println("System tray exiting")
}

func (t *Tray) toggle() {
	t.config.Enabled = !t.config.Enabled
	if t.onToggle != nil {
		t.onToggle(t.config.Enabled)
	}
	t.updateToggleItem()
	config.Save(t.config)
}

func (t *Tray) updateToggleItem() {
	if t.toggleItem == nil {
		return
	}
	t.toggleItem.SetTitle(t.getToggleText())
	systray.SetTooltip(t.getTooltip())
}

func (t *Tray) updateMethodItems() {
	if t.telexItem == nil || t.vniItem == nil {
		return
	}
	if t.config.InputMethod == 0 {
		t.telexItem.Check()
		t.vniItem.Uncheck()
	} else {
		t.telexItem.Uncheck()
		t.vniItem.Check()
	}
}

func (t *Tray) getToggleText() string {
	if t.config.Enabled {
		return "✓ Tiếng Việt"
	}
	return "Bật Tiếng Việt"
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
