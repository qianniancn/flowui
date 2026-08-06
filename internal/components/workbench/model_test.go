package workbench

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/qianniancn/flowui/internal/components/dock"
)

func TestStateTabLifecycleAndMovement(t *testing.T) {
	state := NewState([]Group{
		{Key: "editor", Tabs: []Tab{{Key: "one", Closable: true}, {Key: "two", Closable: true}}, ActiveKey: "one"},
		{Key: "preview", Tabs: []Tab{{Key: "preview", Closable: true}}},
	})
	if !state.ActivateTab("editor", "two") || state.ActiveGroup != "editor" || state.ActiveTab("editor") != "two" {
		t.Fatalf("activation = %#v", state)
	}
	if !state.ReorderTab("editor", "two", 0) || state.Groups[0].Tabs[0].Key != "two" {
		t.Fatalf("reorder = %#v", state.Groups[0].Tabs)
	}
	if !state.MoveTab("editor", "preview", "two", -1) {
		t.Fatal("move failed")
	}
	if state.Groups[0].ActiveKey != "one" || state.ActiveGroup != "preview" || state.ActiveTab("preview") != "two" {
		t.Fatalf("move selection = %#v", state)
	}
	if !state.CloseTab("preview", "two") || state.ActiveTab("preview") != "preview" {
		t.Fatalf("close fallback = %#v", state.Groups[1])
	}
}

func TestMoveTabRejectsInvalidMovesWithoutMutation(t *testing.T) {
	tests := []struct {
		name   string
		groups []Group
		target int
	}{
		{
			name: "out of range target",
			groups: []Group{
				{Key: "source", Tabs: []Tab{{Key: "editor", Closable: true}}},
				{Key: "destination", Tabs: []Tab{{Key: "terminal"}}},
			},
			target: 2,
		},
		{
			name: "disabled tab",
			groups: []Group{
				{Key: "source", Tabs: []Tab{{Key: "editor", Disabled: true}}},
				{Key: "destination", Tabs: []Tab{{Key: "terminal"}}},
			},
			target: -1,
		},
		{
			name: "duplicate destination key",
			groups: []Group{
				{Key: "source", Tabs: []Tab{{Key: "editor", Closable: true}}},
				{Key: "destination", Tabs: []Tab{{Key: "editor"}}},
			},
			target: -1,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState(test.groups)
			before := state.Clone()
			if state.MoveTab("source", "destination", "editor", test.target) {
				t.Fatal("invalid move succeeded")
			}
			if !reflect.DeepEqual(state, before) {
				t.Fatalf("state mutated after rejected move: %#v", state)
			}
		})
	}
}

func TestCloseTabSkipsDisabledNextTabs(t *testing.T) {
	state := NewState([]Group{{
		Key: "editor",
		Tabs: []Tab{
			{Key: "current", Closable: true},
			{Key: "disabled", Disabled: true},
			{Key: "next"},
		},
	}})

	if !state.CloseTab("editor", "current") {
		t.Fatal("closing current tab failed")
	}
	if got := state.ActiveTab("editor"); got != "next" {
		t.Fatalf("active tab after close = %q, want next", got)
	}
}

func TestStateSnapshotRestoreHandlesRenamesAndRemovedNodes(t *testing.T) {
	state := NewState([]Group{
		{Key: "editor", Tabs: []Tab{{Key: "renamed", Closable: true}, {Key: "kept", Closable: true}}},
		{Key: "panel", Tabs: []Tab{{Key: "terminal", Closable: true}}},
	})
	state.Chrome = ChromeState{SidebarVisible: false, BottomPanelVisible: true, StatusBarVisible: false}
	snapshot := Snapshot{
		Version:      0,
		ActiveGroup:  "old-editor",
		FocusedGroup: "old-editor",
		Groups:       []GroupSnapshot{{Key: "old-editor", TabOrder: []string{"old-tab", "kept", "removed"}, ActiveKey: "old-tab"}},
		Dock:         dock.Snapshot{Version: 0, Ratios: map[string]float32{"old-root": .25, "removed": .9}},
	}
	if err := state.Restore(snapshot, Migration{
		GroupAliases: map[string]string{"old-editor": "editor"},
		TabAliases:   map[string]string{"old-tab": "renamed"},
	}); err != nil {
		t.Fatal(err)
	}
	if state.ActiveGroup != "editor" || state.ActiveTab("editor") != "renamed" {
		t.Fatalf("restored selection = %#v", state)
	}
	if got := state.Groups[0].Tabs; len(got) != 2 || got[0].Key != "renamed" || got[1].Key != "kept" {
		t.Fatalf("restored order = %#v", got)
	}
	if state.Snapshot().Version != SnapshotVersion {
		t.Fatalf("snapshot version = %d", state.Snapshot().Version)
	}
	if state.Chrome != (ChromeState{SidebarVisible: false, BottomPanelVisible: true, StatusBarVisible: false}) {
		t.Fatalf("legacy restore overwrote chrome state: %#v", state.Chrome)
	}
}

func TestSnapshotRejectsFutureVersion(t *testing.T) {
	state := NewState([]Group{{Key: "editor"}})
	if err := state.Restore(Snapshot{Version: SnapshotVersion + 1}, Migration{}); err == nil {
		t.Fatal("future snapshot version was accepted")
	}
	if err := state.Restore(Snapshot{Dock: dock.Snapshot{Version: dock.SnapshotVersion + 1}}, Migration{}); err == nil {
		t.Fatal("future dock snapshot version was accepted")
	}
}

func TestSnapshotRestoreRejectsInvalidStructure(t *testing.T) {
	tests := []struct {
		name     string
		snapshot Snapshot
	}{
		{
			name: "duplicate tab order",
			snapshot: Snapshot{
				Version: SnapshotVersion,
				Groups:  []GroupSnapshot{{Key: "editor", TabOrder: []string{"main", "main"}}},
			},
		},
		{
			name: "invalid dock ratio",
			snapshot: Snapshot{
				Version: SnapshotVersion,
				Dock:    dock.Snapshot{Version: dock.SnapshotVersion, Ratios: map[string]float32{"root": 2}},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			state := NewState([]Group{{Key: "editor", Tabs: []Tab{{Key: "main"}}}})
			before := state.Clone()
			if err := state.Restore(test.snapshot, Migration{}); err == nil {
				t.Fatal("invalid snapshot structure was accepted")
			}
			if !reflect.DeepEqual(state, before) {
				t.Fatalf("state mutated after rejected restore: %#v", state)
			}
		})
	}
}

func TestSnapshotJSONRoundTripAndChromeState(t *testing.T) {
	state := NewState([]Group{{Key: "editor", Tabs: []Tab{{Key: "main", Closable: true}}}})
	state.ToggleSidebar()
	state.SetBottomPanelVisible(false)
	state.SetDockSnapshot(dock.Snapshot{
		Version:      dock.SnapshotVersion,
		RootKey:      "root",
		Ratios:       map[string]float32{"root": .35},
		MaximizedKey: "editor",
	})
	snapshot := state.Snapshot()
	data, err := MarshalSnapshot(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, `"activeGroup"`) || !strings.Contains(text, `"chrome"`) || strings.Contains(text, "Content") {
		t.Fatalf("snapshot JSON = %s", text)
	}
	decoded, err := UnmarshalSnapshot(data)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Version != SnapshotVersion || decoded.ActiveGroup != "editor" || decoded.Chrome.SidebarVisible || decoded.Chrome.BottomPanelVisible {
		t.Fatalf("decoded snapshot = %#v", decoded)
	}
	if decoded.Dock.Ratios["root"] != .35 || decoded.Dock.MaximizedKey != "editor" {
		t.Fatalf("decoded dock = %#v", decoded.Dock)
	}
	var viaJSON Snapshot
	if err := json.Unmarshal(data, &viaJSON); err != nil || viaJSON.Version != SnapshotVersion {
		t.Fatalf("json.Unmarshal = %#v/%v", viaJSON, err)
	}
}

func TestSnapshotJSONRejectsInvalidStructure(t *testing.T) {
	for name, data := range map[string]string{
		"duplicate groups": `{"version":1,"groups":[{"key":"editor"},{"key":"editor"}]}`,
		"duplicate tabs":   `{"version":1,"groups":[{"key":"editor","tabOrder":["main","main"]}]}`,
		"invalid ratio":    `{"version":1,"dock":{"version":1,"ratios":{"root":2}}}`,
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := UnmarshalSnapshot([]byte(data)); err == nil {
				t.Fatal("invalid snapshot was accepted")
			}
		})
	}
}

func TestControllerEmitsBindings(t *testing.T) {
	controller := NewController(NewState([]Group{{Key: "editor", Tabs: []Tab{{Key: "one"}, {Key: "two"}}}}))
	var events []Event
	controller.OnEvent(func(event Event) { events = append(events, event) })
	if !controller.ActivateTab("editor", "two") {
		t.Fatal("controller activation failed")
	}
	controller.SetDockSnapshot(dock.Snapshot{Version: dock.SnapshotVersion, RootKey: "root"})
	if len(events) != 2 || events[0].Kind != EventTabActivated || events[1].Kind != EventDockChanged {
		t.Fatalf("events = %#v", events)
	}
}

func TestControllerMoveTabEmitsFinalIndex(t *testing.T) {
	controller := NewController(NewState([]Group{
		{Key: "source", Tabs: []Tab{{Key: "editor", Closable: true}}},
		{Key: "destination", Tabs: []Tab{{Key: "terminal"}}},
	}))
	var events []Event
	controller.OnEvent(func(event Event) { events = append(events, event) })
	if !controller.MoveTab("source", "destination", "editor", -1) {
		t.Fatal("controller move failed")
	}
	if len(events) != 1 || events[0].Kind != EventTabMoved || events[0].Index != 1 || events[0].PreviousIndex != 0 {
		t.Fatalf("move event = %#v", events)
	}
	if source, _ := controller.State().Group("source"); len(source.Tabs) != 0 {
		t.Fatalf("source group after move = %#v", source.Tabs)
	}
	if destination, _ := controller.State().Group("destination"); len(destination.Tabs) != 2 || destination.Tabs[1].Key != "editor" {
		t.Fatalf("destination group after move = %#v", destination.Tabs)
	}
}
