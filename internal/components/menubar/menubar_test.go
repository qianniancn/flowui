package menubar

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/menu"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type fixedMenubarWidget struct {
	size image.Point
}

func (w fixedMenubarWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func menubarTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func menubarTestWidget() Widget {
	trigger := fixedMenubarWidget{size: image.Pt(36, 16)}
	return New("application-menu", []Item{
		NewMenu("file", "File", []menu.Item{{Key: "new", Label: "New"}}).Trigger(trigger).Width(80),
		NewMenu("edit", "Edit", []menu.Item{{Key: "copy", Label: "Copy"}}).Trigger(trigger).Width(80),
		NewMenu("view", "View", []menu.Item{{Key: "zoom", Label: "Zoom"}}).Trigger(trigger).Width(80),
	})
}

func TestMenubarAccessKeyIsValueSemantic(t *testing.T) {
	base := NewMenu("file", "File", nil)
	configured := base.AccessKey('f')
	if base.accessKey != 0 {
		t.Fatalf("AccessKey mutated base: %#v", base)
	}
	if configured.accessKey != 'f' {
		t.Fatalf("AccessKey = %q", configured.accessKey)
	}
}

func TestMenubarItemDelegatesMenuConfiguration(t *testing.T) {
	base := NewMenu("file", "File", []menu.Item{{Key: "new", Label: "New"}})
	configured := base.
		Trigger(fixedMenubarWidget{size: image.Pt(20, 10)}).
		OnAction(func(string) {}).
		OnCheckedChange(func(string, bool) {}).
		OnRadioChange(func(string, string) {}).
		CloseOnSelect(false).
		Width(180).
		Disabled(true)
	if base.trigger != nil || base.disabled || configured.trigger == nil || !configured.disabled {
		t.Fatalf("Menubar item value semantics = base %#v configured %#v", base, configured)
	}
}

func TestMenubarRejectsInvalidItems(t *testing.T) {
	tests := []struct {
		name  string
		items []Item
	}{
		{"empty key", []Item{NewMenu("", "File", nil)}},
		{"empty label", []Item{NewMenu("file", "", nil)}},
		{"duplicate key", []Item{NewMenu("file", "File", nil), NewMenu("file", "Again", nil)}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid Menubar items did not panic")
				}
			}()
			layoutMenubarFrame(menubarTestContext(), new(input.Router), New("invalid", test.items), time.Unix(1, 0))
		})
	}
}

func TestMenubarClickTogglesAndHoverSwitchesOnlyWhileOpen(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(1, 0)
	layoutMenubarFrame(ctx, router, widget, now)

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(90, 16)})
	layoutMenubarFrame(ctx, router, widget, now.Add(time.Millisecond))
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	if state == nil || state.openKey != "" {
		t.Fatalf("closed Menubar hover opened menu: %#v", state)
	}

	now = clickMenubarPoint(ctx, router, widget, now.Add(2*time.Millisecond), f32.Pt(20, 16))
	if state.openKey != "file" {
		t.Fatalf("clicked Menubar open key = %q", state.openKey)
	}
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(90, 16)})
	layoutMenubarFrame(ctx, router, widget, now.Add(time.Millisecond))
	if state.openKey != "edit" {
		t.Fatalf("hover-switched Menubar open key = %q", state.openKey)
	}

	now = clickMenubarPoint(ctx, router, widget, now.Add(2*time.Millisecond), f32.Pt(90, 16))
	if state.openKey != "" {
		t.Fatalf("second trigger click did not close Menubar: %q", state.openKey)
	}
}

func TestMenubarKeyboardNavigationAndOpen(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(2, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	file := state.trigger("file")
	edit := state.trigger("edit")
	view := state.trigger("view")

	router.Source().Execute(key.FocusCmd{Tag: &file.clickable})
	layoutMenubarFrame(ctx, router, widget, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutMenubarFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	if !router.Source().Focused(&edit.clickable) {
		t.Fatal("Right Arrow did not move Menubar focus")
	}
	router.Queue(key.Event{Name: key.NameEnd, State: key.Press})
	layoutMenubarFrame(ctx, router, widget, now.Add(3*time.Millisecond))
	if !router.Source().Focused(&view.clickable) {
		t.Fatal("End did not focus the last Menubar trigger")
	}
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutMenubarFrame(ctx, router, widget, now.Add(4*time.Millisecond))
	if !router.Source().Focused(&file.clickable) {
		t.Fatal("Menubar focus did not loop")
	}
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutMenubarFrame(ctx, router, widget, now.Add(5*time.Millisecond))
	layoutMenubarFrame(ctx, router, widget, now.Add(menubarEnter+6*time.Millisecond))
	if state.openKey != "file" || state.focusPanelKey != "" {
		t.Fatalf("keyboard-opened Menubar = open %q pending focus %q", state.openKey, state.focusPanelKey)
	}
}

func TestMenubarStationaryHoverDoesNotUndoMenuKeyboardSwitch(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(8, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	now = clickMenubarPoint(ctx, router, widget, now.Add(time.Millisecond), f32.Pt(20, 16))
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	if state.openKey != "file" || state.hoveredKey != "file" {
		t.Fatalf("initial Menubar hover state = open %q hovered %q", state.openKey, state.hoveredKey)
	}
	var ops op.Ops
	switchGtx := layout.Context{Ops: &ops, Source: router.Source(), Now: now.Add(time.Millisecond)}
	widget.switchFromMenu(ctx, switchGtx, state, "file", 1)
	layoutMenubarFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	if state.openKey != "edit" {
		t.Fatalf("stationary hover switched Menubar back to %q", state.openKey)
	}
}

func TestMenubarLoopFocusFalseAndDisabledItem(t *testing.T) {
	trigger := fixedMenubarWidget{size: image.Pt(36, 16)}
	widget := New("bounded-menu", []Item{
		NewMenu("file", "File", nil).Trigger(trigger),
		NewMenu("edit", "Edit", nil).Trigger(trigger).Disabled(true),
		NewMenu("view", "View", nil).Trigger(trigger),
	}).LoopFocus(false)
	ctx := menubarTestContext()
	router := new(input.Router)
	now := time.Unix(3, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[menubarState](ctx, "bounded-menu", stateSlotMenubar)
	file := state.trigger("file")
	view := state.trigger("view")
	router.Source().Execute(key.FocusCmd{Tag: &file.clickable})
	layoutMenubarFrame(ctx, router, widget, now.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutMenubarFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	if !router.Source().Focused(&view.clickable) {
		t.Fatal("Menubar navigation did not skip disabled trigger")
	}
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutMenubarFrame(ctx, router, widget, now.Add(3*time.Millisecond))
	if !router.Source().Focused(&view.clickable) {
		t.Fatal("LoopFocus(false) moved beyond the last trigger")
	}
}

func TestMenubarSecondPanelUsesItsTriggerAnchorAndActionCloses(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	action := ""
	widget := menubarTestWidget()
	widget.items[1] = widget.items[1].OnAction(func(key string) { action = key })
	now := time.Unix(4, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	now = clickMenubarPoint(ctx, router, widget, now.Add(time.Millisecond), f32.Pt(90, 16))
	layoutMenubarFrame(ctx, router, widget, now.Add(menubarEnter+time.Millisecond))

	now = clickMenubarPoint(ctx, router, widget, now.Add(2*time.Millisecond), f32.Pt(100, 50))
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	if action != "copy" || state.openKey != "" {
		t.Fatalf("second Menubar action = %q open key %q", action, state.openKey)
	}
	if !router.Source().Focused(&state.trigger("edit").clickable) {
		t.Fatal("Menubar action did not restore trigger focus")
	}
}

func TestMenubarControlledOpenRequestsChange(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	requested := "unchanged"
	widget := menubarTestWidget().OpenKey("").OnOpenChange(func(key string) { requested = key })
	now := time.Unix(5, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	clickMenubarPoint(ctx, router, widget, now.Add(time.Millisecond), f32.Pt(20, 16))
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	if requested != "file" || state.current(widget) != "" {
		t.Fatalf("controlled Menubar request = %q current %q", requested, state.current(widget))
	}
}

func TestMenubarInitialOpenClaimsAndReleasesExclusiveState(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	now := time.Unix(13, 0)
	widget := menubarTestWidget().DefaultOpenKey("file")
	layoutMenubarFrame(ctx, router, widget, now)
	if got := frame.ActiveExclusive(ctx, menubarExclusive); got != "application-menu" {
		t.Fatalf("default-open Menubar exclusive key = %q", got)
	}

	controlled := widget.OpenKey("")
	layoutMenubarFrame(ctx, router, controlled, now.Add(time.Millisecond))
	if got := frame.ActiveExclusive(ctx, menubarExclusive); got != "" {
		t.Fatalf("closed controlled Menubar retained exclusive key %q", got)
	}
}

func TestMenubarDisabledOpenItemClearsUncontrolledState(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	now := time.Unix(14, 0)
	widget := menubarTestWidget().DefaultOpenKey("file")
	layoutMenubarFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	if state == nil || state.openKey != "file" {
		t.Fatalf("default-open Menubar state = %#v", state)
	}

	disabled := widget
	disabled.items = append([]Item(nil), widget.items...)
	disabled.items[0] = disabled.items[0].Disabled(true)
	layoutMenubarFrame(ctx, router, disabled, now.Add(time.Millisecond))
	if state.openKey != "" || frame.ActiveExclusive(ctx, menubarExclusive) != "" {
		t.Fatalf("disabled open item retained state key %q", state.openKey)
	}

	layoutMenubarFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	if state.current(widget) != "" {
		t.Fatalf("re-enabled item unexpectedly reopened Menubar with %q", state.current(widget))
	}
}

func TestMenubarPreservesConfiguredMenuDisabledState(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	action := ""
	content := menu.Menu("disabled-content", []menu.Item{{Key: "action", Label: "Action"}}).
		OnAction(func(key string) { action = key }).
		Disabled(true).
		Width(80)
	widget := New("disabled-content-bar", []Item{
		NewMenuContent("file", "File", content).Trigger(fixedMenubarWidget{size: image.Pt(36, 16)}),
	})
	now := time.Unix(10, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	now = clickMenubarPoint(ctx, router, widget, now.Add(time.Millisecond), f32.Pt(20, 16))
	layoutMenubarFrame(ctx, router, widget, now.Add(menubarEnter+time.Millisecond))
	clickMenubarPoint(ctx, router, widget, now.Add(menubarEnter+2*time.Millisecond), f32.Pt(30, 50))
	if action != "" {
		t.Fatalf("configured disabled Menu invoked %q", action)
	}
}

func TestMenubarModalControlsOutsidePointerPassThrough(t *testing.T) {
	for _, test := range []struct {
		name      string
		modal     bool
		wantClick int
	}{
		{name: "modal", modal: true, wantClick: 0},
		{name: "non-modal", modal: false, wantClick: 1},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := menubarTestContext()
			router := new(input.Router)
			underlying := new(widget.Clickable)
			clicks := 0
			menubar := menubarTestWidget().Modal(test.modal)
			now := time.Unix(11, 0)
			layoutMenubarWithUnderlying(ctx, router, menubar, underlying, &clicks, now)
			now = clickMenubarPointWithUnderlying(ctx, router, menubar, underlying, &clicks, now.Add(time.Millisecond), f32.Pt(20, 16))
			layoutMenubarWithUnderlying(ctx, router, menubar, underlying, &clicks, now.Add(menubarEnter+time.Millisecond))
			clickMenubarPointWithUnderlying(ctx, router, menubar, underlying, &clicks, now.Add(menubarEnter+2*time.Millisecond), f32.Pt(325, 215))
			state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
			if clicks != test.wantClick || state.openKey != "" {
				t.Fatalf("outside click = count %d open %q, want count %d", clicks, state.openKey, test.wantClick)
			}
		})
	}
}

func TestMenubarOutsidePressClosesImmediately(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget()
	now := time.Unix(12, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	now = clickMenubarPoint(ctx, router, widget, now.Add(time.Millisecond), f32.Pt(20, 16))
	layoutMenubarFrame(ctx, router, widget, now.Add(menubarEnter+time.Millisecond))
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 9,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(420, 300),
	})
	layoutMenubarFrame(ctx, router, widget, now.Add(menubarEnter+2*time.Millisecond))
	if state.openKey != "" {
		t.Fatalf("outside press did not close Menubar: %q", state.openKey)
	}
}

func TestMenubarThemeAndSemantics(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Menubar
	if tokens.TriggerHeight != 32 || tokens.TriggerPaddingX != 12 || tokens.TriggerRadius != 8 || tokens.TriggerTextSize != 14 || tokens.PanelGap != 4 {
		t.Fatalf("Menubar theme tokens = %#v", tokens)
	}
	activeTheme := theme.DefaultTheme()
	compact := menubarTestWidget().Compact(true).themeTokens(&activeTheme)
	if compact.TriggerHeight != 28 || compact.TriggerPaddingX != 8 || compact.TriggerRadius != 4 || compact.TriggerTextSize != 13 {
		t.Fatalf("compact Menubar theme tokens = %#v", compact)
	}
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget().Alt("Application menu")
	layoutMenubarFrame(ctx, router, widget, time.Unix(6, 0))
	nodes := router.AppendSemantics(nil)
	if !menubarSemanticDescription(nodes, "Application menu") {
		t.Fatal("Menubar semantic description is missing")
	}
	labels := menubarSemanticButtonLabels(nodes, nil)
	for _, label := range []string{"File", "Edit", "View"} {
		if !labels[label] {
			t.Fatalf("Menubar semantic button %q is missing: %v", label, labels)
		}
	}
}

func TestMenubarResolvesTriggerInsideTransformedBar(t *testing.T) {
	bar := image.Rect(0, 0, 180, 32)
	resolved := image.Rect(20, 40, 380, 104)
	trigger := image.Rect(60, 0, 120, 32)
	if got, want := menubarResolvedChildRect(bar, resolved, trigger), image.Rect(140, 40, 260, 104); got != want {
		t.Fatalf("resolved Menubar trigger = %v, want %v", got, want)
	}
}

func TestMenubarDisabledDoesNotOpen(t *testing.T) {
	ctx := menubarTestContext()
	router := new(input.Router)
	widget := menubarTestWidget().Disabled(true)
	now := time.Unix(7, 0)
	layoutMenubarFrame(ctx, router, widget, now)
	clickMenubarPoint(ctx, router, widget, now.Add(time.Millisecond), f32.Pt(20, 16))
	state, _ := frame.PeekState[menubarState](ctx, "application-menu", stateSlotMenubar)
	if state.openKey != "" || frame.HasTopOverlay(ctx) {
		t.Fatalf("disabled Menubar opened: %#v", state)
	}
}

func clickMenubarPoint(ctx *frame.Context, router *input.Router, widget Widget, now time.Time, position f32.Point) time.Time {
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutMenubarFrame(ctx, router, widget, now)
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutMenubarFrame(ctx, router, widget, now.Add(time.Millisecond))
	return now.Add(time.Millisecond)
}

func clickMenubarPointWithUnderlying(ctx *frame.Context, router *input.Router, menubar Widget, underlying *widget.Clickable, clicks *int, now time.Time, position f32.Point) time.Time {
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: position})
	layoutMenubarWithUnderlying(ctx, router, menubar, underlying, clicks, now)
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: position})
	layoutMenubarWithUnderlying(ctx, router, menubar, underlying, clicks, now.Add(time.Millisecond))
	return now.Add(time.Millisecond)
}

func layoutMenubarFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(480, 360)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	barGtx := gtx
	barGtx.Constraints = layout.Constraints{Max: image.Pt(300, 120)}
	dims := widget.Layout(ctx, barGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func layoutMenubarWithUnderlying(ctx *frame.Context, router *input.Router, menubar Widget, underlying *widget.Clickable, clicks *int, now time.Time) {
	viewport := image.Pt(480, 360)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	for underlying.Clicked(gtx) {
		*clicks++
	}
	buttonGtx := gtx
	buttonGtx.Constraints = layout.Exact(image.Pt(50, 30))
	offset := op.Offset(image.Pt(300, 200)).Push(gtx.Ops)
	underlying.Layout(buttonGtx, func(layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(50, 30)}
	})
	offset.Pop()
	barGtx := gtx
	barGtx.Constraints = layout.Constraints{Max: image.Pt(300, 120)}
	menubar.Layout(ctx, barGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func menubarSemanticDescription(nodes []input.SemanticNode, description string) bool {
	for _, node := range nodes {
		if node.Desc.Description == description || menubarSemanticDescription(node.Children, description) {
			return true
		}
	}
	return false
}

func menubarSemanticButtonLabels(nodes []input.SemanticNode, labels map[string]bool) map[string]bool {
	if labels == nil {
		labels = make(map[string]bool)
	}
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			labels[node.Desc.Label] = true
		}
		menubarSemanticButtonLabels(node.Children, labels)
	}
	return labels
}
