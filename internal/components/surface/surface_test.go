package surface

import (
	"image"
	"image/color"
	"reflect"
	"testing"
	"time"

	"gioui.org/gpu/headless"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/render"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

func newContextWithTheme(_ any, value *theme.Theme) *frame.Context {
	return frame.New(nil, value, locale.LanguageAuto)
}

func DefaultTheme() theme.Theme {
	return theme.DefaultTheme()
}

func DarkTheme() theme.Theme {
	return theme.DarkTheme()
}

func TestSurfaceVariantsUseSemanticPalette(t *testing.T) {
	theme := DefaultTheme()
	tests := []struct {
		name       string
		variant    SurfaceVariant
		background color.NRGBA
		foreground color.NRGBA
	}{
		{name: "default", variant: SurfaceDefault, background: theme.Palette.Surface, foreground: theme.Palette.SurfaceForeground},
		{name: "secondary", variant: SurfaceSecondary, background: theme.Palette.SurfaceSecondary, foreground: theme.Palette.SurfaceSecondaryForeground},
		{name: "tertiary", variant: SurfaceTertiary, background: theme.Palette.SurfaceTertiary, foreground: theme.Palette.SurfaceTertiaryForeground},
		{name: "transparent", variant: SurfaceTransparent, foreground: theme.Palette.Foreground},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			style := surfaceStyleFor(&theme, test.variant)
			if style.background != test.background || style.foreground != test.foreground {
				t.Fatalf("surface style = %#v, want background %#v foreground %#v", style, test.background, test.foreground)
			}
		})
	}
}

func TestSurfaceUnknownVariantFallsBackToDefault(t *testing.T) {
	theme := DefaultTheme()
	got := surfaceStyleFor(&theme, SurfaceVariant(255))
	want := surfaceStyleFor(&theme, SurfaceDefault)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unknown variant style = %#v, want default %#v", got, want)
	}
}

func TestSurfaceLayoutPreservesChildSizeAndScopesForeground(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.Foreground = color.NRGBA{R: 1, A: 0xff}
	theme.Palette.SurfaceSecondaryForeground = color.NRGBA{G: 2, A: 0xff}
	ctx := newContextWithTheme(nil, &theme)
	probe := &surfaceProbeWidget{size: image.Pt(120, 48)}
	var ops op.Ops

	dims := Surface(probe).Variant(SurfaceSecondary).Style(flowstyle.Style{}.Radius(16).Shadow(flowstyle.ShadowSurface)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if dims.Size != probe.size {
		t.Fatalf("surface size = %v, want child size %v", dims.Size, probe.size)
	}
	if probe.foreground != theme.Palette.SurfaceSecondaryForeground {
		t.Fatalf("child foreground = %#v, want %#v", probe.foreground, theme.Palette.SurfaceSecondaryForeground)
	}
	if probe.background != theme.Palette.SurfaceSecondary {
		t.Fatalf("child background = %#v, want %#v", probe.background, theme.Palette.SurfaceSecondary)
	}
	if got := ctx.ForegroundColor(); got != theme.Palette.Foreground {
		t.Fatalf("foreground after layout = %#v, want restored %#v", got, theme.Palette.Foreground)
	}
	if got := ctx.BackgroundColor(); got != theme.Palette.Background {
		t.Fatalf("background after layout = %#v, want restored %#v", got, theme.Palette.Background)
	}
}

func TestSurfaceStyleScopeAndInstancePrecedence(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := newContextWithTheme(nil, &activeTheme)
	probe := &surfaceProbeWidget{size: image.Pt(40, 20)}
	gtx := layout.Context{Constraints: layout.Constraints{Max: image.Pt(100, 100)}, Ops: new(op.Ops)}
	scopeBackground := color.NRGBA{R: 1, A: 0xff}
	scopeForeground := color.NRGBA{G: 2, A: 0xff}
	instanceBackground := color.NRGBA{B: 3, A: 0xff}
	restore := frame.PushStyle(ctx, flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: scopeBackground}).
		TextColor(flowstyle.SolidColor{Color: scopeForeground}),
	)
	defer restore()

	Surface(probe).
		Style(flowstyle.Style{}.Background(flowstyle.SolidColor{Color: instanceBackground})).
		Layout(ctx, gtx)

	if probe.background != instanceBackground {
		t.Fatalf("surface background = %#v, want instance %#v", probe.background, instanceBackground)
	}
	if probe.foreground != scopeForeground {
		t.Fatalf("surface foreground = %#v, want scope %#v", probe.foreground, scopeForeground)
	}
}

func TestSurfaceConditionalStyleTransition(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := newContextWithTheme(nil, &activeTheme)
	start := time.Unix(1, 0)
	from := color.NRGBA{R: 0x10, A: 0xff}
	to := color.NRGBA{B: 0xf0, A: 0xff}
	layoutAt := func(now time.Time, active bool) color.NRGBA {
		probe := &surfaceProbeWidget{size: image.Pt(20, 20)}
		var ops op.Ops
		gtx := layout.Context{Constraints: layout.Exact(image.Pt(20, 20)), Ops: &ops, Now: now}
		frame.BeginFrame(ctx)
		Surface(probe).Style(flowstyle.Style{}.
			Background(flowstyle.SolidColor{Color: from}).
			Transition(flowstyle.PropBackgroundColor, 100*time.Millisecond).
			WhenIf(active, flowstyle.Style{}.Background(flowstyle.SolidColor{Color: to})),
		).Layout(ctx, gtx)
		frame.EndFrame(ctx)
		return probe.background
	}

	if got := layoutAt(start, false); got != from {
		t.Fatalf("initial background = %#v", got)
	}
	if got := layoutAt(start, true); got != from {
		t.Fatalf("transition start = %#v, want %#v", got, from)
	}
	middle := layoutAt(start.Add(50*time.Millisecond), true)
	if middle == from || middle == to {
		t.Fatalf("transition midpoint = %#v", middle)
	}
	if got := layoutAt(start.Add(100*time.Millisecond), true); got != to {
		t.Fatalf("transition end = %#v, want %#v", got, to)
	}
}

func TestNestedSurfaceUsesInnermostForeground(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.SurfaceSecondaryForeground = color.NRGBA{R: 3, A: 0xff}
	theme.Palette.SurfaceTertiaryForeground = color.NRGBA{B: 4, A: 0xff}
	ctx := newContextWithTheme(nil, &theme)
	probe := &surfaceProbeWidget{size: image.Pt(80, 32)}
	var ops op.Ops

	Surface(
		Surface(probe).Variant(SurfaceTertiary),
	).Variant(SurfaceSecondary).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 100)},
		Ops:         &ops,
	})

	if probe.foreground != theme.Palette.SurfaceTertiaryForeground {
		t.Fatalf("nested foreground = %#v, want innermost %#v", probe.foreground, theme.Palette.SurfaceTertiaryForeground)
	}
	if probe.background != theme.Palette.SurfaceTertiary {
		t.Fatalf("nested background = %#v, want innermost %#v", probe.background, theme.Palette.SurfaceTertiary)
	}
}

func TestTransparentSurfacePreservesParentBackground(t *testing.T) {
	theme := DefaultTheme()
	ctx := newContextWithTheme(nil, &theme)
	probe := &surfaceProbeWidget{size: image.Pt(80, 32)}
	var ops op.Ops

	Surface(
		Surface(probe).Variant(SurfaceTransparent),
	).Variant(SurfaceSecondary).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 100)},
		Ops:         &ops,
	})

	if probe.background != theme.Palette.SurfaceSecondary {
		t.Fatalf("transparent surface background = %#v, want inherited %#v", probe.background, theme.Palette.SurfaceSecondary)
	}
}

func TestSurfaceStyleUsesThemeBorder(t *testing.T) {
	activeTheme := DefaultTheme()
	activeTheme.Components.Surface.BorderWidth = 2
	activeTheme.Palette.Border = color.NRGBA{R: 12, G: 34, B: 56, A: 255}

	style := surfaceStyleFor(&activeTheme, SurfaceDefault)
	if style.borderWidth != 2 || style.border != activeTheme.Palette.Border {
		t.Fatalf("surface border = %v/%#v", style.borderWidth, style.border)
	}
}

func TestSurfaceBorderPaintsAboveChild(t *testing.T) {
	window, err := headless.NewWindow(40, 40)
	if err != nil {
		t.Skipf("headless renderer unavailable: %v", err)
	}
	defer window.Release()

	activeTheme := DefaultTheme()
	ctx := newContextWithTheme(nil, &activeTheme)
	var ops op.Ops
	childColor := color.NRGBA{G: 0xff, A: 0xff}
	borderColor := color.NRGBA{R: 0xff, A: 0xff}
	Surface(paintedSurfaceChild{color: childColor}).
		Style(flowstyle.Style{}.BorderWidth(3).BorderColor(flowstyle.SolidColor{Color: borderColor})).
		Layout(ctx, layout.Context{Constraints: layout.Exact(image.Pt(40, 40)), Ops: &ops})
	if err := window.Frame(&ops); err != nil {
		t.Fatal(err)
	}
	pixels := image.NewRGBA(image.Rect(0, 0, 40, 40))
	if err := window.Screenshot(pixels); err != nil {
		t.Fatal(err)
	}
	if got := pixels.RGBAAt(1, 20); got.R <= got.G {
		t.Fatalf("border pixel = %#v, want border above child", got)
	}
	if got := pixels.RGBAAt(20, 20); got.G <= got.R {
		t.Fatalf("center pixel = %#v, want child content", got)
	}
}

func TestGradientSurfaceScopesSampledBackgroundAndForeground(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := newContextWithTheme(nil, &activeTheme)
	probe := &surfaceProbeWidget{size: image.Pt(120, 48)}
	foreground := color.NRGBA{R: 250, G: 250, B: 250, A: 255}
	gradient := flowstyle.LinearGradient(
		flowstyle.ColorStop(0, flowstyle.SolidColor{Color: color.NRGBA{R: 255, A: 255}}),
		flowstyle.ColorStop(1, flowstyle.SolidColor{Color: color.NRGBA{B: 255, A: 255}}),
	)
	var ops op.Ops

	Surface(probe).Style(flowstyle.Style{}.Background(gradient).TextColor(flowstyle.SolidColor{Color: foreground})).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 100)},
		Ops:         &ops,
	})

	if probe.background != (color.NRGBA{R: 128, B: 128, A: 255}) {
		t.Fatalf("gradient context background = %#v", probe.background)
	}
	if probe.foreground != foreground {
		t.Fatalf("gradient foreground = %#v, want %#v", probe.foreground, foreground)
	}
}

func TestSurfaceAndOverlayTokensAreDistinct(t *testing.T) {
	theme := DefaultTheme()
	if theme.Palette.SurfaceSecondary == theme.Palette.Surface || theme.Palette.SurfaceTertiary == theme.Palette.Surface {
		t.Fatal("surface prominence levels should be visually distinct")
	}
	if theme.Palette.SurfaceShadow == theme.Palette.OverlayShadow {
		t.Fatal("surface and overlay shadows should have distinct elevation")
	}
}

func TestSurfacePaletteTokensHonorTransparentValues(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.SurfaceForeground = color.NRGBA{}
	theme.Palette.SurfaceSecondary = color.NRGBA{}
	theme.Palette.SurfaceSecondaryForeground = color.NRGBA{}
	theme.Palette.SurfaceTertiary = color.NRGBA{}
	theme.Palette.SurfaceTertiaryForeground = color.NRGBA{}
	if got := surfaceStyleFor(&theme, SurfaceSecondary); got.background.A != 0 || got.foreground.A != 0 {
		t.Fatalf("secondary style = %#v, want transparent colors", got)
	}
	if got := surfaceStyleFor(&theme, SurfaceTertiary); got.background.A != 0 || got.foreground.A != 0 {
		t.Fatalf("tertiary style = %#v, want transparent colors", got)
	}
}

func TestSurfaceShadowUsesRestrainedElevation(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	layers := render.ThemeShadow(activeTheme.Shadows.Surface, color.NRGBA{A: 100}, 1).EffectiveLayers()
	if len(layers) != 2 {
		t.Fatalf("surface shadow layers = %d, want 2", len(layers))
	}
	if layers[1].Blur >= render.ThemeShadow(activeTheme.Shadows.Overlay, color.NRGBA{A: 100}, 1).EffectiveLayers()[1].Blur {
		t.Fatal("surface shadow should be tighter than overlay shadow")
	}
}

func TestDarkThemeDisablesSurfaceElevation(t *testing.T) {
	activeTheme := DarkTheme()
	if layers := render.ThemeShadow(activeTheme.Shadows.Surface, activeTheme.Palette.SurfaceShadow, 1).EffectiveLayers(); len(layers) != 0 {
		t.Fatalf("dark surface shadow layers = %d, want 0", len(layers))
	}
	if activeTheme.Palette.OverlayShadow.A == 0 {
		t.Fatal("dark overlay shadow should retain separation from surrounding content")
	}
}

type surfaceProbeWidget struct {
	size       image.Point
	foreground color.NRGBA
	background color.NRGBA
	shadowBlur unit.Dp
}

type paintedSurfaceChild struct {
	color color.NRGBA
}

func (w paintedSurfaceChild) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Constrain(gtx.Constraints.Max)
	paint.FillShape(gtx.Ops, w.color, clip.Rect(image.Rectangle{Max: size}).Op())
	return layout.Dimensions{Size: size}
}

func (w *surfaceProbeWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	w.foreground = ctx.ForegroundColor()
	w.background = ctx.BackgroundColor()
	w.shadowBlur = frame.ActiveTheme(ctx).Shadows.Surface.Layers[1].Blur
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}
