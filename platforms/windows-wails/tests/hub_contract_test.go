package tests

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"fkey/core"
	"fkey/services"
)

type hubContract struct {
	Contract int    `json:"contract"`
	Module   string `json:"module"`
	HubFlag  string `json:"hubFlag"`
	Channel  struct {
		Prefix  string   `json:"prefix"`
		Version int      `json:"version"`
		Events  []string `json:"events"`
	} `json:"channel"`
	Settings struct {
		SchemaVersion int      `json:"schemaVersion"`
		Stores        []string `json:"stores"`
		Keys          []string `json:"keys"`
	} `json:"settings"`
}

// TestHubContractMatchesCode fails when a settings key or channel name changes
// without updating hub-contract.json and its golden history.
func TestHubContractMatchesCode(t *testing.T) {
	var c hubContract
	readJSON(t, "../hub-contract.json", &c)
	var history map[string][]string
	readJSON(t, "hub_contract_history.json", &history)

	keys := services.AllSettingKeys()
	if !reflect.DeepEqual(c.Settings.Keys, keys) {
		t.Errorf("contract keys %v != code keys %v (bump schemaVersion and add a history entry)", c.Settings.Keys, keys)
	}
	ver := itoa(c.Settings.SchemaVersion)
	if !reflect.DeepEqual(history[ver], keys) {
		t.Errorf("history[%s] = %v, want %v", ver, history[ver], keys)
	}
	if c.Channel.Prefix != core.HubChannelPrefix {
		t.Errorf("contract prefix %q != %q", c.Channel.Prefix, core.HubChannelPrefix)
	}
	if !strings.HasSuffix(core.HubChannelPrefix, ".v"+itoa(c.Channel.Version)+".") {
		t.Errorf("prefix %q does not carry channel version %d", core.HubChannelPrefix, c.Channel.Version)
	}
	if !reflect.DeepEqual(c.Channel.Events, core.HubEventNames) {
		t.Errorf("contract events %v != %v", c.Channel.Events, core.HubEventNames)
	}
	if c.Module != "fkey" || c.HubFlag != "--hub" || c.Contract != 1 {
		t.Errorf("unexpected contract header: %+v", c)
	}
}

func readJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func itoa(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}
