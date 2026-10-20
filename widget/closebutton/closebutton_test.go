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
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func TestCloseButtonLayoutUsesThemeGeometry(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	activeTheme.Components.CloseButton.Size = 30
	activeTheme.Components.CloseButton.Padding = 5
	activeTheme.Components.CloseButton.IconSize = 22
	probe := new(closeButtonProbe)
	dims := New("close").Icon(probe).Layout(newCloseButtonContext(&activeTheme, locale.LanguageEnglish), closeButtonTestContext())

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
			core.UseStateWith(ctx, "close", "clickable", func() *widget.Clickable { return clickable })
			New("close").Disabled(test.disabled).OnClick(func() { clicked = true }).Layout(ctx, closeButtonTestContext())
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
		button := New("close").Disabled(disabled).OnClick(func() { clicked = true })
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
	button := New("close").Disabled(true)
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

func TestCloseButtonPointerAndKeyboardFocusVisibility(t *testing.T) {
	start := time.Unix(1, 0)
	{
		ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
		router := new(input.Router)
		button := New("close")
		layoutCloseButtonFrame(ctx, router, button, start)
		clickable, ok := core.PeekState[widget.Clickable](ctx, "close", "clickable")
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
		if !router.Source().Focused(clickable) {
			t.Fatal("pointer press did not focus the close button")
		}
		if core.FocusVisible(ctx, clickable, true) {
			t.Fatal("pointer focus should not show the focus ring")
		}
	}

	{
		ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
		router := new(input.Router)
		button := New("close")
		layoutCloseButtonFrame(ctx, router, button, start)
		clickable, ok := core.PeekState[widget.Clickable](ctx, "close", "clickable")
		if !ok {
			t.Fatal("close button clickable state is missing")
		}
		router.Source().Execute(key.FocusCmd{Tag: clickable})
		layoutCloseButtonFrame(ctx, router, button, start.Add(time.Millisecond))
		if !core.FocusVisible(ctx, clickable, true) {
			t.Fatal("keyboard focus should show the focus ring")
		}
	}
}

func TestCloseButtonDefaultLabelUsesContextLanguage(t *testing.T) {
	button := New("close")
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
		button   Widget
		language locale.Language
		label    string
		disabled bool
	}{
		{name: "localized", button: New("close"), language: locale.LanguageChinese, label: "关闭"},
		{name: "custom disabled", button: New("dismiss").Label("Dismiss").Disabled(true), language: locale.LanguageEnglish, label: "Dismiss", disabled: true},
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

func TestCloseButtonStyledOverrideAppliesProviderGeometry(t *testing.T) {
	active := *testtheme.Behavior()
	base := active.VisualProvider.(theme.ComponentVisualProvider)
	provider := &recordingCloseButtonProvider{ComponentVisualProvider: base}
	active.VisualProvider = provider
	ctx := core.New(nil, &active, locale.LanguageEnglish)
	probe := new(closeButtonProbe)
	declaration := style.Style{}.
		Width(40).
		Height(36).
		Radius(3).
		Part(style.PartIcon, style.Style{}.Width(12).Height(10))
	value := (New("styled").Icon(probe)).Style(declaration)
	dimensions := value.Layout(ctx, closeButtonTestContext())

	if dimensions.Size != image.Pt(40, 36) {
		t.Fatalf("styled close button size = %v, want (40,36)", dimensions.Size)
	}
	if probe.constraints != layout.Exact(image.Pt(12, 10)) {
		t.Fatalf("styled icon constraints = %#v, want exact 12x10", probe.constraints)
	}
	if len(provider.requests) != 1 || provider.requests[0].Override == nil {
		t.Fatalf("close button provider requests = %#v, want one instance override", provider.requests)
	}
	root := style.Cascade(style.State{}, provider.requests[0].Override.Declaration)
	if root.Box == nil || root.Box.Width == nil || *root.Box.Width != 40 || root.Box.Height == nil || *root.Box.Height != 36 {
		t.Fatalf("provider instance declaration = %#v, want 40x36 root geometry", root.Box)
	}
}

func TestCloseButtonStyledOverridePreservesBehaviorAndSemantics(t *testing.T) {
	for _, test := range []struct {
		name      string
		disabled  bool
		wantClick bool
		wantLabel string
	}{
		{name: "enabled", wantClick: true, wantLabel: "Dismiss"},
		{name: "disabled", disabled: true, wantLabel: "Dismiss"},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx := newCloseButtonContext(nil, locale.LanguageEnglish)
			router := new(input.Router)
			clicked := false
			button := (New("styled-behavior").
				Label(test.wantLabel).
				Disabled(test.disabled).
				OnClick(func() { clicked = true })).Style(style.Style{}.Width(40).Height(36))
			start := time.Unix(1, 0)
			layoutCloseButtonFrame(ctx, router, button, start)
			router.Queue(
				pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(12, 12)},
				pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(12, 12)},
			)
			layoutCloseButtonFrame(ctx, router, button, start.Add(time.Millisecond))
			if clicked != test.wantClick {
				t.Fatalf("styled click = %v, want %v", clicked, test.wantClick)
			}
			node, ok := closeButtonSemanticNode(router.AppendSemantics(nil))
			if !ok {
				t.Fatal("styled semantic tree does not contain a button")
			}
			if node.Desc.Label != test.wantLabel || node.Desc.Disabled != test.disabled {
				t.Fatalf("styled semantics = label %q disabled %v, want %q/%v", node.Desc.Label, node.Desc.Disabled, test.wantLabel, test.disabled)
			}
		})
	}
}

type recordingCloseButtonProvider struct {
	theme.ComponentVisualProvider
	requests []theme.CloseButtonRequest
}

func (p *recordingCloseButtonProvider) CloseButton(active *theme.Theme, request theme.CloseButtonRequest) theme.CloseButtonVisual {
	p.requests = append(p.requests, request)
	return p.ComponentVisualProvider.CloseButton(active, request)
}

type closeButtonProbe struct {
	constraints layout.Constraints
	foreground  color.NRGBA
}

func (p *closeButtonProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.constraints = gtx.Constraints
	p.foreground = ctx.ForegroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Min}
}

func newCloseButtonContext(activeTheme *theme.Theme, language locale.Language) *core.Context {
	if activeTheme == nil {
		activeTheme = testtheme.Behavior()
	}
	return core.New(nil, activeTheme, language)
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

func layoutCloseButtonFrame(ctx *core.Context, router *input.Router, button core.Widget, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	button.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
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

func layoutDisabledCloseOverBackground(ctx *core.Context, router *input.Router, background *widget.Clickable, button Widget, clicked *bool, now time.Time) {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(24, 24)),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	core.BeginFrame(ctx)
	for background.Clicked(gtx) {
		*clicked = true
	}
	background.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Dimensions{Size: image.Pt(24, 24)}
	})
	button.Layout(ctx, gtx)
	core.ApplyFrameCommands(ctx, gtx)
	core.EndFrame(ctx)
	router.Frame(&ops)
}
