package barchart

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
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/FlowUI/internal/locale"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func TestBarChartOptionsUseValueSemantics(t *testing.T) {
	values := []float64{1, 2, 3}
	baseSeries := Values("sales", "Sales", values)
	itemColors := []color.NRGBA{{R: 2, A: 0xff}, {G: 3, A: 0xff}}
	values[0] = 99
	if baseSeries.values[0] != 1 {
		t.Fatal("BarChart Series retained the caller's values slice")
	}
	configuredSeries := baseSeries.
		Color(color.NRGBA{R: 1, A: 0xff}).
		ItemColors(itemColors).
		Stack("total").
		Width(18).
		MaxWidth(24).
		MinHeight(2).
		Radius(4).
		Background(true).
		ShowLabels(true).
		LabelPosition(LabelInside).
		FormatLabel(func(float64) string { return "label" }).
		Hidden(true)
	itemColors[0] = color.NRGBA{}
	if baseSeries.hasColor || len(baseSeries.itemColors) != 0 || baseSeries.stack != "" || baseSeries.width != 0 || baseSeries.maxWidth != 0 || baseSeries.minHeight != 0 || baseSeries.hasRadius || baseSeries.showBackground || baseSeries.showLabels || baseSeries.formatLabel != nil || baseSeries.hidden {
		t.Fatalf("configuring BarChart Series mutated base: %#v", baseSeries)
	}
	if !configuredSeries.hasColor || len(configuredSeries.itemColors) != 2 || configuredSeries.itemColors[0].R != 2 || configuredSeries.stack != "total" || configuredSeries.width != 18 || configuredSeries.maxWidth != 24 || configuredSeries.minHeight != 2 || configuredSeries.radius != 4 || !configuredSeries.hasRadius || !configuredSeries.showBackground || !configuredSeries.showLabels || configuredSeries.labelPosition != LabelInside || configuredSeries.formatLabel == nil || !configuredSeries.hidden {
		t.Fatalf("configured BarChart Series = %#v", configuredSeries)
	}

	categories := []string{"Mon", "Tue", "Wed"}
	base := New("sales", []Series{baseSeries})
	configured := base.
		Categories(categories).
		Height(280).
		Grid(false).
		Legend(true).
		Tooltip(false).
		IncludeZero(false).
		YRange(10, 20).
		XAxis("Day").
		YAxis("Sales").
		CategoryAxis("Category").
		ValueAxis("Value").
		YTicks(6).
		BarGap(1.2).
		CategoryGap(0.3).
		FormatY(func(float64) string { return "value" }).
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
		Orientation(Horizontal).
		Label("Sales").
		EmptyText("Empty").
		Disabled(true)
	categories[0] = "Changed"
	if len(base.categories) != 0 || base.height != 0 || !base.showGrid || base.hasShowLegend || !base.showTooltip || !base.includeZero || base.hasYRange || base.hasBarGap || base.hasCategoryGap || !base.animation || base.animationDuration != time.Second || base.updateAnimationDuration != 500*time.Millisecond || base.disabled {
		t.Fatalf("configuring BarChart mutated base: %#v", base)
	}
	if configured.categories[0] != "Mon" || configured.height != 280 || configured.showGrid || !configured.hasShowLegend || !configured.showLegend || configured.showTooltip || configured.includeZero || !configured.hasYRange || configured.yTickCount != 6 || configured.barGap != 1.2 || configured.categoryGap != 0.3 || configured.formatY == nil || !configured.hasCategoryAxisLabel || configured.categoryAxisLabel != "Category" || !configured.hasValueAxisLabel || configured.valueAxisLabel != "Value" || configured.animation || configured.animationDuration != 250*time.Millisecond || configured.animationEasing == nil || configured.updateAnimationDuration != 150*time.Millisecond || configured.updateAnimationEasing == nil || configured.onLegendChange == nil || configured.onDataClick == nil || configured.tooltipContent == nil || !configured.hasDataWindow || configured.dataWindow.Start != 0.25 || configured.dataWindow.End != 0.75 || configured.onDataWindowChange == nil || len(configured.markLines) != 1 || len(configured.markAreas) != 1 || len(configured.markPoints) != 1 || configured.orientation != Horizontal || configured.label != "Sales" || configured.emptyText != "Empty" || !configured.disabled {
		t.Fatalf("configured BarChart = %#v", configured)
	}
}

func TestBarChartRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{"series width", func() { Values("a", "A", nil).Width(0) }},
		{"series max width", func() { Values("a", "A", nil).MaxWidth(-1) }},
		{"series min height", func() { Values("a", "A", nil).MinHeight(-1) }},
		{"series radius", func() { Values("a", "A", nil).Radius(-1) }},
		{"label position", func() { Values("a", "A", nil).LabelPosition(LabelPosition(9)) }},
		{"height", func() { New("chart", nil).Height(0) }},
		{"Y range", func() { New("chart", nil).YRange(math.NaN(), 1) }},
		{"Y ticks", func() { New("chart", nil).YTicks(1) }},
		{"negative bar gap", func() { New("chart", nil).BarGap(-0.1) }},
		{"large category gap", func() { New("chart", nil).CategoryGap(1) }},
		{"NaN gap", func() { New("chart", nil).BarGap(float32(math.NaN())) }},
		{"animation duration", func() { New("chart", nil).AnimationDuration(-time.Millisecond) }},
		{"update animation duration", func() { New("chart", nil).UpdateAnimationDuration(-time.Millisecond) }},
		{"animation easing", func() { New("chart", nil).AnimationEasing(nil) }},
		{"update animation easing", func() { New("chart", nil).UpdateAnimationEasing(nil) }},
		{"data window", func() { New("chart", nil).DataWindow(0.8, 0.2) }},
		{"orientation", func() { New("chart", nil).Orientation(Orientation(9)) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid BarChart configuration did not panic")
				}
			}()
			test.run()
		})
	}
}

func TestBarChartDataWindowConstrainsVisibleCategories(t *testing.T) {
	start, end := visibleCategoryRange(10, chart.DataWindow{Start: 0.25, End: 0.75})
	if start != 2 || end != 8 {
		t.Fatalf("BarChart visible category range = %d:%d, want 2:8", start, end)
	}
}

func TestBarChartAnnotationGeometryUsesVisibleScales(t *testing.T) {
	geometry := chartGeometry{
		plot:          image.Rect(10, 20, 110, 120),
		yScale:        newLinearScale(0, 100, 5, false, true),
		bandWidth:     25,
		categoryStart: 2,
		categoryEnd:   6,
	}
	rect, ok := barMarkAreaRect(chart.NewMarkArea(chart.AxisY, 20, 40), geometry)
	if !ok || rect != image.Rect(10, 80, 110, 100) {
		t.Fatalf("BarChart mark area = %v, ok %v", rect, ok)
	}
	from, to, ok := barMarkEndpoints(chart.NewMarkLine(chart.AxisX, 3), geometry)
	if !ok || from != f32.Pt(47.5, 20) || to != f32.Pt(47.5, 120) {
		t.Fatalf("BarChart mark line = %v to %v, ok %v", from, to, ok)
	}
}

func TestHorizontalBarChartGeometryAndRectangles(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("horizontal", []Series{Values("values", "Values", []float64{25, 50, -20, 75})}).
		Categories([]string{"A", "B", "C", "D"}).
		Orientation(Horizontal)
	data := resolveChartData(widget, &activeTheme, testDp)
	geometry := widget.resolveGeometry(data, image.Pt(120, 120), image.Rect(10, 10, 110, 110))
	if !geometry.horizontal || geometry.bandWidth != 25 || geometry.categoryCenter(2) != 72.5 {
		t.Fatalf("horizontal BarChart geometry = %#v", geometry)
	}
	column := columnLayout{offset: -5, width: 10}
	positive := barRectangle(geometry, column, 1, data.series[0].values[1], 0)
	negative := barRectangle(geometry, column, 2, data.series[0].values[2], 0)
	zero := int(math.Round(float64(geometry.mapY(0))))
	if positive.Min.X != zero || positive.Max.X <= positive.Min.X || positive.Min.Y != 43 || positive.Max.Y != 53 {
		t.Fatalf("positive horizontal bar = %v, zero %d", positive, zero)
	}
	if negative.Max.X != zero || negative.Min.X >= negative.Max.X || negative.Min.Y != 68 || negative.Max.Y != 78 {
		t.Fatalf("negative horizontal bar = %v, zero %d", negative, zero)
	}
}

func TestBarChartAxisLabelsFollowOrientation(t *testing.T) {
	widget := New("chart", nil).CategoryAxis("Category").ValueAxis("Value")
	if x, y := widget.axisLabels(); x != "Category" || y != "Value" {
		t.Fatalf("vertical axis labels = %q, %q", x, y)
	}
	if x, y := widget.Orientation(Horizontal).axisLabels(); x != "Value" || y != "Category" {
		t.Fatalf("horizontal axis labels = %q, %q", x, y)
	}
}

func TestBarChartAnimationInterpolatesStackGeometry(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	target := resolveChartData(New("chart", []Series{
		Values("first", "First", []float64{10}).Stack("total"),
		Values("second", "Second", []float64{5}).Stack("total"),
	}), &activeTheme, testDp)
	from := barBaselineData(target)
	midpoint := interpolateBarData(from, target, 0.5)
	first := midpoint.series[0].values[0]
	second := midpoint.series[1].values[0]
	if first.start != 0 || first.end != 5 || second.start != 5 || second.end != 7.5 {
		t.Fatalf("animated BarChart stack midpoint = first %#v second %#v", first, second)
	}
	if target.series[0].values[0].end != 10 || target.series[1].values[0].end != 15 {
		t.Fatalf("BarChart animation mutated target = %#v", target.series)
	}
}

func TestBarChartUpdateAnimationStartsFromDisplayedData(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	previous := resolveChartData(New("chart", []Series{Values("series", "Series", []float64{4})}), &activeTheme, testDp)
	target := resolveChartData(New("chart", []Series{Values("series", "Series", []float64{10})}), &activeTheme, testDp)
	from := barTransitionFrom(previous, target)
	if from.series[0].values[0].value != 4 || from.series[0].values[0].end != 4 {
		t.Fatalf("BarChart update start = %#v", from.series[0].values[0])
	}
}

func TestResolveChartDataGroupsAndStacksSameSignValues(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	data := resolveChartData(New("chart", []Series{
		Values("desktop", "Desktop", []float64{10, -4, math.NaN()}).Stack("traffic"),
		Values("mobile", "Mobile", []float64{5, -3, 2}).Stack("traffic"),
		Values("direct", "Direct", []float64{7, 8}),
	}).Categories([]string{"A", "B", "C"}), &activeTheme, testDp)
	if data.categories != 3 || len(data.series) != 3 || len(data.columns) != 2 {
		t.Fatalf("resolved BarChart data = %#v", data)
	}
	if first, second := data.series[0].values[0], data.series[1].values[0]; first.start != 0 || first.end != 10 || second.start != 10 || second.end != 15 {
		t.Fatalf("positive stack = %#v %#v", first, second)
	}
	if first, second := data.series[0].values[1], data.series[1].values[1]; first.start != 0 || first.end != -4 || second.start != -4 || second.end != -7 {
		t.Fatalf("negative stack = %#v %#v", first, second)
	}
	if data.series[0].values[2].valid || data.series[1].values[2].start != 0 || data.series[1].values[2].end != 2 {
		t.Fatalf("stack with missing value = %#v", data.series)
	}
	if data.yExtent.Minimum != -7 || data.yExtent.Maximum != 15 {
		t.Fatalf("stack extent = %#v", data.yExtent)
	}
}

func TestResolveChartDataAppliesItemColors(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	red := color.NRGBA{R: 0xff, A: 0xff}
	data := resolveChartData(New("chart", []Series{
		Values("values", "Values", []float64{1, 2, 3}).ItemColors([]color.NRGBA{red}),
	}), &activeTheme, testDp)
	if data.series[0].values[0].color != red {
		t.Fatalf("first item color = %#v", data.series[0].values[0].color)
	}
	if data.series[0].values[1].color != data.series[0].color || data.series[0].values[2].color != data.series[0].color {
		t.Fatalf("missing item colors did not use series color: %#v", data.series[0].values)
	}
}

func TestResolveChartDataIgnoresHiddenSeriesForCategoryCount(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	data := resolveChartData(New("chart", []Series{
		Values("visible", "Visible", []float64{1, 2}),
		Values("hidden", "Hidden", []float64{1, 2, 3, 4}).Hidden(true),
	}), &activeTheme, testDp)
	if data.categories != 2 || len(data.series) != 1 || len(data.legend) != 2 || !data.legend[1].hidden || len(data.columns) != 1 {
		t.Fatalf("hidden BarChart series affected layout: %#v", data)
	}
}

func TestIncludeZeroControlsAutomaticExtent(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	data := resolveChartData(New("chart", []Series{Values("values", "Values", []float64{10, 20})}).IncludeZero(false), &activeTheme, testDp)
	if data.yExtent.Minimum != 10 || data.yExtent.Maximum != 20 {
		t.Fatalf("scale extent includes zero when disabled: %#v", data.yExtent)
	}
	data = resolveChartData(New("chart", []Series{Values("values", "Values", []float64{10, 20})}), &activeTheme, testDp)
	if data.yExtent.Minimum != 0 || data.yExtent.Maximum != 20 {
		t.Fatalf("default scale extent excludes zero: %#v", data.yExtent)
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
					t.Fatal("invalid BarChart series keys did not panic")
				}
			}()
			resolveChartData(New("chart", series), &activeTheme, testDp)
		}()
	}
}

func TestColumnLayoutUsesEChartsCategoryAndBarGaps(t *testing.T) {
	columns := []barColumn{{id: "first"}, {id: "second"}}
	layouts := resolveColumnLayouts(columns, 100, 0.1, defaultCategoryGap(len(columns)))
	first, second := layouts["first"], layouts["second"]
	wantWidth := float32(73.0 / 2.1)
	if math.Abs(float64(first.width-wantWidth)) > 1e-4 || first.width != second.width {
		t.Fatalf("automatic bar widths = %#v", layouts)
	}
	if math.Abs(float64(first.offset+36.5)) > 1e-4 || math.Abs(float64(second.offset-(first.offset+first.width*1.1))) > 1e-4 {
		t.Fatalf("automatic bar offsets = %#v", layouts)
	}
	stacked := resolveColumnLayouts([]barColumn{{id: "stack"}}, 100, 0.1, defaultCategoryGap(1))
	if math.Abs(float64(stacked["stack"].width-69)) > 1e-4 || math.Abs(float64(stacked["stack"].offset+34.5)) > 1e-4 {
		t.Fatalf("stacked column layout = %#v", stacked)
	}
	mixed := resolveColumnLayouts([]barColumn{{id: "fixed", width: 30, maxWidth: 20}, {id: "auto"}}, 100, 0.1, 0.2)
	if mixed["fixed"].width != 20 || math.Abs(float64(mixed["auto"].width-58)) > 1e-4 {
		t.Fatalf("mixed explicit and automatic layout = %#v", mixed)
	}
}

func TestBarRectangleHonorsMinimumHeight(t *testing.T) {
	geometry := chartGeometry{
		plot:      image.Rect(0, 0, 100, 100),
		yScale:    newLinearScale(-10, 10, 4, false, true),
		bandWidth: 100,
	}
	column := columnLayout{offset: -10, width: 20}
	positive := barRectangle(geometry, column, 0, resolvedBar{value: 0.01, start: 0, end: 0.01, valid: true}, 4)
	negative := barRectangle(geometry, column, 0, resolvedBar{value: -0.01, start: 0, end: -0.01, valid: true}, 4)
	if positive.Dy() != 4 || negative.Dy() != 4 || positive.Max.Y != negative.Min.Y {
		t.Fatalf("minimum-height bars = positive %v negative %v", positive, negative)
	}
}

func TestBarChartSelectionUsesCategoryIndex(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("chart", []Series{
		Values("first", "First", []float64{10, 20}),
		Values("second", "Second", []float64{12, math.NaN()}),
	})
	data := resolveChartData(widget, &activeTheme, testDp)
	geometry := chartGeometry{plot: image.Rect(0, 0, 100, 100), bandWidth: 50, categoryEnd: 2}
	selection := widget.resolveSelection(data, geometry, 1, true)
	if selection.pixelX != 75 || len(selection.entries) != 1 || selection.entries[0].bar.value != 20 {
		t.Fatalf("BarChart selection = %#v", selection)
	}
}

func TestBarChartPointerSelection(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	widget := New("chart", []Series{Values("series", "Series", []float64{10, 20, 30})}).Categories([]string{"A", "B", "C"})
	now := time.Unix(1, 0)
	layoutBarChartFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotBarChart)
	if state == nil {
		t.Fatal("BarChart state is missing")
	}
	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(300, 100)})
	layoutBarChartFrame(ctx, router, widget, now.Add(time.Millisecond))
	if !state.hovered || state.pointer != f32.Pt(300, 100) {
		t.Fatalf("pointer BarChart state = %#v", state)
	}
}

func TestBarChartRootDoesNotEnterKeyboardFocusOrder(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	widget := New("chart", []Series{
		Values("first", "First", []float64{10, 20}),
		Values("second", "Second", []float64{12, 24}),
	}).OnLegendChange(func(string, bool) {})
	layoutBarChartFrame(ctx, router, widget, time.Unix(1, 0))
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotBarChart)

	router.MoveFocus(key.FocusForward)
	if router.Source().Focused(&state.click) {
		t.Fatal("BarChart root entered keyboard focus order")
	}
	for key, item := range state.legendItems {
		if router.Source().Focused(item) {
			t.Fatalf("BarChart legend item %q entered keyboard focus order", key)
		}
	}
}

func TestBarChartDoubleClickResetsDataWindow(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	requested := chart.DataWindow{}
	widget := New("chart", []Series{Values("series", "Series", []float64{1, 2, 3, 4})}).
		DataWindow(.2, .8).
		OnDataWindowChange(func(window chart.DataWindow) { requested = window })
	start := time.Unix(2, 0)
	layoutBarChartFrame(ctx, router, widget, start)
	queueBarChartClick(router, 1, f32.Pt(300, 140))
	layoutBarChartFrame(ctx, router, widget, start.Add(time.Millisecond))
	queueBarChartClick(router, 2, f32.Pt(300, 140))
	layoutBarChartFrame(ctx, router, widget, start.Add(2*time.Millisecond))
	if requested != chart.FullDataWindow() {
		t.Fatalf("BarChart double-click window = %#v", requested)
	}
}

func TestHorizontalBarChartWheelRequestsControlledDataWindow(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	requested := chart.DataWindow{}
	widget := New("chart", []Series{Values("series", "Series", []float64{1, 2, 3, 4})}).
		Orientation(Horizontal).
		DataWindow(0.2, 0.8).
		OnDataWindowChange(func(window chart.DataWindow) { requested = window })
	now := time.Unix(8, 0)
	layoutBarChartFrame(ctx, router, widget, now)
	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(300, 150), Scroll: f32.Pt(0, -1)})
	layoutBarChartFrame(ctx, router, widget, now.Add(time.Millisecond))
	if requested.End-requested.Start >= 0.6 || requested.Start <= 0.2 || requested.End >= 0.8 {
		t.Fatalf("horizontal BarChart wheel window request = %#v", requested)
	}
}

func TestBarChartWithoutDataWindowDoesNotConsumeParentScroll(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	list := &layout.List{Axis: layout.Vertical}
	widget := New("chart", []Series{Values("series", "Series", []float64{1, 2, 3})})
	now := time.Unix(9, 0)
	layoutBarChartListFrame(ctx, router, list, widget, now)

	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, Position: f32.Pt(300, 100), Scroll: f32.Pt(0, 80)})
	layoutBarChartListFrame(ctx, router, list, widget, now.Add(time.Millisecond))
	if list.Position.First == 0 && list.Position.Offset == 0 {
		t.Fatal("BarChart tooltip interaction consumed the parent list scroll")
	}
}

func TestBarChartLegendClickDataClickAndCustomTooltip(t *testing.T) {
	ctx := barChartTestContext()
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
	layoutBarChartFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotBarChart)
	state.legendItems["hidden"].Click()
	layoutBarChartFrame(ctx, router, widget, now.Add(time.Millisecond))
	if legendKey != "hidden" || legendHidden {
		t.Fatalf("BarChart legend request = key %q hidden %v, want hidden false", legendKey, legendHidden)
	}

	queueBarChartClick(router, 1, f32.Pt(300, 140))
	layoutBarChartFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	if clicked.Label == "" || clicked.Index < 0 || len(clicked.Items) != 1 || clicked.Items[0].SeriesKey != "visible" {
		t.Fatalf("BarChart click selection = %#v", clicked)
	}
	if !tooltipLaidOut {
		t.Fatal("BarChart custom tooltip was not laid out")
	}

	clicked = chart.Selection{}
	tooltipLaidOut = false
	queueBarChartClick(router, 2, f32.Pt(300, 140))
	layoutBarChartFrame(ctx, router, widget.Tooltip(false), now.Add(3*time.Millisecond))
	if len(clicked.Items) != 1 {
		t.Fatalf("BarChart data click with tooltip disabled = %#v", clicked)
	}
	if tooltipLaidOut {
		t.Fatal("BarChart laid out custom tooltip while disabled")
	}
}

func TestBarChartTooltipDisabledClearsInteraction(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	widget := New("chart", []Series{Values("series", "Series", []float64{10, 20})})
	now := time.Unix(4, 0)
	layoutBarChartFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotBarChart)
	state.hovered = true
	layoutBarChartFrame(ctx, router, widget.Tooltip(false), now.Add(time.Millisecond))
	if state.hovered {
		t.Fatalf("disabled BarChart tooltip retained interaction state: %#v", state)
	}
	if state.tooltipTransition.Value() != 0 || len(state.tooltipSelection.entries) != 0 {
		t.Fatalf("disabled BarChart retained tooltip presentation state: %#v", state)
	}
}

func TestBarChartTooltipAnchorTracksPointer(t *testing.T) {
	if got := barTooltipAnchor(f32.Pt(100.4, 79.6)); got != image.Rect(100, 80, 101, 81) {
		t.Fatalf("BarChart tooltip anchor = %v", got)
	}
}

func TestBarChartTooltipAnimatesOutWithLastSelection(t *testing.T) {
	ctx := barChartTestContext()
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
	layoutBarChartFrame(ctx, router, widget, start)
	state, _ := frame.PeekState[chartState](ctx, "chart", stateSlotBarChart)
	state.hovered = true
	state.pointer = f32.Pt(300, 140)
	layoutBarChartFrame(ctx, router, widget, start.Add(time.Millisecond))
	layoutBarChartFrame(ctx, router, widget, start.Add(201*time.Millisecond))

	tooltipLayouts = 0
	state.hovered = false
	layoutBarChartFrame(ctx, router, widget, start.Add(202*time.Millisecond))
	if tooltipLayouts != 1 || !state.tooltipTransition.Exiting() || len(state.tooltipSelection.entries) == 0 {
		t.Fatalf("BarChart exit state = layouts %d exiting %v entries %d", tooltipLayouts, state.tooltipTransition.Exiting(), len(state.tooltipSelection.entries))
	}

	tooltipLayouts = 0
	layoutBarChartFrame(ctx, router, widget, start.Add(302*time.Millisecond))
	if tooltipLayouts != 0 || state.tooltipTransition.Value() != 0 || len(state.tooltipSelection.entries) != 0 {
		t.Fatalf("BarChart completed exit = layouts %d progress %v entries %d", tooltipLayouts, state.tooltipTransition.Value(), len(state.tooltipSelection.entries))
	}
}

func TestBarChartThemeAndSemantics(t *testing.T) {
	tokens := theme.DefaultTheme().Components.BarChart
	if tokens.Height != 320 || tokens.BarRadius != 0 || tokens.SeriesColors[0] != (color.NRGBA{R: 0x50, G: 0x70, B: 0xdd, A: 0xff}) {
		t.Fatalf("BarChart theme tokens = %#v", tokens)
	}
	ctx := barChartTestContext()
	router := new(input.Router)
	layoutBarChartFrame(ctx, router, New("chart", []Series{Values("series", "Series", []float64{1})}).Label("Quarterly sales"), time.Unix(2, 0))
	if !semanticDescriptionExists(router.AppendSemantics(nil), "Quarterly sales, 1 series, 1 categories") {
		t.Fatal("BarChart semantic description is missing")
	}
}

func TestBarChartHandlesSmallConstraintsAndEmptyData(t *testing.T) {
	ctx := barChartTestContext()
	router := new(input.Router)
	widget := New("small", []Series{Values("series", "Series", []float64{math.NaN()})}).XAxis("X").YAxis("Y")
	viewport := image.Pt(24, 18)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: time.Unix(3, 0)}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := widget.Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	if dims.Size != viewport {
		t.Fatalf("small BarChart dimensions = %v, want %v", dims.Size, viewport)
	}
}

func testDp(value unit.Dp) int {
	return int(value)
}

func barChartTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutBarChartFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
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

func queueBarChartClick(router *input.Router, id pointer.ID, position f32.Point) {
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: id, Buttons: pointer.ButtonPrimary, Position: position},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: id, Position: position},
	)
}

func layoutBarChartListFrame(ctx *frame.Context, router *input.Router, list *layout.List, widget Widget, now time.Time) layout.Dimensions {
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
