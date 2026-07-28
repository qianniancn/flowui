package avatar

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type avatarProbe struct {
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *avatarProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(16, 16))}
}

func TestAvatarGeometryMatchesHeroUI(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Avatar
	if tokens.SmallSize != 32 || tokens.MediumSize != 40 || tokens.LargeSize != 48 {
		t.Fatalf("Avatar sizes = %v/%v/%v", tokens.SmallSize, tokens.MediumSize, tokens.LargeSize)
	}
	if tokens.SmallRadius != 16 || tokens.MediumRadius != 24 || tokens.LargeRadius != 24 {
		t.Fatalf("Avatar radii = %v/%v/%v", tokens.SmallRadius, tokens.MediumRadius, tokens.LargeRadius)
	}
	if tokens.SmallTextSize != 14 || tokens.MediumTextSize != 14 || tokens.LargeTextSize != 16 {
		t.Fatalf("Avatar text sizes = %v/%v/%v", tokens.SmallTextSize, tokens.MediumTextSize, tokens.LargeTextSize)
	}
}

func TestAvatarSizesStaySquare(t *testing.T) {
	ctx := avatarTestContext()
	for _, test := range []struct {
		size Size
		want int
	}{{SizeSmall, 32}, {SizeMedium, 40}, {SizeLarge, 48}} {
		gtx := avatarLayoutContext(image.Pt(100, 100))
		if dims := New("AB").Size(test.size).Layout(ctx, gtx); dims.Size != image.Pt(test.want, test.want) {
			t.Fatalf("Avatar size %d = %v", test.size, dims.Size)
		}
	}
	gtx := avatarLayoutContext(image.Pt(24, 60))
	if dims := New("AB").Layout(ctx, gtx); dims.Size != image.Pt(24, 24) {
		t.Fatalf("constrained Avatar = %v", dims.Size)
	}
}

func TestAvatarCustomFallbackReceivesSemanticColors(t *testing.T) {
	probe := new(avatarProbe)
	activeTheme := theme.DefaultTheme()
	ctx := frame.New(nil, &activeTheme, locale.LanguageEnglish)
	New("").Fallback(probe).Color(ColorSuccess).Variant(VariantSoft).Layout(ctx, avatarLayoutContext(image.Pt(100, 100)))
	style := avatarStyleFor(&activeTheme, ColorSuccess, VariantSoft, SizeMedium)
	if probe.layouts != 1 || probe.foreground != style.foreground || probe.background != style.background {
		t.Fatalf("fallback context = layouts %d foreground %v background %v", probe.layouts, probe.foreground, probe.background)
	}
}

func TestAvatarImageSuppressesFallback(t *testing.T) {
	probe := new(avatarProbe)
	img := image.NewNRGBA(image.Rect(0, 0, 80, 40))
	dims := New("").Fallback(probe).Image(paint.NewImageOp(img)).Layout(avatarTestContext(), avatarLayoutContext(image.Pt(100, 100)))
	if dims.Size != image.Pt(40, 40) || probe.layouts != 0 {
		t.Fatalf("image Avatar = size %v fallback layouts %d", dims.Size, probe.layouts)
	}
}

func TestAvatarExposesAccessibleLabel(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Constraints: layout.Constraints{Max: image.Pt(100, 100)}}
	New("AM").Alt("Alex Morgan").Layout(avatarTestContext(), gtx)
	router.Frame(&ops)
	if !avatarSemanticTreeContains(router.AppendSemantics(nil), "Alex Morgan") {
		t.Fatal("Avatar semantics did not expose alt text")
	}
}

func TestAvatarSoftColorsUseSemanticPalette(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	tests := []struct {
		color Color
		bg    color.NRGBA
		fg    color.NRGBA
	}{
		{ColorAccent, activeTheme.Palette.AccentSoft, activeTheme.Palette.AccentSoftForeground},
		{ColorSuccess, activeTheme.Palette.SuccessSoft, activeTheme.Palette.SuccessSoftForeground},
		{ColorWarning, activeTheme.Palette.WarningSoft, activeTheme.Palette.WarningSoftForeground},
		{ColorDanger, activeTheme.Palette.DangerSoft, activeTheme.Palette.DangerSoftForeground},
	}
	for _, test := range tests {
		style := avatarStyleFor(&activeTheme, test.color, VariantSoft, SizeMedium)
		if style.background != test.bg || style.foreground != test.fg {
			t.Fatalf("Avatar color %d = background %v foreground %v", test.color, style.background, style.foreground)
		}
	}
}

func avatarTestContext() *frame.Context {
	activeTheme := theme.DefaultTheme()
	return frame.New(nil, &activeTheme, locale.LanguageEnglish)
}

func avatarLayoutContext(maximum image.Point) layout.Context {
	return layout.Context{Ops: new(op.Ops), Constraints: layout.Constraints{Max: maximum}}
}

func avatarSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || avatarSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
