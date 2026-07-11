package surface

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
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
	if got != want {
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

	dims := Surface(probe).Variant(SurfaceSecondary).Radius(16).Shadow(true).Layout(ctx, layout.Context{
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

func TestSurfaceOptionsKeepValueSemantics(t *testing.T) {
	base := Surface(nil)
	gradient := render.LinearGradient(
		render.GradientStop{Color: color.NRGBA{R: 255, A: 255}},
		render.GradientStop{Offset: 1, Color: color.NRGBA{B: 255, A: 255}},
	)
	foreground := color.NRGBA{G: 255, A: 255}
	styled := base.Variant(SurfaceTertiary).Radius(20).Shadow(true).Background(gradient).Foreground(foreground)
	if base.variant != SurfaceDefault || base.radius != 0 || base.shadow || base.hasBackground || base.hasForeground {
		t.Fatal("surface options mutated the original value")
	}
	if styled.variant != SurfaceTertiary || styled.radius != 20 || !styled.shadow || !styled.hasBackground || !styled.hasForeground || styled.foreground != foreground {
		t.Fatal("surface options did not configure the returned value")
	}
	if got := base.Radius(-10).radius; got != 0 {
		t.Fatalf("negative radius = %v, want 0", got)
	}
}

func TestGradientSurfaceScopesSampledBackgroundAndForeground(t *testing.T) {
	activeTheme := DefaultTheme()
	ctx := newContextWithTheme(nil, &activeTheme)
	probe := &surfaceProbeWidget{size: image.Pt(120, 48)}
	foreground := color.NRGBA{R: 250, G: 250, B: 250, A: 255}
	gradient := render.LinearGradient(
		render.GradientStop{Color: color.NRGBA{R: 255, A: 255}},
		render.GradientStop{Offset: 1, Color: color.NRGBA{B: 255, A: 255}},
	)
	var ops op.Ops

	Surface(probe).Background(gradient).Foreground(foreground).Layout(ctx, layout.Context{
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

func TestNewPaletteTokensFallBackForExistingCustomThemes(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.SurfaceForeground = color.NRGBA{}
	theme.Palette.SurfaceSecondary = color.NRGBA{}
	theme.Palette.SurfaceSecondaryForeground = color.NRGBA{}
	theme.Palette.SurfaceTertiary = color.NRGBA{}
	theme.Palette.SurfaceTertiaryForeground = color.NRGBA{}
	if got := surfaceStyleFor(&theme, SurfaceSecondary); got.background != theme.Palette.SurfaceRaised || got.foreground != theme.Palette.Foreground {
		t.Fatalf("secondary fallback style = %#v", got)
	}
	if got := surfaceStyleFor(&theme, SurfaceTertiary); got.background != theme.Palette.SurfacePressed || got.foreground != theme.Palette.Foreground {
		t.Fatalf("tertiary fallback style = %#v", got)
	}
}

func TestSurfaceShadowUsesRestrainedElevation(t *testing.T) {
	layers := render.SurfaceShadow(color.NRGBA{A: 100}).EffectiveLayers()
	if len(layers) != 2 {
		t.Fatalf("surface shadow layers = %d, want 2", len(layers))
	}
	if layers[1].Blur >= render.PopupShadow(color.NRGBA{A: 100}).EffectiveLayers()[1].Blur {
		t.Fatal("surface shadow should be tighter than overlay shadow")
	}
}

func TestDarkThemeDisablesSurfaceElevation(t *testing.T) {
	theme := DarkTheme()
	if layers := render.SurfaceShadow(theme.Palette.SurfaceShadow).EffectiveLayers(); len(layers) != 0 {
		t.Fatalf("dark surface shadow layers = %d, want 0", len(layers))
	}
	if theme.Palette.OverlayShadow.A == 0 {
		t.Fatal("dark overlay shadow should retain separation from surrounding content")
	}
}

type surfaceProbeWidget struct {
	size       image.Point
	foreground color.NRGBA
	background color.NRGBA
}

func (w *surfaceProbeWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	w.foreground = ctx.ForegroundColor()
	w.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}
