package card

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/description"
)

func TestCardThemeControlsSpacing(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Components.Card.Padding = 20
	activeTheme.Components.Card.Gap = 6
	ctx := cardTestContext(&activeTheme)
	var ops op.Ops
	dims := New(
		fixedWidget{size: image.Pt(40, 10)},
		fixedWidget{size: image.Pt(40, 10)},
	).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(80, 66) {
		t.Fatalf("themed card size = %v, want (80,66)", dims.Size)
	}
}

func TestCardStyledMetricsOverrideUsesProviderBoundary(t *testing.T) {
	ctx := cardTestContext(testtheme.Behavior())
	var ops op.Ops
	dims := (New(
		fixedWidget{size: image.Pt(40, 10)},
		fixedWidget{size: image.Pt(40, 10)},
	)).StyleWithMetrics(style.Style{}, unit.Dp(3), unit.Dp(5)).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})
	if dims.Size != image.Pt(46, 31) {
		t.Fatalf("styled card metrics = %v, want (46,31)", dims.Size)
	}
}

func TestCardInstanceStyleOverridesComponentDefaults(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	ctx := cardTestContext(&activeTheme)
	background := color.NRGBA{R: 0x17, G: 0x72, B: 0x45, A: 0xff}
	foreground := color.NRGBA{R: 0xf1, G: 0xf2, B: 0xf3, A: 0xff}
	probe := &colorProbeWidget{}
	var ops op.Ops
	dims := (New(probe)).Style(style.Style{}.
		Padding(0).
		Radius(0).
		Background(style.SolidColor{Color: background}).
		TextColor(style.SolidColor{Color: foreground})).
		Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 200)},
			Ops:         &ops,
		})

	if dims.Size != image.Pt(40, 10) {
		t.Fatalf("styled card size = %v, want (40,10)", dims.Size)
	}
	if probe.background != background || probe.foreground != foreground {
		t.Fatalf("card colors = %#v/%#v, want %#v/%#v", probe.background, probe.foreground, background, foreground)
	}
	if activeTheme.Components.Card.Padding != 16 || activeTheme.Components.Card.Radius != 24 {
		t.Fatalf("instance style mutated application card theme: %#v", activeTheme.Components.Card)
	}
}

func TestCardScopesVariantColors(t *testing.T) {
	activeTheme := theme.NeutralTheme()
	activeTheme.Palette.SurfaceSecondary = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.SurfaceSecondaryForeground = color.NRGBA{G: 2, A: 0xff}
	activeTheme = paletteTheme(activeTheme)
	ctx := cardTestContext(&activeTheme)
	probe := &colorProbeWidget{}
	var ops op.Ops

	New(probe).Variant(Secondary).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if probe.background != activeTheme.Palette.SurfaceSecondary {
		t.Fatalf("card background = %#v, want %#v", probe.background, activeTheme.Palette.SurfaceSecondary)
	}
	if probe.foreground != activeTheme.Palette.SurfaceSecondaryForeground {
		t.Fatalf("card foreground = %#v, want %#v", probe.foreground, activeTheme.Palette.SurfaceSecondaryForeground)
	}
}

func TestCardDefaultsShadowExceptForTransparentVariant(t *testing.T) {
	activeTheme := *testtheme.Behavior()
	provider, err := theme.ComponentProvider(&activeTheme)
	if err != nil {
		t.Fatal(err)
	}
	defaultVisual := provider.Card(&activeTheme, theme.CardRequest{Variant: theme.CardDefault})
	if shadows := style.Cascade(style.State{}, defaultVisual.Defaults, defaultVisual.Variant, defaultVisual.Override).Paint.Shadows; len(shadows) != 1 || shadows[0].Profile == nil {
		t.Fatal("default card should use surface elevation")
	}
	transparentVisual := provider.Card(&activeTheme, theme.CardRequest{Variant: theme.CardTransparent})
	if shadows := style.Cascade(style.State{}, transparentVisual.Defaults, transparentVisual.Variant, transparentVisual.Override).Paint.Shadows; len(shadows) != 0 {
		t.Fatal("transparent card should not use surface elevation")
	}
}

func TestCardIgnoresNilChildren(t *testing.T) {
	ctx := cardTestContext(testtheme.Behavior())
	var ops op.Ops
	dims := New(nil, fixedWidget{size: image.Pt(40, 10)}, nil).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})
	if dims.Size != image.Pt(72, 42) {
		t.Fatalf("card size with nil children = %v, want (72,42)", dims.Size)
	}
}

func TestCardTracksOverlayThroughPadding(t *testing.T) {
	ctx := cardTestContext(testtheme.Behavior())
	viewport := image.Pt(300, 200)
	core.BeginFrameWithViewport(ctx, viewport)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Ops:         new(op.Ops),
	}
	var anchor image.Rectangle
	New(&cardOverlayProbe{anchor: &anchor}).Layout(ctx, gtx)
	core.LayoutOverlays(ctx, gtx)
	core.EndFrame(ctx)

	want := image.Rect(16, 16, 26, 26)
	if anchor != want {
		t.Fatalf("overlay anchor = %v, want %v", anchor, want)
	}
}

func TestCardPreparesFieldAssociationsAcrossSections(t *testing.T) {
	ctx := cardTestContext(testtheme.Behavior())
	core.BeginFrame(ctx)
	probe := &descriptionProbeWidget{key: "account"}
	var ops op.Ops

	New(
		probe,
		description.New("Account help").For("account"),
	).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if probe.description != "Account help" {
		t.Fatalf("description during content layout = %q, want %q", probe.description, "Account help")
	}
}

func cardTestContext(activeTheme *theme.Theme) *core.Context {
	clone := *activeTheme
	if clone.VisualProvider == nil {
		clone.VisualProvider = testtheme.BehaviorProvider()
	}
	return core.New(nil, &clone, locale.LanguageAuto)
}

type fixedWidget struct {
	size image.Point
}

func (w fixedWidget) Layout(_ *core.Context, gtx layout.Context) layout.Dimensions {
	return layout.Dimensions{Size: gtx.Constraints.Constrain(w.size)}
}

type colorProbeWidget struct {
	foreground color.NRGBA
	background color.NRGBA
}

type cardOverlayProbe struct {
	anchor *image.Rectangle
}

type descriptionProbeWidget struct {
	key         string
	description string
}

func (w *descriptionProbeWidget) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	w.description = core.FieldDescription(ctx, core.FullKey(ctx, w.key))
	return layout.Dimensions{Size: image.Pt(20, 10)}
}

func (w *cardOverlayProbe) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	core.RegisterOverlay(ctx, core.OverlayRequest{
		Key:       "card-overlay-probe",
		Anchor:    image.Rect(0, 0, 10, 10),
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			*w.anchor = anchor
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: image.Pt(20, 10)}
}

// paletteTheme installs a theme's v1 palette as its colour source, so edits to
// the palette reach the widgets: a widget reads the semantic layer, and a theme
// built from a Seed carries one that no longer follows the palette.
func paletteTheme(t theme.Theme) theme.Theme {
	t.UseSemantic(theme.SemanticFromPalette(t.Palette))
	return t
}

func (w *colorProbeWidget) Layout(ctx *core.Context, _ layout.Context) layout.Dimensions {
	w.foreground = ctx.ForegroundColor()
	w.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: image.Pt(40, 10)}
}

func TestDynamicDpConfigurationClampsNegativeValues(t *testing.T) {
	c := (New()).StyleWithMetrics(style.Style{}, unit.Dp(-1.5), unit.Dp(2.5))
	if c.override.Padding == nil || *c.override.Padding != 0 {
		t.Fatalf("negative padding = %v, want 0", c.override.Padding)
	}
	if c.override.Gap == nil || *c.override.Gap != unit.Dp(2.5) {
		t.Fatalf("gap = %v, want 2.5", c.override.Gap)
	}
}
