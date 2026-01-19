package core

import (
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
	"github.com/jezek/xgb/xtest"
)

// X11 virtual keycodes (matching Rust core's macOS-style codes)
// We need to map X11 keysym to our internal codes
var x11ToInternal = map[xproto.Keysym]uint16{
	'a': 0, 'b': 11, 'c': 8, 'd': 2, 'e': 14, 'f': 3,
	'g': 5, 'h': 4, 'i': 34, 'j': 38, 'k': 40, 'l': 37,
	'm': 46, 'n': 45, 'o': 31, 'p': 35, 'q': 12, 'r': 15,
	's': 1, 't': 17, 'u': 32, 'v': 9, 'w': 13, 'x': 7,
	'y': 16, 'z': 6,
	'A': 0, 'B': 11, 'C': 8, 'D': 2, 'E': 14, 'F': 3,
	'G': 5, 'H': 4, 'I': 34, 'J': 38, 'K': 40, 'L': 37,
	'M': 46, 'N': 45, 'O': 31, 'P': 35, 'Q': 12, 'R': 15,
	'S': 1, 'T': 17, 'U': 32, 'V': 9, 'W': 13, 'X': 7,
	'Y': 16, 'Z': 6,
	'0': 29, '1': 18, '2': 19, '3': 20, '4': 21,
	'5': 23, '6': 22, '7': 26, '8': 28, '9': 25,
	' ':  49,  // Space
	0xFF08: 51, // Backspace
	0xFF0D: 36, // Return/Enter
	0xFF1B: 53, // Escape
	'[': 33, ']': 30,
	'.': 47, ',': 43, '/': 44,
	';': 41, '\'': 39,
	'-': 27, '=': 24,
}

// KeyboardHandler manages X11 keyboard interception
type KeyboardHandler struct {
	conn     *xgb.Conn
	bridge   *Bridge
	root     xproto.Window
	running  bool
	stopChan chan struct{}

	// Hotkey config
	HotkeyMod  uint16 // Modifier mask (e.g., ControlMask)
	HotkeyCode xproto.Keycode
}

// NewKeyboardHandler creates X11 keyboard handler
func NewKeyboardHandler(bridge *Bridge) (*KeyboardHandler, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	// Initialize XTest extension for key injection
	if err := xtest.Init(conn); err != nil {
		conn.Close()
		return nil, err
	}

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	return &KeyboardHandler{
		conn:       conn,
		bridge:     bridge,
		root:       root,
		stopChan:   make(chan struct{}),
		HotkeyMod:  xproto.ModMaskControl, // Default: Ctrl+Space
		HotkeyCode: 65,                     // Space keycode (typical)
	}, nil
}

// Start begins keyboard event processing
func (h *KeyboardHandler) Start() error {
	h.running = true

	// Grab the toggle hotkey globally
	h.grabHotkey()

	// Select keyboard events on root window
	mask := uint32(xproto.EventMaskKeyPress | xproto.EventMaskKeyRelease)
	xproto.ChangeWindowAttributes(h.conn, h.root, xproto.CwEventMask, []uint32{mask})

	log.Println("X11 keyboard handler started")

	for h.running {
		select {
		case <-h.stopChan:
			return nil
		default:
			ev, err := h.conn.WaitForEvent()
			if err != nil {
				log.Printf("X11 event error: %v", err)
				continue
			}
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case xproto.KeyPressEvent:
				h.handleKeyPress(e)
			}
		}
	}

	return nil
}

// Stop ends keyboard processing
func (h *KeyboardHandler) Stop() {
	h.running = false
	close(h.stopChan)
	h.ungrabHotkey()
	h.conn.Close()
}

func (h *KeyboardHandler) grabHotkey() {
	xproto.GrabKey(h.conn, true, h.root,
		h.HotkeyMod, h.HotkeyCode,
		xproto.GrabModeAsync, xproto.GrabModeAsync)
}

func (h *KeyboardHandler) ungrabHotkey() {
	xproto.UngrabKey(h.conn, h.HotkeyCode, h.root, h.HotkeyMod)
}

func (h *KeyboardHandler) handleKeyPress(e xproto.KeyPressEvent) {
	// Check for toggle hotkey
	if e.Detail == h.HotkeyCode && (e.State&h.HotkeyMod) != 0 {
		enabled := h.bridge.Toggle()
		log.Printf("IME toggled: %v", enabled)
		// TODO: Notify tray to update icon
		return
	}

	// Get keysym from keycode
	keysym := h.keycodeToKeysym(e.Detail, e.State)
	
	// Map to internal code
	internalCode, ok := x11ToInternal[keysym]
	if !ok {
		return // Not a key we care about
	}

	// Get modifier states
	shift := (e.State & xproto.ModMaskShift) != 0
	capsLock := (e.State & xproto.ModMaskLock) != 0
	ctrl := (e.State & xproto.ModMaskControl) != 0

	// Skip if Ctrl is pressed (except for our processing)
	if ctrl {
		h.bridge.Clear()
		return
	}

	// Process through IME
	result := h.bridge.ProcessKey(internalCode, capsLock, ctrl, shift)
	if result == nil {
		return
	}

	// Send replacement text
	if result.Action == ActionSend {
		h.sendText(result.Text, int(result.Backspace))
	} else if result.Action == ActionRestore {
		h.sendText(result.Text, int(result.Backspace))
	}
}

func (h *KeyboardHandler) keycodeToKeysym(keycode xproto.Keycode, state uint16) xproto.Keysym {
	// Simplified keysym lookup - real implementation needs proper keyboard mapping
	// This is a placeholder - actual implementation needs XKB or manual mapping
	
	// For MVP, assume US keyboard layout
	// Real implementation: use xkb extension or read keyboard mapping
	
	shift := (state & xproto.ModMaskShift) != 0
	
	// Basic mapping for common keys (keycode -> keysym)
	// These keycodes are typical for US QWERTY on X11
	baseMap := map[xproto.Keycode]xproto.Keysym{
		38: 'a', 56: 'b', 54: 'c', 40: 'd', 26: 'e', 41: 'f',
		42: 'g', 43: 'h', 31: 'i', 44: 'j', 45: 'k', 46: 'l',
		58: 'm', 57: 'n', 32: 'o', 33: 'p', 24: 'q', 27: 'r',
		39: 's', 28: 't', 30: 'u', 55: 'v', 25: 'w', 53: 'x',
		29: 'y', 52: 'z',
		10: '1', 11: '2', 12: '3', 13: '4', 14: '5',
		15: '6', 16: '7', 17: '8', 18: '9', 19: '0',
		65: ' ',   // Space
		22: 0xFF08, // Backspace
		36: 0xFF0D, // Return
		9:  0xFF1B, // Escape
		34: '[', 35: ']',
		60: '.', 59: ',', 61: '/',
		47: ';', 48: '\'',
		20: '-', 21: '=',
	}

	keysym, ok := baseMap[keycode]
	if !ok {
		return 0
	}

	// Handle shift for letters
	if shift && keysym >= 'a' && keysym <= 'z' {
		return keysym - 32 // Convert to uppercase
	}

	return keysym
}

func (h *KeyboardHandler) sendText(text string, backspaces int) {
	// Send backspaces first
	for i := 0; i < backspaces; i++ {
		h.sendKey(22, false) // Backspace keycode
		h.sendKey(22, true)
	}

	// Send text using XTest
	// For Unicode, we need to use XIM or commit directly
	// MVP: Use xdotool-style approach or XTest for ASCII
	
	for _, r := range text {
		h.sendUnicodeChar(r)
	}
}

func (h *KeyboardHandler) sendKey(keycode xproto.Keycode, release bool) {
	eventType := byte(xproto.KeyPress)
	if release {
		eventType = byte(xproto.KeyRelease)
	}
	
	xtest.FakeInput(h.conn, eventType, byte(keycode), 0, h.root, 0, 0, 0)
	h.conn.Sync()
}

func (h *KeyboardHandler) sendUnicodeChar(r rune) {
	// For Unicode characters, we use XTest with Unicode input
	// This is simplified - real implementation needs ibus/fcitx or XIM
	
	// MVP approach: Use xdotool via exec (simple but works)
	// Better approach: Direct XTest with proper keysym mapping
	
	// For now, try to find keycode for this rune
	// This won't work for Vietnamese characters without proper setup
	
	// TODO: Implement proper Unicode input via:
	// 1. XIM (X Input Method)
	// 2. ibus client
	// 3. xdotool exec fallback
	
	log.Printf("sendUnicodeChar: %c (0x%X) - needs implementation", r, r)
}
