package core

/*
#cgo LDFLAGS: -L${SRCDIR}/../../../core/target/release -lgonhanh_core -ldl -lm
#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>

// FFI declarations matching Rust core
void ime_init(void);
void ime_method(uint8_t method);
void ime_enabled(bool enabled);
void ime_modern(bool modern);
void ime_esc_restore(bool enabled);
void ime_clear(void);
void ime_free(void* ptr);

typedef struct {
    uint8_t action;      // 0=None, 1=Send, 2=Restore
    uint8_t backspace;
    uint32_t chars[64];
    uint8_t count;
} ImeResult;

ImeResult* ime_key(uint16_t key, bool caps, bool ctrl);
ImeResult* ime_key_ext(uint16_t key, bool caps, bool ctrl, bool shift);
*/
import "C"
import (
	"log"
	"sync"
	"unsafe"
)

// Action types from Rust core
const (
	ActionNone    = 0
	ActionSend    = 1
	ActionRestore = 2
)

// InputMethod types
const (
	MethodTelex = 0
	MethodVNI   = 1
)

// Result represents IME processing result
type Result struct {
	Action    uint8
	Backspace uint8
	Text      string
}

// Bridge wraps the Rust IME core via FFI
type Bridge struct {
	mu      sync.Mutex
	enabled bool
}

// NewBridge creates and initializes the IME bridge
func NewBridge() (*Bridge, error) {
	C.ime_init()
	return &Bridge{enabled: true}, nil
}

// Close cleans up the bridge
func (b *Bridge) Close() {
	// Rust core doesn't need explicit cleanup
}

// ProcessKey sends a key event to the IME core
func (b *Bridge) ProcessKey(keycode uint16, capsLock, ctrl, shift bool) *Result {
	b.mu.Lock()
	defer b.mu.Unlock()

	log.Printf("[DEBUG Bridge] ProcessKey: keycode=%d, caps=%v, ctrl=%v, shift=%v, enabled=%v",
		keycode, capsLock, ctrl, shift, b.enabled)

	if !b.enabled {
		log.Printf("[DEBUG Bridge] IME disabled, returning nil")
		return nil
	}

	var r *C.ImeResult
	if shift {
		r = C.ime_key_ext(C.uint16_t(keycode), C.bool(capsLock), C.bool(ctrl), C.bool(shift))
	} else {
		r = C.ime_key(C.uint16_t(keycode), C.bool(capsLock), C.bool(ctrl))
	}

	if r == nil {
		return nil
	}
	defer C.ime_free(unsafe.Pointer(r))

	if r.action == ActionNone {
		return nil
	}

	// Convert UTF-32 chars to string
	var runes []rune
	for i := 0; i < int(r.count); i++ {
		runes = append(runes, rune(r.chars[i]))
	}

	return &Result{
		Action:    uint8(r.action),
		Backspace: uint8(r.backspace),
		Text:      string(runes),
	}
}

// SetMethod sets input method (0=Telex, 1=VNI)
func (b *Bridge) SetMethod(method int) {
	C.ime_method(C.uint8_t(method))
}

// SetEnabled enables/disables the IME
func (b *Bridge) SetEnabled(enabled bool) {
	b.mu.Lock()
	b.enabled = enabled
	b.mu.Unlock()
	C.ime_enabled(C.bool(enabled))
}

// SetModernTone sets modern tone placement
func (b *Bridge) SetModernTone(modern bool) {
	C.ime_modern(C.bool(modern))
}

// SetEscRestore sets ESC restore behavior
func (b *Bridge) SetEscRestore(enabled bool) {
	C.ime_esc_restore(C.bool(enabled))
}

// Clear resets the input buffer
func (b *Bridge) Clear() {
	C.ime_clear()
}

// IsEnabled returns current enabled state
func (b *Bridge) IsEnabled() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.enabled
}

// Toggle toggles IME on/off
func (b *Bridge) Toggle() bool {
	b.mu.Lock()
	b.enabled = !b.enabled
	enabled := b.enabled
	b.mu.Unlock()
	C.ime_enabled(C.bool(enabled))
	return enabled
}
