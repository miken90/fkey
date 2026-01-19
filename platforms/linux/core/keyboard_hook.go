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
// On Linux, gohook returns ASCII codes as rawcodes
var rawcodeToInternal = map[uint16]uint16{
	// Lowercase letters (ASCII)
	'a': KEY_A, 'b': KEY_B, 'c': KEY_C, 'd': KEY_D, 'e': KEY_E, 'f': KEY_F,
	'g': KEY_G, 'h': KEY_H, 'i': KEY_I, 'j': KEY_J, 'k': KEY_K, 'l': KEY_L,
	'm': KEY_M, 'n': KEY_N, 'o': KEY_O, 'p': KEY_P, 'q': KEY_Q, 'r': KEY_R,
	's': KEY_S, 't': KEY_T, 'u': KEY_U, 'v': KEY_V, 'w': KEY_W, 'x': KEY_X,
	'y': KEY_Y, 'z': KEY_Z,
	// Uppercase letters (ASCII)
	'A': KEY_A, 'B': KEY_B, 'C': KEY_C, 'D': KEY_D, 'E': KEY_E, 'F': KEY_F,
	'G': KEY_G, 'H': KEY_H, 'I': KEY_I, 'J': KEY_J, 'K': KEY_K, 'L': KEY_L,
	'M': KEY_M, 'N': KEY_N, 'O': KEY_O, 'P': KEY_P, 'Q': KEY_Q, 'R': KEY_R,
	'S': KEY_S, 'T': KEY_T, 'U': KEY_U, 'V': KEY_V, 'W': KEY_W, 'X': KEY_X,
	'Y': KEY_Y, 'Z': KEY_Z,
	// Numbers (ASCII)
	'0': KEY_N0, '1': KEY_N1, '2': KEY_N2, '3': KEY_N3, '4': KEY_N4,
	'5': KEY_N5, '6': KEY_N6, '7': KEY_N7, '8': KEY_N8, '9': KEY_N9,
	// Special keys - use ASCII values
	' ': KEY_SPACE, // Space = 32
	8:  KEY_DELETE, // Backspace = ASCII 8
	13: KEY_RETURN, // Enter/Return = ASCII 13
	27: KEY_ESC,    // Escape = ASCII 27
	9:  KEY_TAB,    // Tab = ASCII 9
	// Punctuation (ASCII)
	'.': KEY_DOT, ',': KEY_COMMA, '/': KEY_SLASH,
	';': KEY_SEMICOLON, '\'': KEY_QUOTE,
	'[': KEY_LBRACKET, ']': KEY_RBRACKET,
	'-': KEY_MINUS, '=': KEY_EQUAL,
	'`': KEY_BACKQUOTE,
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
	// Debug: log all key events
	log.Printf("[DEBUG] Key event: rawcode=%d, keychar=%c (%d), mask=0x%x",
		ev.Rawcode, ev.Keychar, ev.Keychar, ev.Mask)

	// Map rawcode to internal code
	internalCode, ok := rawcodeToInternal[ev.Rawcode]
	if !ok {
		log.Printf("[DEBUG] Unmapped rawcode: %d", ev.Rawcode)
		return // Not a key we care about
	}

	log.Printf("[DEBUG] Mapped to internal code: %d", internalCode)

	// Get modifier states from gohook
	// ev.Mask contains modifier info
	shift := (ev.Mask & 0x01) != 0 // Shift
	ctrl := (ev.Mask & 0x04) != 0  // Ctrl
	// capsLock detection via gohook is limited, assume false for now
	capsLock := false

	// Skip if Ctrl is pressed (except already handled by hotkey)
	if ctrl {
		log.Printf("[DEBUG] Ctrl pressed, clearing buffer")
		h.bridge.Clear()
		return
	}

	// Determine if uppercase based on shift or character
	if ev.Keychar >= 'A' && ev.Keychar <= 'Z' {
		shift = true
	}

	log.Printf("[DEBUG] Processing key: code=%d, shift=%v, caps=%v", internalCode, shift, capsLock)

	// Process through IME
	result := h.bridge.ProcessKey(internalCode, capsLock, ctrl, shift)
	if result == nil {
		log.Printf("[DEBUG] IME returned nil (no action)")
		return
	}

	log.Printf("[DEBUG] IME result: action=%d, backspace=%d, text=%q",
		result.Action, result.Backspace, result.Text)

	// Send replacement text
	// IMPORTANT: gohook doesn't suppress original keystrokes, so the original
	// character has already been typed. We need to add 1 extra backspace to
	// delete it before sending our replacement text.
	if result.Action == ActionSend || result.Action == ActionRestore {
		// Add 1 to backspace count to delete the original keystroke
		totalBackspace := int(result.Backspace) + 1
		log.Printf("Sending: backspace=%d (+1 for original key), text=%q", totalBackspace, result.Text)
		h.sendText(result.Text, totalBackspace)
	}
}

func (h *KeyboardHandler) sendText(text string, backspaces int) {
	if err := h.textSender.SendText(text, backspaces); err != nil {
		log.Printf("sendText error: %v", err)
	}
}
