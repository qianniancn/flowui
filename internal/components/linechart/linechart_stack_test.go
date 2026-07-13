package linechart

import (
	"image"
	"math"
	"testing"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestLineChartStackOptionsUseValueSemantics(t *testing.T) {
	base := Values("series", "Series", []float64{1})
	configured := base.
		Stack("total").
		StackStrategy(StackPositive).
		StackOrder(StackSeriesDescending)
	if base.stack != "" || base.stackStrategy != StackSameSign || base.stackOrder != StackSeriesAscending {
		t.Fatalf("configuring stack mutated base series: %#v", base)
	}
	if configured.stack != "total" || configured.stackStrategy != StackPositive || configured.stackOrder != StackSeriesDescending {
		t.Fatalf("configured stack series = %#v", configured)
	}
	if configured.Stack("").stack != "" {
		t.Fatal("empty stack name did not disable stacking")
	}

	for _, run := range []func(){
		func() { base.StackStrategy(StackStrategy(9)) },
		func() { base.StackOrder(StackOrder(9)) },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid stack option did not panic")
				}
			}()
			run()
		}()
	}
}

func TestLineChartDefaultStackSeparatesPositiveAndNegativeValues(t *testing.T) {
	data := resolveStackedLineData([]Series{
		Values("first", "First", []float64{3, -3, 0}).Stack("total"),
		Values("second", "Second", []float64{1, -1, 1}).Stack("total"),
		Values("third", "Third", []float64{-2, 2, 1}).Stack("total"),
	}, true)

	assertLinePointValues(t, data.series[0].points, []float64{3, -3, 0})
	assertLinePointValues(t, data.series[1].points, []float64{4, -4, 1})
	assertLinePointValues(t, data.series[2].points, []float64{-2, 2, 2})
	if data.series[1].points[0].stackBase != 3 || data.series[1].points[1].stackBase != -3 || data.series[1].points[2].hasStackBase {
		t.Fatalf("same-sign stack bases = %#v", data.series[1].points)
	}
	if data.series[2].points[0].hasStackBase || data.series[2].points[1].hasStackBase || data.series[2].points[2].stackBase != 1 {
		t.Fatalf("third same-sign stack bases = %#v", data.series[2].points)
	}
	if data.series[1].points[0].rawY != 1 || data.series[1].points[1].rawY != -1 || data.yExtent.minimum != -4 || data.yExtent.maximum != 4 {
		t.Fatalf("stacked raw values or extent = points %#v extent %#v", data.series[1].points, data.yExtent)
	}
}

func TestLineChartStackStrategiesMatchECharts(t *testing.T) {
	tests := []struct {
		name     string
		strategy StackStrategy
		want     []float64
		valid    []bool
	}{
		{name: "same sign", strategy: StackSameSign, want: []float64{-1, 1, 1}, valid: []bool{true, true, true}},
		{name: "all", strategy: StackAll, want: []float64{2, -2, math.NaN()}, valid: []bool{true, true, false}},
		{name: "positive", strategy: StackPositive, want: []float64{2, 1, 1}, valid: []bool{true, true, true}},
		{name: "negative", strategy: StackNegative, want: []float64{-1, -2, 1}, valid: []bool{true, true, true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data := resolveStackedLineData([]Series{
				Values("base", "Base", []float64{3, -3, math.NaN()}).Stack("total"),
				Values("value", "Value", []float64{-1, 1, 1}).Stack("total").StackStrategy(test.strategy),
			}, true)
			points := data.series[1].points
			for index := range test.want {
				if points[index].valid != test.valid[index] {
					t.Fatalf("point %d validity = %v, want %v", index, points[index].valid, test.valid[index])
				}
				if test.valid[index] && points[index].Y != test.want[index] {
					t.Fatalf("point %d = %v, want %v", index, points[index].Y, test.want[index])
				}
				if !test.valid[index] && !math.IsNaN(points[index].Y) {
					t.Fatalf("point %d = %v, want NaN", index, points[index].Y)
				}
			}
		})
	}
}

func TestLineChartStackOrderAndHiddenSeriesRecalculate(t *testing.T) {
	series := []Series{
		Values("first", "First", []float64{1}).Stack("total").StackOrder(StackSeriesDescending).Smoothness(0.2),
		Values("second", "Second", []float64{2}).Stack("total").Smoothness(0.3),
		Values("third", "Third", []float64{3}).Stack("total").Smoothness(0.4),
	}
	data := resolveStackedLineData(series, true)
	assertLinePointValues(t, data.series[0].points, []float64{6})
	assertLinePointValues(t, data.series[1].points, []float64{5})
	assertLinePointValues(t, data.series[2].points, []float64{3})
	if data.series[0].points[0].stackBase != 5 || data.series[0].stackedOnSmooth != 0.3 || data.series[1].points[0].stackBase != 3 || data.series[1].stackedOnSmooth != 0.4 {
		t.Fatalf("descending stack metadata = %#v", data.series)
	}

	series[0] = series[0].Hidden(true)
	data = resolveStackedLineData(series, true)
	if len(data.series) != 2 || len(data.legend) != 3 || !data.legend[0].hidden {
		t.Fatalf("hidden stack series data = %#v", data)
	}
	assertLinePointValues(t, data.series[0].points, []float64{2})
	assertLinePointValues(t, data.series[1].points, []float64{5})
}

func TestLineChartStackUsesCategoryValuesOrNumericIndexes(t *testing.T) {
	series := []Series{
		XY("first", "First", []Point{{X: 0, Y: 10}, {X: 1, Y: 20}}).Stack("total"),
		XY("second", "Second", []Point{{X: 1, Y: 1}, {X: 0, Y: 2}}).Stack("total"),
	}
	categoryData := resolveStackedLineData(series, true)
	assertLinePointValues(t, categoryData.series[1].points, []float64{21, 12})

	numericData := resolveStackedLineData(series, false)
	assertLinePointValues(t, numericData.series[1].points, []float64{11, 22})
}

func TestLineChartStackGroupsRemainIndependent(t *testing.T) {
	data := resolveStackedLineData([]Series{
		Values("a1", "A1", []float64{1}).Stack("a"),
		Values("b1", "B1", []float64{3}).Stack("b"),
		Values("a2", "A2", []float64{2}).Stack("a"),
		Values("b2", "B2", []float64{4}).Stack("b"),
	}, true)
	assertLinePointValues(t, data.series[0].points, []float64{1})
	assertLinePointValues(t, data.series[1].points, []float64{3})
	assertLinePointValues(t, data.series[2].points, []float64{3})
	assertLinePointValues(t, data.series[3].points, []float64{7})
}

func TestLineChartStackAreaUsesPreviousCumulativeBaseline(t *testing.T) {
	data := resolveStackedLineData([]Series{
		Values("first", "First", []float64{10, 20}).Stack("total").Area(true).Smoothness(0.25),
		Values("second", "Second", []float64{5, 6}).Stack("total").Area(true),
	}, true)
	geometry := chartGeometry{
		plot:   image.Rect(0, 0, 100, 100),
		xScale: newLinearScale(0, 1, 2, false, true),
		yScale: newLinearScale(0, 30, 3, false, true),
	}
	segments := seriesPixelSegments(data.series[1], geometry)
	if len(segments) != 1 || len(segments[0].points) != 2 || len(segments[0].stackedOn) != 13 {
		t.Fatalf("stacked area segments = %#v", segments)
	}
	segment := segments[0]
	assertCloseFloat32(t, segment.points[0].Y, 50)
	assertCloseFloat32(t, segment.points[len(segment.points)-1].Y, float32(100.0/30.0*4.0))
	assertCloseFloat32(t, segment.stackedOn[0].Y, float32(100.0/3.0*2.0))
	assertCloseFloat32(t, segment.stackedOn[len(segment.stackedOn)-1].Y, float32(100.0/3.0))
}

func TestLineChartStackAnimationAndSelectionPreserveRawValues(t *testing.T) {
	widget := New("chart", []Series{
		Values("first", "First", []float64{10}).Stack("total"),
		Values("second", "Second", []float64{5}).Stack("total"),
	}).Categories([]string{"A"})
	activeTheme := theme.DefaultTheme()
	target := resolveChartData(widget, &activeTheme, testDp)
	from := lineBaselineData(target, 0)
	midpoint := interpolateLineData(from, target, 0.5, 0)
	if midpoint.series[1].points[0].Y != 7.5 || midpoint.series[1].points[0].stackBase != 5 {
		t.Fatalf("stacked animation midpoint = %#v", midpoint.series[1].points[0])
	}

	geometry := chartGeometry{
		plot:   image.Rect(0, 0, 100, 100),
		xScale: newLinearScale(0, 1, 2, false, true),
		yScale: newLinearScale(0, 20, 4, false, true),
	}
	selection := widget.resolveSelection(target, geometry, 0, true)
	public := widget.publicSelection(selection, geometry)
	if len(public.Items) != 2 || public.Items[0].Y != 10 || public.Items[1].Y != 5 {
		t.Fatalf("stacked public selection = %#v", public)
	}
	if len(selection.entries) != 2 || selection.entries[1].point.Y != 15 || selection.entries[1].pixel.Y != 25 {
		t.Fatalf("stacked visual selection = %#v", selection)
	}
}

func TestLineChartStackAdditionAvoidsDecimalDrift(t *testing.T) {
	if value := addLineStackValue(0.1, 0.2); value != 0.3 {
		t.Fatalf("safe stack addition = %.17g", value)
	}
}

func resolveStackedLineData(series []Series, categories bool) chartData {
	widget := New("chart", series)
	if categories {
		widget = widget.Categories([]string{"A", "B", "C"})
	}
	activeTheme := theme.DefaultTheme()
	return resolveChartData(widget, &activeTheme, testDp)
}

func assertLinePointValues(t *testing.T, points []resolvedPoint, want []float64) {
	t.Helper()
	if len(points) != len(want) {
		t.Fatalf("point count = %d, want %d", len(points), len(want))
	}
	for index := range want {
		if !points[index].valid || points[index].Y != want[index] {
			t.Fatalf("point %d = %#v, want %v", index, points[index], want[index])
		}
	}
}

func assertCloseFloat32(t *testing.T, got, want float32) {
	t.Helper()
	if math.Abs(float64(got-want)) > 0.001 {
		t.Fatalf("value = %v, want %v", got, want)
	}
}
