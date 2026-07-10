package checkbox

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
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

const stateSlotBool = "bool"

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

func TestCheckboxSyncsValue(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	Checkbox("done", true, "Done").Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 80)},
		Ops:         &ops,
	})

	state := testComponentState[widget.Bool](ctx, "done", stateSlotBool)
	if state == nil {
		t.Fatal("missing bool state")
	}
	if !state.Value {
		t.Fatal("checkbox value = false, want true")
	}
}

func TestCheckboxDisabled(t *testing.T) {
	c := Checkbox("done", true, "Done").Disabled(true)

	if !c.disabled {
		t.Fatal("checkbox was not disabled")
	}
}

func TestCheckboxInvalid(t *testing.T) {
	c := Checkbox("done", false, "Done").Invalid(true)

	if !c.invalid {
		t.Fatal("checkbox was not invalid")
	}
}

func TestCheckboxInvalidStyle(t *testing.T) {
	theme := DefaultTheme()
	style := checkboxStyleFor(&theme, false, false, true)
	danger := color.NRGBA{R: 0xf3, G: 0x12, B: 0x60, A: 0xff}

	if style.border != danger {
		t.Fatalf("invalid border = %v, want %v", style.border, danger)
	}
	if style.accent != danger {
		t.Fatalf("invalid accent = %v, want %v", style.accent, danger)
	}
	if style.focusColor.R != danger.R || style.focusColor.G != danger.G || style.focusColor.B != danger.B {
		t.Fatalf("invalid focus color = %v, want danger rgb", style.focusColor)
	}
}

func TestCheckboxInvalidHoverStyle(t *testing.T) {
	theme := DefaultTheme()
	style := checkboxStyleFor(&theme, true, false, true)
	dangerHover := color.NRGBA{R: 0xf5, G: 0x3a, B: 0x79, A: 0xff}

	if style.border != dangerHover {
		t.Fatalf("invalid hover border = %v, want %v", style.border, dangerHover)
	}
	if style.accent != dangerHover {
		t.Fatalf("invalid hover accent = %v, want %v", style.accent, dangerHover)
	}
}

func TestCheckboxDisabledInvalidStyle(t *testing.T) {
	theme := DefaultTheme()
	style := checkboxStyleFor(&theme, false, true, true)

	if style.border.A != 0x7f {
		t.Fatalf("disabled invalid border alpha = %d, want 127", style.border.A)
	}
	if style.focusColor.A != 0 {
		t.Fatalf("disabled invalid focus alpha = %d, want 0", style.focusColor.A)
	}
}

func TestCheckboxControlOnlyLayout(t *testing.T) {
	dims := Checkbox("done", false, "").Layout(newContext(nil), testLayoutContext())

	if dims.Size != image.Pt(20, 20) {
		t.Fatalf("checkbox size = %v, want (20,20)", dims.Size)
	}
}

func TestCheckboxRespectsConstraints(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(12, 12)}

	dims := Checkbox("done", false, "").Layout(newContext(nil), gtx)

	if dims.Size != image.Pt(12, 12) {
		t.Fatalf("checkbox size = %v, want (12,12)", dims.Size)
	}
}

func TestCheckboxSelectionAnimation(t *testing.T) {
	state := new(checkboxState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.selection(gtx, false); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.selection(gtx, true); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(checkboxSelectDuration / 2)
	mid := state.selection(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(checkboxSelectDuration)
	if got := state.selection(gtx, true); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func TestCheckboxFocusOpacityAnimation(t *testing.T) {
	state := new(checkboxState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	if got := state.focusOpacity(gtx, false); got != 0 {
		t.Fatalf("initial focus = %v, want 0", got)
	}
	if got := state.focusOpacity(gtx, true); got != 0 {
		t.Fatalf("focus start = %v, want 0", got)
	}

	gtx.Now = start.Add(checkboxFocusDuration / 2)
	mid := state.focusOpacity(gtx, true)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("focus midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(checkboxFocusDuration)
	if got := state.focusOpacity(gtx, true); got != 1 {
		t.Fatalf("focus end = %v, want 1", got)
	}
}

func TestCheckboxFocusVisibleIgnoresPointerFocus(t *testing.T) {
	state := new(checkboxState)
	if !state.focusVisible(true, nil) {
		t.Fatal("keyboard focus was not focus-visible")
	}

	state.focusVisible(false, nil)
	if state.focusVisible(true, []widget.Press{{Start: time.Unix(1, 0)}}) {
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

func TestCheckboxFocusesOnPress(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	layoutCheckboxFrame(ctx, router)
	checkbox := testComponentState[widget.Bool](ctx, "done", stateSlotBool)
	if checkbox == nil {
		t.Fatal("missing checkbox state")
	}

	router.Queue(pointer.Event{
		Kind:     pointer.Press,
		Source:   pointer.Mouse,
		Buttons:  pointer.ButtonPrimary,
		Position: f32.Pt(10, 10),
	})
	layoutCheckboxFrame(ctx, router)

	if !router.Source().Focused(checkbox) {
		t.Fatal("checkbox did not gain focus after press")
	}
}

func layoutCheckboxFrame(ctx *frame.Context, router *input.Router) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	Checkbox("done", false, "Done").Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
