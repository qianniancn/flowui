package chip

import (
	"image"
	"image/color"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/locale"
	"github.com/qianniancn/flowui/internal/theme"
)

type chipProbe struct {
	size        image.Point
	layouts     int
	constraints layout.Constraints
	foreground  color.NRGBA
	background  color.NRGBA
}

func (p *chipProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.constraints = gtx.Constraints
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestChipSizesHaveStableHeroUIHeights(t *testing.T) {
	tests := []struct {
		size Size
		want int
	}{
		{SizeSmall, 20},
		{SizeMedium, 24},
		{SizeLarge, 28},
	}
	ctx := chipTestContext(nil)
	for _, test := range tests {
		dims := New("Label").Size(test.size).Layout(ctx, chipLayoutContext(nil, image.Pt(240, 80)))
		if dims.Size.Y != test.want {
			t.Errorf("size %v height = %d, want %d", test.size, dims.Size.Y, test.want)
		}
	}
}

func TestChipStyleMatrixMatchesHeroUI(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	p := activeTheme.Palette
	defaultBackground := p.SurfaceTertiary
	defaultForeground := p.SurfaceTertiaryForeground
	tests := []struct {
		name       string
		color      Color
		variant    Variant
		background color.NRGBA
		foreground color.NRGBA
	}{
		{"default-secondary", ColorDefault, VariantSecondary, defaultBackground, defaultForeground},
		{"accent-secondary", ColorAccent, VariantSecondary, defaultBackground, p.AccentSoftForeground},
		{"success-primary", ColorSuccess, VariantPrimary, p.Success, p.SuccessForeground},
		{"warning-primary", ColorWarning, VariantPrimary, p.Warning, p.WarningForeground},
		{"danger-primary", ColorDanger, VariantPrimary, p.Danger, p.DangerForeground},
		{"accent-soft", ColorAccent, VariantSoft, p.AccentSoft, p.AccentSoftForeground},
		{"success-soft", ColorSuccess, VariantSoft, p.SuccessSoft, p.SuccessSoftForeground},
		{"warning-soft", ColorWarning, VariantSoft, p.WarningSoft, p.WarningSoftForeground},
		{"danger-soft", ColorDanger, VariantSoft, p.DangerSoft, p.DangerSoftForeground},
		{"default-soft", ColorDefault, VariantSoft, defaultSoftColor(defaultBackground), defaultForeground},
		{"accent-tertiary", ColorAccent, VariantTertiary, color.NRGBA{}, p.AccentSoftForeground},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			style := chipStyleFor(&activeTheme, test.color, test.variant)
			if style.background != test.background || style.foreground != test.foreground {
				t.Fatalf("style = %#v/%#v, want %#v/%#v", style.background, style.foreground, test.background, test.foreground)
			}
		})
	}
}

func TestChipDarkThemeUsesHeroUISemanticContrast(t *testing.T) {
	activeTheme := theme.DarkTheme()
	successPrimary := chipStyleFor(&activeTheme, ColorSuccess, VariantPrimary)
	if successPrimary.foreground != activeTheme.Palette.SuccessForeground {
		t.Fatalf("dark success primary foreground = %#v, want %#v", successPrimary.foreground, activeTheme.Palette.SuccessForeground)
	}
	warningPrimary := chipStyleFor(&activeTheme, ColorWarning, VariantPrimary)
	if warningPrimary.foreground != activeTheme.Palette.WarningForeground {
		t.Fatalf("dark warning primary foreground = %#v, want %#v", warningPrimary.foreground, activeTheme.Palette.WarningForeground)
	}
	if activeTheme.Palette.SuccessSoft.A != 0x1f || activeTheme.Palette.WarningSoft.A != 0x1f {
		t.Fatalf("dark soft alpha = success %#x warning %#x, want 0x1f", activeTheme.Palette.SuccessSoft.A, activeTheme.Palette.WarningSoft.A)
	}
}

func TestChipComposedContentInheritsSemanticColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := chipTestContext(&activeTheme)
	start := &chipProbe{size: image.Pt(12, 12)}
	end := &chipProbe{size: image.Pt(12, 12)}
	chip := New("Complete").
		Color(ColorSuccess).
		Variant(VariantSoft).
		StartContent(start).
		EndContent(end)
	chip.Layout(ctx, chipLayoutContext(nil, image.Pt(240, 80)))

	for name, probe := range map[string]*chipProbe{"start": start, "end": end} {
		if probe.layouts != 1 {
			t.Errorf("%s content layouts = %d, want 1", name, probe.layouts)
		}
		if probe.foreground != activeTheme.Palette.SuccessSoftForeground || probe.background != activeTheme.Palette.SuccessSoft {
			t.Errorf("%s colors = %#v/%#v, want success soft colors", name, probe.foreground, probe.background)
		}
		if probe.constraints.Max.Y != 20 {
			t.Errorf("%s max height = %d, want 20", name, probe.constraints.Max.Y)
		}
	}
}

func TestChipTertiaryContentPreservesParentBackground(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	ctx := chipTestContext(&activeTheme)
	probe := &chipProbe{size: image.Pt(12, 12)}
	New("Info").Color(ColorAccent).Variant(VariantTertiary).StartContent(probe).
		Layout(ctx, chipLayoutContext(nil, image.Pt(240, 80)))
	if probe.background != activeTheme.Palette.Background {
		t.Fatalf("tertiary content background = %#v, want parent %#v", probe.background, activeTheme.Palette.Background)
	}
}

func TestChipRespectsNarrowConstraints(t *testing.T) {
	end := &chipProbe{size: image.Pt(12, 12)}
	dims := New("A very long chip label").
		EndContent(end).
		Layout(chipTestContext(nil), chipLayoutContext(nil, image.Pt(80, 40)))
	if dims.Size.X > 80 || dims.Size.Y != 24 {
		t.Fatalf("narrow dimensions = %v, want width <= 80 and height 24", dims.Size)
	}
	if end.layouts != 1 || end.constraints.Max.X < 12 {
		t.Fatalf("narrow trailing content layouts/max width = %d/%d, want one layout with at least 12 px", end.layouts, end.constraints.Max.X)
	}
}

func TestChipExposesLabelSemantics(t *testing.T) {
	router := new(input.Router)
	ctx := chipTestContext(nil)
	var ops op.Ops
	gtx := layout.Context{
		Constraints: layout.Constraints{Max: image.Pt(240, 80)},
		Source:      router.Source(),
		Ops:         &ops,
	}
	New("Completed").Layout(ctx, gtx)
	router.Frame(&ops)
	if !chipSemanticTreeContains(router.AppendSemantics(nil), "Completed") {
		t.Fatal("Chip semantics did not expose its label")
	}
}

func chipTestContext(activeTheme *theme.Theme) *frame.Context {
	return frame.New(nil, activeTheme, locale.LanguageEnglish)
}

func chipLayoutContext(router *input.Router, max image.Point) layout.Context {
	var source input.Source
	if router != nil {
		source = router.Source()
	}
	var ops op.Ops
	return layout.Context{
		Constraints: layout.Constraints{Max: max},
		Source:      source,
		Ops:         &ops,
	}
}

func chipSemanticTreeContains(nodes []input.SemanticNode, label string) bool {
	for _, node := range nodes {
		if node.Desc.Label == label {
			return true
		}
		if chipSemanticTreeContains(node.Children, label) {
			return true
		}
	}
	return false
}
