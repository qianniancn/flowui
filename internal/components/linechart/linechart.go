package linechart

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/animation"
	"github.com/qianniancn/flowui/internal/components/chart"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

// Point is one Cartesian data point. A non-finite Y value creates a gap.
type Point struct {
	X float64
	Y float64
}

// StepMode controls where a stepped line changes value.
type StepMode uint8

const (
	StepNone StepMode = iota
	StepStart
	StepMiddle
	StepEnd
)

// LineStyle controls the stroke pattern.
type LineStyle uint8

const (
	LineSolid LineStyle = iota
	LineDashed
	LineDotted
)

// SamplingMode controls large-series pixel reduction.
type SamplingMode uint8

const (
	SamplingNone SamplingMode = iota
	SamplingMinMax
)

// StackStrategy controls which preceding cumulative values a series stacks on.
type StackStrategy uint8

const (
	StackSameSign StackStrategy = iota
	StackAll
	StackPositive
	StackNegative
)

// StackOrder controls the calculation order inside one stack group.
type StackOrder uint8

const (
	StackSeriesAscending StackOrder = iota
	StackSeriesDescending
)

// Series describes one line and its data.
type Series struct {
	key           string
	label         string
	values        []float64
	points        []Point
	color         color.NRGBA
	hasColor      bool
	width         unit.Dp
	showPoints    bool
	hasShowPoints bool
	pointSize     unit.Dp
	connectNulls  bool
	smooth        float32
	step          StepMode
	lineStyle     LineStyle
	area          bool
	areaColor     color.NRGBA
	hasAreaColor  bool
	sampling      SamplingMode
	stack         string
	stackStrategy StackStrategy
	stackOrder    StackOrder
	hidden        bool
}

// Values creates a category series whose X values are the item indexes.
func Values(key, label string, values []float64) Series {
	return Series{key: key, label: label, values: append([]float64(nil), values...)}
}

// XY creates a series from explicit Cartesian points.
func XY(key, label string, points []Point) Series {
	return Series{key: key, label: label, points: append([]Point(nil), points...)}
}

func (s Series) Color(value color.NRGBA) Series {
	s.color = value
	s.hasColor = true
	return s
}

func (s Series) Width(dp int) Series {
	if dp <= 0 {
		panic("flowui: line chart series width must be positive")
	}
	s.width = unit.Dp(dp)
	return s
}

func (s Series) ShowPoints(show bool) Series {
	s.showPoints = show
	s.hasShowPoints = true
	return s
}

func (s Series) PointSize(dp int) Series {
	if dp <= 0 {
		panic("flowui: line chart point size must be positive")
	}
	s.pointSize = unit.Dp(dp)
	return s
}

func (s Series) ConnectNulls(connect bool) Series {
	s.connectNulls = connect
	return s
}

// Smooth enables ECharts-style cubic smoothing with the default factor 0.5.
func (s Series) Smooth(smooth bool) Series {
	s.smooth = 0
	if smooth {
		s.smooth = 0.5
	}
	return s
}

// Smoothness sets the cubic smoothing factor in the range [0, 1].
func (s Series) Smoothness(value float32) Series {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) || value < 0 || value > 1 {
		panic("flowui: line chart smoothness must be between 0 and 1")
	}
	s.smooth = value
	return s
}

func (s Series) Step(mode StepMode) Series {
	if mode > StepEnd {
		panic("flowui: invalid line chart step mode")
	}
	s.step = mode
	return s
}

func (s Series) LineStyle(style LineStyle) Series {
	if style > LineDotted {
		panic("flowui: invalid line chart line style")
	}
	s.lineStyle = style
	return s
}

func (s Series) Area(show bool) Series {
	s.area = show
	return s
}

func (s Series) AreaColor(value color.NRGBA) Series {
	s.area = true
	s.areaColor = value
	s.hasAreaColor = true
	return s
}

// Sampling reduces large monotonic-X series while retaining local extrema.
func (s Series) Sampling(mode SamplingMode) Series {
	if mode > SamplingMinMax {
		panic("flowui: invalid line chart sampling mode")
	}
	s.sampling = mode
	return s
}

// Stack assigns the series to a stack group. An empty name disables stacking.
func (s Series) Stack(name string) Series {
	s.stack = name
	return s
}

// StackStrategy applies the ECharts stack strategy to this series.
func (s Series) StackStrategy(strategy StackStrategy) Series {
	if strategy > StackNegative {
		panic("flowui: invalid line chart stack strategy")
	}
	s.stackStrategy = strategy
	return s
}

// StackOrder controls the complete group using the first visible series value.
func (s Series) StackOrder(order StackOrder) Series {
	if order > StackSeriesDescending {
		panic("flowui: invalid line chart stack order")
	}
	s.stackOrder = order
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
	xTickCount              int
	yTickCount              int
	xMin                    float64
	xMax                    float64
	hasXRange               bool
	yMin                    float64
	yMax                    float64
	hasYRange               bool
	formatX                 func(float64) string
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
	customStyle             flowstyle.Style
}

func New(key string, series []Series) Widget {
	return Widget{
		key:                     key,
		series:                  append([]Series(nil), series...),
		showGrid:                true,
		showTooltip:             true,
		includeZero:             true,
		xTickCount:              6,
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

// Categories labels index-based series and replaces the automatic X tick labels.
func (w Widget) Categories(categories []string) Widget {
	w.categories = append([]string(nil), categories...)
	return w
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: line chart height must be positive")
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

// IncludeZero controls whether an automatic Y range includes zero.
func (w Widget) IncludeZero(include bool) Widget {
	w.includeZero = include
	return w
}

func (w Widget) XRange(minimum, maximum float64) Widget {
	validateChartRange("X", minimum, maximum)
	w.xMin, w.xMax, w.hasXRange = minimum, maximum, true
	return w
}

func (w Widget) YRange(minimum, maximum float64) Widget {
	validateChartRange("Y", minimum, maximum)
	w.yMin, w.yMax, w.hasYRange = minimum, maximum, true
	return w
}

func (w Widget) XAxis(label string) Widget {
	w.xAxisLabel = label
	return w
}

func (w Widget) YAxis(label string) Widget {
	w.yAxisLabel = label
	return w
}

func (w Widget) XTicks(count int) Widget {
	if count < 2 {
		panic("flowui: line chart X tick count must be at least 2")
	}
	w.xTickCount = count
	return w
}

func (w Widget) YTicks(count int) Widget {
	if count < 2 {
		panic("flowui: line chart Y tick count must be at least 2")
	}
	w.yTickCount = count
	return w
}

func (w Widget) FormatX(format func(float64) string) Widget {
	w.formatX = format
	return w
}

func (w Widget) FormatY(format func(float64) string) Widget {
	w.formatY = format
	return w
}

// Animation enables initial and data-update transitions.
func (w Widget) Animation(enabled bool) Widget {
	w.animation = enabled
	return w
}

// AnimationDuration sets the initial transition duration.
func (w Widget) AnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "line chart")
	w.animationDuration = duration
	return w
}

// AnimationEasing sets the initial transition timing curve.
func (w Widget) AnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "line chart")
	w.animationEasing = easing
	return w
}

// UpdateAnimationDuration sets the data-update transition duration.
func (w Widget) UpdateAnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "line chart")
	w.updateAnimationDuration = duration
	return w
}

// UpdateAnimationEasing sets the data-update transition timing curve.
func (w Widget) UpdateAnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "line chart")
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

// DataWindow sets the controlled normalized visible X range.
func (w Widget) DataWindow(start, end float32) Widget {
	w.dataWindow = chart.NewDataWindow(float64(start), float64(end))
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

// MarkPoints sets controlled Cartesian markers.
func (w Widget) MarkPoints(values []chart.MarkPoint) Widget {
	chart.ValidateAnnotations(nil, nil, values)
	w.markPoints = append([]chart.MarkPoint(nil), values...)
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

func (w Widget) Style(value flowstyle.Style) Widget {
	w.customStyle = value
	return w
}

func (w Widget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return layoutui.Box(frame.WidgetFunc(w.layout)).Style(w.customStyle).Layout(ctx, gtx)
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

func (w Widget) xLabel(value float64, interval float64) string {
	if w.formatX != nil {
		return w.formatX(value)
	}
	if len(w.categories) > 0 {
		index := int(math.Round(value))
		if index >= 0 && index < len(w.categories) && math.Abs(value-float64(index)) < 1e-9 {
			return w.categories[index]
		}
	}
	return chart.FormatAxisNumber(value, interval)
}

func (w Widget) yLabel(value float64, interval float64) string {
	if w.formatY != nil {
		return w.formatY(value)
	}
	return chart.FormatAxisNumber(value, interval)
}

func validateChartRange(axis string, minimum, maximum float64) {
	if !chart.Finite(minimum) || !chart.Finite(maximum) || maximum <= minimum {
		panic(fmt.Sprintf("flowui: line chart %s maximum must be greater than minimum", axis))
	}
}
