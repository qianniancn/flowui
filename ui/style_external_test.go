package ui_test

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
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/uitest"
)

func ExampleLinearGradient() {
	gradient := ui.LinearGradient(
		ui.ColorStop(0, ui.RGB(0x2563eb)),
		ui.ColorStop(1, ui.TokenAccent),
	).Angle(90)
	_ = ui.Surface(ui.Text("Status")).Style(ui.Background(gradient))
}

func TestResolveStyleSupportsCustomWidgets(t *testing.T) {
	activeTheme := ui.DefaultTheme()
	activeTheme.Palette.Surface = color.NRGBA{R: 0x12, G: 0x34, B: 0x56, A: 0xff}
	harness := uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 40), Theme: &activeTheme})

	base := ui.Padding(4).Background(ui.TokenBackground)
	scope := ui.PaddingX(12).Background(ui.TokenSurface)
	instance := ui.PaddingLeft(20).When(ui.Hovered, ui.Opacity(.8))
	var resolved ui.ResolvedStyle
	custom := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		resolved = ui.ResolveStyle(ctx, gtx, "custom", ui.StyleState{Hovered: true}, base, instance)
		return layout.Dimensions{}
	})

	harness.Frame(ui.StyleScope(scope, custom))
	background, ok := resolved.Paint.Background.(ui.SolidColor)
	if !ok || background.Color != activeTheme.Palette.Surface {
		t.Fatalf("background = %#v, want theme surface", resolved.Paint.Background)
	}
	if resolved.Box.Padding.Top != 4 || resolved.Box.Padding.Right != 12 || resolved.Box.Padding.Left != 20 {
		t.Fatalf("padding = %#v, want base/scope/instance cascade", resolved.Box.Padding)
	}
	if resolved.Paint.Opacity == nil || *resolved.Paint.Opacity != .8 {
		t.Fatalf("hover opacity = %#v, want 0.8", resolved.Paint.Opacity)
	}
}

func TestPublicColorConstructors(t *testing.T) {
	if got := ui.RGB(0x9333ea).Color; got != (color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xff}) {
		t.Fatalf("RGB = %#v", got)
	}
	if got := ui.RGBA(0x9333eacc).Color; got != (color.NRGBA{R: 0x93, G: 0x33, B: 0xea, A: 0xcc}) {
		t.Fatalf("RGBA = %#v", got)
	}
	want := color.NRGBA{R: 1, G: 2, B: 3, A: 4}
	if got := ui.Color(want).Color; got != want {
		t.Fatalf("Color = %#v", got)
	}
	if got := ui.Color(color.Gray{Y: 0x7f}).Color; got != (color.NRGBA{R: 0x7f, G: 0x7f, B: 0x7f, A: 0xff}) {
		t.Fatalf("Color(gray) = %#v", got)
	}
}

func TestResolveStylePartSupportsCustomWidgets(t *testing.T) {
	const thumb ui.StylePart = ui.PartThumb
	if ui.PartTrack != "track" {
		t.Fatalf("PartTrack = %q", ui.PartTrack)
	}
	activeTheme := ui.DefaultTheme()
	activeTheme.Palette.Accent = color.NRGBA{R: 1, A: 0xff}
	harness := uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 40), Theme: &activeTheme})
	base := ui.TextColor(ui.TokenForeground)
	instance := ui.Part(thumb, ui.Background(ui.TokenAccent))
	var resolved ui.ResolvedStyle
	custom := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		resolved = ui.ResolveStylePartStatic(ctx, thumb, ui.StyleState{}, base, instance)
		return layout.Dimensions{}
	})

	harness.Frame(custom)
	background, ok := resolved.Paint.Background.(ui.SolidColor)
	if !ok || background.Color != activeTheme.Palette.Accent {
		t.Fatalf("part background = %#v", resolved.Paint.Background)
	}
	if resolved.Text == nil || resolved.Text.Color == nil {
		t.Fatalf("part did not inherit text: %#v", resolved.Text)
	}
}

func TestPublicConditionCompositionAndClearing(t *testing.T) {
	declaration := ui.Background(ui.TokenSurface).
		BoxShadow(0, 2, 4, 0, ui.TokenSurfaceShadow).
		When(ui.All(ui.Hovered, ui.Not(ui.Disabled)), ui.BackgroundNone().BoxShadowNone())

	resolved := ui.Cascade(ui.StyleState{Hovered: true}, declaration)
	if resolved.Paint == nil || resolved.Paint.Background != nil || len(resolved.Paint.Shadows) != 0 {
		t.Fatalf("public clearing = %#v", resolved.Paint)
	}
}

func TestPublicMVUConditionAndResolvedRenderer(t *testing.T) {
	harness := uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 40)})
	base := ui.Width(40).
		Height(20).
		Cursor(ui.CursorPointer).
		When(ui.If(true), ui.Opacity(.5))

	var dims layout.Dimensions
	harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		resolved := ui.ResolveStyleStatic(ctx, ui.StyleState{}, base, ui.Style{})
		gtx.Constraints.Min = image.Point{}
		dims = ui.LayoutResolvedStyle(ctx, gtx, resolved, nil)
		return dims
	}))
	if dims.Size != image.Pt(40, 20) {
		t.Fatalf("resolved renderer size = %v, want (40,20)", dims.Size)
	}
}

func TestPublicThemeShadowProfile(t *testing.T) {
	resolved := ui.Shadow(ui.ShadowSurface).Resolve(ui.StyleState{})
	if resolved.Paint == nil || len(resolved.Paint.Shadows) != 1 || resolved.Paint.Shadows[0].Profile == nil {
		t.Fatalf("theme shadow declaration = %#v", resolved.Paint)
	}
}

func TestPublicThemeMetricTokens(t *testing.T) {
	harness := uitest.NewWithConfig(uitest.Config{Size: image.Pt(100, 40)})
	declaration := ui.Use(ui.TokenControlHeight, ui.TokenControlRadius, ui.TokenControlFontSize)
	var resolved ui.ResolvedStyle
	harness.Frame(ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		resolved = ui.ResolveStyleStatic(ctx, ui.StyleState{}, declaration, ui.Style{})
		return layout.Dimensions{}
	}))
	if resolved.Box == nil || resolved.Box.Height == nil || resolved.Paint == nil || resolved.Paint.Radius == nil || resolved.Text == nil || resolved.Text.FontSize == nil {
		t.Fatalf("metric token resolution = %#v", resolved)
	}
}

func TestInteractiveStyleTransformMovesVisualBounds(t *testing.T) {
	harness := uitest.New(image.Pt(200, 100))
	button := ui.Button("lift", ui.Text("Lift")).
		Label("Lift").
		Style(
			ui.Width(92).Height(36).
				Transition(ui.PropTransform, time.Millisecond).
				When(ui.Hovered, ui.Translate(0, -3).Scale(1.04, 1.04)),
		)
	root := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Min = image.Point{}
			return button.Layout(ctx, gtx)
		})
	})

	harness.Frame(root)
	before, ok := semanticBounds(harness.Router().AppendSemantics(nil), semantic.Button, "Lift")
	if !ok {
		t.Fatal("missing button semantics before transform")
	}

	harness.Router().Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, Position: f32.Pt(100, 50)})
	harness.Frame(root)
	harness.Advance(2 * time.Millisecond)
	harness.Frame(root)
	after, ok := semanticBounds(harness.Router().AppendSemantics(nil), semantic.Button, "Lift")
	if !ok {
		t.Fatal("missing button semantics after transform")
	}
	if after.Min.Y >= before.Min.Y || after.Dx() <= before.Dx() || after.Dy() <= before.Dy() {
		t.Fatalf("transformed bounds = %v, want above and larger than %v", after, before)
	}
}

func TestPublicStyleFluentIsImmutable(t *testing.T) {
	base := ui.Padding(4).Background(ui.TokenSurface)
	changed := base.
		PaddingLeft(20).
		When(ui.Hovered, ui.Opacity(.8)).
		Part(ui.PartLabel, ui.FontWeight(600))

	baseResolved := ui.Cascade(ui.StyleState{Hovered: true}, base)
	if baseResolved.Box.Padding.Left != 4 || baseResolved.Paint.Opacity != nil {
		t.Fatalf("base style was mutated: %#v", baseResolved)
	}

	changedResolved := ui.Cascade(ui.StyleState{Hovered: true}, changed)
	if changedResolved.Box.Padding.Left != 20 || changedResolved.Paint.Opacity == nil || *changedResolved.Paint.Opacity != .8 {
		t.Fatalf("changed style = %#v", changedResolved)
	}
	label := ui.CascadePart(ui.StyleState{}, ui.PartLabel, changed)
	if label.Text == nil || label.Text.FontWeight == nil || *label.Text.FontWeight != 600 {
		t.Fatalf("label part = %#v", label)
	}
}

func ExampleStyleScope() {
	primary := ui.Background(ui.TokenAccent).
		TextColor(ui.TokenAccentForeground).
		Radius(8).
		Transition(ui.PropBackgroundColor, 150*time.Millisecond).
		When(ui.Hovered,
			ui.Background(ui.TokenAccentHover),
		).
		When(ui.Pressed,
			ui.Background(ui.TokenAccentPressed).
				Scale(0.95, 0.95),
		)

	_ = ui.StyleScope(
		ui.Radius(6),
		ui.Button("save", ui.Text("Save")).Style(primary),
	)
}

func semanticBounds(nodes []input.SemanticNode, class semantic.ClassOp, label string) (image.Rectangle, bool) {
	for _, node := range nodes {
		if node.Desc.Class == class && node.Desc.Label == label {
			return node.Desc.Bounds, true
		}
		if bounds, ok := semanticBounds(node.Children, class, label); ok {
			return bounds, true
		}
	}
	return image.Rectangle{}, false
}
