package badge

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
)

type badgeProbe struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

func (p *badgeProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestBadgePlacementsMatchProviderOffset(t *testing.T) {
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

func TestDotBadgeUsesProviderMinimumSize(t *testing.T) {
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

func TestBadgeCustomContentReceivesResolvedProviderColors(t *testing.T) {
	content := &badgeProbe{size: image.Pt(10, 10)}
	ctx := badgeTestContext()
	value := New(nil, "").Content(content).Color(ColorSuccess).Variant(VariantSoft)
	gtx := badgeLayoutContext(nil)
	visual := value.visualWithCache(ctx, nil)
	resolved := value.resolveStyle(ctx, gtx, visual)
	value.layoutBadgeWithVisual(ctx, gtx, visual)
	wantForeground, ok := solidTextColor(resolved.label)
	if !ok {
		t.Fatal("provider label declaration has no solid foreground")
	}
	wantBackground, ok := solidBackground(resolved.root)
	if !ok {
		t.Fatal("provider root declaration has no solid background")
	}
	if content.layouts != 1 || content.foreground != wantForeground || content.background != wantBackground {
		t.Fatalf("Badge content context = layouts %d foreground %v/%v background %v/%v", content.layouts, content.foreground, wantForeground, content.background, wantBackground)
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

func badgeTestContext() *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
}

func badgeLayoutContext(router *input.Router) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	return layout.Context{Ops: new(op.Ops), Source: source, Constraints: layout.Constraints{Max: image.Pt(300, 300)}}
}

func solidTextColor(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Text == nil {
		return color.NRGBA{}, false
	}
	switch source := value.Text.Color.(type) {
	case style.SolidColor:
		return source.Color, true
	case *style.SolidColor:
		return source.Color, source != nil
	default:
		return color.NRGBA{}, false
	}
}

func solidBackground(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Paint == nil {
		return color.NRGBA{}, false
	}
	switch source := value.Paint.Background.(type) {
	case style.SolidColor:
		return source.Color, true
	case *style.SolidColor:
		return source.Color, source != nil
	default:
		return color.NRGBA{}, false
	}
}

func badgeSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label || badgeSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
