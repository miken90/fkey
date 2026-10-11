package core

import (
	"bytes"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// x/sys has no OpenFileMapping wrapper.
var procOpenFileMappingW = windows.NewLazySystemDLL("kernel32.dll").NewProc("OpenFileMappingW")

var hubTestCounter atomic.Uint32

func testHubPrefix() string {
	return fmt.Sprintf("Local\\FKeyIME.HubTest.%d.%d.", windows.GetCurrentProcessId(), hubTestCounter.Add(1))
}

func newTestChannel(t *testing.T) (*HubChannel, string) {
	t.Helper()
	prefix := testHubPrefix()
	ch, err := CreateHubChannel(prefix)
	if err != nil {
		t.Fatalf("CreateHubChannel: %v", err)
	}
	t.Cleanup(ch.Close)
	return ch, prefix
}

func openTestEvent(t *testing.T, name string) windows.Handle {
	t.Helper()
	h, err := windows.OpenEvent(windows.EVENT_MODIFY_STATE|windows.SYNCHRONIZE, false, windows.StringToUTF16Ptr(name))
	if err != nil {
		t.Fatalf("OpenEvent %s: %v", name, err)
	}
	t.Cleanup(func() { windows.CloseHandle(h) })
	return h
}

func readTestState(t *testing.T, prefix string) [hubStateSize]byte {
	t.Helper()
	namePtr := windows.StringToUTF16Ptr(prefix + hubStateName)
	r, _, callErr := procOpenFileMappingW.Call(uintptr(windows.FILE_MAP_READ), 0, uintptr(unsafe.Pointer(namePtr)))
	if r == 0 {
		t.Fatalf("OpenFileMapping: %v", callErr)
	}
	m := windows.Handle(r)
	defer windows.CloseHandle(m)
	addr, err := windows.MapViewOfFile(m, windows.FILE_MAP_READ, 0, 0, hubStateSize)
	if err != nil {
		t.Fatalf("MapViewOfFile: %v", err)
	}
	defer windows.UnmapViewOfFile(addr)
	var out [hubStateSize]byte
	copy(out[:], unsafe.Slice((*byte)(unsafe.Pointer(addr)), hubStateSize))
	return out
}

func waitFlag(f *atomic.Int32, want int32, d time.Duration) bool {
	deadline := time.Now().Add(d)
	for time.Now().Before(deadline) {
		if f.Load() >= want {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return f.Load() >= want
}

func TestParseLaunchArgsHubDefaultsOff(t *testing.T) {
	cases := []struct {
		args []string
		want LaunchArgs
	}{
		{nil, LaunchArgs{}},
		{[]string{"--hub"}, LaunchArgs{Hub: true}},
		{[]string{"--relaunch"}, LaunchArgs{Relaunch: true}},
		{[]string{"--relaunch", "--hub"}, LaunchArgs{Relaunch: true, Hub: true}},
		{[]string{"--HUB", "hub"}, LaunchArgs{}},
	}
	for _, c := range cases {
		if got := ParseLaunchArgs(c.args); got != c.want {
			t.Errorf("ParseLaunchArgs(%v) = %+v, want %+v", c.args, got, c.want)
		}
	}
}

func TestHubStateLayoutGoldenBytes(t *testing.T) {
	want := []byte{0x46, 0x4B, 0x48, 0x42, 0x01, 0x00, 0x01, 0x00, 0x02, 0x00, 0x00, 0x00, 0x34, 0x12, 0x00, 0x00}
	got := encodeHubState(true, 2, 0x1234)
	if !bytes.Equal(got[:], want) {
		t.Fatalf("got % X, want % X", got, want)
	}
}

func TestHubChannelPublishRoundTrip(t *testing.T) {
	ch, prefix := newTestChannel(t)
	changed := openTestEvent(t, prefix+HubEventChanged)

	ch.Publish(true)
	if r, _ := windows.WaitForSingleObject(changed, 1000); r != windows.WAIT_OBJECT_0 {
		t.Fatalf("Changed not signalled after Publish(true): %d", r)
	}
	st := readTestState(t, prefix)
	if st[6] != 1 || st[8]&1 != 0 {
		t.Fatalf("after Publish(true) flags=%d seq=%d, want enabled and even seq", st[6], st[8])
	}
	seqTrue := st[8]

	ch.Publish(false)
	st = readTestState(t, prefix)
	if st[6] != 0 || st[8] != seqTrue+2 {
		t.Fatalf("after Publish(false) flags=%d seq=%d, want disabled and seq %d", st[6], st[8], seqTrue+2)
	}
	if pid := uint32(st[12]) | uint32(st[13])<<8; pid != windows.GetCurrentProcessId()&0xFFFF {
		t.Errorf("pid low bytes %d do not match process", pid)
	}
}

func TestHubChannelCreateFailsWhenNamePreexists(t *testing.T) {
	prefix := testHubPrefix()
	hostile, err := windows.CreateFileMapping(windows.InvalidHandle, nil, windows.PAGE_READWRITE, 0, hubStateSize, windows.StringToUTF16Ptr(prefix+hubStateName))
	if err != nil {
		t.Fatalf("pre-create mapping: %v", err)
	}
	defer windows.CloseHandle(hostile)

	ch, err := CreateHubChannel(prefix)
	if err != ErrHubChannelExists {
		if ch != nil {
			ch.Close()
		}
		t.Fatalf("err = %v, want ErrHubChannelExists", err)
	}
	for _, n := range HubEventNames {
		if h, err := windows.OpenEvent(windows.SYNCHRONIZE, false, windows.StringToUTF16Ptr(prefix+n)); err == nil {
			windows.CloseHandle(h)
			t.Errorf("event %s left behind after failed create", n)
		}
	}
}

func TestHubChannelToggleRunsCallback(t *testing.T) {
	ch, prefix := newTestChannel(t)
	var toggles, reloads atomic.Int32
	done := make(chan struct{})
	go func() {
		ch.RunRequestLoop(func() { toggles.Add(1) }, func() { reloads.Add(1) })
		close(done)
	}()

	windows.SetEvent(openTestEvent(t, prefix+HubEventToggle))
	if !waitFlag(&toggles, 1, time.Second) {
		t.Fatal("toggle callback did not run within 1s")
	}
	if reloads.Load() != 0 {
		t.Error("reload callback ran for a toggle request")
	}
	ch.Close()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Close did not end the request loop")
	}
}

func TestHubChannelReloadRunsCallback(t *testing.T) {
	ch, prefix := newTestChannel(t)
	var toggles, reloads atomic.Int32
	go ch.RunRequestLoop(func() { toggles.Add(1) }, func() { reloads.Add(1) })

	windows.SetEvent(openTestEvent(t, prefix+HubEventReload))
	if !waitFlag(&reloads, 1, time.Second) {
		t.Fatal("reload callback did not run within 1s")
	}
	if toggles.Load() != 0 {
		t.Error("toggle callback ran for a reload request")
	}
}

func TestToggleBeforeLoopStartIsNotLost(t *testing.T) {
	ch, prefix := newTestChannel(t)
	windows.SetEvent(openTestEvent(t, prefix+HubEventToggle))

	var toggles atomic.Int32
	go ch.RunRequestLoop(func() { toggles.Add(1) }, func() {})
	if !waitFlag(&toggles, 1, time.Second) {
		t.Fatal("pending toggle was lost")
	}
	time.Sleep(100 * time.Millisecond)
	if n := toggles.Load(); n != 1 {
		t.Errorf("toggle callback ran %d times, want once", n)
	}
}

func TestHubModeSkipsUpdateCheck(t *testing.T) {
	if ShouldCheckUpdates(true) {
		t.Error("hub mode must skip the update check")
	}
	if !ShouldCheckUpdates(false) {
		t.Error("standalone must check for updates")
	}
}

func TestHubModeOffWhenElevatedOrRunAsAdmin(t *testing.T) {
	cases := []struct{ hub, admin, elevated, want bool }{
		{false, false, false, false},
		{true, false, false, true},
		{true, true, false, false},
		{true, false, true, false},
		{true, true, true, false},
		{false, true, true, false},
	}
	for _, c := range cases {
		if got := HubModeEnabled(c.hub, c.admin, c.elevated); got != c.want {
			t.Errorf("HubModeEnabled(%v,%v,%v) = %v, want %v", c.hub, c.admin, c.elevated, got, c.want)
		}
	}
}

func TestEnabledChangedHandlerOnlyBeepsAndPosts(t *testing.T) {
	var calls []string
	var posted func()
	h := NewEnabledChangedHandler(
		func(e bool) { calls = append(calls, fmt.Sprintf("beep:%v", e)) },
		func(f func()) { calls = append(calls, "post"); posted = f },
		func(e bool) { calls = append(calls, fmt.Sprintf("update:%v", e)) },
	)
	h(true)
	if got := fmt.Sprint(calls); got != "[beep:true post]" {
		t.Fatalf("on the hook thread got %s, want [beep:true post]", got)
	}
	posted()
	if got := fmt.Sprint(calls); got != "[beep:true post update:true]" {
		t.Fatalf("after running the posted func got %s", got)
	}
}

var (
	procQPC     = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryPerformanceCounter")
	procQPF     = windows.NewLazySystemDLL("kernel32.dll").NewProc("QueryPerformanceFrequency")
	qpcFreqOnce sync.Once
	qpcFreq     int64
)

// qpcNow returns a high-resolution tick count; time.Now on Windows can be far
// coarser than the calls measured here.
func qpcNow() int64 {
	qpcFreqOnce.Do(func() { procQPF.Call(uintptr(unsafe.Pointer(&qpcFreq))) })
	var v int64
	procQPC.Call(uintptr(unsafe.Pointer(&v)))
	return v
}

func qpcToDuration(ticks int64) time.Duration {
	return time.Duration(float64(ticks) / float64(qpcFreq) * float64(time.Second))
}

func p99(d []time.Duration) time.Duration {
	sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
	return d[len(d)*99/100]
}

// BenchmarkEnabledChangedHandler measures what runs on the hook thread (beep
// stub + post). The publish runs inside the posted func, so hub-on must not
// differ from hub-off by more than 5 us at p99.
func BenchmarkEnabledChangedHandler(b *testing.B) {
	results := map[string]time.Duration{}
	run := func(name string, update func(bool)) {
		b.Run(name, func(b *testing.B) {
			var posted func()
			h := NewEnabledChangedHandler(func(bool) {}, func(f func()) { posted = f }, update)
			// One call is shorter than a clock tick, so each sample times a
			// batch and is divided back to a per-call figure.
			const batch = 100
			samples := make([]time.Duration, 0, b.N/batch+1)
			b.ResetTimer()
			for i := 0; i < b.N; i += batch {
				n := min(batch, b.N-i)
				start := qpcNow()
				for j := 0; j < n; j++ {
					h(j&1 == 0)
				}
				samples = append(samples, qpcToDuration(qpcNow()-start)/time.Duration(n))
				b.StopTimer()
				posted() // the main-thread half, outside the measured hook-thread part
				b.StartTimer()
			}
			b.StopTimer()
			results[name] = p99(samples)
			b.ReportMetric(float64(results[name].Nanoseconds()), "p99-ns")
		})
	}
	run("hub-off", func(bool) {})
	ch, err := CreateHubChannel(testHubPrefix())
	if err != nil {
		b.Fatal(err)
	}
	defer ch.Close()
	run("hub-on", ch.Publish)

	if delta := results["hub-on"] - results["hub-off"]; delta > 5*time.Microsecond {
		b.Errorf("hook-thread p99 delta %v exceeds 5us", delta)
	}
}
