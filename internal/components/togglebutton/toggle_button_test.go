package togglebutton

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
	"github.com/qianniancn/flowui/internal/components/text"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestToggleButtonUsesHeroUISizes(t *testing.T) {
	for _, test := range []struct {
		name string
		size ToggleButtonSize
		want int
	}{
		{name: "small", size: ToggleButtonSmall, want: 32},
		{name: "medium", size: ToggleButtonMedium, want: 36},
		{name: "large", size: ToggleButtonLarge, want: 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			dims := ToggleButton("size", false, text.New("Toggle")).
				Size(test.size).
				Layout(newToggleButtonContext(nil), toggleButtonTestContext())
			if dims.Size.Y != test.want {
				t.Fatalf("height = %d, want %d", dims.Size.Y, test.want)
			}
		})
	}
}

func TestToggleButtonIconOnlyIsSquare(t *testing.T) {
	for _, test := range []struct {
		size ToggleButtonSize
		want int
	}{
		{size: ToggleButtonSmall, want: 32},
		{size: ToggleButtonMedium, want: 36},
		{size: ToggleButtonLarge, want: 40},
	} {
		dims := ToggleButton("icon", false, text.New("B")).
			Size(test.size).
			IconOnly().
			Layout(newToggleButtonContext(nil), toggleButtonTestContext())
		if dims.Size != image.Pt(test.want, test.want) {
			t.Fatalf("size %d icon dimensions = %v, want square %d", test.size, dims.Size, test.want)
		}
	}
}

func TestToggleButtonUsesThemeHeight(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.ToggleButton.MediumHeight = 42
	dims := ToggleButton("theme", false, text.New("Theme")).Layout(newToggleButtonContext(&activeTheme), toggleButtonTestContext())
	if dims.Size.Y != 42 {
		t.Fatalf("height = %d, want 42", dims.Size.Y)
	}
}

func TestToggleButtonVariantAndSelectedColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	current := color.NRGBA{R: 10, G: 20, B: 30, A: 255}
	defaultIdle := toggleButtonWidgetFor(&activeTheme, current, ToggleButtonDefault, false, false, false, false)
	defaultHover := toggleButtonWidgetFor(&activeTheme, current, ToggleButtonDefault, false, true, false, false)
	ghostIdle := toggleButtonWidgetFor(&activeTheme, current, ToggleButtonGhost, false, false, false, false)
	selected := toggleButtonWidgetFor(&activeTheme, current, ToggleButtonDefault, true, false, false, false)
	selectedHover := toggleButtonWidgetFor(&activeTheme, current, ToggleButtonGhost, true, true, false, false)
	disabled := toggleButtonWidgetFor(&activeTheme, current, ToggleButtonDefault, false, false, false, true)

	if defaultIdle.background != activeTheme.Palette.SurfaceRaised || defaultIdle.foreground != current {
		t.Fatalf("default idle style = %#v", defaultIdle)
	}
	if defaultHover.background != activeTheme.Palette.SurfacePressed {
		t.Fatalf("default hover background = %#v", defaultHover.background)
	}
	if ghostIdle.background.A != 0 || ghostIdle.foreground != activeTheme.Palette.Foreground {
		t.Fatalf("ghost idle style = %#v", ghostIdle)
	}
	if selected.background != activeTheme.Palette.AccentSoft || selected.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("selected style = %#v", selected)
	}
	if selectedHover.background != activeTheme.Palette.AccentSoftHover || selectedHover.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("selected hover style = %#v", selectedHover)
	}
	if disabled.opacity != activeTheme.DisabledOpacityValue() {
		t.Fatalf("disabled opacity = %v, want %v", disabled.opacity, activeTheme.DisabledOpacityValue())
	}
}

func TestToggleButtonReportsInverseControlledValue(t *testing.T) {
	for _, selected := range []bool{false, true} {
		ctx := newToggleButtonContext(nil)
		clickable := new(widget.Clickable)
		clickable.Click()
		frame.UseStateWith(ctx, "toggle", "clickable", func() *widget.Clickable { return clickable })
		var got bool
		called := false
		ToggleButton("toggle", selected, text.New("Toggle")).OnChange(func(value bool) {
			called = true
			got = value
		}).Layout(ctx, toggleButtonTestContext())
		if !called || got == selected {
			t.Fatalf("selected=%v callback called=%v value=%v", selected, called, got)
		}
	}
}

func TestToggleButtonUsesChangedValueInClickFrame(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := newToggleButtonContext(&activeTheme)
	clickable := new(widget.Clickable)
	clickable.Click()
	frame.UseStateWith(ctx, "toggle", "clickable", func() *widget.Clickable { return clickable })
	probe := new(toggleButtonProbe)
	ToggleButton("toggle", false, probe).OnChange(func(bool) {}).Layout(ctx, toggleButtonTestContext())
	if probe.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("click-frame foreground = %#v, want selected %#v", probe.foreground, activeTheme.Palette.AccentSoftForeground)
	}
}

func TestToggleButtonPointerAndKeyboardToggle(t *testing.T) {
	for _, inputKind := range []string{"pointer", "keyboard"} {
		t.Run(inputKind, func(t *testing.T) {
			ctx := newToggleButtonContext(nil)
			router := new(input.Router)
			changed := false
			button := ToggleButton("toggle", false, text.New("Toggle")).OnChange(func(selected bool) { changed = selected })
			start := time.Unix(1, 0)
			layoutToggleButtonFrame(ctx, router, button, start)
			clickable := toggleButtonClickableFromContext(t, ctx, "toggle")
			if inputKind == "pointer" {
				router.Queue(
					pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(18, 18)},
					pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(18, 18)},
				)
			} else {
				router.Source().Execute(key.FocusCmd{Tag: clickable})
				layoutToggleButtonFrame(ctx, router, button, start.Add(time.Millisecond))
				router.Queue(
					key.Event{Name: key.NameSpace, State: key.Press},
					key.Event{Name: key.NameSpace, State: key.Release},
				)
			}
			layoutToggleButtonFrame(ctx, router, button, start.Add(2*time.Millisecond))
			if !changed {
				t.Fatal("toggle input did not report selected=true")
			}
		})
	}
}

func TestToggleButtonDisabledBlocksChangesAndUnderlyingPointerPasses(t *testing.T) {
	ctx := newToggleButtonContext(nil)
	router := new(input.Router)
	background := new(widget.Clickable)
	changed := false
	backgroundClicked := false
	button := ToggleButton("toggle", false, text.New("B")).
		IconOnly().
		Disabled(true).
		OnChange(func(bool) { changed = true })
	start := time.Unix(1, 0)
	layoutDisabledToggleOverBackground(ctx, router, background, button, &backgroundClicked, start)
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(18, 18)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(18, 18)},
	)
	layoutDisabledToggleOverBackground(ctx, router, background, button, &backgroundClicked, start.Add(time.Millisecond))
	if changed || !backgroundClicked {
		t.Fatalf("disabled toggle changed=%v background clicked=%v", changed, backgroundClicked)
	}
}

func TestToggleButtonSemanticsExposeSelectionAndDisabledState(t *testing.T) {
	ctx := newToggleButtonContext(nil)
	router := new(input.Router)
	button := ToggleButton("bold", true, text.New("B")).IconOnly().Label("Bold").Disabled(true)
	layoutToggleButtonFrame(ctx, router, button, time.Unix(1, 0))
	node, ok := toggleButtonSemanticNode(router.AppendSemantics(nil))
	if !ok {
		t.Fatal("semantic tree does not contain a button")
	}
	if node.Desc.Label != "Bold" || !node.Desc.Selected || !node.Desc.Disabled {
		t.Fatalf("semantics = label %q selected %v disabled %v", node.Desc.Label, node.Desc.Selected, node.Desc.Disabled)
	}
	if node.Desc.Bounds.Size() != image.Pt(36, 36) {
		t.Fatalf("semantic bounds = %v, want 36x36", node.Desc.Bounds)
	}
}

func TestToggleButtonOnlyShowsFocusRingForKeyboardFocus(t *testing.T) {
	start := time.Unix(1, 0)
	{
		ctx := newToggleButtonContext(nil)
		router := new(input.Router)
		button := ToggleButton("toggle", false, text.New("Toggle"))
		layoutToggleButtonFrame(ctx, router, button, start)
		clickable := toggleButtonClickableFromContext(t, ctx, "toggle")
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(18, 18)})
		layoutToggleButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		if !router.Source().Focused(clickable) {
			t.Fatal("pointer press did not focus toggle button")
		}
		if frame.FocusVisible(ctx, clickable, true) {
			t.Fatal("pointer focus should not be focus-visible")
		}
	}
	{
		ctx := newToggleButtonContext(nil)
		router := new(input.Router)
		button := ToggleButton("toggle", false, text.New("Toggle"))
		layoutToggleButtonFrame(ctx, router, button, start)
		clickable := toggleButtonClickableFromContext(t, ctx, "toggle")
		router.Source().Execute(key.FocusCmd{Tag: clickable})
		layoutToggleButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		if !frame.FocusVisible(ctx, clickable, true) {
			t.Fatal("keyboard focus should be focus-visible")
		}
	}
}

func TestToggleButtonChildReceivesForegroundAndDisabledContext(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	probe := new(toggleButtonProbe)
	ToggleButton("selected", true, probe).
		Disabled(true).
		Layout(newToggleButtonContext(&activeTheme), toggleButtonTestContext())
	if probe.enabled {
		t.Fatal("disabled toggle laid out child with enabled context")
	}
	if probe.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("child foreground = %#v, want %#v", probe.foreground, activeTheme.Palette.AccentSoftForeground)
	}
}

type toggleButtonProbe struct {
	enabled    bool
	foreground color.NRGBA
}

func (p *toggleButtonProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.enabled = gtx.Enabled()
	p.foreground = ctx.ForegroundColor()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func newToggleButtonContext(activeTheme *theme.Theme) *frame.Context {
	return frame.New(nil, activeTheme, locale.LanguageEnglish)
}

func toggleButtonTestContext() layout.Context {
	var router input.Router
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         time.Unix(1, 0),
	}
}

func layoutToggleButtonFrame(ctx *frame.Context, router *input.Router, button ToggleButtonWidget, now time.Time) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	dims := button.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func toggleButtonClickableFromContext(t *testing.T, ctx *frame.Context, key string) *widget.Clickable {
	t.Helper()
	clickable, ok := frame.PeekState[widget.Clickable](ctx, key, "clickable")
	if !ok {
		t.Fatalf("toggle button clickable %q is missing", key)
	}
	return clickable
}

func toggleButtonSemanticNode(nodes []input.SemanticNode) (input.SemanticNode, bool) {
	for _, node := range nodes {
		if node.Desc.Class == semantic.Button {
			return node, true
		}
		if child, ok := toggleButtonSemanticNode(node.Children); ok {
			return child, true
		}
	}
	return input.SemanticNode{}, false
}

func layoutDisabledToggleOverBackground(ctx *frame.Context, router *input.Router, background *widget.Clickable, button ToggleButtonWidget, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(36, 36)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(36, 36)}
	})
	button.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}
