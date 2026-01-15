package core

// Text injection using Windows SendInput API
// Port of TextSender.cs from .NET implementation

import (
	"time"
	"unsafe"
)

const (
	INPUT_KEYBOARD    = 1
	KEYEVENTF_KEYUP   = 0x0002
	KEYEVENTF_UNICODE = 0x0004
)

// InjectionMethod determines how text is injected
type InjectionMethod int

const (
	MethodFast      InjectionMethod = iota // Batch injection for standard apps
	MethodSlow                             // Per-character with small delays
	MethodExtraSlow                        // Per-character with larger delays (Discord, Wave)
)

// Delay settings (milliseconds)
const (
	// Slow mode - most Electron apps, browsers, terminals
	// Uses per-character sending but no delays
	SlowModeKeyDelay  = 0
	SlowModePreDelay  = 0
	SlowModePostDelay = 0

	// Extra slow mode - problematic apps like Wave
	ExtraSlowModeKeyDelay  = 5
	ExtraSlowModePreDelay  = 10
	ExtraSlowModePostDelay = 8

	FastModeDelay = 0 // Not used anymore
)

// INPUT structure for SendInput
type INPUT struct {
	Type uint32
	Ki   KEYBDINPUT
	_    [8]byte // padding to match 40-byte size on 64-bit
}

// KEYBDINPUT structure
type KEYBDINPUT struct {
	WVk         uint16
	WScan       uint16
	DwFlags     uint32
	Time        uint32
	DwExtraInfo uintptr
}

// Win32 API
var (
	procSendInput = user32.NewProc("SendInput")
)

// inputSize is the size of INPUT structure
var inputSize = unsafe.Sizeof(INPUT{})

// SendText sends text replacement to the active window
func SendText(text string, backspaces int) {
	method := DetectInjectionMethod()
	SendTextWithMethod(text, backspaces, method)
}

// SendTextWithMethod sends text with specific injection method
func SendTextWithMethod(text string, backspaces int, method InjectionMethod) {
	if len(text) == 0 && backspaces == 0 {
		return
	}

	switch method {
	case MethodFast:
		sendFast(text, backspaces)
	case MethodSlow:
		sendSlow(text, backspaces, SlowModePreDelay, SlowModePostDelay, SlowModeKeyDelay)
	case MethodExtraSlow:
		sendSlow(text, backspaces, ExtraSlowModePreDelay, ExtraSlowModePostDelay, ExtraSlowModeKeyDelay)
	}
}

func sendFast(text string, backspaces int) {
	if backspaces > 0 {
		sendBackspaces(backspaces)
	}
	if len(text) > 0 {
		sendUnicodeTextBatch(text)
	}
}

func sendSlow(text string, backspaces int, preDelay, postDelay, keyDelay int) {
	if backspaces > 0 {
		sendBackspaces(backspaces)
		time.Sleep(time.Duration(postDelay) * time.Millisecond)
	}
	if len(text) > 0 {
		time.Sleep(time.Duration(preDelay) * time.Millisecond)
		sendUnicodeTextSlow(text, keyDelay)
	}
}

func sendBackspaces(count int) {
	inputs := make([]INPUT, count*2)

	for i := 0; i < count; i++ {
		// Key down
		inputs[i*2] = INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_BACK,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		}

		// Key up
		inputs[i*2+1] = INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_BACK,
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		}
	}

	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}

func sendUnicodeTextBatch(text string) {
	runes := []rune(text)
	inputs := make([]INPUT, len(runes)*2)
	idx := 0

	for _, r := range runes {
		// Key down
		inputs[idx] = INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         0,
				WScan:       uint16(r),
				DwFlags:     KEYEVENTF_UNICODE,
				DwExtraInfo: InjectedKeyMarker,
			},
		}
		idx++

		// Key up
		inputs[idx] = INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         0,
				WScan:       uint16(r),
				DwFlags:     KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		}
		idx++
	}

	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}

func sendUnicodeTextSlow(text string, delayMs int) {
	runes := []rune(text)

	for _, r := range runes {
		inputs := [2]INPUT{
			// Key down
			{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(r),
					DwFlags:     KEYEVENTF_UNICODE,
					DwExtraInfo: InjectedKeyMarker,
				},
			},
			// Key up
			{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         0,
					WScan:       uint16(r),
					DwFlags:     KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
					DwExtraInfo: InjectedKeyMarker,
				},
			},
		}

		procSendInput.Call(
			2,
			uintptr(unsafe.Pointer(&inputs[0])),
			uintptr(inputSize),
		)

		if delayMs > 0 {
			time.Sleep(time.Duration(delayMs) * time.Millisecond)
		}
	}
}

// DetectInjectionMethod determines the best method for current foreground app
func DetectInjectionMethod() InjectionMethod {
	return GetInjectionMethod()
}
