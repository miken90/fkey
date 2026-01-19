package core

import (
	"log"

	"github.com/jezek/xgb"
	"github.com/jezek/xgb/xproto"
)

// Internal keycodes (matching Rust core's macOS-style codes from core/src/data/keys.rs)
const (
	KEY_A         = 0
	KEY_S         = 1
	KEY_D         = 2
	KEY_F         = 3
	KEY_H         = 4
	KEY_G         = 5
	KEY_Z         = 6
	KEY_X         = 7
	KEY_C         = 8
	KEY_V         = 9
	KEY_B         = 11
	KEY_Q         = 12
	KEY_W         = 13
	KEY_E         = 14
	KEY_R         = 15
	KEY_Y         = 16
	KEY_T         = 17
	KEY_O         = 31
	KEY_U         = 32
	KEY_I         = 34
	KEY_P         = 35
	KEY_L         = 37
	KEY_J         = 38
	KEY_K         = 40
	KEY_N         = 45
	KEY_M         = 46
	KEY_N1        = 18
	KEY_N2        = 19
	KEY_N3        = 20
	KEY_N4        = 21
	KEY_N5        = 23
	KEY_N6        = 22
	KEY_N7        = 26
	KEY_N8        = 28
	KEY_N9        = 25
	KEY_N0        = 29
	KEY_SPACE     = 49
	KEY_DELETE    = 51
	KEY_TAB       = 48
	KEY_RETURN    = 36
	KEY_ESC       = 53
	KEY_DOT       = 47
	KEY_COMMA     = 43
	KEY_SLASH     = 44
	KEY_SEMICOLON = 41
	KEY_QUOTE     = 39
	KEY_LBRACKET  = 33
	KEY_RBRACKET  = 30
	KEY_MINUS     = 27
	KEY_EQUAL     = 24
	KEY_BACKQUOTE = 50
)

// x11ToInternal maps X11 keysyms to internal (macOS-style) keycodes
var x11ToInternal = map[xproto.Keysym]uint16{
	'a': KEY_A, 'b': KEY_B, 'c': KEY_C, 'd': KEY_D, 'e': KEY_E, 'f': KEY_F,
	'g': KEY_G, 'h': KEY_H, 'i': KEY_I, 'j': KEY_J, 'k': KEY_K, 'l': KEY_L,
	'm': KEY_M, 'n': KEY_N, 'o': KEY_O, 'p': KEY_P, 'q': KEY_Q, 'r': KEY_R,
	's': KEY_S, 't': KEY_T, 'u': KEY_U, 'v': KEY_V, 'w': KEY_W, 'x': KEY_X,
	'y': KEY_Y, 'z': KEY_Z,
	'A': KEY_A, 'B': KEY_B, 'C': KEY_C, 'D': KEY_D, 'E': KEY_E, 'F': KEY_F,
	'G': KEY_G, 'H': KEY_H, 'I': KEY_I, 'J': KEY_J, 'K': KEY_K, 'L': KEY_L,
	'M': KEY_M, 'N': KEY_N, 'O': KEY_O, 'P': KEY_P, 'Q': KEY_Q, 'R': KEY_R,
	'S': KEY_S, 'T': KEY_T, 'U': KEY_U, 'V': KEY_V, 'W': KEY_W, 'X': KEY_X,
	'Y': KEY_Y, 'Z': KEY_Z,
	'0': KEY_N0, '1': KEY_N1, '2': KEY_N2, '3': KEY_N3, '4': KEY_N4,
	'5': KEY_N5, '6': KEY_N6, '7': KEY_N7, '8': KEY_N8, '9': KEY_N9,
	' ':    KEY_SPACE,
	0xFF08: KEY_DELETE, // Backspace
	0xFF0D: KEY_RETURN, // Return/Enter
	0xFF1B: KEY_ESC,    // Escape
	0xFF09: KEY_TAB,    // Tab
	'[':    KEY_LBRACKET, ']': KEY_RBRACKET,
	'.': KEY_DOT, ',': KEY_COMMA, '/': KEY_SLASH,
	';': KEY_SEMICOLON, '\'': KEY_QUOTE,
	'-': KEY_MINUS, '=': KEY_EQUAL,
	'`': KEY_BACKQUOTE,
}

// KeyboardHandler manages X11 keyboard interception
type KeyboardHandler struct {
	conn       *xgb.Conn
	bridge     *Bridge
	textSender *TextSender
	root       xproto.Window
	running    bool
	stopChan   chan struct{}

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

	setup := xproto.Setup(conn)
	root := setup.DefaultScreen(conn).Root

	return &KeyboardHandler{
		conn:       conn,
		bridge:     bridge,
		textSender: NewTextSender(),
		root:       root,
		stopChan:   make(chan struct{}),
		HotkeyMod:  xproto.ModMaskControl, // Default: Ctrl+Space
		HotkeyCode: 65,                    // Space keycode (typical)
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
		65: ' ',    // Space
		22: 0xFF08, // Backspace
		36: 0xFF0D, // Return
		9:  0xFF1B, // Escape
		23: 0xFF09, // Tab
		34: '[', 35: ']',
		60: '.', 59: ',', 61: '/',
		47: ';', 48: '\'',
		20: '-', 21: '=',
		49: '`', // Backquote/Grave
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
	if err := h.textSender.SendText(text, backspaces); err != nil {
		log.Printf("sendText error: %v", err)
	}
}
