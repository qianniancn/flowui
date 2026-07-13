package piechart

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestPieChartAllocatesEChartsAngles(t *testing.T) {
	slices := []resolvedSlice{{value: 1}, {value: 3}}
	allocateAngles(slices, 4, -math.Pi/2, 1, 0, 0, true, false)
	assertAngle(t, slices[0].startAngle, -math.Pi/2)
	assertAngle(t, slices[0].endAngle, 0)
	assertAngle(t, slices[1].startAngle, 0)
	assertAngle(t, slices[1].endAngle, 3*math.Pi/2)

	zeros := []resolvedSlice{{radiusRatio: 1}, {radiusRatio: 1}, {radiusRatio: 1}}
	allocateAngles(zeros, 0, -math.Pi/2, 1, 0, 0, true, false)
	for _, slice := range zeros {
		assertAngle(t, slice.rawAngle, 2*math.Pi/3)
	}
	allocateAngles(zeros, 0, -math.Pi/2, 1, 0, 0, false, false)
	if hasVisibleSector(zeros) {
		t.Fatalf("zero-sum PieChart remained visible: %#v", zeros)
	}
}

func TestPieChartMinAndPadAnglesRedistributeRemainder(t *testing.T) {
	slices := []resolvedSlice{{value: 1}, {value: 99}}
	degrees := float32(math.Pi / 180)
	allocateAngles(slices, 100, 0, 1, 2*degrees, 10*degrees, true, false)
	assertAngle(t, slices[0].rawAngle, 12*degrees)
	assertAngle(t, slices[0].sweep(), 10*degrees)
	assertAngle(t, slices[1].rawAngle, 348*degrees)
	assertAngle(t, slices[1].sweep(), 346*degrees)
}

func TestPieChartPercentsUseEChartsLargestRemainder(t *testing.T) {
	slices := []resolvedSlice{{value: 1}, {value: 1}, {value: 1}}
	allocatePercents(slices, 3)
	if slices[0].percent != 33.34 || slices[1].percent != 33.33 || slices[2].percent != 33.33 {
		t.Fatalf("PieChart percentages = %#v", slices)
	}
}

func TestPieChartRoseTypesMatchEChartsLayout(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	values := []Data{Slice("zero", "Zero", 0), Slice("half", "Half", 5), Slice("full", "Full", 10)}
	radius := resolveChartData(New("rose-radius", values).RoseType(RoseRadius), &activeTheme)
	if radius.slices[0].radiusRatio != 0 || radius.slices[1].radiusRatio != .5 || radius.slices[2].radiusRatio != 1 {
		t.Fatalf("radius rose radii = %#v", radius.slices)
	}
	assertAngle(t, radius.slices[1].rawAngle, float32(fullCircle/3))
	assertAngle(t, radius.slices[2].rawAngle, float32(fullCircle*2/3))

	area := resolveChartData(New("rose-area", values).RoseType(RoseArea), &activeTheme)
	for _, slice := range area.slices {
		assertAngle(t, slice.rawAngle, float32(fullCircle/3))
	}
	zeros := resolveChartData(New("rose-zero", []Data{Slice("a", "A", 0), Slice("b", "B", 0)}).
		RoseType(RoseArea).
		StillShowZeroSum(false), &activeTheme)
	if zeros.slices[0].radiusRatio != .5 || zeros.slices[1].radiusRatio != .5 || !hasVisibleSector(zeros.slices) {
		t.Fatalf("zero-sum area rose = %#v", zeros.slices)
	}
}

func TestPieChartRoseAnimationInterpolatesRadius(t *testing.T) {
	from := chartData{slices: []resolvedSlice{{key: "rose", radiusRatio: .25}}}
	target := chartData{slices: []resolvedSlice{{key: "rose", radiusRatio: 1}}}
	midpoint := interpolatePieData(from, target, .5)
	if midpoint.slices[0].radiusRatio != .625 {
		t.Fatalf("animated rose radius = %v", midpoint.slices[0].radiusRatio)
	}
}

func TestResolvePieChartDataFiltersNegativeAndHiddenValues(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	custom := color.NRGBA{R: 0xff, A: 0xff}
	data := resolveChartData(New("pie", []Data{
		Slice("valid", "Valid", 2).Color(custom),
		Slice("negative", "Negative", -1),
		Slice("hidden", "Hidden", 4).Hidden(true),
	}), &activeTheme)
	if len(data.slices) != 1 || len(data.legend) != 3 || data.total != 2 {
		t.Fatalf("resolved PieChart data = %#v", data)
	}
	if data.slices[0].key != "valid" || data.slices[0].color != custom || data.slices[0].percent != 100 {
		t.Fatalf("resolved PieChart slice = %#v", data.slices[0])
	}
	if !data.legend[2].hidden {
		t.Fatalf("hidden PieChart legend item = %#v", data.legend[2])
	}
	selection := New("pie", nil).publicSelection(data.slices[0])
	if selection.Index != 0 || selection.Items[0].Percent != 100 {
		t.Fatalf("PieChart public selection = %#v", selection)
	}
}

func TestPieChartHitTestHonorsDirectionAndDonutHole(t *testing.T) {
	data := chartData{
		dir: 1,
		slices: []resolvedSlice{
			{startAngle: -math.Pi / 2, endAngle: 0, radiusRatio: .25},
			{startAngle: 0, endAngle: 3 * math.Pi / 2, radiusRatio: 1},
		},
	}
	geometry := chartGeometry{center: f32.Pt(50, 50), innerRadius: 10, outerRadius: 40}
	if index := hitTestPie(data, geometry, f32.Pt(60, 40)); index != 0 {
		t.Fatalf("clockwise PieChart hit index = %d", index)
	}
	if index := hitTestPie(data, geometry, f32.Pt(70, 30)); index != -1 {
		t.Fatalf("PieChart hit outside rose radius = %d", index)
	}
	if index := hitTestPie(data, geometry, f32.Pt(50, 50)); index != -1 {
		t.Fatalf("PieChart donut hole hit index = %d", index)
	}

	data.dir = -1
	data.slices = []resolvedSlice{{startAngle: -math.Pi / 2, endAngle: -math.Pi, radiusRatio: 1}}
	if index := hitTestPie(data, geometry, f32.Pt(30, 30)); index != 0 {
		t.Fatalf("counter-clockwise PieChart hit index = %d", index)
	}
}

func TestPieChartLegendAndSemantics(t *testing.T) {
	ctx := pieChartTestContext()
	router := new(input.Router)
	requestedKey := ""
	requestedHidden := false
	widget := New("pie", []Data{Slice("first", "First", 1), Slice("second", "Second", 2).Hidden(true)}).
		Label("Traffic sources").
		OnLegendChange(func(key string, hidden bool) {
			requestedKey, requestedHidden = key, hidden
		})
	layoutPieChartFrame(ctx, router, widget, time.Unix(1, 0))
	state, _ := frame.PeekState[chartState](ctx, "pie", stateSlotPieChart)
	state.legendItems["second"].Click()
	layoutPieChartFrame(ctx, router, widget, time.Unix(1, int64(time.Millisecond)))
	if requestedKey != "second" || requestedHidden {
		t.Fatalf("PieChart legend request = key %q hidden %v", requestedKey, requestedHidden)
	}
	if !semanticDescriptionExists(router.AppendSemantics(nil), "Traffic sources, 1 slice") {
		t.Fatal("PieChart semantic description is missing")
	}

	router.MoveFocus(key.FocusForward)
	if router.Source().Focused(&state.click) {
		t.Fatal("PieChart root entered keyboard focus order")
	}
}

func TestPieChartThemeAndValidation(t *testing.T) {
	tokens := theme.DefaultTheme().Components.PieChart
	if tokens.Height != 360 || tokens.EmphasisSize != 5 || tokens.SeriesColors[0] != (color.NRGBA{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}) {
		t.Fatalf("PieChart theme tokens = %#v", tokens)
	}
	for _, widget := range []Widget{
		New("pie", []Data{Slice("same", "First", 1), Slice("same", "Second", 2)}),
		New("pie", []Data{Slice("", "Empty", 1)}),
		New("pie", nil).InnerRadius(.8).OuterRadius(.5),
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid PieChart configuration did not panic")
				}
			}()
			activeTheme := theme.DefaultTheme()
			resolveChartData(widget, &activeTheme)
		}()
	}
	func() {
		defer func() {
			if recover() == nil {
				t.Fatal("invalid PieChart rose type did not panic")
			}
		}()
		_ = New("pie", nil).RoseType(RoseType(99))
	}()
}

func assertAngle(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 1e-5 {
		t.Fatalf("angle = %v, want %v", got, want)
	}
}

func pieChartTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutPieChartFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(520, 360)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	return dims
}

func semanticDescriptionExists(nodes []input.SemanticNode, description string) bool {
	for _, node := range nodes {
		if node.Desc.Description == description || semanticDescriptionExists(node.Children, description) {
			return true
		}
	}
	return false
}
