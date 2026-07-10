package switches

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func newContext(_ any) *frame.Context {
	return frame.New(nil, nil, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
}

func testComponentState[T any](ctx *frame.Context, key, slot string) *T {
	value, _ := frame.PeekState[T](ctx, key, slot)
	return value
}

func testLayoutContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

type enabledProbeWidget struct {
	enabled bool
}

func (w *enabledProbeWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	w.enabled = gtx.Enabled()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func TestSwitchOptions(t *testing.T) {
	changed := true
	sw := Switch("notifications", true, "Notifications").
		Description("Receive product updates").
		OnChange(func(checked bool) {
			changed = checked
		}).
		Disabled(true).
		Invalid(true).
		Size(SwitchLarge).
		LabelBefore().
		Thumb(func(bool) frame.Widget {
			return text.New("on")
		})

	if sw.key != "notifications" || !sw.checked || sw.label != "Notifications" || sw.description != "Receive product updates" {
		t.Fatal("switch constructor/options did not set text fields")
	}
	if !sw.disabled || !sw.invalid || sw.size != SwitchLarge || !sw.labelBefore || sw.thumb == nil {
		t.Fatal("switch visual options were not set")
	}
	sw.onChange(false)
	if changed {
		t.Fatal("switch onChange did not receive false")
	}
}

func TestSwitchSyncsValue(t *testing.T) {
	ctx := newContext(nil)

	Switch("notifications", true, "Notifications").Layout(ctx, testLayoutContext())

	state := testComponentState[switchState](ctx, "notifications", stateSlotSwitch)
	if state == nil {
		t.Fatal("missing switch state")
	}
	if !state.value.Value {
		t.Fatal("switch value = false, want true")
	}
}

func TestSwitchPointerToggleNotifies(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	checked := false
	var notified bool
	layout := func() {
		layoutSwitchFrameWith(ctx, router, Switch("notifications", checked, "Notifications").
			OnChange(func(next bool) {
				checked = next
				notified = next
			}))
	}

	layout()
	router.Queue(pointer.Event{
		Kind:      pointer.Press,
		Source:    pointer.Mouse,
		PointerID: 1,
		Buttons:   pointer.ButtonPrimary,
		Position:  f32.Pt(10, 10),
	})
	layout()
	router.Queue(pointer.Event{
		Kind:      pointer.Release,
		Source:    pointer.Mouse,
		PointerID: 1,
		Position:  f32.Pt(10, 10),
	})
	layout()

	if !checked || !notified {
		t.Fatalf("switch did not notify pointer toggle, checked=%v notified=%v", checked, notified)
	}
}

func TestSwitchControlOnlyLayout(t *testing.T) {
	dims := Switch("notifications", false, "").Layout(newContext(nil), testLayoutContext())

	if dims.Size != image.Pt(44, 24) {
		t.Fatalf("switch size = %v, want (44,24)", dims.Size)
	}
}

func TestSwitchRespectsConstraints(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(12, 12)}

	dims := Switch("notifications", false, "").Layout(newContext(nil), gtx)

	if dims.Size != image.Pt(12, 12) {
		t.Fatalf("switch size = %v, want (12,12)", dims.Size)
	}
}

func TestSwitchThumbReceivesCheckedState(t *testing.T) {
	got := false
	Switch("power", true, "").
		Thumb(func(checked bool) frame.Widget {
			got = checked
			return text.New("1")
		}).
		Layout(newContext(nil), testLayoutContext())

	if !got {
		t.Fatal("thumb content did not receive checked=true")
	}
}

func TestSwitchPassesDisabledContextToThumb(t *testing.T) {
	probe := &enabledProbeWidget{}

	Switch("power", false, "").
		Disabled(true).
		Thumb(func(bool) frame.Widget {
			return probe
		}).
		Layout(newContext(nil), testLayoutContext())

	if probe.enabled {
		t.Fatal("switch thumb content was laid out with enabled context")
	}
}

func TestSwitchStyleInvalid(t *testing.T) {
	theme := DefaultTheme()
	style := switchStyleFor(&theme, true, false, false, true)

	if style.trackOn != theme.Palette.DangerHover {
		t.Fatalf("invalid hover track = %#v, want %#v", style.trackOn, theme.Palette.DangerHover)
	}
	if style.thumbFg != theme.Palette.Danger {
		t.Fatalf("invalid thumb fg = %#v, want %#v", style.thumbFg, theme.Palette.Danger)
	}
}

func TestSwitchDisabledStyleUsesDisabledOpacity(t *testing.T) {
	theme := DefaultTheme()
	theme.DisabledOpacity = 0.25

	style := switchStyleFor(&theme, false, false, true, false)

	if style.trackOn.A != byte(float32(theme.Palette.Accent.A)*0.25) {
		t.Fatalf("disabled track alpha = %d, want %d", style.trackOn.A, byte(float32(theme.Palette.Accent.A)*0.25))
	}
	if style.focusColor.A != 0 {
		t.Fatalf("disabled focus alpha = %d, want 0", style.focusColor.A)
	}
}

func TestSwitchSizeStyle(t *testing.T) {
	theme := DefaultTheme()

	if got := switchSizeStyleFor(&theme, SwitchSmall).trackWidth; got != theme.Components.Switch.SmallTrackWidth {
		t.Fatalf("small track width = %v, want %v", got, theme.Components.Switch.SmallTrackWidth)
	}
	if got := switchSizeStyleFor(&theme, SwitchLarge).thumbHeight; got != theme.Components.Switch.LargeThumbHeight {
		t.Fatalf("large thumb height = %v, want %v", got, theme.Components.Switch.LargeThumbHeight)
	}
}

func TestSwitchSelectionAnimation(t *testing.T) {
	state := new(switchState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.selection(gtx, false); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.selection(gtx, true); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(switchSelectDuration / 2)
	mid := state.selection(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(switchSelectDuration)
	if got := state.selection(gtx, true); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func TestSwitchFocusOpacityAnimation(t *testing.T) {
	state := new(switchState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.focusOpacity(gtx, false); got != 0 {
		t.Fatalf("initial focus = %v, want 0", got)
	}
	if got := state.focusOpacity(gtx, true); got != 0 {
		t.Fatalf("focus start = %v, want 0", got)
	}

	gtx.Now = start.Add(switchFocusDuration / 2)
	mid := state.focusOpacity(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("focus midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(switchFocusDuration)
	if got := state.focusOpacity(gtx, true); got != 1 {
		t.Fatalf("focus end = %v, want 1", got)
	}
}

func TestSwitchFocusVisibleIgnoresPointerFocus(t *testing.T) {
	state := new(switchState)
	if !state.focusVisible(true, nil) {
		t.Fatal("keyboard focus was not focus-visible")
	}

	state.focusVisible(false, nil)
	if state.focusVisible(true, []widget.Press{widgetPressForSwitchTest()}) {
		t.Fatal("pointer focus was focus-visible")
	}

	if state.focusVisible(true, nil) {
		t.Fatal("pointer focus became focus-visible without losing focus")
	}

	state.focusVisible(false, nil)
	if !state.focusVisible(true, nil) {
		t.Fatal("focus-visible did not reset after blur")
	}
}

func TestSwitchFocusesOnPress(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutSwitchFrame(ctx, router)
	state := testComponentState[switchState](ctx, "notifications", stateSlotSwitch)
	if state == nil {
		t.Fatal("missing switch state")
	}

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(10, 10),
	})
	layoutSwitchFrame(ctx, router)

	if !router.Source().Focused(&state.value) {
		t.Fatal("switch did not gain focus after press")
	}
}

func TestSwitchGroupHorizontalWraps(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Constraints.Max = image.Pt(100, 200)

	dims := SwitchGroup(
		fixedSwitchChild{size: image.Pt(70, 10)},
		fixedSwitchChild{size: image.Pt(70, 10)},
	).Horizontal().Layout(newContext(nil), gtx)

	if dims.Size != image.Pt(70, 36) {
		t.Fatalf("switch group size = %v, want (70,36)", dims.Size)
	}
}

func TestSwitchThumbContentColorUsesThemePalette(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.MutedForeground = color.NRGBA{R: 1, G: 2, B: 3, A: 255}
	theme.Palette.Accent = color.NRGBA{R: 4, G: 5, B: 6, A: 255}
	style := switchStyleFor(&theme, false, false, false, false)
	style.selected = 0

	if got := switchThumbContentColor(style); got != theme.Palette.MutedForeground {
		t.Fatalf("off thumb content color = %#v, want %#v", got, theme.Palette.MutedForeground)
	}
}

func TestSwitchThumbContentColorUsesDisabledStyle(t *testing.T) {
	theme := DefaultTheme()
	theme.DisabledOpacity = 0.25
	style := switchStyleFor(&theme, false, false, true, false)
	style.selected = 0

	if got := switchThumbContentColor(style); got.A != byte(float32(theme.Palette.MutedForeground.A)*0.25) {
		t.Fatalf("disabled off thumb content alpha = %d, want %d", got.A, byte(float32(theme.Palette.MutedForeground.A)*0.25))
	}
}

type fixedSwitchChild struct {
	size image.Point
}

func (f fixedSwitchChild) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(f.size)}
}

func widgetPressForSwitchTest() widget.Press {
	return widget.Press{Start: time.Unix(1, 0)}
}

func layoutSwitchFrame(ctx *frame.Context, router *input.Router) {
	layoutSwitchFrameWith(ctx, router, Switch("notifications", false, "Notifications"))
}

func layoutSwitchFrameWith(ctx *frame.Context, router *input.Router, sw SwitchWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	sw.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
