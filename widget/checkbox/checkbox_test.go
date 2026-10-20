package checkbox

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

const stateSlotBool = "bool"

func newContext(_ any) *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageAuto)
}

func testComponentState[T any](ctx *core.Context, key, slot string) *T {
	value, _ := core.PeekState[T](ctx, key, slot)
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

func (p *checkboxIndicatorProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(10, 10))}
}

func TestCheckboxSyncsValue(t *testing.T) {
	ctx := newContext(nil)
	var ops op.Ops

	New("Done").Key("done").Checked(true).Layout(ctx, layout.Context{
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
	c := New("Done").Key("done").Checked(true).Disabled(true)

	if !c.disabled {
		t.Fatal("checkbox was not disabled")
	}
}

func TestCheckboxInvalid(t *testing.T) {
	c := New("Done").Key("done").Checked(false).Invalid(true)

	if !c.invalid {
		t.Fatal("checkbox was not invalid")
	}
}

func TestCheckboxSemanticsIncludeState(t *testing.T) {
	var router input.Router
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	New("Completed").Key("done").Checked(true).Disabled(true).Description("This item is locked").Layout(newContext(nil), gtx)
	router.Frame(&ops)
	node, ok := checkboxSemanticNode(router.AppendSemantics(nil))
	if !ok {
		t.Fatal("checkbox semantic node missing")
	}
	if node.Desc.Class != semantic.CheckBox || node.Desc.Label != "Completed" {
		t.Fatalf("checkbox semantics = %#v", node.Desc)
	}
	if !node.Desc.Selected || !node.Desc.Disabled || node.Desc.Description != "This item is locked" {
		t.Fatalf("checkbox state semantics = %#v", node.Desc)
	}
}

func TestCheckboxControlOnlyLayout(t *testing.T) {
	dims := New("").Key("done").Checked(false).Layout(newContext(nil), testLayoutContext())

	if dims.Size != image.Pt(20, 20) {
		t.Fatalf("checkbox size = %v, want (20,20)", dims.Size)
	}
}

func TestCheckboxIndicatorPartControlsGeometry(t *testing.T) {
	activeTheme := testtheme.Behavior()
	ctx := core.New(nil, activeTheme, locale.LanguageEnglish)
	gtx := testLayoutContext()
	styled := (New("").Key("compact").Checked(false)).Style(style.Style{}.Part(style.PartIndicator, style.Style{}.Width(12).Height(10)))
	dims := styled.Layout(ctx, gtx)
	visual := testtheme.BehaviorProvider().Checkbox(activeTheme, theme.CheckboxRequest{Variant: Primary})
	focusSpace := max(gtx.Dp(visual.Metrics.FocusSpace), 1)
	want := image.Pt(gtx.Dp(12)+focusSpace*2, gtx.Dp(10)+focusSpace*2)
	if dims.Size != want {
		t.Fatalf("styled checkbox size = %v, want %v", dims.Size, want)
	}
}

func TestCheckboxDescriptionAndErrorLayout(t *testing.T) {
	ctx := newContext(nil)
	dims := New("Email updates").Key("updates").Checked(false).
		Description("Receive product notifications").
		Layout(ctx, testLayoutContext())
	if dims.Size.Y <= 20 {
		t.Fatalf("description height = %d, want more than control height", dims.Size.Y)
	}
	if got := core.FieldDescription(ctx, "updates"); got != "Receive product notifications" {
		t.Fatalf("registered description = %q", got)
	}

	ctx = newContext(nil)
	New("Terms").Key("terms").Checked(false).
		Description("Read the terms").
		Invalid(true).
		ErrorMessage("Terms are required").
		Layout(ctx, testLayoutContext())
	if got := core.FieldDescription(ctx, "terms"); got != "Terms are required" {
		t.Fatalf("registered error = %q", got)
	}
}

func TestCheckboxCustomIndicatorInheritsControlColors(t *testing.T) {
	activeTheme := testtheme.Behavior()
	ctx := core.New(nil, activeTheme, locale.LanguageEnglish)
	probe := &checkboxIndicatorProbe{}
	New("Custom").Key("custom").Checked(true).
		Indicator(func(state IndicatorState) core.Widget {
			if !state.Checked || state.Indeterminate {
				t.Fatalf("indicator state = %+v", state)
			}
			return probe
		}).
		Layout(ctx, testLayoutContext())
	visual := testtheme.BehaviorProvider().Checkbox(activeTheme, theme.CheckboxRequest{Variant: Primary})
	on := style.CascadePart(style.State{Checked: true, Selected: true}, style.PartIndicator, visual.Defaults, visual.Override)
	wantForeground, foregroundOK := computed.Color(on.Text.Color)
	wantBackground, backgroundOK := computed.Brush(on.Paint.Background)
	if probe.layouts != 1 || !foregroundOK || !backgroundOK || probe.foreground != wantForeground || probe.background != wantBackground.ColorAt(.5) {
		t.Fatalf("indicator = layouts %d colors %#v/%#v", probe.layouts, probe.foreground, probe.background)
	}
}

func TestCheckboxCustomIndicatorCanSuppressDefaultCheck(t *testing.T) {
	configured := New("Custom").Key("custom").Checked(true).Indicator(func(IndicatorState) core.Widget { return nil })
	state := IndicatorState{Checked: true}
	if configured.indicator == nil || configured.indicatorWidget(state) != nil {
		t.Fatal("custom Indicator nil result was not preserved")
	}
	options := ControlOptions{Selection: 1, CustomIndicator: true, Indicator: configured.indicatorWidget(state)}
	if !options.CustomIndicator || options.Indicator != nil {
		t.Fatal("custom Indicator presence is ambiguous")
	}
}

func TestCheckboxCustomIndicatorMayReturnNilDuringLayout(t *testing.T) {
	New("Custom").Key("custom").Checked(false).
		Indicator(func(IndicatorState) core.Widget { return nil }).
		Layout(newContext(nil), testLayoutContext())
}

func TestCheckboxPartsResolveStateAndOverrides(t *testing.T) {
	ctx := newContext(nil)
	off := color.NRGBA{R: 0x10, G: 0x20, B: 0x30, A: 0xff}
	on := color.NRGBA{R: 0x40, G: 0x50, B: 0x60, A: 0xff}
	indicatorText := color.NRGBA{R: 0x70, G: 0x80, B: 0x90, A: 0xff}
	label := color.NRGBA{R: 0xa0, G: 0xb0, B: 0xc0, A: 0xff}
	description := color.NRGBA{R: 0xd0, G: 0xe0, B: 0xf0, A: 0xff}
	custom := style.Style{}.
		Part(style.PartIndicator, style.Style{}.
			Background(style.SolidColor{Color: off}).
			BorderWidth(3).
			Radius(9).
			BoxShadow(1, 2, 3, 4, style.RGBA(0x01020380)).
			Translate(2, 3).
			TextColor(style.SolidColor{Color: indicatorText}).
			Opacity(.5).
			On(style.Checked, style.Style{}.Background(style.LinearGradient(
				style.ColorStop(0, style.SolidColor{Color: off}),
				style.ColorStop(1, style.SolidColor{Color: on}),
			)))).
		Part(style.PartLabel, style.Style{}.TextColor(style.SolidColor{Color: label}).FontSize(22)).
		Part(style.PartDescription, style.Style{}.TextColor(style.SolidColor{Color: description}))

	resolved := New("Parts").Key("parts").Checked(true).resolveStyle(ctx, layout.Context{Ops: new(op.Ops)}, "parts", style.State{Checked: true, Selected: true}, &theme.CheckboxOverride{Declaration: custom})
	indicator := resolved.indicator.on
	if indicator.Paint == nil || indicator.Paint.Opacity == nil || *indicator.Paint.Opacity != .5 || indicator.Paint.Border == nil || *indicator.Paint.Border.Width != 3 || *indicator.Paint.Radius != 9 || len(indicator.Paint.Shadows) != 1 {
		t.Fatalf("indicator paint = %#v", indicator.Paint)
	}
	if _, ok := indicator.Paint.Background.(style.Gradient); !ok {
		t.Fatalf("indicator background = %T, want gradient", indicator.Paint.Background)
	}
	if indicator.Trans == nil || indicator.Trans.TranslateX == nil || *indicator.Trans.TranslateX != 2 || indicator.Trans.TranslateY == nil || *indicator.Trans.TranslateY != 3 {
		t.Fatalf("indicator transform = %#v", indicator.Trans)
	}
	if got, ok := computed.Color(indicator.Text.Color); !ok || got != indicatorText {
		t.Fatalf("indicator text = %#v, ok %v", got, ok)
	}
	if resolved.label.Text == nil || resolved.label.Text.FontSize == nil || *resolved.label.Text.FontSize != 22 {
		t.Fatalf("label style = %#v", resolved.label.Text)
	}
	if got, ok := computed.Color(resolved.label.Text.Color); !ok || got != label {
		t.Fatalf("label color = %#v, ok %v", got, ok)
	}
	if got, ok := computed.Color(resolved.description.Text.Color); !ok || got != description {
		t.Fatalf("description color = %#v, ok %v", got, ok)
	}
}

func TestCheckboxIndeterminateAndReadOnlyInteraction(t *testing.T) {
	ctx := newContext(nil)
	router := new(input.Router)
	changed := false
	widget := New("Select all").Key("all").Checked(false).
		Indeterminate(true).
		OnChange(func(value bool) { changed = value })
	clickCheckbox(ctx, router, widget)
	if !changed {
		t.Fatal("indeterminate Checkbox did not request selected state")
	}

	ctx = newContext(nil)
	router = new(input.Router)
	called := false
	clickCheckbox(ctx, router, New("Read only").Key("readonly").Checked(false).ReadOnly(true).OnChange(func(bool) { called = true }))
	if called {
		t.Fatal("read-only Checkbox emitted a change")
	}
}

func TestCheckboxRespectsConstraints(t *testing.T) {
	gtx := testLayoutContext()
	gtx.Constraints = layout.Constraints{Max: image.Pt(12, 12)}

	dims := New("").Key("done").Checked(false).Layout(newContext(nil), gtx)

	if dims.Size != image.Pt(12, 12) {
		t.Fatalf("checkbox size = %v, want (12,12)", dims.Size)
	}
}

func TestCheckboxSelectionAnimation(t *testing.T) {
	state := new(checkboxState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	duration := 100 * time.Millisecond
	if got := state.selection(gtx, false, duration); got != 0 {
		t.Fatalf("initial selection = %v, want 0", got)
	}
	if got := state.selection(gtx, true, duration); got != 0 {
		t.Fatalf("selection start = %v, want 0", got)
	}

	gtx.Now = start.Add(duration / 2)
	mid := state.selection(gtx, true, duration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("selection midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(duration)
	if got := state.selection(gtx, true, duration); got != 1 {
		t.Fatalf("selection end = %v, want 1", got)
	}
}

func TestCheckboxFocusOpacityAnimation(t *testing.T) {
	state := new(checkboxState)
	start := time.Unix(1, 0)
	gtx := testLayoutContext()
	gtx.Now = start

	duration := 100 * time.Millisecond
	if got := state.focusOpacity(gtx, false, duration); got != 0 {
		t.Fatalf("initial focus = %v, want 0", got)
	}
	if got := state.focusOpacity(gtx, true, duration); got != 0 {
		t.Fatalf("focus start = %v, want 0", got)
	}

	gtx.Now = start.Add(duration / 2)
	mid := state.focusOpacity(gtx, true, duration)
	if mid <= 0 || mid >= 1 {
		t.Fatalf("focus midpoint = %v, want between 0 and 1", mid)
	}

	gtx.Now = start.Add(duration)
	if got := state.focusOpacity(gtx, true, duration); got != 1 {
		t.Fatalf("focus end = %v, want 1", got)
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

func layoutCheckboxFrame(ctx *core.Context, router *input.Router) {
	layoutCheckboxWidgetFrame(ctx, router, New("Done").Key("done").Checked(false))
}

func layoutCheckboxWidgetFrame(ctx *core.Context, router *input.Router, checkbox Widget) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	core.BeginFrame(ctx)
	checkbox.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}

func clickCheckbox(ctx *core.Context, router *input.Router, checkbox Widget) {
	layoutCheckboxWidgetFrame(ctx, router, checkbox)
	router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, Buttons: pointer.ButtonPrimary, Position: f32.Pt(10, 10)})
	layoutCheckboxWidgetFrame(ctx, router, checkbox)
	router.Queue(pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, Position: f32.Pt(10, 10)})
	layoutCheckboxWidgetFrame(ctx, router, checkbox)
}

func checkboxSemanticNode(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.CheckBox {
			return node, true
		}
		if child, ok := checkboxSemanticNode(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}
