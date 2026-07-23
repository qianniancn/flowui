package linechart

import (
	"image"
	"image/color"
	"math"
	"runtime"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestLineChartOptionsUseValueSemantics(t *testing.T) {
	baseSeries := Values("api", "API", []float64{1, 2, 3})
	configuredSeries := baseSeries.
		Color(color.NRGBA{R: 1, A: 0xff}).
		Width(3).
		ShowPoints(false).
		PointSize(9).
		ConnectNulls(true).
		Smoothness(0.35).
		Step(StepMiddle).
		LineStyle(LineDashed).
		AreaColor(color.NRGBA{B: 2, A: 0x40}).
		Sampling(SamplingMinMax).
		Hidden(true)
	if baseSeries.hasColor || baseSeries.width != 0 || baseSeries.hasShowPoints || baseSeries.pointSize != 0 || baseSeries.connectNulls || baseSeries.smooth != 0 || baseSeries.step != StepNone || baseSeries.lineStyle != LineSolid || baseSeries.area || baseSeries.sampling != SamplingNone || baseSeries.hidden {
		t.Fatalf("configuring LineChart Series mutated base: %#v", baseSeries)
	}
	if !configuredSeries.hasColor || configuredSeries.width != 3 || !configuredSeries.hasShowPoints || configuredSeries.showPoints || configuredSeries.pointSize != 9 || !configuredSeries.connectNulls || configuredSeries.smooth != 0.35 || configuredSeries.step != StepMiddle || configuredSeries.lineStyle != LineDashed || !configuredSeries.area || !configuredSeries.hasAreaColor || configuredSeries.sampling != SamplingMinMax || !configuredSeries.hidden {
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
		Animation(false).
		AnimationDuration(250*time.Millisecond).
		AnimationEasing(func(value float32) float32 { return value }).
		UpdateAnimationDuration(150*time.Millisecond).
		UpdateAnimationEasing(func(value float32) float32 { return value }).
		OnLegendChange(func(string, bool) {}).
		OnDataClick(func(chart.Selection) {}).
		TooltipContent(func(chart.Selection) frame.Widget { return nil }).
		DataWindow(0.25, 0.75).
		OnDataWindowChange(func(chart.DataWindow) {}).
		MarkLines([]chart.MarkLine{chart.NewMarkLine(chart.AxisY, 15)}).
		MarkAreas([]chart.MarkArea{chart.NewMarkArea(chart.AxisX, 0.5, 1.5)}).
		MarkPoints([]chart.MarkPoint{chart.NewMarkPoint(1, 15)}).
		Label("Traffic").
		EmptyText("Empty").
		Disabled(true).
		Style(flowstyle.Style{}.Radius(4))
	if len(base.categories) != 0 || base.height != 0 || !base.showGrid || base.hasShowLegend || !base.showTooltip || !base.includeZero || base.hasXRange || base.hasYRange || !base.animation || base.animationDuration != time.Second || base.updateAnimationDuration != 500*time.Millisecond || base.disabled {
		t.Fatalf("configuring LineChart mutated base: %#v", base)
	}
	if len(configured.categories) != 3 || configured.height != 280 || configured.showGrid || !configured.hasShowLegend || !configured.showLegend || configured.showTooltip || configured.includeZero || !configured.hasXRange || !configured.hasYRange || configured.xTickCount != 4 || configured.yTickCount != 6 || configured.formatX == nil || configured.formatY == nil || configured.animation || configured.animationDuration != 250*time.Millisecond || configured.animationEasing == nil || configured.updateAnimationDuration != 150*time.Millisecond || configured.updateAnimationEasing == nil || configured.onLegendChange == nil || configured.onDataClick == nil || configured.tooltipContent == nil || !configured.hasDataWindow || configured.dataWindow.Start != 0.25 || configured.dataWindow.End != 0.75 || configured.onDataWindowChange == nil || len(configured.markLines) != 1 || len(configured.markAreas) != 1 || len(configured.markPoints) != 1 || configured.label != "Traffic" || configured.emptyText != "Empty" || !configured.disabled {
		t.Fatalf("configured LineChart = %#v", configured)
	}
	if configured.customStyle.Resolve(flowstyle.StyleState{}).Paint == nil {
		t.Fatal("configured LineChart did not retain its style")
	}
}

func TestLineChartLegendDataRetainsHiddenSeries(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	data := resolveChartData(New("chart", []Series{
		Values("visible", "Visible", []float64{1, 2}),
		Values("hidden", "Hidden", []float64{3, 4}).Hidden(true),
	}), &activeTheme, testDp)
	if len(data.series) != 1 || len(data.legend) != 2 || !data.legend[1].hidden {
		t.Fatalf("LineChart hidden legend data = %#v", data)
	}
}

func TestLineChartDataVersionCachesResolvedData(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	cache := new(chartDataCache)
	metric := unit.Metric{PxPerDp: 1, PxPerSp: 1}
	first := cache.resolve(New("chart", []Series{Values("series", "Series", []float64{1})}).DataVersion(1), &activeTheme, metric)
	second := cache.resolve(New("chart", []Series{Values("series", "Series", []float64{99})}).DataVersion(1), &activeTheme, metric)
	if first.generation == 0 || second.generation != first.generation || second.series[0].points[0].Y != 1 {
		t.Fatalf("same-version LineChart data was not reused: first %#v second %#v", first, second)
	}
	third := cache.resolve(New("chart", []Series{Values("series", "Series", []float64{99})}).DataVersion(2), &activeTheme, metric)
	if third.generation == second.generation || third.series[0].points[0].Y != 99 {
		t.Fatalf("new-version LineChart data was not resolved: %#v", third)
	}
	scaled := cache.resolve(New("chart", []Series{Values("series", "Series", []float64{99})}).DataVersion(2), &activeTheme, unit.Metric{PxPerDp: 2, PxPerSp: 2})
	if scaled.generation == third.generation {
		t.Fatal("LineChart cache ignored a metric change")
	}
	uncached := cache.resolve(New("chart", []Series{Values("series", "Series", []float64{7})}), &activeTheme, metric)
	if uncached.generation != 0 || uncached.series[0].points[0].Y != 7 {
		t.Fatalf("unversioned LineChart data was cached: %#v", uncached)
	}
}

func BenchmarkLineChartDataVersion(b *testing.B) {
	values := make([]float64, 10_000)
	for index := range values {
		values[index] = float64(index)
	}
	widget := New("chart", []Series{Values("series", "Series", values)})
	activeTheme := theme.DefaultTheme()
	metric := unit.Metric{PxPerDp: 1, PxPerSp: 1}
	for _, benchmark := range []struct {
		name   string
		widget Widget
	}{
		{name: "unversioned", widget: widget},
		{name: "versioned", widget: widget.DataVersion(1)},
	} {
		b.Run(benchmark.name, func(b *testing.B) {
			cache := new(chartDataCache)
			b.ReportAllocs()
			for b.Loop() {
				data := cache.resolve(benchmark.widget, &activeTheme, metric)
				runtime.KeepAlive(data)
			}
		})
	}
}

func TestLineChartRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{"series width", func() { Values("a", "A", nil).Width(0) }},
		{"point size", func() { Values("a", "A", nil).PointSize(0) }},
		{"step mode", func() { Values("a", "A", nil).Step(StepMode(9)) }},
		{"line style", func() { Values("a", "A", nil).LineStyle(LineStyle(9)) }},
		{"sampling mode", func() { Values("a", "A", nil).Sampling(SamplingMode(9)) }},
		{"height", func() { New("chart", nil).Height(0) }},
		{"X range", func() { New("chart", nil).XRange(2, 1) }},
		{"Y range", func() { New("chart", nil).YRange(math.NaN(), 1) }},
		{"X ticks", func() { New("chart", nil).XTicks(1) }},
		{"Y ticks", func() { New("chart", nil).YTicks(0) }},
		{"negative smoothness", func() { Values("a", "A", nil).Smoothness(-0.1) }},
		{"large smoothness", func() { Values("a", "A", nil).Smoothness(1.1) }},
		{"NaN smoothness", func() { Values("a", "A", nil).Smoothness(float32(math.NaN())) }},
		{"animation duration", func() { New("chart", nil).AnimationDuration(-time.Millisecond) }},
		{"update animation duration", func() { New("chart", nil).UpdateAnimationDuration(-time.Millisecond) }},
		{"animation easing", func() { New("chart", nil).AnimationEasing(nil) }},
		{"update animation easing", func() { New("chart", nil).UpdateAnimationEasing(nil) }},
		{"data window", func() { New("chart", nil).DataWindow(0.8, 0.2) }},
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

func TestLineChartStepAndSmoothPixelSegments(t *testing.T) {
	points := []f32.Point{f32.Pt(0, 10), f32.Pt(10, 20), f32.Pt(20, 5)}
	middle := steppedPoints(points, StepMiddle)
	want := []f32.Point{f32.Pt(0, 10), f32.Pt(5, 10), f32.Pt(5, 20), f32.Pt(10, 20), f32.Pt(15, 20), f32.Pt(15, 5), f32.Pt(20, 5)}
	if len(middle) != len(want) {
		t.Fatalf("middle step points = %v", middle)
	}
	for index := range want {
		if middle[index] != want[index] {
			t.Fatalf("middle step points = %v", middle)
		}
	}
	smooth := sampledSmoothPoints(points, 0.5)
	if len(smooth) != (len(points)-1)*smoothSamplesPerSegment+1 || smooth[0] != points[0] || smooth[len(smooth)-1] != points[len(points)-1] {
		t.Fatalf("sampled smooth points = %v", smooth)
	}
}

func TestLineChartVisibleMinMaxSamplingRetainsExtrema(t *testing.T) {
	points := make([]f32.Point, 1000)
	for index := range points {
		y := float32(index % 7)
		if index == 500 {
			y = -100
		}
		if index == 501 {
			y = 100
		}
		points[index] = f32.Pt(float32(index), y)
	}
	visible := visiblePixelSegment(points, 400, 600)
	if visible[0].X != 399 || visible[len(visible)-1].X != 601 {
		t.Fatalf("visible points = %v to %v", visible[0], visible[len(visible)-1])
	}
	sampled := minMaxPixelSample(visible, 50)
	if len(sampled) > 102 || sampled[0] != visible[0] || sampled[len(sampled)-1] != visible[len(visible)-1] {
		t.Fatalf("sampled point count/endpoints = %d, %v to %v", len(sampled), sampled[0], sampled[len(sampled)-1])
	}
	foundMinimum, foundMaximum := false, false
	for _, point := range sampled {
		foundMinimum = foundMinimum || point.Y == -100
		foundMaximum = foundMaximum || point.Y == 100
	}
	if !foundMinimum || !foundMaximum {
		t.Fatalf("sampled extrema missing: min %v max %v", foundMinimum, foundMaximum)
	}
}

func TestLineChartDataWindowConstrainsXScale(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("chart", []Series{Values("values", "Values", []float64{1, 2, 3, 4, 5})}).
		Categories([]string{"A", "B", "C", "D", "E"}).
		DataWindow(0.25, 0.75)
	data := resolveChartData(widget, &activeTheme, testDp)
	scale := widget.resolveXScale(data)
	if scale.Minimum != 1 || scale.Maximum != 3 {
		t.Fatalf("LineChart windowed X scale = %#v", scale)
	}
}

func TestLineChartAnnotationGeometryUsesVisibleScales(t *testing.T) {
	geometry := chartGeometry{
		plot:   image.Rect(10, 20, 110, 120),
		xScale: chart.NewLinearScale(0, 10, 5, false, true),
		yScale: chart.NewLinearScale(0, 100, 5, false, true),
	}
	rect, ok := lineMarkAreaRect(chart.NewMarkArea(chart.AxisY, 20, 40), geometry)
	if !ok || rect != image.Rect(10, 80, 110, 100) {
		t.Fatalf("LineChart mark area = %v, ok %v", rect, ok)
	}
	from, to, ok := lineMarkEndpoints(chart.NewMarkLine(chart.AxisX, 5), geometry)
	if !ok || from != f32.Pt(60, 20) || to != f32.Pt(60, 120) {
		t.Fatalf("LineChart mark line = %v to %v, ok %v", from, to, ok)
	}
}

func TestLineChartAnimationInterpolatesPointGeometry(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	target := resolveChartData(New("chart", []Series{
		Values("series", "Series", []float64{10, 20}),
	}), &activeTheme, testDp)
	from := lineBaselineData(target, 0)
	midpoint := interpolateLineData(from, target, 0.5, 0)
	if midpoint.series[0].points[0].Y != 5 || midpoint.series[0].points[1].Y != 10 {
		t.Fatalf("animated LineChart midpoint = %#v", midpoint.series[0].points)
	}
	if target.series[0].points[0].Y != 10 || target.series[0].points[1].Y != 20 {
		t.Fatalf("LineChart animation mutated target = %#v", target.series[0].points)
	}
}

func TestLineChartUpdateAnimationStartsFromDisplayedData(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	previous := resolveChartData(New("chart", []Series{
		Values("series", "Series", []float64{4, 8}),
	}), &activeTheme, testDp)
	target := resolveChartData(New("chart", []Series{
		Values("series", "Series", []float64{10, 20}),
	}), &activeTheme, testDp)
	from := lineTransitionFrom(previous, target, 0)
	if from.series[0].points[0].Y != 4 || from.series[0].points[1].Y != 8 {
		t.Fatalf("LineChart update start = %#v", from.series[0].points)
	}
}

func TestLineChartAnimationBaselineUsesVisibleYRange(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("chart", []Series{Values("series", "Series", []float64{10, 20})}).IncludeZero(false)
	target := resolveChartData(widget, &activeTheme, testDp)
	baseline := widget.animationBaseline(target)
	if baseline != widget.resolveYScale(target).Minimum || baseline == 0 {
		t.Fatalf("non-zero animation baseline = %v", baseline)
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
	if !data.xExtent.Valid || data.xExtent.Minimum != -2 || data.xExtent.Maximum != 2 || data.yExtent.Minimum != 3 || data.yExtent.Maximum != 8 {
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
	scale := chart.NewLinearScale(0, 980, 5, true, false)
	if scale.Minimum != 0 || scale.Maximum != 1000 || scale.Interval != 200 {
		t.Fatalf("nice scale = %#v", scale)
	}
	want := []float64{0, 200, 400, 600, 800, 1000}
	if len(scale.Ticks) != len(want) {
		t.Fatalf("nice ticks = %v", scale.Ticks)
	}
	for index := range want {
		if scale.Ticks[index] != want[index] {
			t.Fatalf("nice ticks = %v", scale.Ticks)
		}
	}

	decimal := chart.NewLinearScale(5.2, 5.8, 5, false, false)
	if math.Abs(decimal.Interval-0.1) > 1e-9 || decimal.Minimum != 5.2 || decimal.Maximum != 5.8 {
		t.Fatalf("decimal nice scale = %#v", decimal)
	}
	constant := chart.NewLinearScale(5, 5, 5, false, false)
	if !constant.Contains(5) || constant.Minimum == constant.Maximum {
		t.Fatalf("constant scale = %#v", constant)
	}
}

func TestLinearScaleHandlesExtremeFiniteValues(t *testing.T) {
	scale := chart.NewLinearScale(-1e308, 1e308, 5, false, false)
	if !chart.Finite(scale.Minimum) || !chart.Finite(scale.Maximum) || !chart.Finite(scale.Interval) || len(scale.Ticks) < 2 {
		t.Fatalf("extreme scale = %#v", scale)
	}
	for index := 1; index < len(scale.Ticks); index++ {
		if scale.Ticks[index] <= scale.Ticks[index-1] {
			t.Fatalf("extreme ticks are not strictly increasing: %v", scale.Ticks)
		}
	}
	if ratio := scale.Ratio(0); math.Abs(ratio-0.5) > 1e-9 {
		t.Fatalf("extreme scale midpoint ratio = %v", ratio)
	}

	tiny := chart.NewLinearScale(1e-12, 5e-12, 4, false, false)
	if len(tiny.Ticks) < 2 || !chart.Finite(tiny.Interval) || chart.FormatAxisNumber(tiny.Ticks[0], tiny.Interval) == "" {
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
		if got := chart.FormatAxisNumber(test.value, test.interval); got != test.want {
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
		xScale: chart.NewLinearScale(0, 1, 2, false, true),
		yScale: chart.NewLinearScale(0, 30, 3, false, true),
	}
	selection := widget.resolveSelection(data, geometry, 1, true)
	if selection.pixelX != 100 || len(selection.entries) != 2 || selection.entries[0].point.Y != 20 || selection.entries[1].point.Y != 24 {
		t.Fatalf("LineChart selection = %#v", selection)
	}
}

func TestLineChartPointerSelection(t *testing.T) {
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
	if !state.hovered || state.pointer != f32.Pt(300, 100) {
		t.Fatalf("pointer LineChart state = %#v", state)
	}
}

func TestLineChartRootDoesNotEnterKeyboardFocusOrder(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	widget := New("chart", []Series{
		Values("first", "First", []float64{10, 20}),
		Values("second", "Second", []float64{12, 24}),
	}).OnLegendChange(func(string, bool) {})
	layoutLineChartFrame(ctx, router, widget, time.Unix(1, 0))
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotLineChart)

	router.MoveFocus(key.FocusForward)
	if router.Source().Focused(&state.click) {
		t.Fatal("LineChart root entered keyboard focus order")
	}
	for key, item := range state.legendItems {
		if router.Source().Focused(item) {
			t.Fatalf("LineChart legend item %q entered keyboard focus order", key)
		}
	}
}

func TestLineChartDoubleClickResetsDataWindow(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	requested := chart.DataWindow{}
	widget := New("chart", []Series{Values("series", "Series", []float64{1, 2, 3, 4})}).
		DataWindow(.2, .8).
		OnDataWindowChange(func(window chart.DataWindow) { requested = window })
	start := time.Unix(2, 0)
	layoutLineChartFrame(ctx, router, widget, start)
	queueLineChartClick(router, 1, f32.Pt(300, 140))
	layoutLineChartFrame(ctx, router, widget, start.Add(time.Millisecond))
	queueLineChartClick(router, 2, f32.Pt(300, 140))
	layoutLineChartFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if requested != chart.FullDataWindow() {
		t.Fatalf("LineChart double-click window = %#v", requested)
	}
}

func TestLineChartWheelRequestsControlledDataWindow(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	requested := chart.DataWindow{}
	widget := New("chart", []Series{Values("series", "Series", []float64{1, 2, 3, 4})}).
		DataWindow(0.2, 0.8).
		OnDataWindowChange(func(window chart.DataWindow) { requested = window })
	now := time.Unix(8, 0)
	layoutLineChartFrame(ctx, router, widget, now)
	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(300, 150), Scroll: f32.Pt(0, -1)})
	layoutLineChartFrame(ctx, router, widget, now.Add(time.Millisecond))
	if requested.End-requested.Start >= 0.6 || requested.Start <= 0.2 || requested.End >= 0.8 {
		t.Fatalf("LineChart wheel window request = %#v", requested)
	}
}

func TestLineChartWithoutDataWindowDoesNotConsumeParentScroll(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	list := &layout.List{Axis: layout.Vertical}
	widget := New("chart", []Series{Values("series", "Series", []float64{1, 2, 3})})
	now := time.Unix(9, 0)
	layoutLineChartListFrame(ctx, router, list, widget, now)

	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(300, 100), Scroll: f32.Pt(0, 80)})
	layoutLineChartListFrame(ctx, router, list, widget, now.Add(time.Millisecond))
	if list.Position.First == 0 && list.Position.Offset == 0 {
		t.Fatal("LineChart tooltip interaction consumed the parent list scroll")
	}
}

func TestLineChartLegendClickDataClickAndCustomTooltip(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	var legendKey string
	var legendHidden bool
	var clicked chart.Selection
	tooltipLaidOut := false
	widget := New("chart", []Series{
		Values("visible", "Visible", []float64{10, 20}),
		Values("hidden", "Hidden", []float64{30, 40}).Hidden(true),
	}).Categories([]string{"A", "B"}).
		OnLegendChange(func(key string, hidden bool) {
			legendKey, legendHidden = key, hidden
		}).
		OnDataClick(func(selection chart.Selection) { clicked = selection }).
		TooltipContent(func(chart.Selection) frame.Widget {
			return frame.WidgetFunc(func(*frame.Context, layout.Context) layout.Dimensions {
				tooltipLaidOut = true
				return layout.Dimensions{Size: image.Pt(80, 24)}
			})
		})
	now := time.Unix(7, 0)
	layoutLineChartFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotLineChart)
	state.legendItems["hidden"].Click()
	layoutLineChartFrame(ctx, router, widget, now.Add(time.Millisecond))
	if legendKey != "hidden" || legendHidden {
		t.Fatalf("LineChart legend request = key %q hidden %v, want hidden false", legendKey, legendHidden)
	}

	queueLineChartClick(router, 1, f32.Pt(300, 140))
	layoutLineChartFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	if clicked.Label == "" || clicked.Index < 0 || len(clicked.Items) != 1 || clicked.Items[0].SeriesKey != "visible" {
		t.Fatalf("LineChart click selection = %#v", clicked)
	}
	if !tooltipLaidOut {
		t.Fatal("LineChart custom tooltip was not laid out")
	}

	clicked = chart.Selection{}
	tooltipLaidOut = false
	queueLineChartClick(router, 2, f32.Pt(300, 140))
	layoutLineChartFrame(ctx, router, widget.Tooltip(false), now.Add(3*time.Millisecond))
	if len(clicked.Items) != 1 {
		t.Fatalf("LineChart data click with tooltip disabled = %#v", clicked)
	}
	if tooltipLaidOut {
		t.Fatal("LineChart laid out custom tooltip while disabled")
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
	layoutLineChartFrame(ctx, router, widget.Tooltip(false), now.Add(time.Millisecond))
	if state.hovered {
		t.Fatalf("disabled LineChart tooltip retained interaction state: %#v", state)
	}
	if state.tooltipTransition.Value() != 0 || len(state.tooltipSelection.entries) != 0 {
		t.Fatalf("disabled LineChart retained tooltip presentation state: %#v", state)
	}
}

func TestLineChartTooltipAnchorTracksPointer(t *testing.T) {
	if got := chart.TooltipAnchor(f32.Pt(100.4, 79.6)); got != image.Rect(100, 80, 101, 81) {
		t.Fatalf("LineChart tooltip anchor = %v", got)
	}
}

func TestLineChartTooltipAnimatesOutWithLastSelection(t *testing.T) {
	ctx := lineChartTestContext()
	router := new(input.Router)
	tooltipLayouts := 0
	widget := New("chart", []Series{Values("series", "Series", []float64{10, 20})}).
		TooltipContent(func(chart.Selection) frame.Widget {
			return frame.WidgetFunc(func(*frame.Context, layout.Context) layout.Dimensions {
				tooltipLayouts++
				return layout.Dimensions{Size: image.Pt(80, 24)}
			})
		})
	start := time.Unix(10, 0)
	layoutLineChartFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotLineChart)
	state.hovered = true
	state.pointer = f32.Pt(300, 140)
	layoutLineChartFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutLineChartFrame(ctx, router, widget, start.Add(201*time.Millisecond))

	tooltipLayouts = 0
	state.hovered = false
	layoutLineChartFrame(ctx, router, widget, start.Add(202*time.Millisecond))
	if tooltipLayouts != 1 || !state.tooltipTransition.Exiting() || len(state.tooltipSelection.entries) == 0 {
		t.Fatalf("LineChart exit state = layouts %d exiting %v entries %d", tooltipLayouts, state.tooltipTransition.Exiting(), len(state.tooltipSelection.entries))
	}

	tooltipLayouts = 0
	layoutLineChartFrame(ctx, router, widget, start.Add(302*time.Millisecond))
	if tooltipLayouts != 0 || state.tooltipTransition.Value() != 0 || len(state.tooltipSelection.entries) != 0 {
		t.Fatalf("LineChart completed exit = layouts %d progress %v entries %d", tooltipLayouts, state.tooltipTransition.Value(), len(state.tooltipSelection.entries))
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

func queueLineChartClick(router *input.Router, id pointer.ID, position f32.Point) {
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: id, Buttons: pointer.ButtonPrimary, Position: position},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: id, Position: position},
	)
}

func layoutLineChartListFrame(ctx *frame.Context, router *input.Router, list *layout.List, widget Widget, now time.Time) layout.Dimensions {
	viewport := image.Pt(520, 180)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: now}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := list.Layout(gtx, 2, func(gtx layout.Context, index int) layout.Dimensions {
		if index == 0 {
			return widget.Layout(ctx, gtx)
		}
		return layout.Dimensions{Size: image.Pt(gtx.Constraints.Max.X, 320)}
	})
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
