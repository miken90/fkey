package core

import (
	"log"

	hook "github.com/robotn/gohook"
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

// gohook rawcode to internal keycode mapping
var rawcodeToInternal = map[uint16]uint16{
	// Letters (gohook rawcodes for Linux)
	38: KEY_A, 56: KEY_B, 54: KEY_C, 40: KEY_D, 26: KEY_E, 41: KEY_F,
	42: KEY_G, 43: KEY_H, 31: KEY_I, 44: KEY_J, 45: KEY_K, 46: KEY_L,
	58: KEY_M, 57: KEY_N, 32: KEY_O, 33: KEY_P, 24: KEY_Q, 27: KEY_R,
	39: KEY_S, 28: KEY_T, 30: KEY_U, 55: KEY_V, 25: KEY_W, 53: KEY_X,
	29: KEY_Y, 52: KEY_Z,
	// Numbers
	10: KEY_N1, 11: KEY_N2, 12: KEY_N3, 13: KEY_N4, 14: KEY_N5,
	15: KEY_N6, 16: KEY_N7, 17: KEY_N8, 18: KEY_N9, 19: KEY_N0,
	// Special keys
	65: KEY_SPACE,
	22: KEY_DELETE, // Backspace
	36: KEY_RETURN,
	9:  KEY_ESC,
	23: KEY_TAB,
	// Punctuation
	34: KEY_LBRACKET, 35: KEY_RBRACKET,
	60: KEY_DOT, 59: KEY_COMMA, 61: KEY_SLASH,
	47: KEY_SEMICOLON, 48: KEY_QUOTE,
	20: KEY_MINUS, 21: KEY_EQUAL,
	49: KEY_BACKQUOTE,
}

// KeyboardHandler manages global keyboard hook using gohook
type KeyboardHandler struct {
	bridge     *Bridge
	textSender *TextSender
	running    bool
	evChan     chan hook.Event
}

// NewKeyboardHandler creates keyboard handler using gohook
func NewKeyboardHandler(bridge *Bridge) (*KeyboardHandler, error) {
	return &KeyboardHandler{
		bridge:     bridge,
		textSender: NewTextSender(),
		running:    false,
	}, nil
}

// Start begins keyboard event processing
func (h *KeyboardHandler) Start() error {
	h.running = true

	// Register Ctrl+Space toggle hotkey
	hook.Register(hook.KeyDown, []string{"space", "ctrl"}, func(e hook.Event) {
		enabled := h.bridge.Toggle()
		log.Printf("IME toggled: %v", enabled)
	})

	log.Println("Global keyboard hook starting...")

	h.evChan = hook.Start()

	for ev := range h.evChan {
		if !h.running {
			break
		}

		// Only process KeyDown events
		if ev.Kind != hook.KeyDown {
			continue
		}

		h.handleKeyEvent(ev)
	}

	return nil
}

// Stop ends keyboard processing
func (h *KeyboardHandler) Stop() {
	h.running = false
	hook.End()
}

func (h *KeyboardHandler) handleKeyEvent(ev hook.Event) {
	// Map rawcode to internal code
	internalCode, ok := rawcodeToInternal[ev.Rawcode]
	if !ok {
		return // Not a key we care about
	}

	// Get modifier states from gohook
	// ev.Mask contains modifier info
	shift := (ev.Mask & 0x01) != 0  // Shift
	ctrl := (ev.Mask & 0x04) != 0   // Ctrl
	// capsLock detection via gohook is limited, assume false for now
	capsLock := false

	// Skip if Ctrl is pressed (except already handled by hotkey)
	if ctrl {
		h.bridge.Clear()
		return
	}

	// Determine if uppercase based on shift or character
	if ev.Keychar >= 'A' && ev.Keychar <= 'Z' {
		shift = true
	}

	// Process through IME
	result := h.bridge.ProcessKey(internalCode, capsLock, ctrl, shift)
	if result == nil {
		return
	}

	// Send replacement text
	if result.Action == ActionSend || result.Action == ActionRestore {
		log.Printf("Sending: backspace=%d, text=%q", result.Backspace, result.Text)
		h.sendText(result.Text, int(result.Backspace))
	}
}

func (h *KeyboardHandler) sendText(text string, backspaces int) {
	if err := h.textSender.SendText(text, backspaces); err != nil {
		log.Printf("sendText error: %v", err)
	}
}
