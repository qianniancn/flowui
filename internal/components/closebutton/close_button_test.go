package closebutton

import (
	"image"
	"image/color"
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
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestCloseButtonOptionsUseValueSemantics(t *testing.T) {
	clicked := false
	icon := new(closeButtonProbe)
	base := CloseButton("dismiss")
	configured := base.
		OnClick(func() { clicked = true }).
		Disabled(true).
		Icon(icon).
		Label("Dismiss")

	if base.onClick != nil || base.disabled || base.icon != nil || base.label != "" {
		t.Fatal("close button options mutated the original value")
	}
	if configured.key != "dismiss" || configured.onClick == nil || !configured.disabled || configured.icon != icon || configured.label != "Dismiss" {
		t.Fatalf("configured close button = %#v", configured)
	}
	configured.onClick()
	if !clicked {
		t.Fatal("click handler was not retained")
	}
}

func TestCloseButtonHeroUIDefaults(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	component := activeTheme.Components.CloseButton
	if component.Size != 24 || component.Radius != 12 || component.Padding != 4 || component.IconSize != 16 {
		t.Fatalf("close button geometry = %#v", component)
	}
	if component.FocusRingWidth != 2 || component.PressedScale != 0.93 {
		t.Fatalf("close button interaction tokens = %#v", component)
	}
}

func TestCloseButtonLayoutUsesThemeGeometry(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.CloseButton.Size = 30
	activeTheme.Components.CloseButton.Padding = 5
	activeTheme.Components.CloseButton.IconSize = 22
	probe := new(closeButtonProbe)
	dims := CloseButton("close").Icon(probe).Layout(newCloseButtonContext(&activeTheme, locale.LanguageEnglish), closeButtonTestContext())

	if dims.Size != image.Pt(30, 30) {
		t.Fatalf("close button size = %v, want (30,30)", dims.Size)
	}
	if probe.constraints != layout.Exact(image.Pt(20, 20)) {
		t.Fatalf("custom icon constraints = %#v, want exact 20x20", probe.constraints)
	}
	if probe.foreground != activeTheme.Palette.MutedForeground {
		t.Fatalf("custom icon foreground = %#v, want %#v", probe.foreground, activeTheme.Palette.MutedForeground)
	}
}

func TestCloseButtonUsesHeroUIPaletteStates(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	idle := closeButtonStyleFor(&activeTheme, false, false)
	hovered := closeButtonStyleFor(&activeTheme, true, false)
	disabled := closeButtonStyleFor(&activeTheme, true, true)

	if idle.background != activeTheme.Palette.SurfaceRaised || idle.foreground != activeTheme.Palette.MutedForeground {
		t.Fatalf("idle style = %#v", idle)
	}
	if hovered.background != activeTheme.Palette.SurfacePressed {
		t.Fatalf("hover background = %#v, want %#v", hovered.background, activeTheme.Palette.SurfacePressed)
	}
	if disabled.background != activeTheme.DisabledColor(activeTheme.Palette.SurfaceRaised) {
		t.Fatalf("disabled background = %#v", disabled.background)
	}
	if disabled.foreground != activeTheme.DisabledColor(activeTheme.Palette.MutedForeground) {
		t.Fatalf("disabled foreground = %#v", disabled.foreground)
	}
}

func TestCloseButtonHandlesClickAndDisabledBlocksIt(t *testing.T) {
	for _, test := range []struct {
		name     string
		disabled bool
		want     bool
	}{
		{name: "enabled", want: true},
		{name: "disabled", disabled: true, want: false},
	} {
		t.Run(test.name, func(t *testing.T) {
			clicked := false
			ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
			clickable := new(widget.Clickable)
			clickable.Click()
			frame.UseStateWith(ctx, "close", "clickable", func() *widget.Clickable { return clickable })
			CloseButton("close").Disabled(test.disabled).OnClick(func() { clicked = true }).Layout(ctx, closeButtonTestContext())
			if clicked != test.want {
				t.Fatalf("clicked = %v, want %v", clicked, test.want)
			}
		})
	}
}

func TestCloseButtonPointerClickAndDisabledBlocking(t *testing.T) {
	for _, disabled := range []bool{false, true} {
		ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
		router := new(input.Router)
		clicked := false
		button := CloseButton("close").Disabled(disabled).OnClick(func() { clicked = true })
		layoutCloseButtonFrame(ctx, router, button, time.Unix(1, 0))
		router.Queue(
			pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(12, 12)},
			pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(12, 12)},
		)
		layoutCloseButtonFrame(ctx, router, button, time.Unix(1, int64(time.Millisecond)))
		if clicked == disabled {
			t.Fatalf("disabled=%v clicked=%v", disabled, clicked)
		}
	}
}

func TestDisabledCloseButtonDoesNotBlockUnderlyingPointerTarget(t *testing.T) {
	ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
	router := new(input.Router)
	background := new(widget.Clickable)
	clicked := false
	button := CloseButton("close").Disabled(true)
	start := time.Unix(1, 0)
	layoutDisabledCloseOverBackground(ctx, router, background, button, &clicked, start)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(12, 12)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(12, 12)},
	)
	layoutDisabledCloseOverBackground(ctx, router, background, button, &clicked, start.Add(time.Millisecond))
	if !clicked {
		t.Fatal("disabled close button blocked the underlying pointer target")
	}
}

func TestCloseIconTransformPreservesHeroUIViewBox(t *testing.T) {
	transform := newCloseIconTransform(image.Pt(20, 16))
	if got := transform.point(0, 0); got != f32.Pt(2, 0) {
		t.Fatalf("viewbox origin = %v, want (2,0)", got)
	}
	if got := transform.point(16, 16); got != f32.Pt(18, 16) {
		t.Fatalf("viewbox extent = %v, want (18,16)", got)
	}
}

func TestCloseButtonFocusGeometryFitsExpandedBounds(t *testing.T) {
	rect, radius := closeButtonFocusGeometry(image.Rect(0, 0, 24, 24), 12, 2)
	if rect != image.Rect(-1, -1, 25, 25) {
		t.Fatalf("focus rect = %v, want (-1,-1)-(25,25)", rect)
	}
	if radius != 13 {
		t.Fatalf("focus radius = %d, want 13", radius)
	}
}

func TestCloseButtonPointerAndKeyboardFocusVisibility(t *testing.T) {
	start := time.Unix(1, 0)
	{
		ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
		router := new(input.Router)
		button := CloseButton("close")
		layoutCloseButtonFrame(ctx, router, button, start)
		clickable, ok := frame.PeekState[widget.Clickable](ctx, "close", "clickable")
		if !ok {
			t.Fatal("close button clickable state is missing")
		}
		router.Queue(pointer.Event{
			Kind:      pointer.Press,
			Source:    pointer.Mouse,
			PointerID: 1,
			Buttons:   pointer.ButtonPrimary,
			Position:  f32.Pt(12, 12),
		})
		layoutCloseButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		buttonState, ok := frame.PeekState[closeButtonState](ctx, "close", stateSlotCloseButton)
		if !ok || !router.Source().Focused(clickable) {
			t.Fatal("pointer press did not focus the close button")
		}
		if buttonState.focus.Visible(true, clickable.History()) {
			t.Fatal("pointer focus should not show the focus ring")
		}
	}

	{
		ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
		router := new(input.Router)
		button := CloseButton("close")
		layoutCloseButtonFrame(ctx, router, button, start)
		clickable, ok := frame.PeekState[widget.Clickable](ctx, "close", "clickable")
		if !ok {
			t.Fatal("close button clickable state is missing")
		}
		router.Source().Execute(key.FocusCmd{Tag: clickable})
		layoutCloseButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		buttonState, ok := frame.PeekState[closeButtonState](ctx, "close", stateSlotCloseButton)
		if !ok || !buttonState.focus.Visible(true, clickable.History()) {
			t.Fatal("keyboard focus should show the focus ring")
		}
	}
}

func TestCloseButtonFocusVisibleIgnoresPointerFocus(t *testing.T) {
	var buttonState closeButtonState
	if !buttonState.focus.Visible(true, nil) {
		t.Fatal("keyboard focus should be visible")
	}
	buttonState.focus.Visible(false, nil)
	if buttonState.focus.Visible(true, []widget.Press{{Start: time.Unix(1, 0)}}) {
		t.Fatal("pointer focus should not be visible")
	}
}

func TestCloseButtonPressedScale(t *testing.T) {
	if got := closeButtonPressedScale(0.93); got != 0.93 {
		t.Fatalf("pressed scale = %v, want 0.93", got)
	}
	if got := closeButtonPressedScale(0); got != 0.93 {
		t.Fatalf("invalid pressed scale fallback = %v, want 0.93", got)
	}
}

func TestCloseButtonScaleTransitionStaysContinuousOnQuickRelease(t *testing.T) {
	var buttonState closeButtonState
	gtx := closeButtonTestContext()
	start := time.Unix(1, 0)
	gtx.Now = start
	if got := buttonState.scale(gtx, 1); got != 1 {
		t.Fatalf("initial scale = %v, want 1", got)
	}
	buttonState.scale(gtx, 0.93)
	gtx.Now = start.Add(closeButtonScaleDuration / 4)
	pressed := buttonState.scale(gtx, 0.93)
	if pressed <= 0.93 || pressed >= 1 {
		t.Fatalf("partial press scale = %v, want between 0.93 and 1", pressed)
	}
	if released := buttonState.scale(gtx, 1); released != pressed {
		t.Fatalf("release scale jumped from %v to %v", pressed, released)
	}
}

func TestCloseButtonDefaultLabelUsesContextLanguage(t *testing.T) {
	button := CloseButton("close")
	if got := button.semanticLabel(newCloseButtonContext(nil, locale.LanguageEnglish)); got != "Close" {
		t.Fatalf("English label = %q, want Close", got)
	}
	if got := button.semanticLabel(newCloseButtonContext(nil, locale.LanguageChinese)); got != "关闭" {
		t.Fatalf("Chinese label = %q, want 关闭", got)
	}
	if got := button.Label("Dismiss").semanticLabel(newCloseButtonContext(nil, locale.LanguageChinese)); got != "Dismiss" {
		t.Fatalf("custom label = %q, want Dismiss", got)
	}
}

func TestCloseButtonSemanticsIncludeLabelAndDisabledState(t *testing.T) {
	for _, test := range []struct {
		name     string
		button   CloseButtonWidget
		language locale.Language
		label    string
		disabled bool
	}{
		{name: "localized", button: CloseButton("close"), language: locale.LanguageChinese, label: "关闭"},
		{name: "custom disabled", button: CloseButton("dismiss").Label("Dismiss").Disabled(true), language: locale.LanguageEnglish, label: "Dismiss", disabled: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newCloseButtonContext(nil, test.language)
			router := new(input.Router)
			layoutCloseButtonFrame(ctx, router, test.button, time.Unix(1, 0))
			node, ok := closeButtonSemanticNode(router.AppendSemantics(nil))
			if !ok {
				t.Fatal("semantic tree does not contain a button")
			}
			if node.Desc.Label != test.label || node.Desc.Disabled != test.disabled {
				t.Fatalf("button semantics = label %q disabled %v, want %q/%v", node.Desc.Label, node.Desc.Disabled, test.label, test.disabled)
			}
			if node.Desc.Bounds.Size() != image.Pt(24, 24) {
				t.Fatalf("semantic bounds = %v, want 24x24", node.Desc.Bounds)
			}
		})
	}
}

type closeButtonProbe struct {
	constraints layout.Constraints
	foreground  color.NRGBA
}

func (p *closeButtonProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	p.foreground = ctx.ForegroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func newCloseButtonContext(activeTheme *theme.Theme, language locale.Language) *frame.Context {
	return frame.New(nil, activeTheme, language)
}

func closeButtonTestContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
	}
}

func layoutCloseButtonFrame(ctx *frame.Context, router *input.Router, button CloseButtonWidget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	button.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func closeButtonSemanticNode(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			return node, true
		}
		if child, ok := closeButtonSemanticNode(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func layoutDisabledCloseOverBackground(ctx *frame.Context, router *input.Router, background *widget.Clickable, button CloseButtonWidget, clicked *bool, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(24, 24)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	for background.Clicked(gtx) {
		*clicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(24, 24)}
	})
	button.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
