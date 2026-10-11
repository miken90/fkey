package services

import (
	"bytes"
	"fmt"
	"math/rand"
	"reflect"
	"sort"
	"testing"

	"golang.org/x/sys/windows/registry"
)

// newTestSettingsService points a service at throwaway keys under
// HKCU\SOFTWARE\FKeyTest and removes them afterwards.
func newTestSettingsService(t *testing.T) *SettingsService {
	t.Helper()
	base := fmt.Sprintf(`SOFTWARE\FKeyTest\%d`, rand.Int63())
	s := NewSettingsService()
	s.keyPath = base
	s.runKeyPath = base + `\Run`
	if k, _, err := registry.CreateKey(registry.CURRENT_USER, s.runKeyPath, registry.SET_VALUE); err != nil {
		t.Fatalf("create test run key: %v", err)
	} else {
		k.Close()
	}
	t.Cleanup(func() {
		registry.DeleteKey(registry.CURRENT_USER, s.runKeyPath)
		registry.DeleteKey(registry.CURRENT_USER, s.keyPath)
		registry.DeleteKey(registry.CURRENT_USER, `SOFTWARE\FKeyTest`) // only succeeds when empty
	})
	return s
}

func snapshotKey(t *testing.T, path string) map[string]string {
	t.Helper()
	k, err := registry.OpenKey(registry.CURRENT_USER, path, registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer k.Close()
	names, _ := k.ReadValueNames(-1)
	sort.Strings(names)
	out := map[string]string{}
	for _, n := range names {
		buf := make([]byte, 1024)
		size, typ, err := k.GetValue(n, buf)
		if err != nil {
			t.Fatalf("get %s: %v", n, err)
		}
		out[n] = fmt.Sprintf("%d:%x", typ, bytes.Clone(buf[:size]))
	}
	return out
}

func TestReloadKeepsInMemoryEnabled(t *testing.T) {
	s := newTestSettingsService(t)
	k, _, err := registry.CreateKey(registry.CURRENT_USER, s.keyPath, registry.SET_VALUE)
	if err != nil {
		t.Fatal(err)
	}
	k.SetDWordValue(KeyEnabled, 0)
	k.SetDWordValue(KeyEscRestore, 0)
	k.SetStringValue(KeyToggleHotkey, "65,2")
	k.Close()

	s.settings.Enabled = true
	s.settings.EscRestore = true
	before := snapshotKey(t, s.keyPath)

	if err := s.Reload(); err != nil {
		t.Fatalf("Reload: %v", err)
	}
	if !s.settings.Enabled {
		t.Error("Reload overwrote the in-memory Enabled")
	}
	if s.settings.EscRestore {
		t.Error("Reload did not apply EscRestore=0")
	}
	if s.settings.ToggleHotkey != "65,2" {
		t.Errorf("ToggleHotkey = %q, want 65,2", s.settings.ToggleHotkey)
	}
	after := snapshotKey(t, s.keyPath)
	if fmt.Sprint(before) != fmt.Sprint(after) {
		t.Errorf("Reload wrote to the registry:\nbefore %v\nafter  %v", before, after)
	}
}

func TestSaveEnabledWritesOnlyEnabled(t *testing.T) {
	s := newTestSettingsService(t)
	rk, _ := registry.OpenKey(registry.CURRENT_USER, s.runKeyPath, registry.SET_VALUE)
	rk.SetStringValue("Other", `"C:\other.exe"`)
	rk.Close()
	runBefore := snapshotKey(t, s.runKeyPath)

	s.settings.Enabled = false
	s.settings.AutoStart = true // would add a Run value if SaveEnabled touched autostart
	if err := s.SaveEnabled(); err != nil {
		t.Fatalf("SaveEnabled: %v", err)
	}
	got := snapshotKey(t, s.keyPath)
	if len(got) != 1 || got[KeyEnabled] == "" {
		t.Errorf("values after SaveEnabled = %v, want only Enabled", got)
	}
	if fmt.Sprint(runBefore) != fmt.Sprint(snapshotKey(t, s.runKeyPath)) {
		t.Error("SaveEnabled touched the Run key")
	}
}

func TestHubManagedSkipsAutoStart(t *testing.T) {
	s := newTestSettingsService(t)
	s.settings.AutoStart = true
	s.SetHubManaged(true)
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if run := snapshotKey(t, s.runKeyPath); len(run) != 0 {
		t.Errorf("Run key has values under hub management: %v", run)
	}
}

func TestAllSettingKeysMatchesSavedValues(t *testing.T) {
	s := newTestSettingsService(t)
	s.SetHubManaged(true)
	if err := s.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}
	snap := snapshotKey(t, s.keyPath)
	var got []string
	for k := range snap {
		got = append(got, k)
	}
	sort.Strings(got)

	want := append([]string(nil), AllSettingKeys()...)
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("saved value names %v != AllSettingKeys() %v", got, want)
	}
}
