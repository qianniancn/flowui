package dropdown

import (
	"image"
	"reflect"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/components/button"
	menupkg "github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/overlay"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
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
		OnOpenChangeEvent(func(OpenChangeEvent) {}).
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

func TestDropdownExtendedOptionsUseValueSemantics(t *testing.T) {
	base := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}})
	var event OpenChangeEvent
	configured := base.
		AutoWidth().
		MinWidth(160).
		MaxWidth(320).
		MatchTriggerWidth(true).
		Arrow(true).
		HoverOpenDelay(20 * time.Millisecond).
		HoverCloseDelay(30 * time.Millisecond).
		LongPressDelay(40 * time.Millisecond).
		TriggerMode(TriggerContextMenu).
		OnOpenChangeEvent(func(value OpenChangeEvent) { event = value })
	if !configured.matchTriggerWidth || !configured.arrow || configured.triggerMode != TriggerContextMenu || reflect.DeepEqual(configured.menu, base.menu) {
		t.Fatalf("extended dropdown options = %#v", configured)
	}
	if configured.hoverOpenDelay != 20*time.Millisecond || configured.hoverCloseDelay != 30*time.Millisecond || configured.longPressDelay != 40*time.Millisecond || !configured.hasHoverOpenDelay || !configured.hasHoverCloseDelay || !configured.hasLongPressDelay || configured.onOpenChangeEvent == nil || event.Open {
		t.Fatalf("extended timing/callback options = %#v", configured)
	}
	if base.matchTriggerWidth || base.arrow || base.triggerMode != TriggerPress {
		t.Fatalf("base dropdown was mutated: %#v", base)
	}
}

func TestDropdownForwardsMenuConfiguration(t *testing.T) {
	checked := func(string, bool) {}
	radio := func(string, string) {}
	configured := dropdownTestWidget(nil).
		AutoSeparateSections(false).
		DataVersion(3).
		OnCheckedChange(checked).
		OnRadioChange(radio).
		Compact(true)
	expected := menupkg.Menu("actions:menu", nil).
		AutoSeparateSections(false).
		DataVersion(3).
		Compact(true)
	actual := configured.menu.OnCheckedChange(nil).OnRadioChange(nil)
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("dropdown menu configuration did not forward all options: %#v", configured.menu)
	}
	if reflect.ValueOf(checked).Pointer() == 0 || reflect.ValueOf(radio).Pointer() == 0 {
		t.Fatal("test callbacks were not initialized")
	}
}

func TestDropdownContextMenuTriggerUsesPointerAnchor(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	var change OpenChangeEvent
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).
		TriggerMode(TriggerContextMenu).
		OnOpenChangeEvent(func(value OpenChangeEvent) { change = value })
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonSecondary, Position: f32.Pt(32, 18)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if state == nil || !state.open || !state.hasContextAnchor || state.contextAnchor.Min != image.Pt(32, 18) {
		t.Fatalf("context trigger state = %#v", state)
	}
	if change.Source != OpenChangeContextMenu || !change.Open {
		t.Fatalf("context open event = %#v", change)
	}
}

func TestDropdownContextAnchorIsNotReusedForProgrammaticOpen(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).TriggerMode(TriggerContextMenu)
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonSecondary, Position: f32.Pt(32, 18)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownEnterDuration+2*time.Millisecond))
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 2, Buttons: pointer.ButtonPrimary, Position: f32.Pt(420, 300)})
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownEnterDuration+3*time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if state == nil || state.open {
		t.Fatalf("context outside close state = %#v", state)
	}
	programmatic := widget.Open(true)
	layoutDropdownFrame(ctx, router, programmatic, start.Add(3*time.Millisecond))
	if state.hasContextAnchor {
		t.Fatal("programmatic open reused the previous context anchor")
	}
}

func TestDropdownButtonUsesIndependentDropdownState(t *testing.T) {
	action := button.Button("create", fixedWidget{size: image.Pt(80, 32)})
	base := Button("split", action, []Item{{Key: "copy", Label: "Copy"}})
	configured := base.
		Variant(button.ButtonDanger).
		Size(button.ButtonSmall).
		AutoWidth().
		Placement(overlay.PopoverBottomEnd).
		OnActionEvent(func(ActionEvent) {}).
		OnOpenChangeEvent(func(OpenChangeEvent) {})
	if configured.key != "split" || configured.dropdown.key != "split:dropdown" || configured.dropdown.placement != overlay.PopoverBottomEnd || reflect.DeepEqual(base.dropdown, configured.dropdown) {
		t.Fatalf("dropdown button configuration = %#v", configured)
	}
	if base.dropdown.trigger == nil || configured.dropdown.trigger == nil {
		t.Fatal("dropdown button lost its trigger")
	}
}

func TestDropdownButtonDisablesPopupWhenActionIsDisabled(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := Button("split", button.Button("action", fixedWidget{size: image.Pt(80, 32)}).Disabled(true), []Item{{Key: "copy", Label: "Copy"}}).DefaultOpen(true)
	layoutDropdownFrame(ctx, router, widget, time.Unix(1, 0))
	state, _ := frame.PeekState[dropdownState](ctx, "split:dropdown", stateSlotDropdown)
	if state == nil || state.open {
		t.Fatalf("disabled dropdown button state = %#v", state)
	}
}

func TestDropdownMenuStyleTargetsPopupOnly(t *testing.T) {
	style := flowstyle.RGBA(0x12345680)
	base := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}})
	configured := base.MenuStyle(flowstyle.Style{}.Background(style))
	expectedMenu := menupkg.Menu("actions:menu", []Item{{Key: "open", Label: "Open"}}).Style(flowstyle.Style{}.Background(style))
	if reflect.DeepEqual(base.menu, expectedMenu) {
		t.Fatal("base dropdown menu unexpectedly has the configured style")
	}
	if !reflect.DeepEqual(configured.menu, expectedMenu) {
		t.Fatal("MenuStyle did not configure the dropdown popup menu")
	}
	if configured.customStyle.Resolve(flowstyle.StyleState{}).Paint != nil {
		t.Fatal("MenuStyle changed dropdown trigger style")
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

func TestDropdownHoverTriggerOpensAndClosesAfterPointerLeaves(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).TriggerMode(TriggerHover)
	start := time.Unix(1, 0)
	layoutDropdownFrame(ctx, router, widget, start)
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(20, 20)})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	if state == nil || state.open {
		t.Fatalf("hover opened before delay: %#v", state)
	}
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownHoverOpen+time.Millisecond))
	if !state.open {
		t.Fatal("hover did not open dropdown")
	}
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(420, 300)})
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownHoverOpen+time.Millisecond+time.Millisecond))
	if !state.open {
		t.Fatal("hover closed before leave delay")
	}
	layoutDropdownFrame(ctx, router, widget, start.Add(dropdownHoverOpen+dropdownHoverClose+2*time.Millisecond))
	if state.open {
		t.Fatal("hover did not close after pointer left")
	}
}

func TestDropdownControlledAndProgrammaticOpenFocusMenu(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	requested := false
	base := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).OnOpenChangeEvent(func(event OpenChangeEvent) { requested = event.Open })
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
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).OnActionEvent(func(event ActionEvent) { action = event.Key })
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

func TestDropdownOutsidePressClosesImmediately(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}})
	start := openDropdownForTest(ctx, router, widget)
	state, _ := frame.PeekState[dropdownState](ctx, "actions", stateSlotDropdown)
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 9,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(420, 300),
	})
	layoutDropdownFrame(ctx, router, widget, start.Add(time.Millisecond))
	if state.open {
		t.Fatal("outside press did not close Dropdown")
	}
	if frame.HasTopOverlay(ctx) {
		t.Fatal("outside press left Dropdown overlay visible")
	}
	if got := state.transition.Current(); got != 0 {
		t.Fatalf("outside press left Dropdown exit progress at %v", got)
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

func TestDropdownUsesMenuRootBackground(t *testing.T) {
	ctx := dropdownTestContext()
	router := new(input.Router)
	got := frame.ActiveTheme(ctx).Palette.Background
	widget := dropdownTestWidget([]Item{{Key: "open", Label: "Open"}}).
		BeforeContent(frame.WidgetFunc(func(ctx *frame.Context, _ layout.Context) layout.Dimensions {
			got = ctx.BackgroundColor()
			return layout.Dimensions{}
		}))

	openDropdownForTest(ctx, router, widget)
	activeTheme := frame.ActiveTheme(ctx)
	if want := theme.ColorOr(activeTheme.Components.Menu.BackgroundColor, activeTheme.Palette.OverlayColor()); got != want {
		t.Fatalf("dropdown background = %#v, want %#v", got, want)
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

func layoutDropdownFrame(ctx *frame.Context, router *input.Router, widget frame.Widget, now time.Time) {
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
