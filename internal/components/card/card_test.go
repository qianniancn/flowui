package card

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/description"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
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

func TestCardSlotLayoutsMatchHeroUI(t *testing.T) {
	ctx := cardTestContext(nil)
	tests := []struct {
		name   string
		widget frame.Widget
		want   image.Point
	}{
		{
			name: "header column without gap",
			widget: CardHeader(
				fixedWidget{size: image.Pt(20, 8)},
				fixedWidget{size: image.Pt(30, 10)},
			),
			want: image.Pt(30, 18),
		},
		{
			name: "content column with four dp gap",
			widget: CardContent(
				fixedWidget{size: image.Pt(20, 8)},
				fixedWidget{size: image.Pt(30, 10)},
			),
			want: image.Pt(30, 22),
		},
		{
			name: "footer row aligned in the middle",
			widget: CardFooter(
				fixedWidget{size: image.Pt(20, 8)},
				fixedWidget{size: image.Pt(30, 10)},
			),
			want: image.Pt(50, 10),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ops op.Ops
			dims := test.widget.Layout(ctx, layout.Context{
				Constraints: layout.Constraints{Max: image.Pt(300, 200)},
				Ops:         &ops,
			})
			if dims.Size != test.want {
				t.Fatalf("slot size = %v, want %v", dims.Size, test.want)
			}
		})
	}
}

func TestCardDefaultsShadowExceptForTransparentVariant(t *testing.T) {
	if !Card().resolvedShadow() {
		t.Fatal("default card should use surface elevation")
	}
	if Card().Variant(CardTransparent).resolvedShadow() {
		t.Fatal("transparent card should not use surface elevation")
	}
	if !Card().Variant(CardTransparent).Shadow(true).resolvedShadow() {
		t.Fatal("explicit shadow option should override the transparent default")
	}
}

func TestCardOptionsKeepValueSemantics(t *testing.T) {
	base := Card()
	styled := base.
		Variant(CardTertiary).
		Padding(20).
		Gap(8).
		Radius(16).
		Shadow(false)

	if base.variant != CardDefault || base.hasPadding || base.hasGap || base.hasRadius || base.hasShadow {
		t.Fatal("card options mutated the original value")
	}
	if styled.variant != CardTertiary || styled.padding != 20 || styled.gap != 8 || styled.radius != 16 {
		t.Fatal("card options did not configure the returned value")
	}
	if !styled.hasShadow || styled.shadow {
		t.Fatal("explicit shadow option was not retained")
	}
	if got := base.Padding(-1).padding; got != 0 {
		t.Fatalf("negative padding = %v, want 0", got)
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
		CardContent(probe),
		CardFooter(description.Description("Account help").For("account")),
	).Layout(ctx, layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(300, 200)},
		Ops:         &ops,
	})

	if probe.description != "Account help" {
		t.Fatalf("description during content layout = %q, want %q", probe.description, "Account help")
	}
}

func TestCardSlotOptionsKeepValueSemantics(t *testing.T) {
	header := CardHeader()
	content := CardContent()
	footer := CardFooter()
	if got := header.Gap(3); header.hasGap || !got.hasGap || got.gap != 3 {
		t.Fatal("header gap did not preserve value semantics")
	}
	if got := content.Gap(5); content.hasGap || !got.hasGap || got.gap != 5 {
		t.Fatal("content gap did not preserve value semantics")
	}
	if got := footer.Gap(7); footer.hasGap || !got.hasGap || got.gap != 7 {
		t.Fatal("footer gap did not preserve value semantics")
	}

	title := CardTitle("Title")
	styledTitle := title.Size(16).LineHeight(22).Color(color.NRGBA{R: 1}).Weight(500)
	if title.size != 0 || title.lineHeight != 0 || title.hasColor || title.weight != 0 {
		t.Fatal("title options mutated the original value")
	}
	if styledTitle.size != 16 || styledTitle.lineHeight != 22 || !styledTitle.hasColor || styledTitle.weight != 500 {
		t.Fatal("title options did not configure the returned value")
	}

	descriptionWidget := CardDescription("Description")
	styledDescription := descriptionWidget.Size(13).LineHeight(18).Color(color.NRGBA{G: 1})
	if descriptionWidget.size != 0 || descriptionWidget.lineHeight != 0 || descriptionWidget.hasColor {
		t.Fatal("description options mutated the original value")
	}
	if styledDescription.size != 13 || styledDescription.lineHeight != 18 || !styledDescription.hasColor {
		t.Fatal("description options did not configure the returned value")
	}
}

func TestCardSlotGapOverrides(t *testing.T) {
	ctx := cardTestContext(nil)
	tests := []struct {
		name   string
		widget frame.Widget
		want   image.Point
	}{
		{name: "header", widget: CardHeader(fixedWidget{image.Pt(20, 8)}, fixedWidget{image.Pt(30, 10)}).Gap(5), want: image.Pt(30, 23)},
		{name: "content", widget: CardContent(fixedWidget{image.Pt(20, 8)}, fixedWidget{image.Pt(30, 10)}).Gap(7), want: image.Pt(30, 25)},
		{name: "footer", widget: CardFooter(fixedWidget{image.Pt(20, 8)}, fixedWidget{image.Pt(30, 10)}).Gap(9), want: image.Pt(59, 10)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var ops op.Ops
			dims := test.widget.Layout(ctx, layout.Context{
				Constraints: layout.Constraints{Max: image.Pt(300, 200)},
				Ops:         &ops,
			})
			if dims.Size != test.want {
				t.Fatalf("slot size = %v, want %v", dims.Size, test.want)
			}
		})
	}
}

func TestCardTextSlotsLayout(t *testing.T) {
	ctx := cardTestContext(nil)
	widgets := []frame.Widget{
		CardTitle("Default title"),
		CardTitle("Styled title").
			Size(16).
			LineHeight(22).
			Color(color.NRGBA{R: 1, A: 0xff}).
			Weight(font.Bold),
		CardDescription("Default description"),
		CardDescription("Styled description").
			Size(13).
			LineHeight(18).
			Color(color.NRGBA{G: 1, A: 0xff}),
	}
	for _, widget := range widgets {
		var ops op.Ops
		dims := widget.Layout(ctx, layout.Context{
			Constraints: layout.Constraints{Max: image.Pt(300, 100)},
			Ops:         &ops,
		})
		if dims.Size.X <= 0 || dims.Size.Y <= 0 {
			t.Fatalf("text slot size = %v, want non-zero dimensions", dims.Size)
		}
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
