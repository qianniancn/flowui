package dock

import (
	"encoding/json"
	"image"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/state"
)

func TestDockLayoutRecursivelySplitsPanels(t *testing.T) {
	left := new(dockProbe)
	top := new(dockProbe)
	bottom := new(dockProbe)
	root := Split("workspace", Horizontal,
		Panel("sidebar", left),
		Split("editor-panel", Vertical,
			Panel("editor", top),
			Panel("panel", bottom),
		).Ratio(.5),
	).Ratio(.25)
	widget := New("workbench", root)

	widget.Layout(dockTestContext(), layout.Context{Constraints: layout.Exact(image.Pt(400, 300)), Ops: new(op.Ops)})
	if left.constraints != layout.Exact(image.Pt(100, 300)) {
		t.Fatalf("left constraints = %#v, want 100x300", left.constraints)
	}
	if top.constraints != layout.Exact(image.Pt(299, 150)) {
		t.Fatalf("top constraints = %#v, want 299x150", top.constraints)
	}
	if bottom.constraints != layout.Exact(image.Pt(299, 149)) {
		t.Fatalf("bottom constraints = %#v, want 299x149", bottom.constraints)
	}
}

func TestDockLayoutReportsAndRestoresRatioSnapshot(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	left := new(dockProbe)
	right := new(dockProbe)
	var changed Snapshot
	widget := func(snapshot *Snapshot) DockLayoutWidget {
		value := New("workbench", Split("workspace", Horizontal,
			Panel("left", left),
			Panel("right", right),
		).Ratio(.5)).OnChange(func(value Snapshot) { changed = value })
		if snapshot != nil {
			value = value.Snapshot(*snapshot)
		}
		return value
	}
	start := time.Unix(1, 0)
	layoutDockFrame(ctx, router, widget(nil), start, image.Pt(400, 200))
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(200, 100)})
	layoutDockFrame(ctx, router, widget(nil), start.Add(time.Millisecond), image.Pt(400, 200))
	router.Queue(
		pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(280, 100)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(280, 100)},
	)
	layoutDockFrame(ctx, router, widget(nil), start.Add(2*time.Millisecond), image.Pt(400, 200))
	if ratio := changed.Ratios["workspace"]; math.Abs(float64(ratio-.7)) > .02 {
		t.Fatalf("snapshot ratio = %v, want approximately .7", ratio)
	}

	newCtx := dockTestContext()
	newRouter := new(input.Router)
	restoredLeft := new(dockProbe)
	restoredRight := new(dockProbe)
	restored := New("restored-workbench", Split("workspace", Horizontal,
		Panel("left", restoredLeft),
		Panel("right", restoredRight),
	).Ratio(.5)).Snapshot(changed)
	layoutDockFrame(newCtx, newRouter, restored, start, image.Pt(400, 200))
	if got := restoredLeft.constraints.Max.X; math.Abs(float64(got-279)) > 2 {
		t.Fatalf("restored first width = %d, want approximately 279", got)
	}
}

func TestDockLayoutCollapseUsesFullRegionAndRetainsHiddenBranch(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	left := new(dockProbe)
	right := new(dockProbe)
	root := func(collapsed bool) Node {
		return Split("workspace", Horizontal,
			Panel("left", left),
			Panel("right", right),
		).Ratio(.25).FirstCollapsed(collapsed)
	}
	start := time.Unix(1, 0)
	layoutDockFrame(ctx, router, New("workbench", root(true)).KeepAlive(true), start, image.Pt(400, 200))
	if left.layouts != 1 || right.layouts != 1 {
		t.Fatalf("collapsed layouts = %d/%d, want hidden left and visible right", left.layouts, right.layouts)
	}
	if right.constraints != layout.Exact(image.Pt(400, 200)) {
		t.Fatalf("visible collapsed branch = %#v, want full layout", right.constraints)
	}
	firstState := left.state
	layoutDockFrame(ctx, router, New("workbench", root(false)).KeepAlive(true), start.Add(time.Millisecond), image.Pt(400, 200))
	if left.layouts != 2 || left.state != firstState {
		t.Fatal("expanded branch did not restore its retained state")
	}
}

func TestDockLayoutControlledCollapseState(t *testing.T) {
	left := new(dockProbe)
	right := new(dockProbe)
	widget := New("workbench", Split("workspace", Horizontal,
		Panel("left", left),
		Panel("right", right),
	)).Snapshot(Snapshot{Collapsed: map[string]CollapseState{
		"workspace": {Second: true},
	}})
	widget.Layout(dockTestContext(), layout.Context{Constraints: layout.Exact(image.Pt(320, 200)), Ops: new(op.Ops)})
	if left.constraints != layout.Exact(image.Pt(320, 200)) || right.layouts != 0 {
		t.Fatalf("controlled collapse left=%#v right layouts=%d", left.constraints, right.layouts)
	}
}

func TestDockLayoutMaximizesAndRestoresNode(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	left := new(dockProbe)
	right := new(dockProbe)
	root := Split("workspace", Horizontal,
		Panel("left", left),
		Panel("right", right),
	)
	start := time.Unix(1, 0)
	layoutDockFrame(ctx, router, New("workbench", root).MaximizedKey("right").KeepAlive(true), start, image.Pt(320, 200))
	if left.layouts != 1 || right.layouts != 1 {
		t.Fatalf("maximized layouts = %d/%d, want hidden/visible", left.layouts, right.layouts)
	}
	if right.constraints != layout.Exact(image.Pt(320, 200)) {
		t.Fatalf("maximized panel constraints = %#v, want full region", right.constraints)
	}
	leftState := left.state
	layoutDockFrame(ctx, router, New("workbench", root).MaximizedKey("").KeepAlive(true), start.Add(time.Millisecond), image.Pt(320, 200))
	if left.layouts != 2 || right.layouts != 2 || left.state != leftState {
		t.Fatalf("restored layouts/state = %d/%d/%v, want 2/2/true", left.layouts, right.layouts, left.state == leftState)
	}
}

func TestDockLayoutScopesNodeStateToLayoutKey(t *testing.T) {
	ctx := dockTestContext()
	firstLeft := new(dockProbe)
	firstRight := new(dockProbe)
	secondLeft := new(dockProbe)
	secondRight := new(dockProbe)
	root := func(left, right *dockProbe) Node {
		return Split("workspace", Horizontal,
			Panel("left", left),
			Panel("right", right),
		)
	}

	var ops op.Ops
	size := image.Pt(320, 200)
	gtx := layout.Context{Constraints: layout.Exact(size), Ops: &ops}
	frame.BeginFrameWithViewport(ctx, size)
	New("first-workbench", root(firstLeft, firstRight)).Layout(ctx, gtx)
	New("second-workbench", root(secondLeft, secondRight)).Layout(ctx, gtx)
	frame.EndFrame(ctx)

	if firstLeft.key == secondLeft.key || firstLeft.key == "" {
		t.Fatalf("scoped panel keys = %q/%q, want distinct non-empty keys", firstLeft.key, secondLeft.key)
	}
}

func TestDockLayoutKeepsStateAcrossRepeatedCollapses(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	left := new(dockProbe)
	right := new(dockProbe)
	root := func(collapsed bool) Node {
		return Split("workspace", Horizontal,
			Panel("left", left),
			Panel("right", right),
		).FirstCollapsed(collapsed)
	}
	start := time.Unix(1, 0)

	layoutDockFrame(ctx, router, New("workbench", root(false)).KeepAlive(true), start, image.Pt(400, 200))
	firstState := left.state
	layoutDockFrame(ctx, router, New("workbench", root(true)).KeepAlive(true), start.Add(time.Millisecond), image.Pt(400, 200))
	layoutDockFrame(ctx, router, New("workbench", root(false)).KeepAlive(true), start.Add(2*time.Millisecond), image.Pt(400, 200))
	layoutDockFrame(ctx, router, New("workbench", root(true)).KeepAlive(true), start.Add(3*time.Millisecond), image.Pt(400, 200))
	layoutDockFrame(ctx, router, New("workbench", root(false)).KeepAlive(true), start.Add(4*time.Millisecond), image.Pt(400, 200))

	if left.state != firstState {
		t.Fatal("repeatedly collapsed panel did not retain its state")
	}
}

func TestDockLayoutInitializesNewMaximizedHiddenBranch(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	oldLeft := new(dockProbe)
	newLeft := new(dockProbe)
	right := new(dockProbe)
	root := func(leftKey string, left *dockProbe) Node {
		return Split("workspace", Horizontal,
			Panel(leftKey, left),
			Panel("right", right),
		)
	}
	start := time.Unix(1, 0)

	layoutDockFrame(ctx, router, New("workbench", root("old-left", oldLeft)).MaximizedKey("right").KeepAlive(true), start, image.Pt(320, 200))
	layoutDockFrame(ctx, router, New("workbench", root("new-left", newLeft)).MaximizedKey("right").KeepAlive(true), start.Add(time.Millisecond), image.Pt(320, 200))
	if newLeft.layouts != 1 {
		t.Fatalf("new hidden branch layouts = %d, want 1", newLeft.layouts)
	}
}

func TestDockLayoutInitializesBothCollapsedBranches(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	left := new(dockProbe)
	right := new(dockProbe)
	root := Split("workspace", Horizontal,
		Panel("left", left),
		Panel("right", right),
	).FirstCollapsed(true).SecondCollapsed(true)

	layoutDockFrame(ctx, router, New("workbench", root).KeepAlive(true), time.Unix(1, 0), image.Pt(320, 200))
	if left.layouts != 1 || right.layouts != 1 {
		t.Fatalf("collapsed branch layouts = %d/%d, want 1/1", left.layouts, right.layouts)
	}
}

func TestDockLayoutReleasesRemovedNodeState(t *testing.T) {
	ctx := dockTestContext()
	router := new(input.Router)
	left := new(dockProbe)
	replacement := new(dockProbe)
	right := new(dockProbe)
	root := func(leftKey string, leftPanel *dockProbe) Node {
		return Split("workspace", Horizontal,
			Panel(leftKey, leftPanel),
			Panel("right", right),
		)
	}
	start := time.Unix(1, 0)

	layoutDockFrame(ctx, router, New("workbench", root("left", left)).KeepAlive(true), start, image.Pt(320, 200))
	leftKey := left.key
	layoutDockFrame(ctx, router, New("workbench", root("replacement", replacement)).KeepAlive(true), start.Add(time.Millisecond), image.Pt(320, 200))
	if _, ok := frame.PeekState[widget.Clickable](ctx, leftKey, "probe"); ok {
		t.Fatal("removed dock node state was retained")
	}
}

func TestDockLayoutRejectsDuplicateNodeKeys(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("duplicate dock node keys did not panic")
		}
	}()
	New("workbench", Split("workspace", Horizontal,
		Panel("shared", nil),
		Panel("shared", nil),
	)).Layout(dockTestContext(), layout.Context{Ops: new(op.Ops)})
}

func TestSnapshotMigrateRenamesAndDropsNodes(t *testing.T) {
	snapshot := Snapshot{
		Version:      0,
		RootKey:      "old-root",
		Ratios:       map[string]float32{"old-root": .2, "removed": .9},
		Collapsed:    map[string]CollapseState{"old-root": {First: true}, "removed": {Second: true}},
		MaximizedKey: "removed",
	}
	got := snapshot.Migrate("root", map[string]struct{}{"root": {}}, map[string]string{"old-root": "root"})
	if got.Version != SnapshotVersion || got.RootKey != "root" {
		t.Fatalf("migrated header = %#v", got)
	}
	if got.Ratios["root"] != .2 || len(got.Ratios) != 1 || len(got.Collapsed) != 1 || got.MaximizedKey != "" {
		t.Fatalf("migrated state = %#v", got)
	}

	snapshot.Ratios["root"] = .7
	snapshot.Collapsed["root"] = CollapseState{Second: true}
	got = snapshot.Migrate("root", map[string]struct{}{"root": {}}, map[string]string{"old-root": "root"})
	if got.Ratios["root"] != .7 || !got.Collapsed["root"].Second {
		t.Fatalf("canonical keys did not win alias collisions: %#v", got)
	}
}

func TestSnapshotJSONRoundTrip(t *testing.T) {
	snapshot := Snapshot{
		Version:      SnapshotVersion,
		RootKey:      "root",
		Ratios:       map[string]float32{"root": .4},
		Collapsed:    map[string]CollapseState{"root": {Second: true}},
		MaximizedKey: "editor",
	}
	data, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatal(err)
	}
	var decoded Snapshot
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Version != SnapshotVersion || decoded.RootKey != "root" || decoded.Ratios["root"] != .4 || !decoded.Collapsed["root"].Second || decoded.MaximizedKey != "editor" {
		t.Fatalf("decoded snapshot = %#v", decoded)
	}
}

type dockProbe struct {
	layouts     int
	constraints layout.Constraints
	key         string
	state       *widget.Clickable
}

func (p *dockProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.constraints = gtx.Constraints
	p.key = frame.ClaimKey(ctx, state.KindCustom, "control")
	p.state = frame.UseState[widget.Clickable](ctx, p.key, "probe")
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func dockTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutDockFrame(ctx *frame.Context, router *input.Router, widget DockLayoutWidget, now time.Time, size image.Point) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(size), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, size)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}
