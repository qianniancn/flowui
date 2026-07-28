package heatmap

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/components/chart"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/components/tooltip"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/overlay"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	"github.com/qianniancn/flowui/internal/theme"
)

type Widget struct {
	key                             string
	xLabels, yLabels                []string
	values                          [][]float64
	height                          unit.Dp
	showGrid, showTooltip, disabled bool
	min, max                        float64
	hasRange                        bool
	minColor, maxColor              color.NRGBA
	hasMinColor, hasMaxColor        bool
	label, emptyText                string
	onDataClick                     func(chart.Selection)
	tooltipContent                  func(chart.Selection) frame.Widget
	customStyle                     flowstyle.Style
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
	xLabels := make([]string, weeks)
	for week := range xLabels {
		date := gridStart.AddDate(0, 0, week*7)
		weekEnd := date.AddDate(0, 0, 6)
		for day := date; !day.After(weekEnd); day = day.AddDate(0, 0, 1) {
			monthStart := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
			if day.Day() == 1 && !monthStart.Before(start) && !monthStart.After(end) {
				xLabels[week] = monthStart.Format("Jan")
				break
			}
		}
	}
	return Widget{key: key, xLabels: xLabels, yLabels: []string{"", "Mon", "", "Wed", "", "Fri", ""}, values: matrix, cellLabels: labels, showGrid: true, showTooltip: true}
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
func (w Widget) MinColor(value color.NRGBA) Widget           { w.minColor, w.hasMinColor = value, true; return w }
func (w Widget) MaxColor(value color.NRGBA) Widget           { w.maxColor, w.hasMaxColor = value, true; return w }
func (w Widget) Label(label string) Widget                   { w.label = label; return w }
func (w Widget) EmptyText(text string) Widget                { w.emptyText = text; return w }
func (w Widget) OnDataClick(fn func(chart.Selection)) Widget { w.onDataClick = fn; return w }
func (w Widget) TooltipContent(fn func(chart.Selection) frame.Widget) Widget {
	w.tooltipContent = fn
	return w
}
func (w Widget) Disabled(disabled bool) Widget      { w.disabled = disabled; return w }
func (w Widget) Style(value flowstyle.Style) Widget { w.customStyle = value; return w }

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

func stateFor(ctx *frame.Context, key string) *heatState {
	key = frame.ClaimKey(ctx, stateutil.KindHeatmap, key)
	return frame.UseState[heatState](ctx, key, "heatmap")
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutui.Box(frame.WidgetFunc(w.layout)).Style(w.customStyle).Layout(ctx, gtx)
}

func (w Widget) layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	s := stateFor(ctx, w.key)
	enabled := gtx.Enabled() && !w.disabled
	tokens := frame.ActiveTheme(ctx).Components.Heatmap
	height := tokens.Height
	if w.height > 0 {
		height = w.height
	}
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(height)))
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}
	rows, cols := len(w.yLabels), len(w.xLabels)
	for _, row := range w.values {
		if len(row) > cols {
			cols = len(row)
		}
	}
	if rows == 0 || cols == 0 {
		w.drawEmpty(ctx, gtx, size, w.emptyText)
		return layout.Dimensions{Size: size}
	}
	plotLeft := gtx.Dp(tokens.PlotPaddingLeft)
	plotTop := gtx.Dp(tokens.PlotPaddingTop)
	maxPlotWidth := max(size.X-plotLeft-gtx.Dp(tokens.PlotPaddingRight), 1)
	maxPlotHeight := max(size.Y-plotTop-gtx.Dp(tokens.PlotPaddingBottom), 1)
	cellSize := max(gtx.Dp(tokens.CellSize), 1)
	cellSize = min(cellSize, maxPlotWidth/cols)
	cellSize = min(cellSize, maxPlotHeight/rows)
	if cellSize <= 0 {
		return layout.Dimensions{Size: size}
	}
	plot := image.Rect(plotLeft, plotTop, plotLeft+cellSize*cols, plotTop+cellSize*rows)
	minValue, maxValue := w.valueExtent(rows, cols)
	minColor, maxColor := w.colors(frame.ActiveTheme(ctx))
	cellW, cellH := float32(cellSize), float32(cellSize)
	updatePointer(gtx, enabled, plot, &s.tag, &s.hovered, &s.pointer)
	s.selection = w.selection(s.pointer, s.hovered, plot, cellW, cellH, rows, cols, minValue, maxValue, minColor, maxColor)
	activated := updateClick(gtx, enabled, &s.click)
	if activated && s.selection.valid && w.onDataClick != nil {
		w.onDataClick(w.publicSelection(s.selection, w.xLabel(s.selection.x), w.yLabel(s.selection.y)))
	}
	if enabled && w.showTooltip && s.selection.valid {
		s.tooltipTransition.Progress(gtx, true, frame.ActiveTheme(ctx).Motion)
	} else if !s.selection.valid || !w.showTooltip {
		s.tooltipTransition.Progress(gtx, false, frame.ActiveTheme(ctx).Motion)
	}
	opacity := paint.PushOpacity(gtx.Ops, 1)
	w.draw(ctx, gtx, plot, rows, cols, cellW, cellH, minValue, maxValue, minColor, maxColor, s.selection, tokens)
	if w.showTooltip && (s.selection.valid || s.tooltipTransition.Value() > 0) && enabled {
		w.drawTooltip(ctx, gtx, s, s.selection, chart.TooltipAnchor(s.pointer))
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

func (w Widget) colors(t *theme.Theme) (color.NRGBA, color.NRGBA) {
	minColor, maxColor := t.Components.Heatmap.MinColor, t.Components.Heatmap.MaxColor
	if w.hasMinColor {
		minColor = w.minColor
	}
	if w.hasMaxColor {
		maxColor = w.maxColor
	}
	return minColor, maxColor
}
func (w Widget) xLabel(index int) string {
	if index >= 0 && index < len(w.xLabels) && w.xLabels[index] != "" {
		return w.xLabels[index]
	}
	return ""
}
func (w Widget) yLabel(index int) string {
	if index >= 0 && index < len(w.yLabels) && w.yLabels[index] != "" {
		return w.yLabels[index]
	}
	return ""
}

func (w Widget) selection(pos f32.Point, hovered bool, plot image.Rectangle, cellW, cellH float32, rows, cols int, minValue, maxValue float64, minColor, maxColor color.NRGBA) cellSelection {
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
	return cellSelection{x: x, y: y, value: w.values[y][x], label: label, color: interpolate(minColor, maxColor, float32((w.values[y][x]-minValue)/(maxValue-minValue))), valid: true}
}

func (w Widget) publicSelection(s cellSelection, xLabel, yLabel string) chart.Selection {
	label := s.label
	if label == "" {
		label = xLabel + ", " + yLabel
	}
	return chart.Selection{Label: label, Index: s.x, X: float64(s.x), Items: []chart.Datum{{SeriesKey: "value", SeriesLabel: yLabel, X: float64(s.x), Y: s.value, Color: s.color}}}
}

func (w Widget) draw(ctx *frame.Context, gtx layout.Context, plot image.Rectangle, rows, cols int, cellW, cellH float32, minValue, maxValue float64, minColor, maxColor color.NRGBA, selected cellSelection, tokens theme.HeatmapTheme) {
	gap := gtx.Dp(tokens.CellGap)
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
			col := tokens.EmptyColor
			if col.A == 0 {
				col = frame.ActiveTheme(ctx).Palette.SurfaceSecondary
			}
			if y < len(w.values) && x < len(w.values[y]) && chart.Finite(w.values[y][x]) {
				col = interpolate(minColor, maxColor, float32((w.values[y][x]-minValue)/(maxValue-minValue)))
			}
			paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, min(gtx.Dp(tokens.CellRadius), min(rect.Dx(), rect.Dy())/2)).Op(gtx.Ops))
			if selected.valid && selected.x == x && selected.y == y {
				border := clip.Stroke{Path: clip.UniformRRect(rect, min(gtx.Dp(tokens.CellRadius), min(rect.Dx(), rect.Dy())/2)).Path(gtx.Ops), Width: float32(max(gtx.Dp(2), 1))}
				paint.FillShape(gtx.Ops, frame.ActiveTheme(ctx).Palette.Foreground, border.Op())
			}
		}
	}
	for x := range cols {
		label := w.xLabel(x)
		if label == "" {
			continue
		}
		call, dims := chart.RecordText(ctx, gtx, label, tokens.AxisTextSize, font.Normal, frame.ActiveTheme(ctx).Palette.MutedForeground, max(int(cellW*2), 1))
		chart.PlaceRecorded(gtx, call, dims, image.Pt(int(float32(plot.Min.X)+float32(x)*cellW), max(plot.Min.Y-gtx.Dp(tokens.TickLabelGap)-dims.Size.Y, 0)))
	}
	for y := range rows {
		call, dims := chart.RecordText(ctx, gtx, w.yLabel(y), tokens.AxisTextSize, font.Normal, frame.ActiveTheme(ctx).Palette.MutedForeground, max(plot.Min.X-gtx.Dp(tokens.TickLabelGap), 1))
		chart.PlaceRecorded(gtx, call, dims, image.Pt(max(plot.Min.X-gtx.Dp(tokens.TickLabelGap)-dims.Size.X, 0), int(float32(plot.Min.Y)+float32(y)*cellH+(cellH-float32(dims.Size.Y))/2)))
	}
}

func (w Widget) drawEmpty(ctx *frame.Context, gtx layout.Context, size image.Point, text string) {
	if text == "" {
		text = "No data"
	}
	call, dims := chart.RecordText(ctx, gtx, text, 14, font.Normal, frame.ActiveTheme(ctx).Palette.MutedForeground, size.X)
	chart.PlaceRecorded(gtx, call, dims, image.Pt((size.X-dims.Size.X)/2, (size.Y-dims.Size.Y)/2))
}
func (w Widget) drawTooltip(ctx *frame.Context, gtx layout.Context, state *heatState, selected cellSelection, anchor image.Rectangle) {
	if !selected.valid {
		return
	}
	content := w.tooltipContent
	var widget frame.Widget
	if content != nil {
		widget = content(w.publicSelection(selected, w.xLabel(selected.x), w.yLabel(selected.y)))
	} else {
		widget = frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
			title := w.xLabel(selected.x)
			if selected.label != "" {
				title = selected.label
			}
			rows := []chart.TooltipRow{{Text: fmt.Sprintf("%s  %.2f", w.yLabel(selected.y), selected.value), Color: selected.color}}
			return chart.LayoutTooltipRows(ctx, gtx, title, rows, frame.ActiveTheme(ctx).Components.Tooltip.TextSize, frame.ActiveTheme(ctx).Palette.OverlayForeground, gtx.Dp(7), gtx.Dp(4), chart.TooltipMarkerSquare)
		})
	}
	tooltip.NewPopup(widget).Placement(overlay.PopoverRightStart).Offset(max(frame.ActiveTheme(ctx).Components.Heatmap.TooltipGap, 0)).Progress(1).Layout(ctx, gtx, anchor)
}

func interpolate(a, b color.NRGBA, t float32) color.NRGBA {
	if t < 0 {
		t = 0
	}
	if t > 1 {
		t = 1
	}
	return color.NRGBA{R: byte(float32(a.R) + (float32(b.R)-float32(a.R))*t), G: byte(float32(a.G) + (float32(b.G)-float32(a.G))*t), B: byte(float32(a.B) + (float32(b.B)-float32(a.B))*t), A: byte(float32(a.A) + (float32(b.A)-float32(a.A))*t)}
}
func updatePointer(gtx layout.Context, enabled bool, plot image.Rectangle, tag event.Tag, hovered *bool, position *f32.Point) {
	for {
		value, ok := gtx.Event(pointer.Filter{Target: tag, Kinds: pointer.Enter | pointer.Leave | pointer.Move | pointer.Drag | pointer.Press | pointer.Cancel})
		if !ok {
			break
		}
		e, ok := value.(pointer.Event)
		if !ok {
			continue
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
