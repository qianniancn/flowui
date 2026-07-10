package flowui

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
)

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
	if got := ctx.foregroundColor(); got != theme.Palette.Foreground {
		t.Fatalf("foreground after layout = %#v, want restored %#v", got, theme.Palette.Foreground)
	}
	if got := ctx.backgroundColor(); got != theme.Palette.Background {
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
	styled := base.Variant(SurfaceTertiary).Radius(20).Shadow(true)
	if base.variant != SurfaceDefault || base.radius != 0 || base.shadow {
		t.Fatal("surface options mutated the original value")
	}
	if styled.variant != SurfaceTertiary || styled.radius != 20 || !styled.shadow {
		t.Fatal("surface options did not configure the returned value")
	}
	if got := base.Radius(-10).radius; got != 0 {
		t.Fatalf("negative radius = %v, want 0", got)
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

	theme.Palette.Overlay = color.NRGBA{R: 9, A: 0xff}
	theme.Palette.OverlayForeground = color.NRGBA{G: 8, A: 0xff}
	style := popoverStyleFor(&theme)
	if style.surface != theme.Palette.Overlay || style.text != theme.Palette.OverlayForeground {
		t.Fatal("popover style does not use overlay tokens")
	}
}

func TestNewPaletteTokensFallBackForExistingCustomThemes(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.SurfaceForeground = color.NRGBA{}
	theme.Palette.SurfaceSecondary = color.NRGBA{}
	theme.Palette.SurfaceSecondaryForeground = color.NRGBA{}
	theme.Palette.SurfaceTertiary = color.NRGBA{}
	theme.Palette.SurfaceTertiaryForeground = color.NRGBA{}
	theme.Palette.Overlay = color.NRGBA{}
	theme.Palette.OverlayForeground = color.NRGBA{}
	theme.Palette.OverlayShadow = color.NRGBA{}

	if got := surfaceStyleFor(&theme, SurfaceSecondary); got.background != theme.Palette.SurfaceRaised || got.foreground != theme.Palette.Foreground {
		t.Fatalf("secondary fallback style = %#v", got)
	}
	if got := surfaceStyleFor(&theme, SurfaceTertiary); got.background != theme.Palette.SurfacePressed || got.foreground != theme.Palette.Foreground {
		t.Fatalf("tertiary fallback style = %#v", got)
	}
	if got := popoverStyleFor(&theme); got.surface != theme.Palette.Surface || got.text != theme.Palette.Foreground {
		t.Fatalf("overlay fallback style = %#v", got)
	}
	if got := theme.Palette.overlayShadowColor(); got != theme.Palette.Shadow {
		t.Fatalf("overlay shadow fallback = %#v, want %#v", got, theme.Palette.Shadow)
	}
}

func TestSurfaceShadowUsesRestrainedElevation(t *testing.T) {
	layers := SurfaceShadow(color.NRGBA{A: 100}).EffectiveLayers()
	if len(layers) != 2 {
		t.Fatalf("surface shadow layers = %d, want 2", len(layers))
	}
	if layers[1].Blur >= PopupShadow(color.NRGBA{A: 100}).EffectiveLayers()[1].Blur {
		t.Fatal("surface shadow should be tighter than overlay shadow")
	}
}

func TestDarkThemeDisablesSurfaceElevation(t *testing.T) {
	theme := DarkTheme()
	if layers := SurfaceShadow(theme.Palette.SurfaceShadow).EffectiveLayers(); len(layers) != 0 {
		t.Fatalf("dark surface shadow layers = %d, want 0", len(layers))
	}
	if theme.Palette.OverlayShadow.A == 0 {
		t.Fatal("dark overlay shadow should retain separation from surrounding content")
	}
}

func TestPopoverScopesOverlayForegroundToContent(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.OverlayForeground = color.NRGBA{R: 7, G: 8, A: 0xff}
	ctx := newContextWithTheme(nil, &theme)
	probe := &surfaceProbeWidget{size: image.Pt(40, 20)}
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(200, 120)},
		Ops:         &ops,
	}

	Popover("help", true, Spacer(20, 20), probe).layoutPanel(ctx, gtx, PopoverBottom)

	if probe.foreground != theme.Palette.OverlayForeground {
		t.Fatalf("popover content foreground = %#v, want overlay foreground %#v", probe.foreground, theme.Palette.OverlayForeground)
	}
	if probe.background != theme.Palette.Overlay {
		t.Fatalf("popover content background = %#v, want overlay background %#v", probe.background, theme.Palette.Overlay)
	}
}

func TestModalScopesOverlayForegroundToContent(t *testing.T) {
	theme := DefaultTheme()
	theme.Palette.OverlayForeground = color.NRGBA{B: 9, A: 0xff}
	ctx := newContextWithTheme(nil, &theme)
	probe := &surfaceProbeWidget{size: image.Pt(40, 20)}
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(240, 160)},
		Ops:         &ops,
	}

	Modal("dialog", true, "", probe).CloseButton(false).layoutDialogSurface(ctx, gtx, new(modalState))

	if probe.foreground != theme.Palette.OverlayForeground {
		t.Fatalf("modal content foreground = %#v, want overlay foreground %#v", probe.foreground, theme.Palette.OverlayForeground)
	}
	if probe.background != theme.Palette.Overlay {
		t.Fatalf("modal content background = %#v, want overlay background %#v", probe.background, theme.Palette.Overlay)
	}
}

type surfaceProbeWidget struct {
	size       image.Point
	foreground color.NRGBA
	background color.NRGBA
}

func (w *surfaceProbeWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	w.foreground = ctx.foregroundColor()
	w.background = ctx.backgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}
