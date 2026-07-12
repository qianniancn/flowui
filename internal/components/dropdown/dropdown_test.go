package dropdown

import (
	"image"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/button"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/overlay"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type fixedWidget struct {
	size image.Point
}

func (w fixedWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type layoutProbe struct {
	count int
}

func (p *layoutProbe) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	p.count++
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(120, 24))}
}

func dropdownTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func dropdownTestWidget(items []Item) Widget {
	return New("actions", fixedWidget{size: image.Pt(120, 40)}, items)
}

func TestDropdownOptionsUseValueSemantics(t *testing.T) {
	base := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}})
	if base.placement != overlay.PopoverBottomStart {
		t.Fatalf("default placement = %v, want bottom-start", base.placement)
	}
	configured := base.
		Open(true).
		DefaultOpen(true).
		OnOpenChange(func(bool) {}).
		TriggerMode(TriggerLongPress).
		Offset(8).
		ShouldFlip(false).
		AvoidOverflow(false).
		CloseOnSelect(false).
		Disabled(true).
		Width(256)
	if base.hasOpen || base.disabled || base.hasOffset || base.triggerMode != TriggerPress {
		t.Fatalf("base dropdown mutated: %#v", base)
	}
	if !configured.hasOpen || !configured.open || !configured.hasDefaultOpen || !configured.defaultOpen || configured.triggerMode != TriggerLongPress || !configured.disabled {
		t.Fatalf("configured dropdown = %#v", configured)
	}
	if !configured.hasOffset || configured.offset != 8 || !configured.hasShouldFlip || configured.shouldFlip || !configured.hasAvoidOverflow || configured.avoidOverflow {
		t.Fatal("dropdown overlay options were not retained")
	}
}

func TestDropdownUsesMenuThemeAndKeepsOnlyTriggerTokens(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	dropdownTokens := activeTheme.Components.Dropdown
	menuTokens := activeTheme.Components.Menu
	if dropdownTokens.PanelGap != 4 || dropdownTokens.TriggerFocusRingWidth != 2 || dropdownTokens.TriggerFocusRingOffset != 2 || dropdownTokens.TriggerPressedScale != 0.97 {
		t.Fatalf("dropdown trigger tokens = %#v", dropdownTokens)
	}
	if menuTokens.Width != 220 || menuTokens.Padding != 6 || menuTokens.ItemMinHeight != 36 || menuTokens.ItemRadius != 16 || menuTokens.ShortcutTextSize != 14 {
		t.Fatalf("dropdown menu tokens = %#v", menuTokens)
	}
}

func TestDropdownFocusRingStaysInsideCustomTrigger(t *testing.T) {
	trigger := image.Rect(0, 0, 120, 40)
	ring, radius := dropdownFocusRingGeometry(trigger, 12, 2, 2)
	if ring != image.Rect(3, 3, 117, 37) || radius != 9 {
		t.Fatalf("Dropdown focus ring = %v radius %d", ring, radius)
	}
	if !ring.In(trigger) {
		t.Fatalf("Dropdown focus ring %v escapes trigger %v", ring, trigger)
	}
}

func TestDropdownPressOpensAndFocusesFirstItem(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}, {Key: "delete", Label: "Delete"}})
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownEnterDuration+3*time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if state == nil || !state.open || !frame.HasTopOverlay(ctx) {
		t.Fatalf("dropdown open state = %#v", state)
	}
	if state.focusFirst {
		t.Fatal("Dropdown did not hand initial focus to Menu")
	}
}

func TestDropdownButtonTriggerUsesDropdownClickable(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := New("button-trigger", button.Button("trigger", fixedWidget{size: image.Pt(80, 24)}), []Item{{Key: "open", Label: "Open"}})
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "button-trigger", stateSlotDropdown)
	if state == nil || !state.open || len(state.trigger.History()) == 0 {
		t.Fatalf("button trigger state = %#v", state)
	}
	if frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("pointer-clicked Button trigger exposed keyboard-visible focus")
	}
}

func TestDropdownKeyboardArrowOpens(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}})
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	router.Queue(key.Event{Name: key.NameDownArrow, State: key.Press})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	if !state.open || !state.focusVisible {
		t.Fatalf("keyboard dropdown = open %v focus visible %v", state.open, state.focusVisible)
	}
}

func TestDropdownKeyboardClickPreservesKeyboardFocusOrigin(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}})
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	router.Source().Execute(key.FocusCmd{Tag: &state.trigger})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(key.Event{Name: key.NameReturn, State: key.Press})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	router.Queue(key.Event{Name: key.NameReturn, State: key.Release})
	layoutDropdownFrame(ctx, router, widget, start.Add(3*time.Millisecond))
	if !state.open || !state.focusVisible {
		t.Fatalf("keyboard click = open %v focus visible %v", state.open, state.focusVisible)
	}
	if !frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("keyboard-clicked Dropdown trigger was recorded as pointer focus")
	}
}

func TestDropdownLongPressTrigger(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).TriggerMode(TriggerLongPress)
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 7, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if state == nil || !state.touchTracking {
		t.Fatalf("long press did not start tracking: %#v", state)
	}
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownLongPress+2*time.Millisecond))
	if !state.open {
		t.Fatal("long press did not open dropdown")
	}
}

func TestDropdownLongPressMovementCancelsOpen(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).TriggerMode(TriggerLongPress)
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Touch, PointerID: 7, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Touch, PointerID: 7, Position: f32.Pt(40, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownLongPress+3*time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if state == nil || state.open || state.touchTracking {
		t.Fatalf("moved long press state = %#v", state)
	}
}

func TestDropdownControlledAndProgrammaticOpenFocusMenu(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	requested := false
	base := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).OnOpenChange(func(open bool) { requested = open })
	widget := base.Open(false)
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if !requested || state == nil || !state.focusFirst {
		t.Fatalf("controlled open request = %v state %#v", requested, state)
	}

	widget = base.Open(true)
	layoutDropdownFrame(ctx, router, widget, start.Add(3*time.Millisecond))
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownEnterDuration+4*time.Millisecond))
	if state.focusFirst {
		t.Fatal("controlled Dropdown did not hand focus to Menu")
	}
}

func TestDropdownMenuActionClosesAndRestoresTriggerFocus(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	action := ""
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).OnAction(func(key string) { action = key })
	start := openDropdownForTest(ctx, router, widget)
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	clickFirstDropdownItem(ctx, router, widget, start)
	if action != "open" || state.open || !router.Source().Focused(&state.trigger) {
		t.Fatalf("dropdown action = %q open %v focused %v", action, state.open, router.Source().Focused(&state.trigger))
	}
	if frame.FocusVisible(ctx, &state.trigger, true) {
		t.Fatal("pointer-selected Menu item exposed a Tab-style Dropdown focus ring")
	}
}

func TestDropdownCloseOnSelectIsDelegatedToMenu(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "one", Label: "One"}}).CloseOnSelect(false)
	start := openDropdownForTest(ctx, router, widget)
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	clickFirstDropdownItem(ctx, router, widget, start)
	if !state.open {
		t.Fatal("Menu ignored Dropdown CloseOnSelect(false)")
	}
}

func TestDropdownCustomPopoverContentIsLaidOutByMenu(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	before := new(layoutProbe)
	after := new(layoutProbe)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).BeforeContent(before).AfterContent(after)
	openDropdownForTest(ctx, router, widget)
	if before.count == 0 || after.count == 0 {
		t.Fatalf("custom Menu content layouts = before %d after %d", before.count, after.count)
	}
}

func openDropdownForTest(ctx *frame.Context, router *input.Router, widget Widget) time.Time {
	start := time.Unix(2, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownEnterDuration+3*time.Millisecond))
	return start.Add(dropdownEnterDuration + 3*time.Millisecond)
}

func clickFirstDropdownItem(ctx *frame.Context, router *input.Router, widget Widget, start time.Time) {
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(30, 60)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(30, 60)})
	layoutDropdownFrame(ctx, router, widget, start.Add(2*time.Millisecond))
}

func layoutDropdownFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) {
	viewport := image.Pt(480, 360)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	triggerGtx := gtx
	triggerGtx.Constraints = layout.Constraints{Max: image.Pt(160, 60)}
	widget.Layout(ctx, triggerGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
