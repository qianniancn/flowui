package titlebar

import (
	"image"
	"image/color"
	"testing"
	"time"

	"gioui.org/app"
	"gioui.org/f32"
	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/pointer"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/components/menu"
	"github.com/qianniancn/flowui/internal/components/menubar"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestTitleBarStyleUsesValueSemantics(t *testing.T) {
	base := New("workspace", "FlowUI", nil)
	styled := base.
		Leading(fixedMenu{size: image.Pt(20, 20)}).
		ShowClose(false).
		Style(flowstyle.Style{}.Height(40))
	if base.customStyle.Resolve(flowstyle.StyleState{}).Box != nil || styled.customStyle.Resolve(flowstyle.StyleState{}).Box == nil {
		t.Fatal("TitleBar style did not preserve value semantics")
	}
	if base.leading != nil || !base.showClose || styled.leading == nil || styled.showClose {
		t.Fatal("TitleBar configuration did not preserve value semantics")
	}
}

func TestTitleBarLabelPartUsesCommonBoxRenderer(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageEnglish)
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(300, 100)}, Ops: new(op.Ops)}
	base := New("base", "FlowUI", nil).resolveStyle(ctx, gtx, "base", false)
	styled := New("styled", "FlowUI", nil).
		Style(flowstyle.Style{}.Part(flowstyle.PartLabel, flowstyle.Style{}.PaddingY(7))).
		resolveStyle(ctx, gtx, "styled", false)
	baseDims := layoutTitle(ctx, gtx, "FlowUI", base.title)
	gtx.Ops = new(op.Ops)
	styledDims := layoutTitle(ctx, gtx, "FlowUI", styled.title)
	if styledDims.Size.Y != baseDims.Size.Y+14 {
		t.Fatalf("styled title height = %d, want %d", styledDims.Size.Y, baseDims.Size.Y+14)
	}
}

func TestTitleBarCloseControlIsFlushAndUsesLightHoverGlyph(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(600, 100)}, Ops: new(op.Ops)}
	bar := New("workspace", "FlowUI", nil)
	resolved := bar.resolveStyle(ctx, gtx, "workspace", false)
	if resolved.root.Box == nil || resolved.root.Box.Padding == nil {
		t.Fatal("title bar root padding is missing")
	}
	padding := resolved.root.Box.Padding
	if padding.Left != activeTheme.Components.TitleBar.PaddingX || padding.Right != 0 {
		t.Fatalf("title bar horizontal padding = left %v, right %v", padding.Left, padding.Right)
	}
	if resolved.leading.Box == nil || resolved.leading.Box.Padding == nil || resolved.leading.Box.Padding.Right != activeTheme.Components.TitleBar.LeadingGap {
		t.Fatalf("title bar leading style = %#v", resolved.leading.Box)
	}
	restoreStyle := frame.PushInheritedStyle(ctx, flowstyle.TextDeclaration(resolved.root.Text))
	defer restoreStyle()

	closeHover := bar.resolveControlStyle(
		ctx,
		gtx,
		"workspace",
		system.ActionClose,
		true,
		flowstyle.StyleState{Hovered: true},
	)
	if closeHover.Text == nil {
		t.Fatal("close hover text style is missing")
	}
	if got, ok := styleruntime.Color(closeHover.Text.Color); !ok || got != activeTheme.Palette.AccentForeground {
		t.Fatalf("close hover glyph color = %v, want %v", got, activeTheme.Palette.AccentForeground)
	}
	if got := resolvedControlForeground(closeHover, color.NRGBA{}); got != activeTheme.Palette.AccentForeground {
		t.Fatalf("drawn close hover glyph color = %v, want %v", got, activeTheme.Palette.AccentForeground)
	}
}

func TestTitleBarFillsWidthAndLimitsMoveRegion(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	bar := New("workspace", "main.go - FlowUI", fixedMenu{size: image.Pt(120, 32)})

	dims := layoutTitleBarFrame(ctx, router, bar, time.Unix(1, 0), image.Pt(600, 35))
	if dims.Size != image.Pt(600, 35) {
		t.Fatalf("title bar size = %v, want (600,35)", dims.Size)
	}
	for _, point := range []f32.Point{f32.Pt(240, 1), f32.Pt(240, 16), f32.Pt(240, 33)} {
		if action, ok := router.ActionAt(point); !ok || action != system.ActionMove {
			t.Fatalf("title move action at %v = %v, found %v", point, action, ok)
		}
	}
	for _, point := range []f32.Point{f32.Pt(20, 16), f32.Pt(580, 16)} {
		if action, ok := router.ActionAt(point); ok && action == system.ActionMove {
			t.Fatalf("non-title point %v was marked movable", point)
		}
	}
	semantics := router.AppendSemantics(nil)
	for _, label := range []string{"Minimize", "Maximize", "Close"} {
		if !hasSemanticLabel(semantics, label) {
			t.Fatalf("missing %q control semantics", label)
		}
	}
}

func TestTitleBarTracksMaximizedWindowState(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	bar := New("workspace", "FlowUI", nil)

	layoutTitleBarFrame(ctx, router, bar, time.Unix(2, 0), image.Pt(500, 35))
	state, ok := frame.PeekState[titleBarState](ctx, "workspace", stateSlotTitleBar)
	if !ok {
		t.Fatal("missing title bar state")
	}
	if state.decorations.Maximized {
		t.Fatal("windowed title bar reported maximized")
	}

	frame.UpdateWindowConfig(ctx, app.Config{Mode: app.Maximized})
	layoutTitleBarFrame(ctx, router, bar, time.Unix(2, 0).Add(time.Millisecond), image.Pt(500, 35))
	if !state.decorations.Maximized {
		t.Fatal("maximized title bar did not switch to restore state")
	}
}

func TestTitleBarSlotsStayOutsideMoveRegionsAndControlsCanBeHidden(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	bar := New("workspace", "FlowUI", fixedMenu{size: image.Pt(80, 32)}).
		Leading(fixedMenu{size: image.Pt(24, 24)}).
		Center(fixedMenu{size: image.Pt(120, 30)}).
		Trailing(fixedMenu{size: image.Pt(40, 30)}).
		ShowMinimize(false).
		ShowMaximize(false).
		Style(flowstyle.Style{}.Height(44))

	dims := layoutTitleBarFrame(ctx, router, bar, time.Unix(3, 0), image.Pt(640, 44))
	if dims.Size != image.Pt(640, 44) {
		t.Fatalf("title bar size = %v, want (640,44)", dims.Size)
	}
	for _, point := range []f32.Point{f32.Pt(180, 22), f32.Pt(470, 22)} {
		if action, ok := router.ActionAt(point); !ok || action != system.ActionMove {
			t.Fatalf("move action at %v = %v, found %v", point, action, ok)
		}
	}
	for _, point := range []f32.Point{f32.Pt(16, 22), f32.Pt(320, 22), f32.Pt(565, 22), f32.Pt(620, 22)} {
		if action, ok := router.ActionAt(point); ok && action == system.ActionMove {
			t.Fatalf("interactive point %v was marked movable", point)
		}
	}
	semantics := router.AppendSemantics(nil)
	if hasSemanticLabel(semantics, "Minimize") || hasSemanticLabel(semantics, "Maximize") {
		t.Fatal("hidden title bar controls remained in the semantic tree")
	}
	if !hasSemanticLabel(semantics, "Close") {
		t.Fatal("visible close control is missing from the semantic tree")
	}
}

func TestTitleBarNativeDecorationFallbackOmitsWindowInteractions(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	bar := New("workspace", "FlowUI", fixedMenu{size: image.Pt(80, 32)}).
		Leading(fixedMenu{size: image.Pt(24, 24)}).
		Center(fixedMenu{size: image.Pt(120, 30)}).
		Trailing(fixedMenu{size: image.Pt(40, 30)}).
		Style(flowstyle.Style{}.Height(44))
	bar.clientDecorations = false

	dims := layoutTitleBarFrame(ctx, router, bar, time.Unix(4, 0), image.Pt(640, 44))
	if dims.Size != image.Pt(640, 44) {
		t.Fatalf("application header size = %v, want (640,44)", dims.Size)
	}
	for _, point := range []f32.Point{f32.Pt(16, 22), f32.Pt(180, 22), f32.Pt(320, 22), f32.Pt(610, 22)} {
		if action, ok := router.ActionAt(point); ok && action == system.ActionMove {
			t.Fatalf("native-decoration header point %v was marked movable", point)
		}
	}
	semantics := router.AppendSemantics(nil)
	for _, label := range []string{"Minimize", "Maximize", "Close"} {
		if hasSemanticLabel(semantics, label) {
			t.Fatalf("native-decoration header exposed %q control", label)
		}
	}
}

func TestTitleBarPlatformModeMatchesSupport(t *testing.T) {
	if got := NewPlatform("workspace", "FlowUI", nil).clientDecorations; got != Supported() {
		t.Fatalf("platform title bar client decorations = %v, want %v", got, Supported())
	}
}

func TestTitleBarCustomHeightAndCloseHandler(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	closeRequests := 0
	bar := New("workspace", "FlowUI", nil).
		ShowMinimize(false).
		ShowMaximize(false).
		OnClose(func() { closeRequests++ }).
		Style(flowstyle.Style{}.Height(48))

	dims := layoutTitleBarFlexibleFrame(ctx, router, bar, time.Unix(4, 0), image.Pt(600, 100))
	if dims.Size != image.Pt(600, 48) {
		t.Fatalf("custom title bar size = %v, want (600,48)", dims.Size)
	}
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(580, 46)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(580, 46)},
	)
	layoutTitleBarFlexibleFrame(ctx, router, bar, time.Unix(4, 0).Add(time.Millisecond), image.Pt(600, 100))
	if closeRequests != 1 {
		t.Fatalf("close handler calls = %d, want 1", closeRequests)
	}
}

func TestTitleBarDefaultCloseUsesWindowLifecycle(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	closeRequests := 0
	frame.SetWindowCloseRequest(ctx, func() { closeRequests++ })
	router := new(input.Router)
	bar := New("workspace", "FlowUI", nil).
		ShowMinimize(false).
		ShowMaximize(false).
		Style(flowstyle.Style{}.Height(44))

	layoutTitleBarFrame(ctx, router, bar, time.Unix(5, 0), image.Pt(600, 44))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(580, 42)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(580, 42)},
	)
	layoutTitleBarFrame(ctx, router, bar, time.Unix(5, 0).Add(time.Millisecond), image.Pt(600, 44))
	if closeRequests != 1 {
		t.Fatalf("window lifecycle close requests = %d, want 1", closeRequests)
	}
}

func TestTitleBarPreservesWindowControlsWhenCenterIsWide(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	closeRequests := 0
	bar := New("workspace", "FlowUI", nil).
		Center(constrainedFixed{size: image.Pt(500, 30)}).
		ShowMinimize(false).
		ShowMaximize(false).
		OnClose(func() { closeRequests++ }).
		Style(flowstyle.Style{}.Height(44))

	layoutTitleBarFrame(ctx, router, bar, time.Unix(5, 0), image.Pt(320, 44))
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1, Buttons: pointer.ButtonPrimary, Position: f32.Pt(300, 22)},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(300, 22)},
	)
	layoutTitleBarFrame(ctx, router, bar, time.Unix(5, 0).Add(time.Millisecond), image.Pt(320, 44))
	if closeRequests != 1 {
		t.Fatalf("close handler calls = %d, want 1", closeRequests)
	}
}

func TestTitleBarCenterKeepsItsOwnHeight(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	var centerConstraints layout.Constraints
	center := frame.WidgetFunc(func(_ *frame.Context, gtx layout.Context) layout.Dimensions {
		centerConstraints = gtx.Constraints
		return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(280, 30))}
	})
	bar := New("workspace", "FlowUI", nil).
		Center(center).
		Style(flowstyle.Style{}.Height(44))

	dims := layoutTitleBarFrame(ctx, router, bar, time.Unix(6, 0), image.Pt(640, 44))
	if dims.Size != image.Pt(640, 44) {
		t.Fatalf("title bar size = %v, want (640,44)", dims.Size)
	}
	if centerConstraints.Min.Y != 0 || centerConstraints.Max.Y != 44 {
		t.Fatalf("center height constraints = %v, want min 0 and max 44", centerConstraints)
	}
}

func TestTitleBarTracksMenubarPopupAfterLeadingContent(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	router := new(input.Router)
	action := ""
	applicationMenu := menubar.New("application-menu", []menubar.Item{
		menubar.NewMenu("file", "File", []menu.Item{{Key: "open", Label: "Open"}}).
			Width(180).
			OnActionEvent(func(event menu.ActionEvent) { action = event.Key }),
	}).Compact(true)
	bar := New("workspace", "FlowUI", applicationMenu).
		Leading(fixedMenu{size: image.Pt(100, 32)}).
		ShowMinimize(false).
		ShowMaximize(false).
		ShowClose(false)

	now := time.Unix(7, 0)
	layoutTitleBarOverlayFrame(ctx, router, bar, now, image.Pt(640, 44))
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(120, 20),
	})
	layoutTitleBarOverlayFrame(ctx, router, bar, now.Add(time.Millisecond), image.Pt(640, 44))
	router.Queue(pointer.Event{
		Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
		Position: f32.Pt(120, 20),
	})
	layoutTitleBarOverlayFrame(ctx, router, bar, now.Add(2*time.Millisecond), image.Pt(640, 44))
	menuOpenDuration := 150 * time.Millisecond
	layoutTitleBarOverlayFrame(ctx, router, bar, now.Add(menuOpenDuration+time.Millisecond), image.Pt(640, 44))

	// The leading slot moves the menu to roughly x=116. This point is inside
	// the correctly anchored popup but outside a popup incorrectly left at x=8.
	router.Queue(pointer.Event{
		Kind: pointer.Press, Source: pointer.Mouse, PointerID: 1,
		Buttons: pointer.ButtonPrimary, Position: f32.Pt(190, 60),
	})
	layoutTitleBarOverlayFrame(ctx, router, bar, now.Add(menuOpenDuration+2*time.Millisecond), image.Pt(640, 44))
	router.Queue(pointer.Event{
		Kind: pointer.Release, Source: pointer.Mouse, PointerID: 1,
		Position: f32.Pt(190, 60),
	})
	layoutTitleBarOverlayFrame(ctx, router, bar, now.Add(menuOpenDuration+3*time.Millisecond), image.Pt(640, 44))
	if action != "open" {
		t.Fatalf("menu action at leading-adjusted popup position = %q, want open", action)
	}
}

func TestTitleBarControlLabelsAndGlyphs(t *testing.T) {
	ctx := frame.New(nil, nil, locale.LanguageChinese)
	if got := controlLabel(ctx, system.ActionMaximize, false); got != "最大化" {
		t.Fatalf("maximize label = %q", got)
	}
	if got := controlLabel(ctx, system.ActionMaximize, true); got != "还原" {
		t.Fatalf("restore label = %q", got)
	}
	if glyph := controlGlyphFor(system.ActionMinimize, false); glyph != controlGlyphMinimize {
		t.Fatalf("minimize glyph = %v", glyph)
	}
	if glyph := controlGlyphFor(system.ActionMaximize, false); glyph != controlGlyphMaximize {
		t.Fatalf("maximize glyph = %v", glyph)
	}
	if glyph := controlGlyphFor(system.ActionMaximize, true); glyph != controlGlyphRestore {
		t.Fatalf("restore glyph = %v", glyph)
	}
	if glyph := controlGlyphFor(system.ActionClose, false); glyph != controlGlyphClose {
		t.Fatalf("close glyph = %v", glyph)
	}
	if glyph := controlGlyphFor(system.ActionRaise, false); glyph != controlGlyphNone {
		t.Fatalf("unsupported glyph = %v", glyph)
	}
}

func TestControlGlyphsRenderAtOneX(t *testing.T) {
	window, err := headless.NewWindow(46, 44)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	tokens := theme.DefaultTheme().Components.TitleBar
	tests := []struct {
		name        string
		glyph       controlGlyph
		filled      image.Point
		transparent image.Point
	}{
		{name: "minimize", glyph: controlGlyphMinimize, filled: image.Pt(23, 21), transparent: image.Pt(23, 20)},
		{name: "maximize", glyph: controlGlyphMaximize, filled: image.Pt(18, 17), transparent: image.Pt(23, 22)},
		{name: "restore", glyph: controlGlyphRestore, filled: image.Pt(22, 17), transparent: image.Pt(22, 22)},
		{name: "close", glyph: controlGlyphClose, filled: image.Pt(23, 22), transparent: image.Pt(23, 17)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ops op.Ops
			gtx := layout.Context{
				Constraints: layout.Exact(image.Pt(46, 44)),
				Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
				Ops:         &ops,
			}
			drawControlGlyph(gtx, image.Pt(46, 44), test.glyph, color.NRGBA{A: 0xff}, tokens)
			if err := window.Frame(&ops); err != nil {
				t.Fatal(err)
			}
			pixels := image.NewRGBA(image.Rect(0, 0, 46, 44))
			if err := window.Screenshot(pixels); err != nil {
				t.Fatal(err)
			}
			if alpha := pixels.RGBAAt(test.filled.X, test.filled.Y).A; alpha == 0 {
				t.Fatalf("expected filled pixel at %v", test.filled)
			}
			if alpha := pixels.RGBAAt(test.transparent.X, test.transparent.Y).A; alpha != 0 {
				t.Fatalf("pixel at %v alpha = %d, want 0", test.transparent, alpha)
			}
		})
	}
}

func TestControlGlyphStrokePreservesFractionalDp(t *testing.T) {
	tests := []struct {
		name   string
		metric unit.Metric
		want   float32
	}{
		{name: "default metric", metric: unit.Metric{}, want: 1.25},
		{name: "one x", metric: unit.Metric{PxPerDp: 1}, want: 1.25},
		{name: "one and a half x", metric: unit.Metric{PxPerDp: 1.5}, want: 1.875},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := controlGlyphStroke(test.metric, 1.25, 12); got != test.want {
				t.Fatalf("control glyph stroke = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCloseHoverGlyphRendersLightOnDangerBackground(t *testing.T) {
	window, err := headless.NewWindow(46, 44)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	activeTheme := theme.DefaultTheme()
	var ops op.Ops
	paint.Fill(&ops, activeTheme.Components.TitleBar.CloseHover)
	gtx := layout.Context{
		Constraints: layout.Exact(image.Pt(46, 44)),
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Ops:         &ops,
	}
	drawControlGlyph(
		gtx,
		image.Pt(46, 44),
		controlGlyphClose,
		activeTheme.Palette.AccentForeground,
		activeTheme.Components.TitleBar,
	)
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 46, 44))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	if got := pixels.RGBAAt(20, 19); got.G < 200 || got.B < 200 {
		t.Fatalf("close hover glyph pixel = %v, want a light glyph", got)
	}
}

func layoutTitleBarFrame(ctx *frame.Context, router *input.Router, bar Widget, now time.Time, size image.Point) layout.Dimensions {
	return layoutTitleBarWithConstraints(ctx, router, bar, now, layout.Exact(size))
}

func layoutTitleBarFlexibleFrame(ctx *frame.Context, router *input.Router, bar Widget, now time.Time, size image.Point) layout.Dimensions {
	return layoutTitleBarWithConstraints(ctx, router, bar, now, layout.Constraints{Max: size})
}

func layoutTitleBarOverlayFrame(ctx *frame.Context, router *input.Router, bar Widget, now time.Time, size image.Point) {
	var ops op.Ops
	viewport := image.Pt(size.X, 360)
	gtx := layout.Context{
		Constraints: layout.Exact(viewport),
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	barGtx := gtx
	barGtx.Constraints = layout.Exact(size)
	frame.BeginFrameWithViewport(ctx, viewport)
	bar.Layout(ctx, barGtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
}

func layoutTitleBarWithConstraints(ctx *frame.Context, router *input.Router, bar Widget, now time.Time, constraints layout.Constraints) layout.Dimensions {
	var ops op.Ops
	gtx := layout.Context{
		Constraints: constraints,
		Source:      router.Source(),
		Ops:         &ops,
		Now:         now,
	}
	frame.BeginFrame(ctx)
	dims := bar.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

type fixedMenu struct {
	size image.Point
}

func (m fixedMenu) Layout(*frame.Context, layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: m.size}
}

type constrainedFixed struct {
	size image.Point
}

func (w constrainedFixed) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

func hasSemanticLabel(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || hasSemanticLabel(node.Children, label) {
			return true
		}
	}
	return false
}
