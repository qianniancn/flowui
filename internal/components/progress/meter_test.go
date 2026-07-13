package progress

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"testing"

	"gioui.org/io/input"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/text"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type meterProbe struct {
	size       image.Point
	layouts    int
	foreground color.NRGBA
}

type meterOverlayProbe struct {
	anchor image.Rectangle
}

func (p *meterProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	p.layouts++
	p.foreground = ctx.ForegroundColor()
	return layout.Dimensions{Size: gtx.Constraints.Constrain(p.size)}
}

func (p *meterOverlayProbe) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	frame.RegisterOverlay(ctx, frame.OverlayRequest{
		Key:       "meter-value-overlay",
		Anchor:    image.Rect(0, 0, 10, 10),
		HasAnchor: true,
		Layout: func(_ layout.Context, anchor image.Rectangle, _ bool) layout.Dimensions {
			p.anchor = anchor
			return layout.Dimensions{}
		},
	})
	return layout.Dimensions{Size: gtx.Constraints.Constrain(image.Pt(48, 18))}
}

func TestMeterOptionsUseValueSemantics(t *testing.T) {
	formatter := func(value float64) string { return "formatted" }
	base := Meter("storage", 60)
	configured := base.
		Label("Storage").
		Alt("Storage usage").
		ShowValue().
		ValueText("60 GB").
		ValueFormatter(formatter).
		ValueContent(text.New("Custom")).
		Range(0, 1000).
		Color(MeterWarning).
		Size(MeterLarge).
		Disabled(true)
	if base.label != "" || base.alt != "" || base.showValue || base.hasValueText || base.valueFormatter != nil || base.valueContent != nil || base.minValue != 0 || base.maxValue != 100 || base.color != MeterAccent || base.size != MeterMedium || base.disabled {
		t.Fatalf("configuring Meter mutated base: %#v", base)
	}
	if configured.label != "Storage" || configured.alt != "Storage usage" || !configured.showValue || configured.valueText != "60 GB" || configured.valueFormatter == nil || configured.valueContent == nil || configured.maxValue != 1000 || configured.color != MeterWarning || configured.size != MeterLarge || !configured.disabled {
		t.Fatalf("configured Meter = %#v", configured)
	}
}

func TestMeterDoesNotExposeIndeterminateMode(t *testing.T) {
	if _, ok := reflect.TypeFor[MeterWidget]().MethodByName("Indeterminate"); ok {
		t.Fatal("Meter unexpectedly exposes Indeterminate")
	}
}

func TestMeterRatioClampsKnownRange(t *testing.T) {
	tests := []struct {
		meter MeterWidget
		want  float32
	}{
		{Meter("meter", 50), 0.5},
		{Meter("meter", -10), 0},
		{Meter("meter", 140), 1},
		{Meter("meter", 75).Range(50, 100), 0.5},
		{Meter("meter", 10).Range(10, 10), 0},
		{Meter("meter", math.NaN()), 0},
		{Meter("meter", math.Inf(1)), 0},
	}
	for _, test := range tests {
		if got := test.meter.ratio(); got != test.want {
			t.Fatalf("Meter ratio = %v, want %v", got, test.want)
		}
	}
}

func TestMeterOutputFormatting(t *testing.T) {
	if got := Meter("meter", 42).outputText(); got != "42%" {
		t.Fatalf("default output = %q", got)
	}
	if got := Meter("meter", 750).Range(0, 1000).ValueFormatter(func(value float64) string { return "$750" }).outputText(); got != "$750" {
		t.Fatalf("formatted output = %q", got)
	}
	if got := Meter("meter", 42).ValueFormatter(func(float64) string { return "formatted" }).ValueText("42 GB").outputText(); got != "42 GB" {
		t.Fatalf("explicit output = %q", got)
	}
}

func TestMeterFormatterRunsOncePerLayout(t *testing.T) {
	calls := 0
	meter := Meter("revenue", 750).
		Label("Revenue").
		ValueFormatter(func(value float64) string {
			calls++
			return "$750"
		})
	meter.Layout(newContext(nil), progressBarTestContext())
	if calls != 1 {
		t.Fatalf("formatter calls = %d, want 1", calls)
	}
}

func TestMeterSemanticDescription(t *testing.T) {
	if got := Meter("storage", 60).Label("Storage").semanticDescription("60%"); got != "Storage 60%" {
		t.Fatalf("description = %q", got)
	}
	if got := Meter("storage", 60).Label("Storage").Alt("Storage usage").semanticDescription("60%"); got != "Storage usage 60%" {
		t.Fatalf("alt description = %q", got)
	}
	if got := Meter("storage", 60).semanticDescription("60%"); got != "Meter 60%" {
		t.Fatalf("fallback description = %q", got)
	}
	if got := Meter("storage", 60).Label("Storage").ValueText("").semanticDescription(""); got != "Storage 60%" {
		t.Fatalf("empty visual output description = %q", got)
	}
}

func TestMeterHeroUIThemeTokens(t *testing.T) {
	tokens := theme.DefaultTheme().Components.Meter
	if tokens.SmallHeight != 4 || tokens.MediumHeight != 8 || tokens.LargeHeight != 12 {
		t.Fatalf("Meter heights = %v/%v/%v", tokens.SmallHeight, tokens.MediumHeight, tokens.LargeHeight)
	}
	if tokens.SmallRadius != 2 || tokens.MediumRadius != 4 || tokens.LargeRadius != 6 || tokens.HeaderGap != 4 || tokens.TextSize != 14 {
		t.Fatalf("Meter tokens = %#v", tokens)
	}
}

func TestMeterStyleUsesHeroUISemanticColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	if style := meterStyleFor(&activeTheme, MeterDefault, false); style.track != activeTheme.Palette.Default || style.fill != activeTheme.Palette.DefaultForeground {
		t.Fatalf("default Meter style = %#v", style)
	}
	if style := meterStyleFor(&activeTheme, MeterSuccess, false); style.fill != activeTheme.Palette.Success {
		t.Fatalf("success Meter fill = %v", style.fill)
	}
	if style := meterStyleFor(&activeTheme, MeterDanger, true); style.fill != activeTheme.DisabledColor(activeTheme.Palette.Danger) {
		t.Fatalf("disabled Meter fill = %v", style.fill)
	}
}

func TestMeterLayoutTrackAndHeader(t *testing.T) {
	if dims := Meter("storage", 60).Layout(newContext(nil), progressBarTestContext()); dims.Size != image.Pt(300, 8) {
		t.Fatalf("track-only Meter size = %v", dims.Size)
	}
	dims := Meter("storage-header", 60).
		Label("Storage").
		ShowValue().
		Layout(newContext(nil), progressBarTestContext())
	if dims.Size.X != 300 || dims.Size.Y <= 8 {
		t.Fatalf("header Meter size = %v", dims.Size)
	}
}

func TestMeterValueContentRendersWithEmptyValueText(t *testing.T) {
	content := &meterProbe{size: image.Pt(48, 18)}
	ctx := newContext(nil)
	dims := Meter("custom-output", 60).
		ValueText("").
		ValueContent(content).
		Layout(ctx, progressBarTestContext())
	if content.layouts != 1 {
		t.Fatalf("custom value content layouts = %d, want 1", content.layouts)
	}
	if dims.Size != image.Pt(300, 30) {
		t.Fatalf("custom value Meter size = %v, want (300,30)", dims.Size)
	}
	if content.foreground != frame.ActiveTheme(ctx).Palette.Foreground {
		t.Fatalf("custom value foreground = %v", content.foreground)
	}
}

func TestMeterTracksOverlayInsideRightAlignedValueContent(t *testing.T) {
	probe := new(meterOverlayProbe)
	ctx := newContext(nil)
	gtx := progressBarTestContext()
	frame.BeginFrameWithViewport(ctx, gtx.Constraints.Max)
	Meter("tracked-output", 60).ValueContent(probe).Layout(ctx, gtx)
	frame.LayoutOverlays(ctx, gtx)
	frame.EndFrame(ctx)

	if want := image.Rect(252, 0, 262, 10); probe.anchor != want {
		t.Fatalf("custom value overlay anchor = %v, want %v", probe.anchor, want)
	}
}

func TestMeterExposesAccessibleDescription(t *testing.T) {
	router := new(input.Router)
	var ops op.Ops
	gtx := layout.Context{Ops: &ops, Source: router.Source(), Constraints: layout.Constraints{Max: image.Pt(300, 120)}}
	Meter("storage", 60).Alt("Storage usage").Layout(newContext(nil), gtx)
	router.Frame(&ops)
	if !meterSemanticTreeContains(router.AppendSemantics(nil), "Storage usage 60%") {
		t.Fatal("Meter semantics did not expose description")
	}
}

func meterSemanticTreeContains(nodes []input.SemanticNode, description string) bool {
	for _, node := range nodes {
		if node.Desc.Description == description || meterSemanticTreeContains(node.Children, description) {
			return true
		}
	}
	return false
}
