package avatar

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/testtheme"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type avatarStyle struct {
	background, foreground     color.NRGBA
	diameter, radius, iconSize unit.Dp
	textSize                   unit.Sp
}

func avatarStyleFor(activeTheme *theme.Theme, c Color, v Variant, s Size) avatarStyle {
	provider, err := theme.ComponentProvider(activeTheme)
	if err != nil {
		panic(err)
	}
	visual := provider.Avatar(activeTheme, theme.AvatarRequest{Color: c, Variant: v, Size: s})
	theme.ValidateAvatarVisual(visual)
	variant := style.Cascade(style.State{}, visual.Variant)
	size := style.Cascade(style.State{}, visual.Size)
	result := avatarStyle{}
	if variant.Paint != nil {
		if value, ok := variant.Paint.Background.(style.SolidColor); ok {
			result.background = value.Color
		}
	}
	if variant.Text != nil {
		if value, ok := variant.Text.Color.(style.SolidColor); ok {
			result.foreground = value.Color
		}
	}
	if size.Box != nil && size.Box.Width != nil {
		result.diameter = *size.Box.Width
	}
	if size.Paint != nil && size.Paint.Radius != nil {
		result.radius = *size.Paint.Radius
	}
	if size.Text != nil && size.Text.FontSize != nil {
		result.textSize = *size.Text.FontSize
	}
	part := style.CascadePart(style.State{}, style.PartIcon, visual.Defaults, visual.Variant, visual.Size, visual.Override)
	if part.Box != nil && part.Box.Width != nil {
		result.iconSize = *part.Box.Width
	}
	return result
}

type avatarProbe struct {
	layouts    int
	foreground color.NRGBA
	background color.NRGBA
}

type avatarIconRendererProvider struct {
	theme.ComponentVisualProvider
	calls *int
}

func (p avatarIconRendererProvider) Avatar(active *theme.Theme, request theme.AvatarRequest) theme.AvatarVisual {
	visual := p.ComponentVisualProvider.Avatar(active, request)
	visual.IconRenderer = func(gtx layout.Context, _ color.NRGBA) layout.Dimensions {
		(*p.calls)++
		return layout.Dimensions{Size: gtx.Constraints.Min}
	}
	return visual
}

func (p *avatarProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(16, 16))}
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
	activeTheme := *testtheme.Behavior()
	ctx := core.New(nil, &activeTheme, locale.LanguageEnglish)
	New("").Fallback(probe).Color(ColorSuccess).Variant(VariantSoft).Layout(ctx, avatarLayoutContext(image.Pt(100, 100)))
	style := avatarStyleFor(&activeTheme, ColorSuccess, VariantSoft, SizeMedium)
	if probe.layouts != 1 || probe.foreground != style.foreground || probe.background != style.background {
		t.Fatalf("fallback context = layouts %d foreground %v background %v", probe.layouts, probe.foreground, probe.background)
	}
}

func TestAvatarDefaultFallbackUsesProviderIconRenderer(t *testing.T) {
	calls := 0
	active := testtheme.Behavior()
	active.VisualProvider = avatarIconRendererProvider{
		ComponentVisualProvider: testtheme.BehaviorProvider(),
		calls:                   &calls,
	}
	ctx := core.New(nil, active, locale.LanguageEnglish)

	dims := New("").Layout(ctx, avatarLayoutContext(image.Pt(40, 40)))
	if dims.Size != image.Pt(40, 40) {
		t.Fatalf("provider-rendered Avatar size = %v", dims.Size)
	}
	if calls != 1 {
		t.Fatalf("provider icon renderer calls = %d, want 1", calls)
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
	activeTheme := *testtheme.Behavior()
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

func avatarTestContext() *core.Context {
	return core.New(nil, testtheme.Behavior(), locale.LanguageEnglish)
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
