package chip

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
	"github.com/qianniancn/flowui/theme"
)

type chipProbe struct {
	size        image.Point
	layouts     int
	constraints layout.Constraints
	foreground  color.NRGBA
	background  color.NRGBA
}

func (p *chipProbe) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.constraints = gtx.Constraints
	p.foreground = ctx.ForegroundColor()
	p.background = ctx.BackgroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func TestChipSizesHaveStableBehaviorHeights(t *testing.T) {
	tests := []struct {
		size Size
		want int
	}{
		{SizeSmall, 20},
		{SizeMedium, 24},
		{SizeLarge, 28},
	}
	ctx := chipTestContext(testtheme.Behavior())
	for _, test := range tests {
		dims := New("Label").Size(test.size).Layout(ctx, chipLayoutContext(nil, image.Pt(240, 80)))
		if dims.Size.Y != test.want {
			t.Errorf("size %v height = %d, want %d", test.size, dims.Size.Y, test.want)
		}
	}
}

func TestChipComposedContentInheritsSemanticColors(t *testing.T) {
	activeTheme := testtheme.Behavior()
	ctx := chipTestContext(activeTheme)
	start := &chipProbe{size: image.Pt(12, 12)}
	end := &chipProbe{size: image.Pt(12, 12)}
	New("Complete").
		Color(ColorSuccess).
		Variant(VariantSoft).
		StartContent(start).
		EndContent(end).
		Layout(ctx, chipLayoutContext(nil, image.Pt(240, 80)))

	tone := activeTheme.Semantic().Tone(theme.ToneSuccess)
	for name, probe := range map[string]*chipProbe{"start": start, "end": end} {
		if probe.layouts != 1 {
			t.Errorf("%s content layouts = %d, want 1", name, probe.layouts)
		}
		if probe.foreground != tone.OnSoft || probe.background != tone.Soft {
			t.Errorf("%s colors = %#v/%#v, want success soft colors", name, probe.foreground, probe.background)
		}
		if probe.constraints.Max.Y != 20 {
			t.Errorf("%s max height = %d, want 20", name, probe.constraints.Max.Y)
		}
	}
}

func TestChipTertiaryContentPreservesParentBackground(t *testing.T) {
	activeTheme := testtheme.Behavior()
	ctx := chipTestContext(activeTheme)
	probe := &chipProbe{size: image.Pt(12, 12)}
	New("Info").Color(ColorAccent).Variant(VariantTertiary).StartContent(probe).
		Layout(ctx, chipLayoutContext(nil, image.Pt(240, 80)))
	if probe.background != activeTheme.Semantic().Background {
		t.Fatalf("tertiary content background = %#v, want parent %#v", probe.background, activeTheme.Semantic().Background)
	}
}

func TestChipRespectsNarrowConstraints(t *testing.T) {
	end := &chipProbe{size: image.Pt(12, 12)}
	dims := New("A very long chip label").
		EndContent(end).
		Layout(chipTestContext(testtheme.Behavior()), chipLayoutContext(nil, image.Pt(80, 40)))
	if dims.Size.X > 80 || dims.Size.Y != 24 {
		t.Fatalf("narrow dimensions = %v, want width <= 80 and height 24", dims.Size)
	}
	if end.layouts != 1 || end.constraints.Max.X < 12 {
		t.Fatalf("narrow trailing content layouts/max width = %d/%d, want one layout with at least 12 px", end.layouts, end.constraints.Max.X)
	}
}

func TestChipExposesLabelSemantics(t *testing.T) {
	router := new(input.Router)
	ctx := chipTestContext(testtheme.Behavior())
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

func chipTestContext(activeTheme *theme.Theme) *core.Context {
	return core.New(nil, activeTheme, locale.LanguageEnglish)
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
