package core

// Hub channel: a small named-object channel (state mapping + events) that lets
// the my-utilities hub read FKey's V/E state and ask it to toggle or reload.
// Layout and names are versioned (v1); see hub-contract.json.

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	HubChannelPrefix = "Local\\FKeyIME.Hub.v1."

	hubStateSize    = 16
	hubStateMagic   = 0x42484B46
	hubStateVersion = 1
	hubStateName    = "State"
	hubSDDLFormat   = "D:P(A;;GA;;;SY)(A;;GA;;;BA)(A;;GA;;;%s)S:(ML;;NW;;;ME)"
)

// Event names under HubChannelPrefix: Changed is manual-reset (state flipped);
// Toggle and Reload are auto-reset requests from the hub.
const (
	HubEventChanged = "Changed"
	HubEventToggle  = "Toggle"
	HubEventReload  = "Reload"
)

// HubEventNames lists the events CreateHubChannel creates.
var HubEventNames = []string{HubEventChanged, HubEventToggle, HubEventReload}

// ErrHubChannelExists means a channel object with our name already exists
// (a previous instance or a hostile creator); creation fails closed.
var ErrHubChannelExists = errors.New("hub channel already exists")

// encodeHubState lays out the 16-byte state page, little-endian.
func encodeHubState(enabled bool, seq, pid uint32) [hubStateSize]byte {
	var b [hubStateSize]byte
	binary.LittleEndian.PutUint32(b[0:4], hubStateMagic)
	binary.LittleEndian.PutUint16(b[4:6], hubStateVersion)
	var flags uint16
	if enabled {
		flags = 1
	}
	binary.LittleEndian.PutUint16(b[6:8], flags)
	binary.LittleEndian.PutUint32(b[8:12], seq)
	binary.LittleEndian.PutUint32(b[12:16], pid)
	return b
}

// HubChannel owns the mapping and events created under --hub.
type HubChannel struct {
	mapping, changed, toggle, reload windows.Handle
	view                             *[hubStateSize]byte
	mu                               sync.Mutex
	seq                              uint32
	closed                           atomic.Bool
	loopRunning                      atomic.Bool
	loopDone                         chan struct{}
}

// CreateHubChannel creates the state mapping and the Changed/Toggle/Reload
// events under prefix. If any object already exists it closes everything and
// returns ErrHubChannelExists.
func CreateHubChannel(prefix string) (*HubChannel, error) {
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil {
		return nil, fmt.Errorf("token user: %w", err)
	}
	sd, err := windows.SecurityDescriptorFromString(fmt.Sprintf(hubSDDLFormat, user.User.Sid.String()))
	if err != nil {
		return nil, fmt.Errorf("security descriptor: %w", err)
	}
	sa := &windows.SecurityAttributes{
		Length:             uint32(unsafe.Sizeof(windows.SecurityAttributes{})),
		SecurityDescriptor: sd,
	}

	c := &HubChannel{loopDone: make(chan struct{})}
	fail := func(err error) (*HubChannel, error) {
		c.closeHandles()
		return nil, err
	}
	// x/sys returns ERROR_ALREADY_EXISTS as err alongside a valid handle.
	create := func(h *windows.Handle, v windows.Handle, err error) error {
		*h = v
		if err == windows.ERROR_ALREADY_EXISTS {
			return ErrHubChannelExists
		}
		if err != nil {
			return err
		}
		return nil
	}

	m, err := windows.CreateFileMapping(windows.InvalidHandle, sa, windows.PAGE_READWRITE, 0, hubStateSize, windows.StringToUTF16Ptr(prefix+hubStateName))
	if err := create(&c.mapping, m, err); err != nil {
		return fail(err)
	}
	for _, ev := range []struct {
		h      *windows.Handle
		name   string
		manual uint32
	}{{&c.changed, HubEventChanged, 1}, {&c.toggle, HubEventToggle, 0}, {&c.reload, HubEventReload, 0}} {
		h, err := windows.CreateEvent(sa, ev.manual, 0, windows.StringToUTF16Ptr(prefix+ev.name))
		if err := create(ev.h, h, err); err != nil {
			return fail(err)
		}
	}

	addr, err := windows.MapViewOfFile(c.mapping, windows.FILE_MAP_WRITE, 0, 0, hubStateSize)
	if err != nil {
		return fail(fmt.Errorf("map view: %w", err))
	}
	// The one uintptr->pointer conversion; the view is not Go memory. go vet may flag it.
	c.view = (*[hubStateSize]byte)(unsafe.Pointer(addr))
	init := encodeHubState(false, 0, uint32(os.Getpid()))
	copy(c.view[:], init[:])
	return c, nil
}

// Publish writes the state under a seqlock (odd while writing) and signals Changed.
func (c *HubChannel) Publish(enabled bool) {
	c.mu.Lock()
	if c.view == nil {
		c.mu.Unlock()
		return
	}
	seq := (*uint32)(unsafe.Pointer(&c.view[8]))
	c.seq++
	atomic.StoreUint32(seq, c.seq)
	var flags uint16
	if enabled {
		flags = 1
	}
	binary.LittleEndian.PutUint16(c.view[6:8], flags)
	c.seq++
	atomic.StoreUint32(seq, c.seq)
	c.mu.Unlock()
	windows.SetEvent(c.changed)
}

// RunRequestLoop blocks, calling onToggle / onReload as the hub signals them,
// until Close. Requests signalled before the loop starts stay pending in the
// auto-reset events.
func (c *HubChannel) RunRequestLoop(onToggle, onReload func()) {
	handles := []windows.Handle{c.toggle, c.reload}
	c.loopRunning.Store(true)
	defer close(c.loopDone)
	for {
		idx, err := windows.WaitForMultipleObjects(handles, false, windows.INFINITE)
		if c.closed.Load() || err != nil {
			return
		}
		switch idx - windows.WAIT_OBJECT_0 {
		case 0:
			onToggle()
		case 1:
			onReload()
		}
	}
}

// Close releases the request loop and every handle.
func (c *HubChannel) Close() {
	if c.closed.Swap(true) {
		return
	}
	windows.SetEvent(c.toggle)
	// Never close handles under a thread that is waiting on them.
	if c.loopRunning.Load() {
		select {
		case <-c.loopDone:
		case <-time.After(2 * time.Second):
		}
	}
	c.closeHandles()
}

func (c *HubChannel) closeHandles() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.view != nil {
		windows.UnmapViewOfFile(uintptr(unsafe.Pointer(c.view)))
		c.view = nil
	}
	for _, h := range []*windows.Handle{&c.changed, &c.toggle, &c.reload, &c.mapping} {
		if *h != 0 {
			windows.CloseHandle(*h)
			*h = 0
		}
	}
}
