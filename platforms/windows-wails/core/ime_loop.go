package core

// IME Loop - orchestrates keyboard hook, Rust engine, and text injection
// This is the main integration point for Vietnamese input processing

import (
	"sync"
)

// ImeLoop manages the complete IME processing pipeline
type ImeLoop struct {
	hook     *KeyboardHook
	bridge   *Bridge
	settings *ImeSettings
	running  bool
	mu       sync.Mutex

	// Callbacks for UI notification
	OnEnabledChanged func(enabled bool)
}

// ImeSettings holds runtime IME configuration
type ImeSettings struct {
	Enabled            bool
	InputMethod        InputMethod
	ModernTone         bool
	SkipWShortcut      bool
	BracketShortcut    bool
	EscRestore         bool
	FreeTone           bool
	EnglishAutoRestore bool
	AutoCapitalize     bool
}

// DefaultImeSettings returns default settings
func DefaultImeSettings() *ImeSettings {
	return &ImeSettings{
		Enabled:            true,
		InputMethod:        Telex,
		ModernTone:         true,
		SkipWShortcut:      false,
		BracketShortcut:    false,
		EscRestore:         true,
		FreeTone:           false,
		EnglishAutoRestore: false,
		AutoCapitalize:     true,
	}
}

// NewImeLoop creates a new IME loop
func NewImeLoop() (*ImeLoop, error) {
	bridge, err := GetBridge()
	if err != nil {
		return nil, err
	}

	hook := NewKeyboardHook()
	settings := DefaultImeSettings()

	loop := &ImeLoop{
		hook:     hook,
		bridge:   bridge,
		settings: settings,
	}

	// Set up key processing callback
	hook.OnKeyPressed = loop.processKey

	return loop, nil
}

// Start begins the IME loop
func (l *ImeLoop) Start() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.running {
		return nil
	}

	// Initialize Rust engine
	l.bridge.Initialize()
	l.applySettings()

	// Start keyboard hook
	if err := l.hook.Start(); err != nil {
		return err
	}

	l.running = true
	return nil
}

// Stop ends the IME loop
func (l *ImeLoop) Stop() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.running {
		return
	}

	l.hook.Stop()
	l.running = false
}

// IsRunning returns whether the IME loop is active
func (l *ImeLoop) IsRunning() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.running
}

// SetEnabled enables or disables IME processing
func (l *ImeLoop) SetEnabled(enabled bool) {
	l.settings.Enabled = enabled
	l.bridge.SetEnabled(enabled)

	if l.OnEnabledChanged != nil {
		l.OnEnabledChanged(enabled)
	}
}

// Toggle toggles IME enabled state
func (l *ImeLoop) Toggle() bool {
	newState := !l.settings.Enabled
	l.SetEnabled(newState)
	return newState
}

// SetHotkey sets the toggle hotkey
func (l *ImeLoop) SetHotkey(keyCode uint16, ctrl, alt, shift bool) {
	l.hook.Hotkey = &KeyboardShortcut{
		KeyCode: keyCode,
		Ctrl:    ctrl,
		Alt:     alt,
		Shift:   shift,
	}
	l.hook.OnHotkey = func() {
		l.Toggle()
	}
}

// UpdateSettings applies new settings to the engine
func (l *ImeLoop) UpdateSettings(settings *ImeSettings) {
	l.settings = settings
	l.applySettings()
}

// applySettings syncs settings to Rust engine
func (l *ImeLoop) applySettings() {
	l.bridge.SetEnabled(l.settings.Enabled)
	l.bridge.SetMethod(l.settings.InputMethod)
	l.bridge.SetModernTone(l.settings.ModernTone)
	l.bridge.SetSkipWShortcut(l.settings.SkipWShortcut)
	l.bridge.SetBracketShortcut(l.settings.BracketShortcut)
	l.bridge.SetEscRestore(l.settings.EscRestore)
	l.bridge.SetFreeTone(l.settings.FreeTone)
	l.bridge.SetEnglishAutoRestore(l.settings.EnglishAutoRestore)
	l.bridge.SetAutoCapitalize(l.settings.AutoCapitalize)
}

// ClearBuffer clears the IME buffer
func (l *ImeLoop) ClearBuffer() {
	l.bridge.Clear()
}

// processKey handles a keystroke through the IME pipeline
// Returns true if the key was handled (should be blocked)
func (l *ImeLoop) processKey(keyCode uint16, shift, capsLock bool) bool {
	if !l.settings.Enabled {
		return false
	}

	// Translate Windows VK to macOS keycode for Rust engine
	macKeycode := TranslateToMacKeycode(keyCode)
	if macKeycode == 0xFFFF {
		// Key not mapped, pass through
		return false
	}

	// Process through Rust engine
	result := l.bridge.ProcessKey(macKeycode, capsLock, false, shift)

	switch result.Action {
	case ActionNone:
		// No action needed, pass through
		return false

	case ActionSend:
		// Send replacement text
		text := result.GetText()
		backspaces := int(result.Backspace)
		SendText(text, backspaces)
		return true

	case ActionRestore:
		// Restore original text (ESC pressed)
		text := result.GetText()
		backspaces := int(result.Backspace)
		SendText(text, backspaces)
		return true
	}

	return false
}

// AddShortcut adds a text expansion shortcut
func (l *ImeLoop) AddShortcut(trigger, replacement string) {
	l.bridge.AddShortcut(trigger, replacement)
}

// RemoveShortcut removes a shortcut
func (l *ImeLoop) RemoveShortcut(trigger string) {
	l.bridge.RemoveShortcut(trigger)
}

// ClearShortcuts removes all shortcuts
func (l *ImeLoop) ClearShortcuts() {
	l.bridge.ClearShortcuts()
}
