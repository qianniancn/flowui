package badge

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type badgeProbe struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *badgeProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestBadgeGeometryMatchesHeroUI(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Badge
	if tokens.SmallMinSize != 16 || tokens.MediumMinSize != 28 || tokens.LargeMinSize != 32 {
		t.Fatalf("Badge sizes = %v/%v/%v", tokens.SmallMinSize, tokens.MediumMinSize, tokens.LargeMinSize)
	}
	if tokens.SmallRadius != 12 || tokens.MediumRadius != 24 || tokens.LargeRadius != 16 || tokens.BorderWidth != 1 || tokens.PlacementOffsetRatio != 0.25 {
		t.Fatalf("Badge geometry = %#v", tokens)
	}
	if tokens.SmallTextSize != 10 || tokens.MediumTextSize != 12 || tokens.LargeTextSize != 14 {
		t.Fatalf("Badge text sizes = %v/%v/%v", tokens.SmallTextSize, tokens.MediumTextSize, tokens.LargeTextSize)
	}
}

func TestBadgePlacementsMatchHeroUIQuarterOffset(t *testing.T) {
	anchor := image.Pt(40, 40)
	badge := image.Pt(16, 16)
	tests := []struct {
		placement Placement
		want      image.Point
	}{
		{PlacementTopRight, image.Pt(28, -4)},
		{PlacementTopLeft, image.Pt(-4, -4)},
		{PlacementBottomRight, image.Pt(28, 28)},
		{PlacementBottomLeft, image.Pt(-4, 28)},
	}
	for _, test := range tests {
		if got := badgePosition(anchor, badge, test.placement, 0.25); got != test.want {
			t.Fatalf("Badge placement %d = %v, want %v", test.placement, got, test.want)
		}
	}
}

func TestBadgeDoesNotChangeAnchorDimensions(t *testing.T) {
	anchor := &badgeProbe{size: image.Pt(48, 32)}
	dims := New(anchor, "99+").Color(ColorDanger).Size(SizeSmall).Layout(badgeTestContext(), badgeLayoutContext(nil))
	if dims.Size != anchor.size || anchor.layouts != 1 {
		t.Fatalf("Badge anchor = size %v layouts %d", dims.Size, anchor.layouts)
	}
}

func TestDotBadgeUsesSquareMinimumSize(t *testing.T) {
	ctx := badgeTestContext()
	for _, test := range []struct {
		size Size
		want int
	}{{SizeSmall, 16}, {SizeMedium, 28}, {SizeLarge, 32}} {
		dims := New(nil, "").Size(test.size).layoutBadge(ctx, badgeLayoutContext(nil))
		if dims.Size != image.Pt(test.want, test.want) {
			t.Fatalf("dot Badge size %d = %v", test.size, dims.Size)
		}
	}
}

func TestBadgeCustomContentReceivesSemanticColors(t *testing.T) {
	content := &badgeProbe{size: image.Pt(10, 10)}
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	New(nil, "").Content(content).Color(ColorSuccess).Variant(VariantSoft).layoutBadge(ctx, badgeLayoutContext(nil))
	style := badgeStyleFor(&activeTheme, ColorSuccess, VariantSoft, SizeMedium, ctx.BackgroundColor())
	if content.layouts != 1 || content.foreground != style.foreground || content.background != style.background {
		t.Fatalf("Badge content context = layouts %d foreground %v background %v", content.layouts, content.foreground, content.background)
	}
}

func TestBadgeVariantColorsUseSemanticPalette(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	primary := badgeStyleFor(&activeTheme, ColorDanger, VariantPrimary, SizeMedium, activeTheme.Palette.Background)
	if primary.background != activeTheme.Palette.Danger || primary.foreground != activeTheme.Palette.DangerForeground {
		t.Fatalf("primary danger Badge = background %v foreground %v", primary.background, primary.foreground)
	}
	secondary := badgeStyleFor(&activeTheme, ColorAccent, VariantSecondary, SizeMedium, activeTheme.Palette.Background)
	if secondary.background != activeTheme.Palette.Default || secondary.foreground != activeTheme.Palette.AccentSoftForeground {
		t.Fatalf("secondary accent Badge = background %v foreground %v", secondary.background, secondary.foreground)
	}
	soft := badgeStyleFor(&activeTheme, ColorSuccess, VariantSoft, SizeMedium, activeTheme.Palette.Background)
	if soft.background != activeTheme.Palette.SuccessSoft || soft.foreground != activeTheme.Palette.SuccessSoftForeground {
		t.Fatalf("soft success Badge = background %v foreground %v", soft.background, soft.foreground)
	}
	if soft.border != activeTheme.Palette.Background {
		t.Fatalf("Badge border = %v, want HeroUI background token %v", soft.border, activeTheme.Palette.Background)
	}
}

func TestBadgeExposesAccessibleLabel(t *testing.T) {
	router := new(input.Router)
	gtx := badgeLayoutContext(router)
	New(&badgeProbe{size: image.Pt(40, 40)}, "").Alt("Online").Color(ColorSuccess).Size(SizeSmall).Layout(badgeTestContext(), gtx)
	router.Frame(gtx.Ops)
	if !badgeSemanticTreeContains(router.AppendSemantics(nil), "Online") {
		t.Fatal("Badge semantics did not expose alt text")
	}
}

func badgeTestContext() *frame.Context {
	activeTheme := theme.DefaultTheme()
	return frame.New(nil, &activeTheme, locale.LanguageEnglish)
}

func badgeLayoutContext(router *input.Router) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{Ops: new(op.Ops), Source: source, Constraints: layout.Constraints{Max: image.Pt(300, 300)}}
}

func badgeSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || badgeSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
