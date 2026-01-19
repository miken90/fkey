package core

import (
	"encoding/binary"
	"fmt"
	"log"
	"os"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// uinput event types
const (
	EV_SYN = 0x00
	EV_KEY = 0x01
)

// uinput sync codes
const (
	SYN_REPORT = 0
)

// Key state values
const (
	KEY_RELEASE = 0
	KEY_PRESS   = 1
)

// Linux keycodes for common keys
const (
	LINUX_KEY_BACKSPACE = 14
	LINUX_KEY_TAB       = 15
	LINUX_KEY_ENTER     = 28
	LINUX_KEY_LEFTSHIFT = 42
	LINUX_KEY_RIGHTSHIFT = 54
	LINUX_KEY_LEFTCTRL  = 29
	LINUX_KEY_SPACE     = 57
	LINUX_KEY_U         = 22

	// Letter keys (a-z)
	LINUX_KEY_A = 30
	LINUX_KEY_B = 48
	LINUX_KEY_C = 46
	LINUX_KEY_D = 32
	LINUX_KEY_E = 18
	LINUX_KEY_F = 33
	LINUX_KEY_G = 34
	LINUX_KEY_H = 35
	LINUX_KEY_I = 23
	LINUX_KEY_J = 36
	LINUX_KEY_K = 37
	LINUX_KEY_L = 38
	LINUX_KEY_M = 50
	LINUX_KEY_N = 49
	LINUX_KEY_O = 24
	LINUX_KEY_P = 25
	LINUX_KEY_Q = 16
	LINUX_KEY_R = 19
	LINUX_KEY_S = 31
	LINUX_KEY_T = 20
	LINUX_KEY_V = 47
	LINUX_KEY_W = 17
	LINUX_KEY_X = 45
	LINUX_KEY_Y = 21
	LINUX_KEY_Z = 44

	// Number keys
	LINUX_KEY_0 = 11
	LINUX_KEY_1 = 2
	LINUX_KEY_2 = 3
	LINUX_KEY_3 = 4
	LINUX_KEY_4 = 5
	LINUX_KEY_5 = 6
	LINUX_KEY_6 = 7
	LINUX_KEY_7 = 8
	LINUX_KEY_8 = 9
	LINUX_KEY_9 = 10
)

// uinput ioctl constants
const (
	UINPUT_MAX_NAME_SIZE = 80
	UI_SET_EVBIT         = 0x40045564
	UI_SET_KEYBIT        = 0x40045565
	UI_DEV_CREATE        = 0x5501
	UI_DEV_DESTROY       = 0x5502
	UI_DEV_SETUP         = 0x405c5503
	BUS_USB              = 0x03
)

// uinputSetup is the uinput_setup struct for UI_DEV_SETUP
type uinputSetup struct {
	ID struct {
		Bustype uint16
		Vendor  uint16
		Product uint16
		Version uint16
	}
	Name       [UINPUT_MAX_NAME_SIZE]byte
	FFEffects  uint32
}

// inputEvent matches the kernel's input_event struct
type inputEvent struct {
	Time  syscall.Timeval
	Type  uint16
	Code  uint16
	Value int32
}

// UInputDevice manages a virtual keyboard via /dev/uinput
type UInputDevice struct {
	fd       int
	mu       sync.Mutex
	ready    bool
}

var (
	uinputDevice     *UInputDevice
	uinputDeviceOnce sync.Once
	uinputDeviceErr  error
)

// GetUInputDevice returns the singleton uinput device, initializing on first call
func GetUInputDevice() (*UInputDevice, error) {
	uinputDeviceOnce.Do(func() {
		uinputDevice, uinputDeviceErr = newUInputDevice()
	})
	return uinputDevice, uinputDeviceErr
}

// newUInputDevice creates and configures a virtual keyboard
func newUInputDevice() (*UInputDevice, error) {
	fd, err := syscall.Open("/dev/uinput", syscall.O_WRONLY|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open /dev/uinput: %w (ensure user is in 'input' group)", err)
	}

	dev := &UInputDevice{fd: fd}

	// Enable EV_KEY events
	if err := dev.ioctl(UI_SET_EVBIT, EV_KEY); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("UI_SET_EVBIT failed: %w", err)
	}

	// Enable all keys we might need (0-255 covers standard keyboard)
	for key := 0; key < 256; key++ {
		if err := dev.ioctl(UI_SET_KEYBIT, uintptr(key)); err != nil {
			syscall.Close(fd)
			return nil, fmt.Errorf("UI_SET_KEYBIT failed for key %d: %w", key, err)
		}
	}

	// Setup device info
	var setup uinputSetup
	setup.ID.Bustype = BUS_USB
	setup.ID.Vendor = 0x1234
	setup.ID.Product = 0x5678
	setup.ID.Version = 1
	copy(setup.Name[:], "FKey Virtual Keyboard")

	if err := dev.ioctlPtr(UI_DEV_SETUP, unsafe.Pointer(&setup)); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("UI_DEV_SETUP failed: %w", err)
	}

	// Create the device
	if err := dev.ioctl(UI_DEV_CREATE, 0); err != nil {
		syscall.Close(fd)
		return nil, fmt.Errorf("UI_DEV_CREATE failed: %w", err)
	}

	// Give udev time to create device node
	time.Sleep(100 * time.Millisecond)

	dev.ready = true
	log.Println("uinput virtual keyboard created successfully")
	return dev, nil
}

func (d *UInputDevice) ioctl(req, val uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(d.fd), req, val)
	if errno != 0 {
		return errno
	}
	return nil
}

func (d *UInputDevice) ioctlPtr(req uintptr, ptr unsafe.Pointer) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(d.fd), req, uintptr(ptr))
	if errno != 0 {
		return errno
	}
	return nil
}

// Close destroys the virtual device
func (d *UInputDevice) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.ready {
		return nil
	}

	d.ioctl(UI_DEV_DESTROY, 0)
	syscall.Close(d.fd)
	d.ready = false
	return nil
}

// writeEvent writes a single input event
func (d *UInputDevice) writeEvent(evType, code uint16, value int32) error {
	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)

	ev := inputEvent{
		Time:  tv,
		Type:  evType,
		Code:  code,
		Value: value,
	}

	buf := make([]byte, unsafe.Sizeof(ev))
	*(*inputEvent)(unsafe.Pointer(&buf[0])) = ev

	_, err := syscall.Write(d.fd, buf)
	return err
}

// sync sends a SYN_REPORT to flush pending events
func (d *UInputDevice) sync() error {
	return d.writeEvent(EV_SYN, SYN_REPORT, 0)
}

// pressKey sends key down + key up + sync
func (d *UInputDevice) pressKey(code uint16) error {
	if err := d.writeEvent(EV_KEY, code, KEY_PRESS); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}
	if err := d.writeEvent(EV_KEY, code, KEY_RELEASE); err != nil {
		return err
	}
	return d.sync()
}

// pressKeyWithShift sends shift+key
func (d *UInputDevice) pressKeyWithShift(code uint16) error {
	// Shift down
	if err := d.writeEvent(EV_KEY, LINUX_KEY_LEFTSHIFT, KEY_PRESS); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}

	// Key press
	if err := d.writeEvent(EV_KEY, code, KEY_PRESS); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}

	// Key release
	if err := d.writeEvent(EV_KEY, code, KEY_RELEASE); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}

	// Shift up
	if err := d.writeEvent(EV_KEY, LINUX_KEY_LEFTSHIFT, KEY_RELEASE); err != nil {
		return err
	}
	return d.sync()
}

// SendBackspaces sends n backspace key presses
func (d *UInputDevice) SendBackspaces(count int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.ready {
		return fmt.Errorf("uinput device not ready")
	}

	for i := 0; i < count; i++ {
		if err := d.pressKey(LINUX_KEY_BACKSPACE); err != nil {
			return err
		}
	}
	return nil
}

// SendText sends backspaces followed by text using uinput
// For Unicode characters, uses the Ctrl+Shift+U hex input method (GTK/IBus compatible)
func (d *UInputDevice) SendText(text string, backspaces int) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.ready {
		return fmt.Errorf("uinput device not ready")
	}

	log.Printf("[uinput] SendText: backspaces=%d, text=%q", backspaces, text)

	// Small delay for stability
	time.Sleep(5 * time.Millisecond)

	// Send backspaces
	for i := 0; i < backspaces; i++ {
		if err := d.pressKey(LINUX_KEY_BACKSPACE); err != nil {
			return fmt.Errorf("backspace failed: %w", err)
		}
	}

	// Send each character
	for _, r := range text {
		if err := d.sendRune(r); err != nil {
			return fmt.Errorf("sendRune failed for %c: %w", r, err)
		}
	}

	return nil
}

// sendRune sends a single character
func (d *UInputDevice) sendRune(r rune) error {
	// Try direct keycode for ASCII
	if code, shift, ok := runeToLinuxKey(r); ok {
		if shift {
			return d.pressKeyWithShift(code)
		}
		return d.pressKey(code)
	}

	// For Unicode characters, use Ctrl+Shift+U method (GTK/IBus)
	return d.sendUnicodeViaCtrlShiftU(r)
}

// sendUnicodeViaCtrlShiftU sends Unicode using GTK's Ctrl+Shift+U input method
// Sequence: Ctrl+Shift+U, hex digits, Space (to commit)
func (d *UInputDevice) sendUnicodeViaCtrlShiftU(r rune) error {
	// Press Ctrl+Shift+U
	if err := d.writeEvent(EV_KEY, LINUX_KEY_LEFTCTRL, KEY_PRESS); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}
	if err := d.writeEvent(EV_KEY, LINUX_KEY_LEFTSHIFT, KEY_PRESS); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}
	if err := d.pressKey(LINUX_KEY_U); err != nil {
		return err
	}
	if err := d.writeEvent(EV_KEY, LINUX_KEY_LEFTSHIFT, KEY_RELEASE); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}
	if err := d.writeEvent(EV_KEY, LINUX_KEY_LEFTCTRL, KEY_RELEASE); err != nil {
		return err
	}
	if err := d.sync(); err != nil {
		return err
	}

	// Type hex digits
	hex := fmt.Sprintf("%x", r)
	for _, h := range hex {
		code, shift := hexDigitToKey(h)
		if shift {
			if err := d.pressKeyWithShift(code); err != nil {
				return err
			}
		} else {
			if err := d.pressKey(code); err != nil {
				return err
			}
		}
	}

	// Space to commit (or Enter in some systems)
	return d.pressKey(LINUX_KEY_SPACE)
}

// hexDigitToKey converts hex digit to Linux keycode
func hexDigitToKey(h rune) (uint16, bool) {
	switch h {
	case '0':
		return LINUX_KEY_0, false
	case '1':
		return LINUX_KEY_1, false
	case '2':
		return LINUX_KEY_2, false
	case '3':
		return LINUX_KEY_3, false
	case '4':
		return LINUX_KEY_4, false
	case '5':
		return LINUX_KEY_5, false
	case '6':
		return LINUX_KEY_6, false
	case '7':
		return LINUX_KEY_7, false
	case '8':
		return LINUX_KEY_8, false
	case '9':
		return LINUX_KEY_9, false
	case 'a', 'A':
		return LINUX_KEY_A, false
	case 'b', 'B':
		return LINUX_KEY_B, false
	case 'c', 'C':
		return LINUX_KEY_C, false
	case 'd', 'D':
		return LINUX_KEY_D, false
	case 'e', 'E':
		return LINUX_KEY_E, false
	case 'f', 'F':
		return LINUX_KEY_F, false
	}
	return 0, false
}

// runeToLinuxKey maps ASCII characters to Linux keycodes
func runeToLinuxKey(r rune) (code uint16, shift bool, ok bool) {
	switch r {
	// Lowercase letters
	case 'a':
		return LINUX_KEY_A, false, true
	case 'b':
		return LINUX_KEY_B, false, true
	case 'c':
		return LINUX_KEY_C, false, true
	case 'd':
		return LINUX_KEY_D, false, true
	case 'e':
		return LINUX_KEY_E, false, true
	case 'f':
		return LINUX_KEY_F, false, true
	case 'g':
		return LINUX_KEY_G, false, true
	case 'h':
		return LINUX_KEY_H, false, true
	case 'i':
		return LINUX_KEY_I, false, true
	case 'j':
		return LINUX_KEY_J, false, true
	case 'k':
		return LINUX_KEY_K, false, true
	case 'l':
		return LINUX_KEY_L, false, true
	case 'm':
		return LINUX_KEY_M, false, true
	case 'n':
		return LINUX_KEY_N, false, true
	case 'o':
		return LINUX_KEY_O, false, true
	case 'p':
		return LINUX_KEY_P, false, true
	case 'q':
		return LINUX_KEY_Q, false, true
	case 'r':
		return LINUX_KEY_R, false, true
	case 's':
		return LINUX_KEY_S, false, true
	case 't':
		return LINUX_KEY_T, false, true
	case 'u':
		return LINUX_KEY_U, false, true
	case 'v':
		return LINUX_KEY_V, false, true
	case 'w':
		return LINUX_KEY_W, false, true
	case 'x':
		return LINUX_KEY_X, false, true
	case 'y':
		return LINUX_KEY_Y, false, true
	case 'z':
		return LINUX_KEY_Z, false, true
	// Uppercase letters
	case 'A':
		return LINUX_KEY_A, true, true
	case 'B':
		return LINUX_KEY_B, true, true
	case 'C':
		return LINUX_KEY_C, true, true
	case 'D':
		return LINUX_KEY_D, true, true
	case 'E':
		return LINUX_KEY_E, true, true
	case 'F':
		return LINUX_KEY_F, true, true
	case 'G':
		return LINUX_KEY_G, true, true
	case 'H':
		return LINUX_KEY_H, true, true
	case 'I':
		return LINUX_KEY_I, true, true
	case 'J':
		return LINUX_KEY_J, true, true
	case 'K':
		return LINUX_KEY_K, true, true
	case 'L':
		return LINUX_KEY_L, true, true
	case 'M':
		return LINUX_KEY_M, true, true
	case 'N':
		return LINUX_KEY_N, true, true
	case 'O':
		return LINUX_KEY_O, true, true
	case 'P':
		return LINUX_KEY_P, true, true
	case 'Q':
		return LINUX_KEY_Q, true, true
	case 'R':
		return LINUX_KEY_R, true, true
	case 'S':
		return LINUX_KEY_S, true, true
	case 'T':
		return LINUX_KEY_T, true, true
	case 'U':
		return LINUX_KEY_U, true, true
	case 'V':
		return LINUX_KEY_V, true, true
	case 'W':
		return LINUX_KEY_W, true, true
	case 'X':
		return LINUX_KEY_X, true, true
	case 'Y':
		return LINUX_KEY_Y, true, true
	case 'Z':
		return LINUX_KEY_Z, true, true
	// Numbers
	case '0':
		return LINUX_KEY_0, false, true
	case '1':
		return LINUX_KEY_1, false, true
	case '2':
		return LINUX_KEY_2, false, true
	case '3':
		return LINUX_KEY_3, false, true
	case '4':
		return LINUX_KEY_4, false, true
	case '5':
		return LINUX_KEY_5, false, true
	case '6':
		return LINUX_KEY_6, false, true
	case '7':
		return LINUX_KEY_7, false, true
	case '8':
		return LINUX_KEY_8, false, true
	case '9':
		return LINUX_KEY_9, false, true
	// Special
	case ' ':
		return LINUX_KEY_SPACE, false, true
	case '\t':
		return LINUX_KEY_TAB, false, true
	case '\n':
		return LINUX_KEY_ENTER, false, true
	}
	return 0, false, false
}

// IsUInputAvailable checks if uinput can be used
func IsUInputAvailable() bool {
	// Check if /dev/uinput exists and is writable
	f, err := os.OpenFile("/dev/uinput", os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	f.Close()
	return true
}

// WriteEventsRaw writes raw input_event bytes (for batch operations)
func (d *UInputDevice) WriteEventsRaw(events []byte) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.ready {
		return fmt.Errorf("uinput device not ready")
	}

	_, err := syscall.Write(d.fd, events)
	return err
}

// BuildBatchEvents creates a batch of events for atomic sending
// This is similar to Windows SendInput - all events in one write
func BuildBatchEvents(text string, backspaces int) []byte {
	eventSize := int(unsafe.Sizeof(inputEvent{}))
	// Each key needs press + sync + release + sync = 4 events
	// Estimate: backspaces*4 + runes*4 (for ASCII) or more for Unicode
	runes := []rune(text)
	estimatedEvents := (backspaces + len(runes)) * 4
	buf := make([]byte, 0, estimatedEvents*eventSize)

	var tv syscall.Timeval
	syscall.Gettimeofday(&tv)

	appendEvent := func(evType, code uint16, value int32) {
		ev := inputEvent{Time: tv, Type: evType, Code: code, Value: value}
		evBytes := make([]byte, eventSize)
		// Copy struct to bytes
		binary.LittleEndian.PutUint64(evBytes[0:8], uint64(ev.Time.Sec))
		binary.LittleEndian.PutUint64(evBytes[8:16], uint64(ev.Time.Usec))
		binary.LittleEndian.PutUint16(evBytes[16:18], ev.Type)
		binary.LittleEndian.PutUint16(evBytes[18:20], ev.Code)
		binary.LittleEndian.PutUint32(evBytes[20:24], uint32(ev.Value))
		buf = append(buf, evBytes...)
	}

	appendKeyPress := func(code uint16) {
		appendEvent(EV_KEY, code, KEY_PRESS)
		appendEvent(EV_SYN, SYN_REPORT, 0)
		appendEvent(EV_KEY, code, KEY_RELEASE)
		appendEvent(EV_SYN, SYN_REPORT, 0)
	}

	// Backspaces
	for i := 0; i < backspaces; i++ {
		appendKeyPress(LINUX_KEY_BACKSPACE)
	}

	// Text - only handles ASCII in batch mode
	for _, r := range runes {
		if code, shift, ok := runeToLinuxKey(r); ok {
			if shift {
				appendEvent(EV_KEY, LINUX_KEY_LEFTSHIFT, KEY_PRESS)
				appendEvent(EV_SYN, SYN_REPORT, 0)
				appendKeyPress(code)
				appendEvent(EV_KEY, LINUX_KEY_LEFTSHIFT, KEY_RELEASE)
				appendEvent(EV_SYN, SYN_REPORT, 0)
			} else {
				appendKeyPress(code)
			}
		}
		// Skip non-ASCII in batch mode (would need Ctrl+Shift+U sequence)
	}

	return buf
}
