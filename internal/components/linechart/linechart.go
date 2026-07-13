package linechart

import (
	"fmt"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
)

// Point is one Cartesian data point. A non-finite Y value creates a gap.
type Point struct {
	X float64
	Y float64
}

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
	connectNulls  bool
	smooth        float32
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

func (s Series) Hidden(hidden bool) Series {
	s.hidden = hidden
	return s
}

type Widget struct {
	key           string
	series        []Series
	categories    []string
	height        unit.Dp
	showGrid      bool
	showLegend    bool
	hasShowLegend bool
	showTooltip   bool
	includeZero   bool
	disabled      bool
	label         string
	emptyText     string
	xAxisLabel    string
	yAxisLabel    string
	xTickCount    int
	yTickCount    int
	xMin          float64
	xMax          float64
	hasXRange     bool
	yMin          float64
	yMax          float64
	hasYRange     bool
	formatX       func(float64) string
	formatY       func(float64) string
}

func New(key string, series []Series) Widget {
	return Widget{
		key:         key,
		series:      append([]Series(nil), series...),
		showGrid:    true,
		showTooltip: true,
		includeZero: true,
		xTickCount:  6,
		yTickCount:  5,
	}
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
	return len(data.series) > 1
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
	return formatAxisNumber(value, interval)
}

func (w Widget) yLabel(value float64, interval float64) string {
	if w.formatY != nil {
		return w.formatY(value)
	}
	return formatAxisNumber(value, interval)
}

func validateChartRange(axis string, minimum, maximum float64) {
	if !finite(minimum) || !finite(maximum) || maximum <= minimum {
		panic(fmt.Sprintf("flowui: line chart %s maximum must be greater than minimum", axis))
	}
}
