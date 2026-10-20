package heatmap

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/i18n"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
	"github.com/qianniancn/flowui/widget/tooltip"
)

type Widget struct {
	override                        *theme.HeatmapOverride
	key                             string
	xLabels, yLabels                []string
	calendarStart, calendarEnd      time.Time
	values                          [][]float64
	height                          unit.Dp
	showGrid, showTooltip, disabled bool
	min, max                        float64
	hasRange                        bool
	heatmapOverride                 theme.HeatmapOverride
	label, emptyText                string
	hasEmptyText                    bool
	onDataClick                     func(chart.Selection)
	tooltipContent                  func(chart.Selection) core.Widget
	visualCache                     *heatmapVisualCache
	cellLabels                      [][]string
}

type CalendarValue struct {
	Date  time.Time
	Value float64
}

// New constructs an ECharts-style calendar heatmap from date/value pairs.
func New(key string, start, end time.Time, values []CalendarValue) Widget {
	if key == "" {
		panic("flowui: empty heatmap key")
	}
	if start.IsZero() || end.IsZero() {
		panic("flowui: heatmap calendar end must not be before start")
	}
	start = calendarDate(start)
	endYear, endMonth, endDay := end.Date()
	end = time.Date(endYear, endMonth, endDay, 0, 0, 0, 0, start.Location())
	if end.Before(start) {
		panic("flowui: heatmap calendar end must not be before start")
	}
	gridStart := start.AddDate(0, 0, -int(start.Weekday()))
	gridEnd := end.AddDate(0, 0, 6-int(end.Weekday()))
	weeks := int(gridEnd.Sub(gridStart).Hours()/24/7) + 1
	matrix := make([][]float64, 7)
	labels := make([][]string, 7)
	for weekday := range matrix {
		matrix[weekday] = make([]float64, weeks)
		labels[weekday] = make([]string, weeks)
		for week := range matrix[weekday] {
			matrix[weekday][week] = math.NaN()
		}
	}
	startKey, endKey := dayKeyOf(start), dayKeyOf(end)
	byDate := make(map[dayKey]float64, len(values))
	for _, item := range values {
		if item.Date.IsZero() {
			continue
		}
		key := dayKeyOf(item.Date)
		if !key.before(startKey) && !endKey.before(key) {
			byDate[key] = item.Value
		}
	}
	for offset := 0; offset < weeks*7; offset++ {
		date := gridStart.AddDate(0, 0, offset)
		weekday, week := int(date.Weekday()), offset/7
		labels[weekday][week] = date.Format("2006-01-02")
		if value, ok := byDate[dayKeyOf(date)]; ok {
			matrix[weekday][week] = value
		}
	}
	return Widget{
		key:           key,
		xLabels:       make([]string, weeks),
		yLabels:       make([]string, 7),
		calendarStart: start,
		calendarEnd:   end,
		values:        matrix,
		cellLabels:    labels,
		showGrid:      true,
		showTooltip:   true,
		visualCache:   &heatmapVisualCache{},
	}
}

func calendarDate(value time.Time) time.Time {
	if value.IsZero() {
		return time.Time{}
	}
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

// dayKey identifies a civil calendar day independent of *time.Location, so a
// value whose Date carries a different location than the range bounds still maps
// to the correct cell. A time.Time map key would otherwise compare its location,
// dropping same-day values across zones.
type dayKey struct {
	year  int
	month time.Month
	day   int
}

func dayKeyOf(value time.Time) dayKey {
	year, month, day := value.Date()
	return dayKey{year: year, month: month, day: day}
}

func (k dayKey) before(other dayKey) bool {
	if k.year != other.year {
		return k.year < other.year
	}
	if k.month != other.month {
		return k.month < other.month
	}
	return k.day < other.day
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: heatmap height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}
func (w Widget) Grid(show bool) Widget    { w.showGrid = show; return w }
func (w Widget) Tooltip(show bool) Widget { w.showTooltip = show; return w }
func (w Widget) ValueRange(minimum, maximum float64) Widget {
	if !chart.Finite(minimum) || !chart.Finite(maximum) || maximum <= minimum {
		panic("flowui: heatmap maximum must be greater than minimum")
	}
	w.min, w.max, w.hasRange = minimum, maximum, true
	return w
}
func (w Widget) MinColor(value color.NRGBA) Widget {
	w.heatmapOverride.MinColor = value
	w.heatmapOverride.HasMinColor = true
	return w
}
func (w Widget) MaxColor(value color.NRGBA) Widget {
	w.heatmapOverride.MaxColor = value
	w.heatmapOverride.HasMaxColor = true
	return w
}
func (w Widget) Label(label string) Widget { w.label = label; return w }
func (w Widget) EmptyText(text string) Widget {
	w.emptyText = text
	w.hasEmptyText = true
	return w
}
func (w Widget) OnDataClick(fn func(chart.Selection)) Widget { w.onDataClick = fn; return w }
func (w Widget) TooltipContent(fn func(chart.Selection) core.Widget) Widget {
	w.tooltipContent = fn
	return w
}
func (w Widget) Disabled(disabled bool) Widget { w.disabled = disabled; return w }

type heatState struct {
	click             gesture.Click
	tag               struct{}
	hovered           bool
	pointer           f32.Point
	tooltipTransition tooltip.PopupTransition
	selection         cellSelection
}
type cellSelection struct {
	x, y  int
	value float64
	label string
	color color.NRGBA
	valid bool
}

func stateFor(ctx *core.Context, key string) *heatState {
	key = core.ClaimKey(ctx, core.KindHeatmap, key)
	return core.UseState[heatState](ctx, key, "heatmap")
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.HeatmapOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	active := core.ActiveTheme(ctx)
	visual := w.resolveVisual(active, override)
	resolved := computed.ResolveStatic(ctx, style.State{Disabled: w.disabled}, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	return layoutui.Resolved(ctx, gtx, resolved, heatmapVisualWidget{value: w, visual: visual})
}

type heatmapVisualWidget struct {
	value  Widget
	visual *theme.HeatmapVisual
}

func (v heatmapVisualWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return v.value.layout(ctx, gtx, *v.visual)
}

func (w Widget) resolveVisual(active *theme.Theme, override *theme.HeatmapOverride) *theme.HeatmapVisual {
	requestOverride := w.providerOverride(override)
	load := func(value *theme.HeatmapOverride) *theme.HeatmapVisual {
		visual := theme.ResolveVisual(active, theme.HeatmapRequest{Override: value},
			theme.ComponentVisualProvider.Heatmap, theme.ValidateHeatmapVisual)
		return &visual
	}
	if requestOverride != nil || w.visualCache == nil {
		return load(requestOverride)
	}
	return w.visualCache.cache.Resolve(active, struct{}{}, func() *theme.HeatmapVisual {
		return load(nil)
	})
}

func (w Widget) providerOverride(override *theme.HeatmapOverride) *theme.HeatmapOverride {
	if !w.heatmapOverride.HasMinColor && !w.heatmapOverride.HasMaxColor {
		return override
	}
	combined := w.heatmapOverride
	if override != nil {
		combined.Declaration = override.Declaration
		if override.HasMinColor {
			combined.MinColor = override.MinColor
			combined.HasMinColor = true
		}
		if override.HasMaxColor {
			combined.MaxColor = override.MaxColor
			combined.HasMaxColor = true
		}
	}
	return &combined
}

func (w Widget) layout(ctx *core.Context, gtx layout.Context, visual theme.HeatmapVisual) layout.Dimensions {
	s := stateFor(ctx, w.key)
	enabled := gtx.Enabled() && !w.disabled
	metrics := visual.Metrics
	height := metrics.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}
	xLabels, yLabels := w.calendarLabels(ctx)
	rows, cols := len(yLabels), len(xLabels)
	for _, row := range w.values {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if rows == 0 || cols == 0 {
		emptyText := w.emptyText
		if !w.hasEmptyText {
			emptyText = ctx.Text(i18n.ChartNoData)
		}
		w.drawEmpty(ctx, gtx, size, emptyText, visual)
		return layout.Dimensions{Size: size}
	}
	plotLeft := gtx.Dp(metrics.PlotPaddingLeft)
	plotTop := gtx.Dp(metrics.PlotPaddingTop)
	maxPlotWidth := max(size.X-plotLeft-gtx.Dp(metrics.PlotPaddingRight), 1)
	maxPlotHeight := max(size.Y-plotTop-gtx.Dp(metrics.PlotPaddingBottom), 1)
	cellSize := max(gtx.Dp(metrics.CellSize), 1)
	cellSize = min(cellSize, maxPlotWidth/cols)
	cellSize = min(cellSize, maxPlotHeight/rows)
	if cellSize <= 0 {
		return layout.Dimensions{Size: size}
	}
	plot := image.Rect(plotLeft, plotTop, plotLeft+cellSize*cols, plotTop+cellSize*rows)
	minValue, maxValue := w.valueExtent(rows, cols)
	cellW, cellH := float32(cellSize), float32(cellSize)
	updatePointer(gtx, enabled, plot, &s.tag, &s.hovered, &s.pointer)
	s.selection = w.selection(s.pointer, s.hovered, plot, cellW, cellH, rows, cols, minValue, maxValue, visual.Scale)
	activated := updateClick(gtx, enabled, &s.click)
	if activated && s.selection.valid && w.onDataClick != nil {
		w.onDataClick(w.publicSelection(s.selection, labelAt(xLabels, s.selection.x), labelAt(yLabels, s.selection.y)))
	}
	if enabled && w.showTooltip && s.selection.valid {
		s.tooltipTransition.ProgressWithContext(ctx, gtx, true)
	} else if !s.selection.valid || !w.showTooltip {
		s.tooltipTransition.ProgressWithContext(ctx, gtx, false)
	}
	opacity := paint.PushOpacity(gtx.Ops, 1)
	w.draw(ctx, gtx, plot, rows, cols, cellW, cellH, minValue, maxValue, visual.Scale, s.selection, metrics, visual.Colors, xLabels, yLabels)
	if w.showTooltip && (s.selection.valid || s.tooltipTransition.Value() > 0) && enabled {
		w.drawTooltip(ctx, s, s.selection, chart.TooltipAnchor(s.pointer), xLabels, yLabels, visual)
	}
	opacity.Pop()
	addPointerInput(gtx, plot, enabled && (w.showTooltip || w.onDataClick != nil), &s.tag)
	addClickInput(gtx, size, enabled && w.onDataClick != nil, &s.click)
	return layout.Dimensions{Size: size}
}

func (w Widget) valueExtent(rows, cols int) (float64, float64) {
	if w.hasRange {
		return w.min, w.max
	}
	minValue, maxValue := math.Inf(1), math.Inf(-1)
	for y := range rows {
		if y >= len(w.values) {
			continue
		}
		for x := 0; x < cols && x < len(w.values[y]); x++ {
			v := w.values[y][x]
			if chart.Finite(v) {
				minValue = math.Min(minValue, v)
				maxValue = math.Max(maxValue, v)
			}
		}
	}
	if math.IsInf(minValue, 1) {
		return 0, 1
	}
	if minValue == maxValue {
		return minValue, minValue + 1
	}
	return minValue, maxValue
}

func (w Widget) calendarLabels(ctx *core.Context) ([]string, []string) {
	if w.calendarStart.IsZero() || w.calendarEnd.IsZero() {
		return w.xLabels, w.yLabels
	}
	xLabels := make([]string, len(w.xLabels))
	for week := range xLabels {
		date := w.calendarStart.AddDate(0, 0, week*7-int(w.calendarStart.Weekday()))
		weekEnd := date.AddDate(0, 0, 6)
		for day := date; !day.After(weekEnd); day = day.AddDate(0, 0, 1) {
			monthStart := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
			if day.Day() == 1 && !monthStart.Before(w.calendarStart) && !monthStart.After(w.calendarEnd) {
				xLabels[week] = monthLabel(ctx, monthStart.Month())
				break
			}
		}
	}
	yLabels := make([]string, 7)
	for index := range yLabels {
		yLabels[index] = weekdayLabel(ctx, index)
	}
	return xLabels, yLabels
}

func monthLabel(ctx *core.Context, month time.Month) string {
	keys := [...]i18n.Key{
		i18n.CalendarMonthJanuary, i18n.CalendarMonthFebruary, i18n.CalendarMonthMarch,
		i18n.CalendarMonthApril, i18n.CalendarMonthMay, i18n.CalendarMonthJune,
		i18n.CalendarMonthJuly, i18n.CalendarMonthAugust, i18n.CalendarMonthSeptember,
		i18n.CalendarMonthOctober, i18n.CalendarMonthNovember, i18n.CalendarMonthDecember,
	}
	if month < time.January || month > time.December {
		return ""
	}
	return ctx.Text(keys[month-time.January])
}

func weekdayLabel(ctx *core.Context, weekday int) string {
	keys := [...]i18n.Key{
		i18n.CalendarWeekdaySunday, i18n.CalendarWeekdayMonday, i18n.CalendarWeekdayTuesday,
		i18n.CalendarWeekdayWednesday, i18n.CalendarWeekdayThursday, i18n.CalendarWeekdayFriday,
		i18n.CalendarWeekdaySaturday,
	}
	if weekday < int(time.Sunday) || weekday > int(time.Saturday) {
		return ""
	}
	if weekday != int(time.Monday) && weekday != int(time.Wednesday) && weekday != int(time.Friday) {
		return ""
	}
	return ctx.Text(keys[weekday])
}

func (w Widget) selection(pos f32.Point, hovered bool, plot image.Rectangle, cellW, cellH float32, rows, cols int, minValue, maxValue float64, scale theme.HeatmapColorScale) cellSelection {
	if !hovered || !pos.Round().In(plot) {
		return cellSelection{}
	}
	x, y := int((pos.X-float32(plot.Min.X))/cellW), int((pos.Y-float32(plot.Min.Y))/cellH)
	if x < 0 || x >= cols || y < 0 || y >= rows || y >= len(w.values) || x >= len(w.values[y]) || !chart.Finite(w.values[y][x]) {
		return cellSelection{}
	}
	label := ""
	if y < len(w.cellLabels) && x < len(w.cellLabels[y]) {
		label = w.cellLabels[y][x]
	}
	return cellSelection{x: x, y: y, value: w.values[y][x], label: label, color: scale(minValue, maxValue, w.values[y][x]), valid: true}
}

func (w Widget) publicSelection(s cellSelection, xLabel, yLabel string) chart.Selection {
	label := s.label
	if label == "" {
		label = xLabel + ", " + yLabel
	}
	return chart.Selection{Label: label, Index: s.x, X: float64(s.x), Items: []chart.Datum{{SeriesKey: "value", SeriesLabel: yLabel, X: float64(s.x), Y: s.value, Color: s.color}}}
}

func (w Widget) draw(ctx *core.Context, gtx layout.Context, plot image.Rectangle, rows, cols int, cellW, cellH float32, minValue, maxValue float64, scale theme.HeatmapColorScale, selected cellSelection, metrics theme.HeatmapMetrics, colors theme.HeatmapColors, xLabels, yLabels []string) {
	gap := gtx.Dp(metrics.CellGap)
	if !w.showGrid {
		gap = 0
	}
	for y := range rows {
		for x := range cols {
			left := int(math.Round(float64(float32(plot.Min.X) + float32(x)*cellW + float32(gap)/2)))
			top := int(math.Round(float64(float32(plot.Min.Y) + float32(y)*cellH + float32(gap)/2)))
			right := int(math.Round(float64(float32(plot.Min.X) + float32(x+1)*cellW - float32(gap)/2)))
			bottom := int(math.Round(float64(float32(plot.Min.Y) + float32(y+1)*cellH - float32(gap)/2)))
			rect := image.Rect(left, top, right, bottom)
			if rect.Empty() {
				continue
			}
			col := colors.Empty
			if y < len(w.values) && x < len(w.values[y]) && chart.Finite(w.values[y][x]) {
				col = scale(minValue, maxValue, w.values[y][x])
			}
			paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, min(gtx.Dp(metrics.CellRadius), min(rect.Dx(), rect.Dy())/2)).Op(gtx.Ops))
			if selected.valid && selected.x == x && selected.y == y {
				border := clip.Stroke{Path: clip.UniformRRect(rect, min(gtx.Dp(metrics.CellRadius), min(rect.Dx(), rect.Dy())/2)).Path(gtx.Ops), Width: float32(max(gtx.Dp(metrics.SelectionBorderWidth), 1))}
				paint.FillShape(gtx.Ops, colors.Selection, border.Op())
			}
		}
	}
	for x := range cols {
		label := labelAt(xLabels, x)
		if label == "" {
			continue
		}
		call, dims := chart.RecordText(ctx, gtx, label, metrics.AxisTextSize, metrics.AxisFontWeight, colors.AxisText, max(int(cellW*2), 1))
		chart.PlaceRecorded(gtx, call, dims, image.Pt(int(float32(plot.Min.X)+float32(x)*cellW), max(plot.Min.Y-gtx.Dp(metrics.TickLabelGap)-dims.Size.Y, 0)))
	}
	for y := range rows {
		call, dims := chart.RecordText(ctx, gtx, labelAt(yLabels, y), metrics.AxisTextSize, metrics.AxisFontWeight, colors.AxisText, max(plot.Min.X-gtx.Dp(metrics.TickLabelGap), 1))
		chart.PlaceRecorded(gtx, call, dims, image.Pt(max(plot.Min.X-gtx.Dp(metrics.TickLabelGap)-dims.Size.X, 0), int(float32(plot.Min.Y)+float32(y)*cellH+(cellH-float32(dims.Size.Y))/2)))
	}
}

func labelAt(labels []string, index int) string {
	if index >= 0 && index < len(labels) {
		return labels[index]
	}
	return ""
}

func (w Widget) drawEmpty(ctx *core.Context, gtx layout.Context, size image.Point, text string, visual theme.HeatmapVisual) {
	call, dims := chart.RecordText(ctx, gtx, text, visual.Metrics.EmptyTextSize, visual.Metrics.EmptyFontWeight, visual.Colors.Muted, size.X)
	chart.PlaceRecorded(gtx, call, dims, image.Pt((size.X-dims.Size.X)/2, (size.Y-dims.Size.Y)/2))
}
func (w Widget) drawTooltip(ctx *core.Context, state *heatState, selected cellSelection, anchor image.Rectangle, xLabels, yLabels []string, visual theme.HeatmapVisual) {
	if !selected.valid {
		return
	}
	content := w.tooltipContent
	var widget core.Widget
	if content != nil {
		widget = content(w.publicSelection(selected, labelAt(xLabels, selected.x), labelAt(yLabels, selected.y)))
	} else {
		widget = core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
			title := labelAt(xLabels, selected.x)
			if selected.label != "" {
				title = selected.label
			}
			rows := []chart.TooltipRow{{Text: fmt.Sprintf("%s  %.2f", labelAt(yLabels, selected.y), selected.value), Color: selected.color}}
			return chart.LayoutTooltipRows(ctx, gtx, title, rows, visual.Metrics.TooltipTextSize, visual.Colors.OverlayForeground, gtx.Dp(visual.Metrics.TooltipMarkerSize), gtx.Dp(visual.Metrics.TooltipRowGap), chart.TooltipMarkerSquare)
		})
	}
	tooltip.NewPopup(widget).Placement(core.PopoverRightStart).Offset(max(visual.Metrics.TooltipGap, 0)).Progress(1).Overlay(ctx, core.FullKey(ctx, "tooltip:"+w.key), anchor)
}

func updatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, tag event.Tag, hovered *bool, position *f32.Point) {
	for {
		e, ok := core.NextPointerEvent(gtx, tag, pointer.Enter|pointer.Leave|pointer.Move|pointer.Drag|pointer.Press|pointer.Cancel)
		if !ok {
			break
		}
		if !enabled {
			*hovered = false
			continue
		}
		switch e.Kind {
		case pointer.Enter, pointer.Move, pointer.Drag, pointer.Press:
			*hovered = e.Position.Round().In(plot)
			*position = e.Position
		case pointer.Leave, pointer.Cancel:
			*hovered = false
		}
	}
}
func addPointerInput(gtx layout.Context, plot image.Rectangle, enabled bool, tag event.Tag) {
	if !enabled || plot.Empty() {
		return
	}
	area := clip.Rect(plot).Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	pointer.CursorPointer.Add(gtx.Ops)
	event.Op(gtx.Ops, tag)
	pass.Pop()
	area.Pop()
}
func updateClick(gtx layout.Context, enabled bool, click *gesture.Click) bool {
	activated := false
	for {
		value, ok := click.Update(gtx.Source)
		if !ok {
			break
		}
		if value.Kind == gesture.KindClick {
			activated = true
		}
	}
	return activated && enabled
}
func addClickInput(gtx layout.Context, size image.Point, enabled bool, click *gesture.Click) {
	if !enabled || size.X <= 0 || size.Y <= 0 {
		return
	}
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	click.Add(gtx.Ops)
	pass.Pop()
	area.Pop()
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.HeatmapOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
