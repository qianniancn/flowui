package linechart

import (
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestLineChartOptionsUseValueSemantics(t *testing.T) {
	baseSeries := Values("api", "API", []float64{1, 2, 3})
	configuredSeries := baseSeries.
		Color(color.NRGBA{R: 1, A: 0xff}).
		Width(3).
		ShowPoints(false).
		ConnectNulls(true).
		Smoothness(0.35).
		Hidden(true)
	if baseSeries.hasColor || baseSeries.width != 0 || baseSeries.hasShowPoints || baseSeries.connectNulls || baseSeries.smooth != 0 || baseSeries.hidden {
		t.Fatalf("configuring LineChart Series mutated base: %#v", baseSeries)
	}
	if !configuredSeries.hasColor || configuredSeries.width != 3 || !configuredSeries.hasShowPoints || configuredSeries.showPoints || !configuredSeries.connectNulls || configuredSeries.smooth != 0.35 || !configuredSeries.hidden {
		t.Fatalf("configured LineChart Series = %#v", configuredSeries)
	}
	if smooth := baseSeries.Smooth(true); smooth.smooth != 0.5 || baseSeries.smooth != 0 {
		t.Fatalf("Smooth(true) = %#v, base %#v", smooth, baseSeries)
	}

	base := New("traffic", []Series{baseSeries})
	configured := base.
		Categories([]string{"Mon", "Tue", "Wed"}).
		Height(280).
		Grid(false).
		Legend(true).
		Tooltip(false).
		IncludeZero(false).
		XRange(0, 2).
		YRange(10, 20).
		XAxis("Day").
		YAxis("Requests").
		XTicks(4).
		YTicks(6).
		FormatX(func(float64) string { return "x" }).
		FormatY(func(float64) string { return "y" }).
		Label("Traffic").
		EmptyText("Empty").
		Disabled(true)
	if len(base.categories) != 0 || base.height != 0 || !base.showGrid || base.hasShowLegend || !base.showTooltip || !base.includeZero || base.hasXRange || base.hasYRange || base.disabled {
		t.Fatalf("configuring LineChart mutated base: %#v", base)
	}
	if len(configured.categories) != 3 || configured.height != 280 || configured.showGrid || !configured.hasShowLegend || !configured.showLegend || configured.showTooltip || configured.includeZero || !configured.hasXRange || !configured.hasYRange || configured.xTickCount != 4 || configured.yTickCount != 6 || configured.formatX == nil || configured.formatY == nil || configured.label != "Traffic" || configured.emptyText != "Empty" || !configured.disabled {
		t.Fatalf("configured LineChart = %#v", configured)
	}
}

func TestLineChartRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{"series width", func() { Values("a", "A", nil).Width(0) }},
		{"height", func() { New("chart", nil).Height(0) }},
		{"X range", func() { New("chart", nil).XRange(2, 1) }},
		{"Y range", func() { New("chart", nil).YRange(math.NaN(), 1) }},
		{"X ticks", func() { New("chart", nil).XTicks(1) }},
		{"Y ticks", func() { New("chart", nil).YTicks(0) }},
		{"negative smoothness", func() { Values("a", "A", nil).Smoothness(-0.1) }},
		{"large smoothness", func() { Values("a", "A", nil).Smoothness(1.1) }},
		{"NaN smoothness", func() { Values("a", "A", nil).Smoothness(float32(math.NaN())) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid LineChart configuration did not panic")
				}
			}()
			test.run()
		})
	}
}

func TestResolveChartDataHandlesCategoryValuesAndInvalidPoints(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	data := resolveChartData(New("chart", []Series{
		Values("values", "Values", []float64{4, math.NaN(), 8}),
		XY("xy", "XY", []Point{{X: -2, Y: 3}, {X: math.Inf(1), Y: 9}}),
	}).Categories([]string{"A", "B", "C"}), &activeTheme, testDp)
	if len(data.series) != 2 || len(data.series[0].points) != 3 || data.series[0].points[1].valid {
		t.Fatalf("resolved LineChart data = %#v", data)
	}
	if !data.xExtent.valid || data.xExtent.minimum != -2 || data.xExtent.maximum != 2 || data.yExtent.minimum != 3 || data.yExtent.maximum != 8 {
		t.Fatalf("resolved extents = X %#v Y %#v", data.xExtent, data.yExtent)
	}
	wantX := []float64{-2, 0, 2}
	if len(data.xValues) != len(wantX) {
		t.Fatalf("resolved X values = %v", data.xValues)
	}
	for index := range wantX {
		if data.xValues[index] != wantX[index] {
			t.Fatalf("resolved X values = %v", data.xValues)
		}
	}
}

func TestResolveChartDataRejectsSeriesKeys(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	for _, series := range [][]Series{
		{Values("", "Empty", nil)},
		{Values("same", "First", nil), Values("same", "Second", nil)},
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid LineChart series keys did not panic")
				}
			}()
			resolveChartData(New("chart", series), &activeTheme, testDp)
		}()
	}
}

func TestLinearScaleUsesEChartsStyleNiceTicks(t *testing.T) {
	scale := newLinearScale(0, 980, 5, true, false)
	if scale.minimum != 0 || scale.maximum != 1000 || scale.interval != 200 {
		t.Fatalf("nice scale = %#v", scale)
	}
	want := []float64{0, 200, 400, 600, 800, 1000}
	if len(scale.ticks) != len(want) {
		t.Fatalf("nice ticks = %v", scale.ticks)
	}
	for index := range want {
		if scale.ticks[index] != want[index] {
			t.Fatalf("nice ticks = %v", scale.ticks)
		}
	}

	decimal := newLinearScale(5.2, 5.8, 5, false, false)
	if math.Abs(decimal.interval-0.1) > 1e-9 || decimal.minimum != 5.2 || decimal.maximum != 5.8 {
		t.Fatalf("decimal nice scale = %#v", decimal)
	}
	constant := newLinearScale(5, 5, 5, false, false)
	if !constant.contains(5) || constant.minimum == constant.maximum {
		t.Fatalf("constant scale = %#v", constant)
	}
}

func TestLinearScaleHandlesExtremeFiniteValues(t *testing.T) {
	scale := newLinearScale(-1e308, 1e308, 5, false, false)
	if !finite(scale.minimum) || !finite(scale.maximum) || !finite(scale.interval) || len(scale.ticks) < 2 {
		t.Fatalf("extreme scale = %#v", scale)
	}
	for index := 1; index < len(scale.ticks); index++ {
		if scale.ticks[index] <= scale.ticks[index-1] {
			t.Fatalf("extreme ticks are not strictly increasing: %v", scale.ticks)
		}
	}
	if ratio := scale.ratio(0); math.Abs(ratio-0.5) > 1e-9 {
		t.Fatalf("extreme scale midpoint ratio = %v", ratio)
	}

	tiny := newLinearScale(1e-12, 5e-12, 4, false, false)
	if len(tiny.ticks) < 2 || !finite(tiny.interval) || formatAxisNumber(tiny.ticks[0], tiny.interval) == "" {
		t.Fatalf("tiny scale = %#v", tiny)
	}
}

func TestAxisNumberFormatting(t *testing.T) {
	for _, test := range []struct {
		value    float64
		interval float64
		want     string
	}{
		{12345, 1000, "12,345"},
		{-1234.5, 0.5, "-1,234.5"},
		{0.00000001, 0.1, "0"},
	} {
		if got := formatAxisNumber(test.value, test.interval); got != test.want {
			t.Fatalf("formatAxisNumber(%v, %v) = %q, want %q", test.value, test.interval, got, test.want)
		}
	}
}

func TestWalkLineHonorsConnectNulls(t *testing.T) {
	points := []resolvedPoint{
		{Point: Point{X: 0, Y: 1}, valid: true},
		{Point: Point{X: 1, Y: math.NaN()}},
		{Point: Point{X: 2, Y: 3}, valid: true},
	}
	for _, test := range []struct {
		connect bool
		moves   int
	}{
		{connect: false, moves: 2},
		{connect: true, moves: 1},
	} {
		moves := 0
		walkLine(points, test.connect, func(_ resolvedPoint, move bool) {
			if move {
				moves++
			}
		})
		if moves != test.moves {
			t.Fatalf("connectNulls %v produced %d path moves, want %d", test.connect, moves, test.moves)
		}
	}
}

func TestSmoothCubicsStayWithinAdjacentPointBounds(t *testing.T) {
	points := []f32.Point{f32.Pt(0, 0), f32.Pt(10, 100), f32.Pt(100, 0)}
	cubics := smoothCubics(points, 0.5)
	if len(cubics) != 2 || cubics[len(cubics)-1].to != points[len(points)-1] {
		t.Fatalf("smooth cubics = %#v", cubics)
	}
	previous := points[0]
	for index, cubic := range cubics {
		if !pointInsideBounds(cubic.control0, previous, cubic.to) || !pointInsideBounds(cubic.control1, previous, cubic.to) {
			t.Fatalf("cubic %d exceeds adjacent bounds: %#v", index, cubic)
		}
		previous = cubic.to
	}
	straight := smoothCubics([]f32.Point{f32.Pt(0, 0), f32.Pt(20, 20)}, 0.5)
	if len(straight) != 1 || straight[0].control0 != (f32.Pt(0, 0)) || straight[0].control1 != (f32.Pt(20, 20)) {
		t.Fatalf("two-point smooth line = %#v", straight)
	}
}

func TestSplitSmoothLineHonorsConnectNulls(t *testing.T) {
	points := []resolvedPoint{
		{Point: Point{X: 0, Y: 1}, valid: true},
		{Point: Point{X: 1, Y: math.NaN()}},
		{Point: Point{X: 2, Y: 3}, valid: true},
	}
	transform := func(point resolvedPoint) f32.Point { return f32.Pt(float32(point.X), float32(point.Y)) }
	if segments := splitSmoothLine(points, false, transform); len(segments) != 2 {
		t.Fatalf("disconnected smooth segments = %v", segments)
	}
	if segments := splitSmoothLine(points, true, transform); len(segments) != 1 || len(segments[0]) != 2 {
		t.Fatalf("connected smooth segments = %v", segments)
	}
}

func pointInsideBounds(point, first, second f32.Point) bool {
	return point.X >= min(first.X, second.X) && point.X <= max(first.X, second.X) &&
		point.Y >= min(first.Y, second.Y) && point.Y <= max(first.Y, second.Y)
}

func TestLineChartSelectionUsesSharedNearestX(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("chart", []Series{
		XY("first", "First", []Point{{X: 0, Y: 10}, {X: 1, Y: 20}}),
		XY("second", "Second", []Point{{X: 0, Y: 12}, {X: 1, Y: 24}}),
	})
	data := resolveChartData(widget, &activeTheme, testDp)
	geometry := chartGeometry{
		plot:   image.Rect(0, 0, 100, 100),
		xScale: newLinearScale(0, 1, 2, false, true),
		yScale: newLinearScale(0, 30, 3, false, true),
	}
	selection := widget.resolveSelection(data, geometry, 1, true)
	if selection.pixelX != 100 || len(selection.entries) != 2 || selection.entries[0].point.Y != 20 || selection.entries[1].point.Y != 24 {
		t.Fatalf("LineChart selection = %#v", selection)
	}
}

func TestLineChartPointerAndKeyboardSelection(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	widget := New("chart", []Series{Values("series", "Series", []float64{10, 20, 30})}).Categories([]string{"A", "B", "C"})
	now := time.Unix(1, 0)
	layoutLineChartFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotLineChart)
	if state == nil {
		t.Fatal("LineChart state is missing")
	}
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(300, 100)})
	layoutLineChartFrame(ctx, router, widget, now.Add(time.Millisecond))
	if !state.hovered || state.keyboard {
		t.Fatalf("pointer LineChart state = %#v", state)
	}

	router.Source().Execute(key.FocusCmd{Tag: &state.root})
	layoutLineChartFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	router.Queue(key.Event{Name: key.NameRightArrow, State: key.Press})
	layoutLineChartFrame(ctx, router, widget, now.Add(3*time.Millisecond))
	if !state.keyboard || state.keyboardIndex != 2 {
		t.Fatalf("keyboard LineChart state = %#v", state)
	}
	if !frame.FocusVisible(ctx, &state.root, router.Source().Focused(&state.root)) {
		t.Fatal("keyboard LineChart navigation did not restore visible focus")
	}
}

func TestLineChartTooltipDisabledClearsInteraction(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	widget := New("chart", []Series{Values("series", "Series", []float64{10, 20})})
	now := time.Unix(4, 0)
	layoutLineChartFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotLineChart)
	state.hovered = true
	state.keyboard = true
	state.keyboardIndex = 1
	state.pointerIndex = 1
	layoutLineChartFrame(ctx, router, widget.Tooltip(false), now.Add(time.Millisecond))
	if state.hovered || state.keyboard || state.keyboardIndex != -1 || state.pointerIndex != -1 {
		t.Fatalf("disabled LineChart tooltip retained interaction state: %#v", state)
	}
}

func TestLineChartThemeAndSemantics(t *testing.T) {
	tokens := theme.DefaultTheme().Components.LineChart
	if tokens.Height != 320 || tokens.LineWidth != 2 || tokens.PointSize != 6 || tokens.SeriesColors[0] != (color.NRGBA{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}) {
		t.Fatalf("LineChart theme tokens = %#v", tokens)
	}
	ctx := lineChartTestContext()
	router := new(input.Router)
	layoutLineChartFrame(ctx, router, New("chart", []Series{Values("series", "Series", []float64{1})}).Label("Traffic trend"), time.Unix(2, 0))
	if !semanticDescriptionExists(router.AppendSemantics(nil), "Traffic trend, 1 series") {
		t.Fatal("LineChart semantic description is missing")
	}
}

func TestLineChartHandlesSmallConstraints(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	widget := New("small", []Series{Values("series", "Series", []float64{1, 2})}).
		XAxis("X").
		YAxis("Y")
	viewport := image.Pt(24, 18)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: time.Unix(3, 0)}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	if dims.Size != viewport {
		t.Fatalf("small LineChart dimensions = %v, want %v", dims.Size, viewport)
	}
}

func TestLineChartLaysOutLegendAndEmptyState(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	widget := New("empty", []Series{
		Values("first", "First", []float64{math.NaN()}),
		Values("second", "Second", nil),
	}).EmptyText("No samples")
	if dims := layoutLineChartFrame(ctx, router, widget, time.Unix(5, 0)); dims.Size != image.Pt(520, 320) {
		t.Fatalf("empty LineChart dimensions = %v", dims.Size)
	}
}

func testDp(value unit.Dp) int {
	return int(value)
}

func lineChartTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutLineChartFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(520, 320)
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
