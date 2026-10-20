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
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/theme"
)

func TestToggleButtonUsesStandardSizes(t *testing.T) {
	for _, test := range []struct {
		name string
		size Size
		want int
	}{
		{name: "small", size: Small, want: 32},
		{name: "medium", size: Medium, want: 36},
		{name: "large", size: Large, want: 40},
	} {
		t.Run(test.name, func(t *testing.T) {
			dims := New("Toggle").Key("size").Selected(false).
				Size(test.size).
				Layout(newToggleButtonContext(testtheme.Behavior()), toggleButtonTestContext())
			if dims.Size.Y != test.want {
				t.Fatalf("height = %d, want %d", dims.Size.Y, test.want)
			}
		})
	}
}

func TestToggleButtonIconOnlyIsSquare(t *testing.T) {
	for _, test := range []struct {
		size Size
		want int
	}{
		{size: Small, want: 32},
		{size: Medium, want: 36},
		{size: Large, want: 40},
	} {
		dims := New("B").Key("icon").Selected(false).
			Size(test.size).
			IconOnly().
			Layout(newToggleButtonContext(testtheme.Behavior()), toggleButtonTestContext())
		if dims.Size != image.Pt(test.want, test.want) {
			t.Fatalf("size %d icon dimensions = %v, want square %d", test.size, dims.Size, test.want)
		}
	}
}

func TestToggleButtonUsesThemeHeight(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	activeTheme.Components.ToggleButton.MediumHeight = 42
	dims := New("Theme").Key("theme").Selected(false).Layout(newToggleButtonContext(&activeTheme), toggleButtonTestContext())
	if dims.Size.Y != 42 {
		t.Fatalf("height = %d, want 42", dims.Size.Y)
	}
}

func TestToggleButtonReportsInverseControlledValue(t *testing.T) {
	for _, selected := range []bool{false, true} {
		ctx := newToggleButtonContext(testtheme.Behavior())
		clickable := new(widget.Clickable)
		clickable.Click()
		core.UseStateWith(ctx, "toggle", "clickable", func() *widget.Clickable { return clickable })
		var got bool
		called := false
		New("Toggle").Key("toggle").Selected(selected).OnChange(func(value bool) {
			called = true
			got = value
		}).Layout(ctx, toggleButtonTestContext())
		if !called || got == selected {
			t.Fatalf("selected=%v callback called=%v value=%v", selected, called, got)
		}
	}
}

func TestToggleButtonUsesChangedValueInClickFrame(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	ctx := newToggleButtonContext(&activeTheme)
	clickable := new(widget.Clickable)
	clickable.Click()
	core.UseStateWith(ctx, "toggle", "clickable", func() *widget.Clickable { return clickable })
	probe := new(toggleButtonProbe)
	New("").Key("toggle").Selected(false).Content(probe).OnChange(func(bool) {}).Layout(ctx, toggleButtonTestContext())
	if probe.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("click-frame foreground = %#v, want selected %#v", probe.foreground, activeTheme.Palette.AccentSoftForeground)
	}
}

func TestToggleButtonPointerAndKeyboardToggle(t *testing.T) {
	for _, inputKind := range []string{"pointer", "keyboard"} {
		t.Run(inputKind, func(t *testing.T) {
			ctx := newToggleButtonContext(testtheme.Behavior())
			router := new(input.Router)
			changed := false
			button := New("Toggle").Key("toggle").Selected(false).OnChange(func(selected bool) { changed = selected })
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
	ctx := newToggleButtonContext(testtheme.Behavior())
	router := new(input.Router)
	background := new(widget.Clickable)
	changed := false
	backgroundClicked := false
	button := New("B").Key("toggle").Selected(false).
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
	ctx := newToggleButtonContext(testtheme.Behavior())
	router := new(input.Router)
	button := New("B").Key("bold").Selected(true).IconOnly().Label("Bold").Disabled(true)
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

func TestToggleButtonTextAndCustomContentKeepAccessibleLabel(t *testing.T) {
	for _, custom := range []bool{false, true} {
		ctx := newToggleButtonContext(testtheme.Behavior())
		router := new(input.Router)
		value := New("Bold").Key("bold").Selected(true)
		probe := new(toggleButtonProbe)
		if custom {
			value = value.Content(probe).IconOnly()
		}
		layoutToggleButtonFrame(ctx, router, value, time.Unix(1, 0))
		node, ok := toggleButtonSemanticNode(router.AppendSemantics(nil))
		if !ok || node.Desc.Label != "Bold" || !node.Desc.Selected {
			t.Fatalf("custom=%v semantics = %#v", custom, node.Desc)
		}
		if custom && probe.foreground == (color.NRGBA{}) {
			t.Fatal("custom toggle content was not laid out")
		}
	}
}

func TestToggleButtonOnlyShowsFocusRingForKeyboardFocus(t *testing.T) {
	start := time.Unix(1, 0)
	{
		ctx := newToggleButtonContext(testtheme.Behavior())
		router := new(input.Router)
		button := New("Toggle").Key("toggle").Selected(false)
		layoutToggleButtonFrame(ctx, router, button, start)
		clickable := toggleButtonClickableFromContext(t, ctx, "toggle")
		router.Queue(pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(18, 18)})
		layoutToggleButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		if !router.Source().Focused(clickable) {
			t.Fatal("pointer press did not focus toggle button")
		}
		if core.FocusVisible(ctx, clickable, true) {
			t.Fatal("pointer focus should not be focus-visible")
		}
	}
	{
		ctx := newToggleButtonContext(testtheme.Behavior())
		router := new(input.Router)
		button := New("Toggle").Key("toggle").Selected(false)
		layoutToggleButtonFrame(ctx, router, button, start)
		clickable := toggleButtonClickableFromContext(t, ctx, "toggle")
		router.Source().Execute(key.FocusCmd{Tag: clickable})
		layoutToggleButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		if !core.FocusVisible(ctx, clickable, true) {
			t.Fatal("keyboard focus should be focus-visible")
		}
	}
}

func TestToggleButtonChildReceivesForegroundAndDisabledContext(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	probe := new(toggleButtonProbe)
	New("").Key("selected").Selected(true).Content(probe).
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

func (p *toggleButtonProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.enabled = gtx.Enabled()
	p.foreground = ctx.ForegroundColor()
	return layout.Dimensions{Size: image.Pt(16, 16)}
}

func newToggleButtonContext(activeTheme *theme.Theme) *core.Context {
	return core.New(nil, activeTheme, locale.LanguageEnglish)
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

func layoutToggleButtonFrame(ctx *core.Context, router *input.Router, button Widget, now time.Time) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	dims := button.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func toggleButtonClickableFromContext(t *testing.T, ctx *core.Context, key string) *widget.Clickable {
	t.Helper()
	clickable, ok := core.PeekState[widget.Clickable](ctx, key, "clickable")
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

func layoutDisabledToggleOverBackground(ctx *core.Context, router *input.Router, background *widget.Clickable, button Widget, backgroundClicked *bool, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(36, 36)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	for background.Clicked(gtx) {
		*backgroundClicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(36, 36)}
	})
	button.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
