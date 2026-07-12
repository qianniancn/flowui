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

type checkboxIndicatorProbe struct {
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *checkboxIndicatorProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(10, 10))}
}

func TestCheckboxOptionsUseValueSemantics(t *testing.T) {
	probe := &checkboxIndicatorProbe{}
	base := Checkbox("done", false, "Done")
	configured := base.
		Variant(CheckboxSecondary).
		Indeterminate(true).
		ReadOnly(true).
		Required(true).
		Description("Supporting text").
		ErrorMessage("Required").
		Disabled(true).
		Invalid(true).
		Indicator(func(IndicatorState) frame.Widget { return probe }).
		OnChange(func(bool) {})

	if base.variant != CheckboxPrimary || base.indeterminate || base.readOnly || base.required || base.description != "" {
		t.Fatal("configuring a Checkbox mutated the base value")
	}
	if configured.variant != CheckboxSecondary || !configured.indeterminate || !configured.readOnly || !configured.required {
		t.Fatal("variant or behavior options were not retained")
	}
	if configured.description != "Supporting text" || configured.errorMessage != "Required" || !configured.disabled || !configured.invalid {
		t.Fatal("field state options were not retained")
	}
	if configured.indicator == nil || configured.onChange == nil {
		t.Fatal("indicator or callback was not retained")
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
	style := checkboxStyleFor(&theme, CheckboxPrimary, false, false, false, true)
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

func TestCheckboxVariantsMatchHeroUI(t *testing.T) {
	activeTheme := DefaultTheme()
	primary := checkboxStyleFor(&activeTheme, CheckboxPrimary, false, false, false, false)
	secondary := checkboxStyleFor(&activeTheme, CheckboxSecondary, false, false, false, false)
	if primary.bg != activeTheme.Palette.Surface || primary.shadow != 1 {
		t.Fatalf("primary style = bg %#v shadow %v", primary.bg, primary.shadow)
	}
	if secondary.bg != activeTheme.Palette.SurfaceTertiary || secondary.shadow != 0 {
		t.Fatalf("secondary style = bg %#v shadow %v", secondary.bg, secondary.shadow)
	}
	pressed := checkboxStyleFor(&activeTheme, CheckboxPrimary, false, true, false, false)
	if pressed.accent != activeTheme.Palette.AccentHover {
		t.Fatalf("pressed accent = %#v", pressed.accent)
	}
}

func TestCheckboxThemeMatchesHeroUIGeometry(t *testing.T) {
	activeTheme := DefaultTheme()
	tokens := activeTheme.Components.Checkbox
	if tokens.Size != 16 || tokens.IndicatorSize != 12 || tokens.LabelGap != 8 || tokens.DescriptionIndent != 28 {
		t.Fatalf("Checkbox geometry = %+v", tokens)
	}
	if tokens.CheckStroke != 1.5 || tokens.IndeterminateStroke != 1.5 || tokens.FocusRingWidth != 2 {
		t.Fatalf("Checkbox stroke geometry = %+v", tokens)
	}
	if dark := theme.DarkTheme().Components.Checkbox.ShadowOpacity; dark != 0 {
		t.Fatalf("dark Checkbox shadow opacity = %v, want 0", dark)
	}
}

func TestCheckboxInvalidHoverStyle(t *testing.T) {
	theme := DefaultTheme()
	style := checkboxStyleFor(&theme, CheckboxPrimary, true, false, false, true)
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
	style := checkboxStyleFor(&theme, CheckboxPrimary, false, false, true, true)

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

func TestCheckboxDescriptionAndErrorLayout(t *testing.T) {
	ctx := newContext(nil)
	dims := Checkbox("updates", false, "Email updates").
		Description("Receive product notifications").
		Layout(ctx, testLayoutContext())
	if dims.Size.Y <= 20 {
		t.Fatalf("description height = %d, want more than control height", dims.Size.Y)
	}
	if got := frame.FieldDescription(ctx, "updates"); got != "Receive product notifications" {
		t.Fatalf("registered description = %q", got)
	}

	ctx = newContext(nil)
	Checkbox("terms", false, "Terms").
		Description("Read the terms").
		Invalid(true).
		ErrorMessage("Terms are required").
		Layout(ctx, testLayoutContext())
	if got := frame.FieldDescription(ctx, "terms"); got != "Terms are required" {
		t.Fatalf("registered error = %q", got)
	}
}

func TestCheckboxCustomIndicatorInheritsControlColors(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	probe := &checkboxIndicatorProbe{}
	Checkbox("custom", true, "Custom").
		Indicator(func(state IndicatorState) frame.Widget {
			if !state.Checked || state.Indeterminate {
				t.Fatalf("indicator state = %+v", state)
			}
			return probe
		}).
		Layout(ctx, testLayoutContext())
	if probe.layouts != 1 || probe.foreground != activeTheme.Palette.AccentForeground || probe.background != activeTheme.Palette.Accent {
		t.Fatalf("indicator = layouts %d colors %#v/%#v", probe.layouts, probe.foreground, probe.background)
	}
}

func TestCheckboxCustomIndicatorCanSuppressDefaultCheck(t *testing.T) {
	configured := Checkbox("custom", true, "Custom").Indicator(func(IndicatorState) frame.Widget { return nil })
	state := IndicatorState{Checked: true}
	if configured.indicator == nil || configured.indicatorWidget(state) != nil {
		t.Fatal("custom Indicator nil result was not preserved")
	}
	options := ControlOptions{Selection: 1, CustomIndicator: true, Indicator: configured.indicatorWidget(state)}
	if !options.CustomIndicator || options.Indicator != nil {
		t.Fatal("custom Indicator presence is ambiguous")
	}
}

func TestCheckboxIndeterminateAndReadOnlyInteraction(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	changed := false
	widget := Checkbox("all", false, "Select all").
		Indeterminate(true).
		OnChange(func(value bool) { changed = value })
	clickCheckbox(ctx, router, widget)
	if !changed {
		t.Fatal("indeterminate Checkbox did not request selected state")
	}

	ctx = newContext(nil)
	router = new(input.Router)
	called := false
	clickCheckbox(ctx, router, Checkbox("readonly", false, "Read only").ReadOnly(true).OnChange(func(bool) { called = true }))
	if called {
		t.Fatal("read-only Checkbox emitted a change")
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
	layoutCheckboxWidgetFrame(ctx, router, Checkbox("done", false, "Done"))
}

func layoutCheckboxWidgetFrame(ctx *frame.Context, router *input.Router, checkbox CheckboxWidget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	frame.BeginFrame(ctx)
	checkbox.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func clickCheckbox(ctx *frame.Context, router *input.Router, checkbox CheckboxWidget) {
	layoutCheckboxWidgetFrame(ctx, router, checkbox)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(10, 10)})
	layoutCheckboxWidgetFrame(ctx, router, checkbox)
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(10, 10)})
	layoutCheckboxWidgetFrame(ctx, router, checkbox)
}
