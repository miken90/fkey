package main

import (
	"fkey/core"
	"fkey/services"
)

// AppBindings exposes methods to the frontend via Wails bindings
type AppBindings struct {
	imeLoop     *core.ImeLoop
	settingsSvc *services.SettingsService
	updaterSvc  *services.UpdaterService
}

// NewAppBindings creates a new AppBindings instance
func NewAppBindings(imeLoop *core.ImeLoop, settingsSvc *services.SettingsService) *AppBindings {
	return &AppBindings{
		imeLoop:     imeLoop,
		settingsSvc: settingsSvc,
		updaterSvc:  services.NewUpdaterService(Version),
	}
}

// --- Frontend Bindings (exposed to JavaScript) ---

// GetEnabled returns whether Vietnamese input is enabled
func (a *AppBindings) GetEnabled() bool {
	return a.settingsSvc.Settings().Enabled
}

// SetEnabled toggles Vietnamese input on/off
func (a *AppBindings) SetEnabled(enabled bool) {
	a.imeLoop.SetEnabled(enabled)
	a.settingsSvc.Settings().Enabled = enabled
	a.settingsSvc.Save()
	// Note: OnEnabledChanged callback in ImeLoop will update tray icon
}

// Toggle toggles IME and returns new state
func (a *AppBindings) Toggle() bool {
	enabled := a.imeLoop.Toggle()
	a.settingsSvc.Settings().Enabled = enabled
	a.settingsSvc.Save()
	return enabled
}

// GetInputMethod returns current input method (0=Telex, 1=VNI)
func (a *AppBindings) GetInputMethod() int {
	return a.settingsSvc.Settings().InputMethod
}

// SetInputMethod sets input method
func (a *AppBindings) SetInputMethod(method int) {
	a.settingsSvc.Settings().InputMethod = method
	a.imeLoop.UpdateSettings(&core.ImeSettings{
		Enabled:     a.settingsSvc.Settings().Enabled,
		InputMethod: core.InputMethod(method),
	})
	a.settingsSvc.Save()
}

// GetSettings returns all settings
func (a *AppBindings) GetSettings() map[string]interface{} {
	s := a.settingsSvc.Settings()
	return map[string]interface{}{
		"enabled":            s.Enabled,
		"inputMethod":        s.InputMethod,
		"modernTone":         s.ModernTone,
		"autoStart":          s.AutoStart,
		"skipWShortcut":      s.SkipWShortcut,
		"escRestore":         s.EscRestore,
		"freeTone":           s.FreeTone,
		"englishAutoRestore": s.EnglishAutoRestore,
		"autoCapitalize":     s.AutoCapitalize,
		"toggleHotkey":       s.ToggleHotkey,
	}
}

// SaveSettings saves all settings
func (a *AppBindings) SaveSettings(settings map[string]interface{}) error {
	s := a.settingsSvc.Settings()

	if v, ok := settings["enabled"].(bool); ok {
		s.Enabled = v
	}
	if v, ok := settings["inputMethod"].(float64); ok {
		s.InputMethod = int(v)
	}
	if v, ok := settings["modernTone"].(bool); ok {
		s.ModernTone = v
	}
	if v, ok := settings["autoStart"].(bool); ok {
		s.AutoStart = v
	}
	if v, ok := settings["skipWShortcut"].(bool); ok {
		s.SkipWShortcut = v
	}
	if v, ok := settings["escRestore"].(bool); ok {
		s.EscRestore = v
	}
	if v, ok := settings["freeTone"].(bool); ok {
		s.FreeTone = v
	}
	if v, ok := settings["englishAutoRestore"].(bool); ok {
		s.EnglishAutoRestore = v
	}
	if v, ok := settings["autoCapitalize"].(bool); ok {
		s.AutoCapitalize = v
	}
	if v, ok := settings["toggleHotkey"].(string); ok {
		s.ToggleHotkey = v
	}

	// Apply to IME loop
	a.imeLoop.UpdateSettings(&core.ImeSettings{
		Enabled:            s.Enabled,
		InputMethod:        core.InputMethod(s.InputMethod),
		ModernTone:         s.ModernTone,
		SkipWShortcut:      s.SkipWShortcut,
		EscRestore:         s.EscRestore,
		FreeTone:           s.FreeTone,
		EnglishAutoRestore: s.EnglishAutoRestore,
		AutoCapitalize:     s.AutoCapitalize,
	})

	// Update hotkey
	keyCode, ctrl, alt, shift := services.ParseHotkey(s.ToggleHotkey)
	a.imeLoop.SetHotkey(keyCode, ctrl, alt, shift)

	return a.settingsSvc.Save()
}

// GetVersion returns app version
func (a *AppBindings) GetVersion() string {
	return Version
}

// GetShortcuts returns all shortcuts
func (a *AppBindings) GetShortcuts() ([]map[string]string, error) {
	shortcuts, err := a.settingsSvc.LoadShortcuts()
	if err != nil {
		return nil, err
	}

	result := make([]map[string]string, len(shortcuts))
	for i, sc := range shortcuts {
		result[i] = map[string]string{
			"trigger":     sc.Trigger,
			"replacement": sc.Replacement,
		}
	}
	return result, nil
}

// SaveShortcuts saves all shortcuts
func (a *AppBindings) SaveShortcuts(shortcuts []map[string]string) error {
	// Clear existing shortcuts in engine
	a.imeLoop.ClearShortcuts()

	// Convert and save
	scs := make([]services.Shortcut, len(shortcuts))
	for i, sc := range shortcuts {
		scs[i] = services.Shortcut{
			Trigger:     sc["trigger"],
			Replacement: sc["replacement"],
		}
		// Add to engine
		a.imeLoop.AddShortcut(sc["trigger"], sc["replacement"])
	}

	return a.settingsSvc.SaveShortcuts(scs)
}

// AddShortcut adds a single shortcut
func (a *AppBindings) AddShortcut(trigger, replacement string) {
	a.imeLoop.AddShortcut(trigger, replacement)
}

// RemoveShortcut removes a shortcut
func (a *AppBindings) RemoveShortcut(trigger string) {
	a.imeLoop.RemoveShortcut(trigger)
}

// --- Update Methods ---

// CheckForUpdates checks GitHub for a newer version
func (a *AppBindings) CheckForUpdates(force bool) (*services.UpdateInfo, error) {
	return a.updaterSvc.CheckForUpdates(force)
}

// OpenReleasePage opens the release page in browser
func (a *AppBindings) OpenReleasePage(url string) error {
	return a.updaterSvc.OpenReleasePage(url)
}
