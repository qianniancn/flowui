package card

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/components/description"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

func TestCardVariantsMapToSemanticSurfaces(t *testing.T) {
	tests := []struct {
		variant CardVariant
		want    uint8
	}{
		{variant: CardDefault, want: 0},
		{variant: CardSecondary, want: 1},
		{variant: CardTertiary, want: 2},
		{variant: CardTransparent, want: 3},
		{variant: CardVariant(255), want: 0},
	}
	for _, test := range tests {
		if got := uint8(cardSurfaceVariant(test.variant)); got != test.want {
			t.Fatalf("surface variant for %d = %d, want %d", test.variant, got, test.want)
		}
	}
}

func TestCardUsesHeroUIDefaultSpacing(t *testing.T) {
	ctx := cardTestContext(nil)
	var ops op.Ops
	dims := Card(
		fixedWidget{size: image.Pt(40, 10)},
		fixedWidget{size: image.Pt(40, 10)},
	).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if dims.Size != image.Pt(72, 64) {
		t.Fatalf("card size = %v, want (72,64)", dims.Size)
	}
}

func TestCardThemeControlsSpacing(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	activeTheme.Components.Card.Padding = 20
	activeTheme.Components.Card.Gap = 6
	ctx := cardTestContext(&activeTheme)
	var ops op.Ops
	dims := Card(
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

func TestCardInstanceStyleOverridesComponentDefaults(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := cardTestContext(&activeTheme)
	background := color.NRGBA{R: 0x17, G: 0x72, B: 0x45, A: 0xff}
	foreground := color.NRGBA{R: 0xf1, G: 0xf2, B: 0xf3, A: 0xff}
	probe := &colorProbeWidget{}
	var ops op.Ops
	dims := Card(probe).
		Style(flowstyle.Style{}.
			Padding(0).
			Radius(0).
			Background(flowstyle.SolidColor{Color: background}).
			TextColor(flowstyle.SolidColor{Color: foreground}),
		).
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
	activeTheme := theme.DefaultTheme()
	activeTheme.Palette.SurfaceSecondary = color.NRGBA{R: 1, A: 0xff}
	activeTheme.Palette.SurfaceSecondaryForeground = color.NRGBA{G: 2, A: 0xff}
	ctx := cardTestContext(&activeTheme)
	probe := &colorProbeWidget{}
	var ops op.Ops

	Card(probe).Variant(CardSecondary).Layout(ctx, layout.Context{
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
	activeTheme := theme.DefaultTheme()
	if shadows := cardRootDeclaration(&activeTheme, CardDefault).Resolve(flowstyle.StyleState{}).Paint.Shadows; len(shadows) != 1 || shadows[0].Profile == nil {
		t.Fatal("default card should use surface elevation")
	}
	if shadows := cardRootDeclaration(&activeTheme, CardTransparent).Resolve(flowstyle.StyleState{}).Paint.Shadows; len(shadows) != 0 {
		t.Fatal("transparent card should not use surface elevation")
	}
}

func TestCardIgnoresNilChildren(t *testing.T) {
	ctx := cardTestContext(nil)
	var ops op.Ops
	dims := Card(nil, fixedWidget{size: image.Pt(40, 10)}, nil).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})
	if dims.Size != image.Pt(72, 42) {
		t.Fatalf("card size with nil children = %v, want (72,42)", dims.Size)
	}
}

func TestCardTracksOverlayThroughPadding(t *testing.T) {
	ctx := cardTestContext(nil)
	viewport := image.Pt(300, 200)
	frame.BeginFrameWithViewport(ctx, viewport)
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: viewport},
		Ops:         new(op.Ops),
	}
	var anchor image.Rectangle
	Card(&cardOverlayProbe{anchor: &anchor}).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	want := image.Rect(16, 16, 26, 26)
	if anchor != want {
		t.Fatalf("overlay anchor = %v, want %v", anchor, want)
	}
}

func TestCardPreparesFieldAssociationsAcrossSections(t *testing.T) {
	ctx := cardTestContext(nil)
	frame.BeginFrame(ctx)
	probe := &descriptionProbeWidget{key: "account"}
	var ops op.Ops

	Card(
		probe,
		description.Description("Account help").For("account"),
	).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if probe.description != "Account help" {
		t.Fatalf("description during content layout = %q, want %q", probe.description, "Account help")
	}
}

func cardTestContext(activeTheme *theme.Theme) *frame.Context {
	return frame.New(nil, activeTheme, locale.LanguageAuto)
}

type fixedWidget struct {
	size image.Point
}

func (w fixedWidget) Layout(_ *frame.Context, gtx layout.Context) layout.Dimensions {
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

func (w *descriptionProbeWidget) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	w.description = frame.FieldDescription(ctx, frame.FullKey(ctx, w.key))
	return layout.Dimensions{Size: image.Pt(20, 10)}
}

func (w *cardOverlayProbe) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
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

func (w *colorProbeWidget) Layout(ctx *frame.Context, _ layout.Context) layout.Dimensions {
	w.foreground = ctx.ForegroundColor()
	w.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: image.Pt(40, 10)}
}
