package barchart

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// Series describes one category-based bar series.
type Series struct {
	key            string
	label          string
	values         []float64
	color          color.NRGBA
	hasColor       bool
	itemColors     []color.NRGBA
	stack          string
	width          unit.Dp
	maxWidth       unit.Dp
	minHeight      unit.Dp
	radius         unit.Dp
	hasRadius      bool
	showBackground bool
	showLabels     bool
	labelPosition  LabelPosition
	formatLabel    func(float64) string
	hidden         bool
}

// LabelPosition controls bar value label placement.
type LabelPosition uint8

const (
	LabelAuto LabelPosition = iota
	LabelInside
	LabelOutside
)

// Orientation controls whether values grow vertically or horizontally.
type Orientation uint8

const (
	Vertical Orientation = iota
	Horizontal
)

// Values creates a bar series whose values correspond to category indexes.
// A non-finite value leaves the category empty.
func Values(key, label string, values []float64) Series {
	return Series{key: key, label: label, values: append([]float64(nil), values...)}
}

func (s Series) Color(value color.NRGBA) Series {
	s.color = value
	s.hasColor = true
	return s
}

// ItemColors overrides individual bar colors by category index. Missing
// entries continue to use the series color.
func (s Series) ItemColors(values []color.NRGBA) Series {
	s.itemColors = append([]color.NRGBA(nil), values...)
	return s
}

// Stack places series with the same non-empty name in one column. Positive
// and negative values accumulate independently, matching ECharts' default.
func (s Series) Stack(name string) Series {
	s.stack = name
	return s
}

func (s Series) Width(dp int) Series {
	if dp <= 0 {
		panic("flowui: bar chart series width must be positive")
	}
	s.width = unit.Dp(dp)
	return s
}

func (s Series) MaxWidth(dp int) Series {
	if dp <= 0 {
		panic("flowui: bar chart series maximum width must be positive")
	}
	s.maxWidth = unit.Dp(dp)
	return s
}

func (s Series) MinHeight(dp int) Series {
	if dp < 0 {
		panic("flowui: bar chart series minimum height must not be negative")
	}
	s.minHeight = unit.Dp(dp)
	return s
}

func (s Series) Radius(dp int) Series {
	if dp < 0 {
		panic("flowui: bar chart series radius must not be negative")
	}
	s.radius = unit.Dp(dp)
	s.hasRadius = true
	return s
}

func (s Series) Background(show bool) Series {
	s.showBackground = show
	return s
}

func (s Series) ShowLabels(show bool) Series {
	s.showLabels = show
	return s
}

func (s Series) LabelPosition(position LabelPosition) Series {
	if position > LabelOutside {
		panic("flowui: invalid bar chart label position")
	}
	s.labelPosition = position
	return s
}

func (s Series) FormatLabel(format func(float64) string) Series {
	s.formatLabel = format
	return s
}

func (s Series) Hidden(hidden bool) Series {
	s.hidden = hidden
	return s
}

type Widget struct {
	key                     string
	series                  []Series
	categories              []string
	dataVersion             uint64
	hasDataVersion          bool
	height                  unit.Dp
	showGrid                bool
	showLegend              bool
	hasShowLegend           bool
	showTooltip             bool
	includeZero             bool
	disabled                bool
	label                   string
	emptyText               string
	xAxisLabel              string
	yAxisLabel              string
	categoryAxisLabel       string
	valueAxisLabel          string
	hasCategoryAxisLabel    bool
	hasValueAxisLabel       bool
	yTickCount              int
	yMin                    float64
	yMax                    float64
	hasYRange               bool
	barGap                  float32
	hasBarGap               bool
	categoryGap             float32
	hasCategoryGap          bool
	formatY                 func(float64) string
	animation               bool
	animationDuration       time.Duration
	animationEasing         animation.Easing
	updateAnimationDuration time.Duration
	updateAnimationEasing   animation.Easing
	onLegendChange          func(string, bool)
	onDataClick             func(chart.Selection)
	tooltipContent          func(chart.Selection) frame.Widget
	dataWindow              chart.DataWindow
	hasDataWindow           bool
	onDataWindowChange      func(chart.DataWindow)
	markLines               []chart.MarkLine
	markAreas               []chart.MarkArea
	markPoints              []chart.MarkPoint
	orientation             Orientation
}

func New(key string, series []Series) Widget {
	return Widget{
		key:                     key,
		series:                  append([]Series(nil), series...),
		showGrid:                true,
		showTooltip:             true,
		includeZero:             true,
		yTickCount:              5,
		animation:               true,
		animationDuration:       time.Second,
		animationEasing:         animation.EaseCubicOut,
		updateAnimationDuration: 500 * time.Millisecond,
		updateAnimationEasing:   animation.EaseCubicInOut,
	}
}

// DataVersion enables resolved-data reuse. Increase version whenever the
// series, categories, or series rendering options change.
func (w Widget) DataVersion(version uint64) Widget {
	w.dataVersion = version
	w.hasDataVersion = true
	return w
}

func (w Widget) Categories(categories []string) Widget {
	w.categories = append([]string(nil), categories...)
	return w
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: bar chart height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}

func (w Widget) Grid(show bool) Widget {
	w.showGrid = show
	return w
}

func (w Widget) Legend(show bool) Widget {
	w.showLegend = show
	w.hasShowLegend = true
	return w
}

func (w Widget) Tooltip(show bool) Widget {
	w.showTooltip = show
	return w
}

func (w Widget) IncludeZero(include bool) Widget {
	w.includeZero = include
	return w
}

func (w Widget) YRange(minimum, maximum float64) Widget {
	validateChartRange(minimum, maximum)
	w.yMin, w.yMax, w.hasYRange = minimum, maximum, true
	return w
}

// ValueRange sets the numeric value-axis range in either orientation.
func (w Widget) ValueRange(minimum, maximum float64) Widget {
	return w.YRange(minimum, maximum)
}

func (w Widget) XAxis(label string) Widget {
	w.xAxisLabel = label
	return w
}

func (w Widget) YAxis(label string) Widget {
	w.yAxisLabel = label
	return w
}

// CategoryAxis sets the category-axis label in either orientation.
func (w Widget) CategoryAxis(label string) Widget {
	w.categoryAxisLabel = label
	w.hasCategoryAxisLabel = true
	return w
}

// ValueAxis sets the numeric value-axis label in either orientation.
func (w Widget) ValueAxis(label string) Widget {
	w.valueAxisLabel = label
	w.hasValueAxisLabel = true
	return w
}

func (w Widget) YTicks(count int) Widget {
	if count < 2 {
		panic("flowui: bar chart Y tick count must be at least 2")
	}
	w.yTickCount = count
	return w
}

// ValueTicks sets the numeric value-axis tick target in either orientation.
func (w Widget) ValueTicks(count int) Widget {
	return w.YTicks(count)
}

// BarGap sets the gap between adjacent columns as a ratio of their width.
func (w Widget) BarGap(ratio float32) Widget {
	validateNonnegative("bar gap", ratio)
	w.barGap, w.hasBarGap = ratio, true
	return w
}

// CategoryGap sets the empty portion of each category band.
func (w Widget) CategoryGap(ratio float32) Widget {
	validateRatio("category gap", ratio)
	w.categoryGap, w.hasCategoryGap = ratio, true
	return w
}

func (w Widget) FormatY(format func(float64) string) Widget {
	w.formatY = format
	return w
}

// FormatValue formats numeric value-axis labels in either orientation.
func (w Widget) FormatValue(format func(float64) string) Widget {
	return w.FormatY(format)
}

// Animation enables initial and data-update transitions.
func (w Widget) Animation(enabled bool) Widget {
	w.animation = enabled
	return w
}

// AnimationDuration sets the initial transition duration.
func (w Widget) AnimationDuration(duration time.Duration) Widget {
	validateAnimationDuration(duration)
	w.animationDuration = duration
	return w
}

// AnimationEasing sets the initial transition timing curve.
func (w Widget) AnimationEasing(easing animation.Easing) Widget {
	validateAnimationEasing(easing)
	w.animationEasing = easing
	return w
}

// UpdateAnimationDuration sets the data-update transition duration.
func (w Widget) UpdateAnimationDuration(duration time.Duration) Widget {
	validateAnimationDuration(duration)
	w.updateAnimationDuration = duration
	return w
}

// UpdateAnimationEasing sets the data-update transition timing curve.
func (w Widget) UpdateAnimationEasing(easing animation.Easing) Widget {
	validateAnimationEasing(easing)
	w.updateAnimationEasing = easing
	return w
}

// OnLegendChange registers a controlled series visibility request. Hidden is
// the next value to pass to Series.Hidden.
func (w Widget) OnLegendChange(fn func(seriesKey string, hidden bool)) Widget {
	w.onLegendChange = fn
	return w
}

// OnDataClick registers a callback for activation of the current chart selection.
func (w Widget) OnDataClick(fn func(chart.Selection)) Widget {
	w.onDataClick = fn
	return w
}

// TooltipContent replaces the default tooltip body with custom content.
func (w Widget) TooltipContent(fn func(chart.Selection) frame.Widget) Widget {
	w.tooltipContent = fn
	return w
}

// DataWindow sets the controlled normalized visible category range.
func (w Widget) DataWindow(start, end float32) Widget {
	w.dataWindow = chart.NewDataWindow(start, end)
	w.hasDataWindow = true
	return w
}

// OnDataWindowChange enables wheel zoom, drag pan, and double-click reset.
func (w Widget) OnDataWindowChange(fn func(chart.DataWindow)) Widget {
	w.onDataWindowChange = fn
	return w
}

// MarkLines sets controlled reference lines.
func (w Widget) MarkLines(values []chart.MarkLine) Widget {
	chart.ValidateAnnotations(values, nil, nil)
	w.markLines = append([]chart.MarkLine(nil), values...)
	return w
}

// MarkAreas sets controlled reference bands.
func (w Widget) MarkAreas(values []chart.MarkArea) Widget {
	chart.ValidateAnnotations(nil, values, nil)
	w.markAreas = append([]chart.MarkArea(nil), values...)
	return w
}

// MarkPoints sets controlled Cartesian markers. X is a category index.
func (w Widget) MarkPoints(values []chart.MarkPoint) Widget {
	chart.ValidateAnnotations(nil, nil, values)
	w.markPoints = append([]chart.MarkPoint(nil), values...)
	return w
}

func (w Widget) Orientation(value Orientation) Widget {
	if value > Horizontal {
		panic("flowui: invalid bar chart orientation")
	}
	w.orientation = value
	return w
}

func (w Widget) Label(label string) Widget {
	w.label = label
	return w
}

func (w Widget) EmptyText(text string) Widget {
	w.emptyText = text
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return w.layout(ctx, gtx)
}

func (w Widget) legendVisible(data chartData) bool {
	if w.hasShowLegend {
		return w.showLegend
	}
	return len(data.legend) > 1
}

func (w Widget) effectiveDataWindow() chart.DataWindow {
	if w.hasDataWindow {
		return w.dataWindow
	}
	return chart.FullDataWindow()
}

func (w Widget) categoryLabel(index int) string {
	if index >= 0 && index < len(w.categories) && w.categories[index] != "" {
		return w.categories[index]
	}
	return fmt.Sprint(index + 1)
}

func (w Widget) yLabel(value, interval float64) string {
	if w.formatY != nil {
		return w.formatY(value)
	}
	return chart.FormatAxisNumber(value, interval)
}

func validateChartRange(minimum, maximum float64) {
	if !finite(minimum) || !finite(maximum) || maximum <= minimum {
		panic("flowui: bar chart Y maximum must be greater than minimum")
	}
}

func validateRatio(name string, value float32) {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value < 0 || value >= 1 {
		panic(fmt.Sprintf("flowui: bar chart %s must be between 0 and 1", name))
	}
}

func validateNonnegative(name string, value float32) {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value < 0 {
		panic(fmt.Sprintf("flowui: bar chart %s must be finite and nonnegative", name))
	}
}

func validateAnimationDuration(duration time.Duration) {
	if duration < 0 {
		panic("flowui: bar chart animation duration must not be negative")
	}
}

func validateAnimationEasing(easing animation.Easing) {
	if easing == nil {
		panic("flowui: bar chart animation easing must not be nil")
	}
}
