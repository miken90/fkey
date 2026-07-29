package core

// Text injection using Windows SendInput API
// Port of TextSender.cs from .NET implementation

import (
	"log"
	"sync"
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
	MethodFast        InjectionMethod = iota // Separate calls with small delay (most apps)
	MethodSlow                               // Per-character with delays (Electron, browsers)
	MethodAtomic                             // Single atomic SendInput (Discord - no flicker)
	MethodPaste                              // Clipboard + Ctrl+V (Warp terminal workaround)
	MethodPassthrough                        // Skip IME processing entirely (remote desktop apps like Parsec)
	MethodPasteShiftV                        // Clipboard + Ctrl+Shift+V (WSLg Linux terminals via msrdc.exe)
)

// Delay settings (milliseconds)
const (
	// Fast mode - standard apps, small delay between backspaces and text
	FastModeDelay = 5

	// Slow mode - Electron apps, browsers, terminals
	SlowModeKeyDelay  = 5
	SlowModePreDelay  = 20
	SlowModePostDelay = 15

	// WSLg (msrdc.exe) - the RDP virtual channel does not forward KEYEVENTF_UNICODE
	// events to the Linux guest, so text is injected via clipboard + Ctrl+Shift+V.
	// The Windows clipboard is mirrored to the WSLg guest over the RDP clipboard
	// channel, which has some latency; wait before pasting and before restoring.
	// These run synchronously inside the low-level keyboard hook, so the total
	// (sync + restore) must stay well under the ~300ms LowLevelHooksTimeout.
	WSLgClipboardSyncDelay = 90 // wait for Windows clipboard format list to reach the WSLg guest
	WSLgPasteRestoreDelay  = 40 // wait for guest to consume paste before restoring clipboard

	// Clipboard restore is deferred to a background goroutine (see sendPaste/sendPasteShiftV)
	// because Ctrl+V / Ctrl+Shift+V is delivered asynchronously: the target app reads the
	// clipboard on its own message loop, which can take longer than the hook's synchronous
	// budget allows. Restoring too early races the paste and the app ends up reading back
	// whatever was on the clipboard *before* we set our text (only observable when the
	// clipboard was non-empty beforehand, since an empty saved value is never restored).
	ClipboardRestoreDelay = 300
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
	case MethodAtomic:
		sendAtomic(text, backspaces)
	case MethodPaste:
		sendPaste(text, backspaces)
	case MethodPasteShiftV:
		sendPasteShiftV(text, backspaces)
	}
}

// Virtual key code for V (Ctrl+V paste)
const VK_V = 0x56

// Clipboard-restore coordination for paste-based injection (sendPaste / sendPasteShiftV).
//
// Each composing keystroke overwrites the clipboard with the composed text and then pastes
// it. The user's real clipboard must be restored afterwards. Restoring per keystroke (one
// goroutine each, fixed delay) races the target's ASYNCHRONOUS paste consumption — under
// WSLg the Ctrl+Shift+V travels the RDP channel and the guest reads the clipboard well
// after our SetClipboardText returns. A restore fired in that gap makes the guest paste the
// stale saved content instead of the diacritic (only observable with a non-empty clipboard,
// which is why the bug was intermittent). Overlapping bursts made it worse: each keystroke's
// restore captured the *previous* keystroke's composed text and its timer could fire mid-paste
// of a later keystroke.
//
// Fix: coordinate a SINGLE debounced restore of the user's ORIGINAL clipboard. The original
// is captured only at the start of a burst; the timer is reset on every keystroke so it only
// fires once typing goes idle (no paste can still be in flight); and it restores only if our
// injected text is still the clipboard owner — verified via the clipboard sequence number, so
// a fresh user copy (or the guest touching the clipboard) is never clobbered.
var (
	pasteRestoreMu     sync.Mutex
	pasteRestoreTimer  *time.Timer
	pasteSavedOriginal string
	pasteOwnClipboard  bool
	pasteLastSetSeq    uint32
)

// scheduleClipboardRestore records that we just set our composed text on the clipboard and
// (re)arms the single debounced restore of the user's original clipboard. savedBefore is the
// clipboard content captured immediately before this keystroke's SetClipboardText; it is only
// adopted as the original on the first keystroke of a burst (otherwise it is our own
// previously-injected text and must not become what we restore).
func scheduleClipboardRestore(savedBefore string) {
	pasteRestoreMu.Lock()
	defer pasteRestoreMu.Unlock()

	if !pasteOwnClipboard {
		pasteSavedOriginal = savedBefore
		pasteOwnClipboard = true
	}

	// Record the sequence number of our just-set clipboard so the deferred restore can tell
	// whether our text is still there (reads do not bump the sequence number; only writes do).
	pasteLastSetSeq = GetClipboardSequenceNumber()

	if pasteRestoreTimer != nil {
		pasteRestoreTimer.Stop()
	}
	pasteRestoreTimer = time.AfterFunc(ClipboardRestoreDelay*time.Millisecond, func() {
		pasteRestoreMu.Lock()
		defer pasteRestoreMu.Unlock()
		// Only restore if our injected text is still on the clipboard untouched. If the
		// sequence number changed, the guest/app or a fresh user copy replaced it and
		// restoring would clobber legitimate content.
		if pasteSavedOriginal != "" && GetClipboardSequenceNumber() == pasteLastSetSeq {
			SetClipboardText(pasteSavedOriginal)
		}
		pasteOwnClipboard = false
		pasteSavedOriginal = ""
	})
}

// sendPasteShiftV injects text via clipboard + Ctrl+Shift+V.
// Used for WSLg-hosted Linux terminals (wezterm, etc.) shown on Windows via msrdc.exe.
// The RDP virtual channel forwards VK keystrokes (VK_BACK, Ctrl+Shift+V) that carry real
// scancodes to the Linux guest, but drops KEYEVENTF_UNICODE events, so normal Unicode
// injection loses characters. The Windows clipboard is mirrored into the guest, so pasting
// the composed text is reliable. Ctrl+Shift+V is the standard paste shortcut for Linux
// terminal emulators (Ctrl+V is a control code there).
func sendPasteShiftV(text string, backspaces int) {
	// Delete the raw characters already echoed by the guest.
	if backspaces > 0 {
		sendBackspaces(backspaces)
		time.Sleep(FastModeDelay * time.Millisecond)
	}

	if len(text) == 0 {
		return
	}

	savedClipboard, _ := GetClipboardText()

	if err := SetClipboardText(text); err != nil {
		// Fallback: try direct Unicode (may drop chars, but better than nothing)
		sendUnicodeTextSlow(text, SlowModeKeyDelay)
		return
	}

	// Wait for the Windows clipboard to propagate to the WSLg guest over RDP.
	time.Sleep(WSLgClipboardSyncDelay * time.Millisecond)

	sendCtrlShiftV()

	// Restore the user's original clipboard via a single debounced, sequence-guarded restore
	// (see scheduleClipboardRestore) so a stale restore can never clobber the clipboard while
	// the guest is still consuming this or a later paste.
	scheduleClipboardRestore(savedClipboard)
}

// sendCtrlShiftV sends a Ctrl+Shift+V keystroke
func sendCtrlShiftV() {
	inputs := [6]INPUT{
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_CONTROL, DwFlags: 0, DwExtraInfo: InjectedKeyMarker}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_SHIFT, DwFlags: 0, DwExtraInfo: InjectedKeyMarker}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_V, DwFlags: 0, DwExtraInfo: InjectedKeyMarker}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_V, DwFlags: KEYEVENTF_KEYUP, DwExtraInfo: InjectedKeyMarker}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_SHIFT, DwFlags: KEYEVENTF_KEYUP, DwExtraInfo: InjectedKeyMarker}},
		{Type: INPUT_KEYBOARD, Ki: KEYBDINPUT{WVk: VK_CONTROL, DwFlags: KEYEVENTF_KEYUP, DwExtraInfo: InjectedKeyMarker}},
	}

	procSendInput.Call(
		6,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}

// sendPaste injects text via clipboard + Ctrl+V
// Used for apps that don't render KEYEVENTF_UNICODE properly (e.g., Warp terminal)
func sendPaste(text string, backspaces int) {
	// Send backspaces first
	if backspaces > 0 {
		sendBackspaces(backspaces)
		time.Sleep(FastModeDelay * time.Millisecond)
	}

	if len(text) == 0 {
		return
	}

	// Save current clipboard content
	savedClipboard, _ := GetClipboardText()

	// Set new text to clipboard
	if err := SetClipboardText(text); err != nil {
		// Fallback to slow mode if clipboard fails
		sendUnicodeTextSlow(text, SlowModeKeyDelay)
		return
	}

	// Small delay to ensure clipboard is set
	time.Sleep(10 * time.Millisecond)

	// Send Ctrl+V
	sendCtrlV()

	// Restore the user's original clipboard via a single debounced, sequence-guarded restore
	// (see scheduleClipboardRestore) so a stale restore can never clobber the clipboard while
	// the app is still consuming this or a later paste.
	scheduleClipboardRestore(savedClipboard)
}

// sendCtrlV sends Ctrl+V keystroke
func sendCtrlV() {
	inputs := [4]INPUT{
		// Ctrl down
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_CONTROL,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// V down
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_V,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// V up
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_V,
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
		// Ctrl up
		{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_CONTROL,
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		},
	}

	procSendInput.Call(
		4,
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}

func sendFast(text string, backspaces int) {
	// Previously: separate SendInput calls (backspaces, sleep, text). Under fast typing
	// this split let the target's message loop fall behind between the two calls, and the
	// next keystroke's replacement (computed by the Rust engine assuming the prior one had
	// already landed) could then desync from what was actually on screen — diacritics land
	// on the wrong char or get dropped. A single atomic SendInput call removes the gap:
	// Windows guarantees backspace+text events within one call are delivered together and
	// in order, with no artificial delay for other input to interleave.
	sendAtomic(text, backspaces)
}

// sendSlow separates backspaces from text injection with a settling delay for apps that
// mishandle rapid/batched Unicode input (Electron, browsers, terminals).
//
// Bug fix: this used to sleep(postDelay) after the backspaces AND sleep(preDelay) again
// before the text (35ms combined, stacked back-to-back with no injection in between).
// That's pure unnecessary hook-thread blocking time — it doesn't make either delay more
// effective, it just doubles the window during which a fast typist's next physical
// keystroke can enter the input queue while we're still mid-replacement for the previous
// one, worsening exactly the fast-typing desync this function exists to avoid. A single
// delay between the backspace phase and the text phase preserves the same pacing intent
// apps were tuned against, at roughly half the synchronous cost per replacement.
func sendSlow(text string, backspaces int, preDelay, postDelay, keyDelay int) {
	if backspaces > 0 {
		sendBackspaces(backspaces)
	}
	if len(text) > 0 {
		if backspaces > 0 {
			time.Sleep(time.Duration(maxInt(preDelay, postDelay)) * time.Millisecond)
		}
		sendUnicodeTextSlow(text, keyDelay)
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// sendAtomic combines backspaces and text into a single SendInput call
// This prevents flicker in Discord and other rich-text editors
func sendAtomic(text string, backspaces int) {
	runes := []rune(text)
	totalEvents := backspaces*2 + len(runes)*2

	if totalEvents == 0 {
		return
	}

	inputs := make([]INPUT, totalEvents)
	idx := 0

	// Add backspace events first
	for i := 0; i < backspaces; i++ {
		inputs[idx] = INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_BACK,
				DwFlags:     0,
				DwExtraInfo: InjectedKeyMarker,
			},
		}
		idx++
		inputs[idx] = INPUT{
			Type: INPUT_KEYBOARD,
			Ki: KEYBDINPUT{
				WVk:         VK_BACK,
				DwFlags:     KEYEVENTF_KEYUP,
				DwExtraInfo: InjectedKeyMarker,
			},
		}
		idx++
	}

	// Add Unicode text events
	for _, r := range runes {
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

	// Single atomic SendInput call
	sent, _, _ := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
	checkSendInputResult(sent, len(inputs), "sendAtomic")
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

// checkSendInputResult verifies that SendInput queued every event we asked for.
// A partial injection (target dropped some events, e.g. under load during fast typing)
// leaves the on-screen text out of sync with the Rust engine's internal buffer, which
// would corrupt every subsequent backspace/insert computed from that point on. When we
// detect this, clear the engine buffer so the next keystroke starts fresh instead of
// compounding the desync into further misplaced diacritics.
func checkSendInputResult(sent uintptr, expected int, label string) {
	if int(sent) == expected {
		return
	}
	log.Printf("[TextSender] %s: SendInput queued %d/%d events, resetting IME buffer to avoid desync", label, sent, expected)
	if bridge, err := GetBridge(); err == nil && bridge != nil {
		bridge.Clear()
	}
}

// DetectInjectionMethod determines the best method for current foreground app
func DetectInjectionMethod() InjectionMethod {
	return GetInjectionMethod()
}

// SendTextWithProfile sends text with full app profile (includes backspace mode)
func SendTextWithProfile(text string, backspaces int, profile AppProfile) {
	if len(text) == 0 && backspaces == 0 {
		return
	}

	switch profile.Method {
	case MethodFast:
		sendFastWithProfile(text, backspaces, profile.BackspaceMode)
	case MethodSlow:
		sendSlowWithProfile(text, backspaces, SlowModePreDelay, SlowModePostDelay, SlowModeKeyDelay, profile.BackspaceMode)
	case MethodAtomic:
		sendAtomicWithProfile(text, backspaces, profile.BackspaceMode)
	case MethodPaste:
		sendPaste(text, backspaces)
	case MethodPasteShiftV:
		sendPasteShiftV(text, backspaces)
	}
}

func sendFastWithProfile(text string, backspaces int, bsMode BackspaceMode) {
	// See sendFast: atomic injection avoids the fast-typing desync caused by splitting
	// backspaces and text into two SendInput calls with a sleep in between.
	sendAtomicWithProfile(text, backspaces, bsMode)
}

// sendSlowWithProfile mirrors sendSlow's single-settling-delay fix; see sendSlow for rationale.
func sendSlowWithProfile(text string, backspaces int, preDelay, postDelay, keyDelay int, bsMode BackspaceMode) {
	if backspaces > 0 {
		sendBackspacesWithMode(backspaces, bsMode)
	}
	if len(text) > 0 {
		if backspaces > 0 {
			time.Sleep(time.Duration(maxInt(preDelay, postDelay)) * time.Millisecond)
		}
		sendUnicodeTextSlow(text, keyDelay)
	}
}

// sendAtomicWithProfile combines backspaces and text into a single SendInput call
// Supports both VK_BACK and Unicode BS modes
func sendAtomicWithProfile(text string, backspaces int, bsMode BackspaceMode) {
	runes := []rune(text)
	totalEvents := backspaces*2 + len(runes)*2

	if totalEvents == 0 {
		return
	}

	inputs := make([]INPUT, totalEvents)
	idx := 0

	// Add backspace events first
	for i := 0; i < backspaces; i++ {
		if bsMode == BackspaceUnicode {
			// Unicode BS (0x08) - for CLI apps that don't handle DEL
			inputs[idx] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         0,
					WScan:       0x0008, // Unicode BS
					DwFlags:     KEYEVENTF_UNICODE,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
			idx++
			inputs[idx] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         0,
					WScan:       0x0008,
					DwFlags:     KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
		} else {
			// VK_BACK (default)
			inputs[idx] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         VK_BACK,
					DwFlags:     0,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
			idx++
			inputs[idx] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         VK_BACK,
					DwFlags:     KEYEVENTF_KEYUP,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
		}
		idx++
	}

	// Add Unicode text events
	for _, r := range runes {
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

	// Single atomic SendInput call
	sent, _, _ := procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
	checkSendInputResult(sent, len(inputs), "sendAtomicWithProfile")
}

// sendBackspacesWithMode sends backspaces using specified mode
func sendBackspacesWithMode(count int, bsMode BackspaceMode) {
	inputs := make([]INPUT, count*2)

	for i := 0; i < count; i++ {
		if bsMode == BackspaceUnicode {
			// Unicode BS (0x08)
			inputs[i*2] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         0,
					WScan:       0x0008,
					DwFlags:     KEYEVENTF_UNICODE,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
			inputs[i*2+1] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         0,
					WScan:       0x0008,
					DwFlags:     KEYEVENTF_UNICODE | KEYEVENTF_KEYUP,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
		} else {
			// VK_BACK (default)
			inputs[i*2] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         VK_BACK,
					DwFlags:     0,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
			inputs[i*2+1] = INPUT{
				Type: INPUT_KEYBOARD,
				Ki: KEYBDINPUT{
					WVk:         VK_BACK,
					DwFlags:     KEYEVENTF_KEYUP,
					DwExtraInfo: InjectedKeyMarker,
				},
			}
		}
	}

	procSendInput.Call(
		uintptr(len(inputs)),
		uintptr(unsafe.Pointer(&inputs[0])),
		uintptr(inputSize),
	)
}
