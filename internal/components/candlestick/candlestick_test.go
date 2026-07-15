package candlestick

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
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

func TestCandlestickChartOptionsUseValueSemantics(t *testing.T) {
	values := []Candle{OHLC(10, 12, 9, 13)}
	base := New("market", values)
	values[0].open = 99
	if base.data[0].open != 10 {
		t.Fatal("CandlestickChart retained the caller's data slice")
	}

	categories := []string{"Mon"}
	configured := base.
		Categories(categories).
		Height(280).
		Grid(false).
		Tooltip(false).
		Crosshair(false).
		YRange(8, 14).
		YTicks(6).
		FormatY(func(float64) string { return "price" }).
		XAxis("Date").
		YAxis("Price").
		Width(10).
		MaxWidth(16).
		MinWidth(2).
		UpColor(color.NRGBA{R: 1, A: 0xff}).
		DownColor(color.NRGBA{G: 2, A: 0xff}).
		DojiColor(color.NRGBA{B: 3, A: 0xff}).
		Animation(false).
		AnimationDuration(250*time.Millisecond).
		AnimationEasing(func(value float32) float32 { return value }).
		UpdateAnimationDuration(150*time.Millisecond).
		UpdateAnimationEasing(func(value float32) float32 { return value }).
		OnDataClick(func(chart.Selection) {}).
		TooltipContent(func(chart.Selection) frame.Widget { return nil }).
		DataWindow(.25, .75).
		OnDataWindowChange(func(chart.DataWindow) {}).
		MarkLines([]chart.MarkLine{chart.NewMarkLine(chart.AxisY, 10)}).
		MarkAreas([]chart.MarkArea{chart.NewMarkArea(chart.AxisX, 0, 1)}).
		MarkPoints([]chart.MarkPoint{chart.NewMarkPoint(0, 10)}).
		Label("Market").
		EmptyText("Empty").
		Disabled(true)
	categories[0] = "Changed"
	timestamps := []time.Time{time.Date(2026, time.July, 13, 22, 0, 0, 0, time.UTC)}
	timeConfigured := base.Times(timestamps).FormatTime(func(time.Time) string { return "custom" })
	timestamps[0] = time.Time{}
	if len(timeConfigured.categories) != 0 || len(timeConfigured.times) != 1 || timeConfigured.times[0].IsZero() || timeConfigured.categoryLabel(0) != "custom" {
		t.Fatalf("configured CandlestickChart times = %#v", timeConfigured.times)
	}
	categoryConfigured := timeConfigured.Categories([]string{"Tue"})
	if len(categoryConfigured.times) != 0 || categoryConfigured.categoryLabel(0) != "Tue" {
		t.Fatalf("Categories did not replace CandlestickChart times: %#v", categoryConfigured)
	}

	if len(base.categories) != 0 || base.height != 0 || !base.showGrid || !base.showTooltip || !base.showCrosshair || base.hasYRange || base.width != 0 || !base.animation || base.animationDuration != 300*time.Millisecond || base.disabled {
		t.Fatalf("configuring CandlestickChart mutated base: %#v", base)
	}
	if configured.categories[0] != "Mon" || configured.height != 280 || configured.showGrid || configured.showTooltip || configured.showCrosshair || !configured.hasYRange || configured.yMin != 8 || configured.yMax != 14 || configured.yTickCount != 6 || configured.formatY == nil || configured.xAxisLabel != "Date" || configured.yAxisLabel != "Price" || configured.width != 10 || configured.maxWidth != 16 || configured.minWidth != 2 || !configured.hasUpColor || !configured.hasDownColor || !configured.hasDojiColor || configured.animation || configured.animationDuration != 250*time.Millisecond || configured.animationEasing == nil || configured.updateAnimationDuration != 150*time.Millisecond || configured.updateAnimationEasing == nil || configured.onDataClick == nil || configured.tooltipContent == nil || !configured.hasDataWindow || configured.dataWindow.Start != .25 || configured.dataWindow.End != .75 || configured.onDataWindowChange == nil || len(configured.markLines) != 1 || len(configured.markAreas) != 1 || len(configured.markPoints) != 1 || configured.label != "Market" || configured.emptyText != "Empty" || !configured.disabled {
		t.Fatalf("configured CandlestickChart = %#v", configured)
	}
}

func TestCandlestickChartRejectsInvalidConfiguration(t *testing.T) {
	tests := []struct {
		name string
		run  func()
	}{
		{"height", func() { New("chart", nil).Height(0) }},
		{"Y range", func() { New("chart", nil).YRange(math.NaN(), 1) }},
		{"Y ticks", func() { New("chart", nil).YTicks(1) }},
		{"width", func() { New("chart", nil).Width(0) }},
		{"maximum width", func() { New("chart", nil).MaxWidth(0) }},
		{"minimum width", func() { New("chart", nil).MinWidth(0) }},
		{"animation duration", func() { New("chart", nil).AnimationDuration(-time.Millisecond) }},
		{"update animation duration", func() { New("chart", nil).UpdateAnimationDuration(-time.Millisecond) }},
		{"animation easing", func() { New("chart", nil).AnimationEasing(nil) }},
		{"update animation easing", func() { New("chart", nil).UpdateAnimationEasing(nil) }},
		{"data window", func() { New("chart", nil).DataWindow(.8, .2) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid CandlestickChart configuration did not panic")
				}
			}()
			test.run()
		})
	}
}

func TestCandlestickChartMatchesEChartsSignsAndExtent(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	values := []Candle{
		OHLC(10, 12, 8, 13),
		OHLC(12, 10, 9, 14),
		OHLC(11, 11, 10, 12),
		OHLC(9, 9, 8.5, 10),
		OHLC(math.NaN(), 1, 0, 2),
	}
	data := resolveChartData(New("market", values), &activeTheme)
	if data.extent.Minimum != 8 || data.extent.Maximum != 14 || !data.extent.Valid {
		t.Fatalf("CandlestickChart extent = %#v", data.extent)
	}
	wantSigns := []int{signUp, signDown, signUp, signDown}
	for index, want := range wantSigns {
		if data.candles[index].sign != want {
			t.Fatalf("candle %d sign = %d, want %d", index, data.candles[index].sign, want)
		}
	}
	if data.candles[4].valid {
		t.Fatal("non-finite candle remained visible")
	}

	doji := color.NRGBA{R: 0x80, G: 0x80, B: 0x80, A: 0xff}
	data = resolveChartData(New("market", values).DojiColor(doji), &activeTheme)
	if data.candles[2].sign != signDoji || data.candles[3].sign != signDoji || data.candles[2].color != doji {
		t.Fatalf("configured doji candles = %#v %#v", data.candles[2], data.candles[3])
	}
}

func TestCandlestickChartUsesEChartsWidthAndWindowRules(t *testing.T) {
	if got := New("chart", nil).resolveCandleWidth(20, testDp); got != 10 {
		t.Fatalf("automatic candle width = %v, want 10", got)
	}
	if got := New("chart", nil).MaxWidth(6).resolveCandleWidth(20, testDp); got != 6 {
		t.Fatalf("maximum candle width = %v, want 6", got)
	}
	if got := New("chart", nil).MinWidth(12).resolveCandleWidth(20, testDp); got != 12 {
		t.Fatalf("minimum candle width = %v, want 12", got)
	}
	if got := New("chart", nil).Width(18).resolveCandleWidth(20, testDp); got != 18 {
		t.Fatalf("explicit candle width = %v, want 18", got)
	}
	start, end := visibleCategoryRange(10, chart.DataWindow{Start: .25, End: .75})
	if start != 2 || end != 8 {
		t.Fatalf("visible candle range = %d:%d, want 2:8", start, end)
	}
}

func TestCandlestickChartAutoScalesVisibleWindow(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("market", []Candle{
		OHLC(90, 95, 80, 100),
		OHLC(92, 88, 85, 98),
		OHLC(42, 46, 40, 48),
		OHLC(46, 44, 41, 50),
	}).DataWindow(.5, 1)
	data := resolveChartData(widget, &activeTheme)
	scale := widget.resolveYScale(data)
	if scale.minimum > 40 || scale.maximum < 50 || scale.maximum >= 80 {
		t.Fatalf("visible CandlestickChart Y scale = %#v", scale)
	}
	fixed := widget.YRange(0, 120).resolveYScale(data)
	if fixed.minimum != 0 || fixed.maximum != 120 {
		t.Fatalf("fixed CandlestickChart Y scale = %#v", fixed)
	}
}

func TestCandlestickChartFormatsVisibleTimePeriods(t *testing.T) {
	base := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name   string
		values []time.Time
		want   string
	}{
		{"seconds", []time.Time{base, base.Add(30 * time.Second), base.Add(time.Minute)}, "15:04:05"},
		{"hours", []time.Time{base, base.Add(time.Hour), base.Add(48 * time.Hour)}, "01-02 15:04"},
		{"days", []time.Time{base, base.Add(24 * time.Hour), base.Add(10 * 24 * time.Hour)}, "01-02"},
		{"short DST day", []time.Time{base, base.Add(23 * time.Hour), base.Add(47 * time.Hour)}, "01-02"},
		{"months", []time.Time{base, time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)}, "2006-01"},
		{"years", []time.Time{base, time.Date(2027, time.January, 1, 0, 0, 0, 0, time.UTC), time.Date(2028, time.January, 1, 0, 0, 0, 0, time.UTC)}, "2006"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			widget := New("market", nil).Times(test.values)
			if got := widget.timeAxisFormat(0, len(test.values)); got != test.want {
				t.Fatalf("time axis format = %q, want %q", got, test.want)
			}
		})
	}

	widget := New("market", nil).Times([]time.Time{base.Add(3*time.Hour + 4*time.Minute)})
	if got := widget.categoryLabel(0); got != "2026-01-01 03:04:00" {
		t.Fatalf("full time label = %q", got)
	}
}

func TestCandlestickChartScaleHandlesExtremeFiniteValues(t *testing.T) {
	scale := newLinearScale(-1e308, 1e308, 5, false)
	if !finite(scale.minimum) || !finite(scale.maximum) || !finite(scale.interval) || len(scale.ticks) < 2 {
		t.Fatalf("extreme CandlestickChart scale = %#v", scale)
	}
	if ratio := scale.ratio(0); math.Abs(ratio-.5) > 1e-9 {
		t.Fatalf("extreme CandlestickChart midpoint ratio = %v", ratio)
	}
	if value := scale.valueAt(.5); math.Abs(value) > 1e292 {
		t.Fatalf("extreme CandlestickChart midpoint value = %v", value)
	}
}

func TestCandlestickChartAnimationAndSelectionPreserveOHLC(t *testing.T) {
	activeTheme := theme.DefaultTheme()
	widget := New("market", []Candle{OHLC(10, 14, 8, 16)}).Categories([]string{"Mon"}).Label("Price")
	target := resolveChartData(widget, &activeTheme)
	baseline := candleBaselineData(target)
	if baseline.candles[0].open != 10 || baseline.candles[0].close != 10 || baseline.candles[0].low != 10 || baseline.candles[0].high != 10 {
		t.Fatalf("entry animation baseline = %#v", baseline.candles[0])
	}
	middle := interpolateCandleData(baseline, target, .5)
	if middle.candles[0].open != 10 || middle.candles[0].close != 12 || middle.candles[0].low != 9 || middle.candles[0].high != 13 {
		t.Fatalf("interpolated candle = %#v", middle.candles[0])
	}

	geometry := chartGeometry{plot: image.Rect(0, 0, 100, 100), categoryEnd: 1, bandWidth: 100}
	selection := widget.publicSelection(widget.resolveSelection(target, geometry, 0))
	if selection.Label != "Mon" || selection.Index != 0 || len(selection.Items) != 1 {
		t.Fatalf("CandlestickChart selection = %#v", selection)
	}
	item := selection.Items[0]
	if item.SeriesKey != "market" || item.SeriesLabel != "Price" || item.Y != 14 || item.Open != 10 || item.Close != 14 || item.Low != 8 || item.High != 16 {
		t.Fatalf("CandlestickChart selection item = %#v", item)
	}
}

func TestCandlestickChartLaysOutCustomMarkPointContent(t *testing.T) {
	ctx := candlestickTestContext()
	router := new(input.Router)
	laidOut := false
	markerSize := image.Point{}
	markerColor := color.NRGBA{}
	wantColor := color.NRGBA{R: 0x12, G: 0x80, B: 0x40, A: 0xff}
	content := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		laidOut = true
		markerSize = gtx.Constraints.Min
		markerColor = ctx.ForegroundColor()
		return layout.Dimensions{Size: markerSize}
	})
	widget := New("market", []Candle{OHLC(10, 12, 9, 13)}).
		MarkPoints([]chart.MarkPoint{chart.NewMarkPoint(0, 11).Size(15).Color(wantColor).Content(content)})
	layoutCandlestickFrame(ctx, router, widget, time.Unix(4, 0))
	if !laidOut || markerSize != image.Pt(15, 15) || markerColor != wantColor {
		t.Fatalf("custom mark point = laid out %v size %v color %#v", laidOut, markerSize, markerColor)
	}
}

func TestCandlestickChartTooltipUsesConventionalOHLCOrder(t *testing.T) {
	widget := New("market", nil).FormatY(func(value float64) string { return fmt.Sprintf("$%.2f", value) })
	rows := widget.candlestickTooltipRows(chartSelection{candle: resolvedCandle{open: 10, high: 13, low: 9, close: 12}}, 1)
	if len(rows) != 4 || rows[0] != "Open  $10.00" || rows[1] != "High  $13.00" || rows[2] != "Low  $9.00" || rows[3] != "Close  $12.00" {
		t.Fatalf("CandlestickChart tooltip rows = %#v", rows)
	}
}

func TestCandlestickChartPrunesLabelsAfterClampingToLeftEdge(t *testing.T) {
	ctx := candlestickTestContext()
	viewport := image.Pt(520, 360)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Ops: &ops}
	frame.BeginFrameWithViewport(ctx, viewport)
	tokens := frame.ActiveTheme(ctx).Components.CandlestickChart
	label := "07-07 07:00"
	width := recordChartText(ctx, gtx, label, tokens.AxisTextSize, font.Normal, color.NRGBA{A: 0xff}, 400).dims.Size.X
	plot := image.Rect(100, 0, 500, 300)
	first := float32(plot.Min.X + 2)
	geometry := chartGeometry{plot: plot, xTicks: []categoryTick{
		{label: label, pixel: first},
		{label: label, pixel: first + float32(width+16)},
	}}

	ticks := New("market", nil).pruneCategoryTicks(ctx, gtx, geometry, candlestickStyleFor(frame.ActiveTheme(ctx), false))
	frame.EndFrame(ctx)
	if len(ticks) != 1 {
		t.Fatalf("left-edge time labels were not pruned: %#v", ticks)
	}
}

func TestCandlestickChartPointerSemanticsAndFocus(t *testing.T) {
	ctx := candlestickTestContext()
	router := new(input.Router)
	widget := New("market", []Candle{OHLC(10, 12, 9, 13)}).Label("Market price")
	now := time.Unix(1, 0)
	layoutCandlestickFrame(ctx, router, widget, now)
	state, _ := frame.PeekState[chartState](ctx, "market", stateSlotCandlestickChart)
	if state == nil {
		t.Fatal("CandlestickChart state is missing")
	}
	if !semanticDescriptionExists(router.AppendSemantics(nil), "Market price, 1 candles") {
		t.Fatal("CandlestickChart semantic description is missing")
	}
	router.MoveFocus(key.FocusForward)
	if router.Source().Focused(&state.click) {
		t.Fatal("CandlestickChart entered keyboard focus order")
	}

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(300, 100)})
	layoutCandlestickFrame(ctx, router, widget, now.Add(time.Millisecond))
	if !state.hovered || state.pointer != f32.Pt(300, 100) {
		t.Fatalf("pointer CandlestickChart state = %#v", state)
	}
}

func TestCandlestickChartClickAndDataWindowInteractions(t *testing.T) {
	ctx := candlestickTestContext()
	router := new(input.Router)
	requested := chart.DataWindow{}
	clicked := chart.Selection{}
	widget := New("market", []Candle{
		OHLC(10, 12, 9, 13),
		OHLC(12, 11, 10, 14),
		OHLC(11, 15, 10, 16),
		OHLC(15, 14, 13, 16),
	}).
		Categories([]string{"A", "B", "C", "D"}).
		DataWindow(.2, .8).
		OnDataWindowChange(func(window chart.DataWindow) { requested = window }).
		OnDataClick(func(selection chart.Selection) { clicked = selection })
	now := time.Unix(3, 0)
	layoutCandlestickFrame(ctx, router, widget, now)

	router.Queue(pointer.Event{Kind: pointer.Scroll, Source: pointer.Mouse, PointerID: 1, Position: f32.Pt(300, 150), Scroll: f32.Pt(0, -1)})
	layoutCandlestickFrame(ctx, router, widget, now.Add(time.Millisecond))
	if requested.End-requested.Start >= .6 || requested.Start <= .2 || requested.End >= .8 {
		t.Fatalf("CandlestickChart wheel window request = %#v", requested)
	}

	router.Queue(pointer.Event{Kind: pointer.Move, Source: pointer.Mouse, PointerID: 2, Position: f32.Pt(300, 120)})
	layoutCandlestickFrame(ctx, router, widget, now.Add(2*time.Millisecond))
	queueCandlestickClick(router, 2, f32.Pt(300, 120))
	layoutCandlestickFrame(ctx, router, widget, now.Add(3*time.Millisecond))
	if clicked.Index < 0 || len(clicked.Items) != 1 || clicked.Items[0].Open == 0 {
		t.Fatalf("CandlestickChart click selection = %#v", clicked)
	}

	queueCandlestickClick(router, 3, f32.Pt(300, 120))
	layoutCandlestickFrame(ctx, router, widget, now.Add(4*time.Millisecond))
	if requested != chart.FullDataWindow() {
		t.Fatalf("CandlestickChart double-click window = %#v", requested)
	}
}

func TestCandlestickChartThemeAndSmallLayout(t *testing.T) {
	tokens := theme.DefaultTheme().Components.CandlestickChart
	if tokens.Height != 360 || tokens.CrosshairWidth != 1 || tokens.UpColor != (color.NRGBA{R: 0xeb, G: 0x54, B: 0x54, A: 0xff}) || tokens.DownColor != (color.NRGBA{R: 0x47, G: 0xb2, B: 0x62, A: 0xff}) {
		t.Fatalf("CandlestickChart theme tokens = %#v", tokens)
	}

	ctx := candlestickTestContext()
	router := new(input.Router)
	viewport := image.Pt(24, 18)
	var ops op.Ops
	gtx := layout.Context{Constraints: layout.Exact(viewport), Source: router.Source(), Ops: &ops, Now: time.Unix(2, 0)}
	frame.BeginFrameWithViewport(ctx, viewport)
	dims := New("small", []Candle{OHLC(math.NaN(), 0, 0, 0)}).XAxis("Date").YAxis("Price").Layout(ctx, gtx)
	frame.ApplyFrameCommands(ctx, gtx)
	frame.EndFrame(ctx)
	router.Frame(&ops)
	if dims.Size != viewport {
		t.Fatalf("small CandlestickChart dimensions = %v, want %v", dims.Size, viewport)
	}
}

func testDp(value unit.Dp) int {
	return int(value)
}

func candlestickTestContext() *frame.Context {
	return frame.New(nil, nil, locale.LanguageEnglish)
}

func layoutCandlestickFrame(ctx *frame.Context, router *input.Router, widget Widget, now time.Time) layout.Dimensions {
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

func queueCandlestickClick(router *input.Router, id pointer.ID, position f32.Point) {
	router.Queue(
		pointer.Event{Kind: pointer.Press, Source: pointer.Mouse, PointerID: id, Buttons: pointer.ButtonPrimary, Position: position},
		pointer.Event{Kind: pointer.Release, Source: pointer.Mouse, PointerID: id, Position: position},
	)
}

func semanticDescriptionExists(nodes []input.SemanticNode, description string) bool {
	for _, node := range nodes {
		if node.Desc.Description == description || semanticDescriptionExists(node.Children, description) {
			return true
		}
	}
	return false
}
