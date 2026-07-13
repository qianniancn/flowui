package barchart

import (
	"fmt"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/unit"
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
	hidden         bool
}

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

func (s Series) Hidden(hidden bool) Series {
	s.hidden = hidden
	return s
}

type Widget struct {
	key            string
	series         []Series
	categories     []string
	height         unit.Dp
	showGrid       bool
	showLegend     bool
	hasShowLegend  bool
	showTooltip    bool
	includeZero    bool
	disabled       bool
	label          string
	emptyText      string
	xAxisLabel     string
	yAxisLabel     string
	yTickCount     int
	yMin           float64
	yMax           float64
	hasYRange      bool
	barGap         float32
	hasBarGap      bool
	categoryGap    float32
	hasCategoryGap bool
	formatY        func(float64) string
}

func New(key string, series []Series) Widget {
	return Widget{
		key:         key,
		series:      append([]Series(nil), series...),
		showGrid:    true,
		showTooltip: true,
		includeZero: true,
		yTickCount:  5,
	}
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

func (w Widget) XAxis(label string) Widget {
	w.xAxisLabel = label
	return w
}

func (w Widget) YAxis(label string) Widget {
	w.yAxisLabel = label
	return w
}

func (w Widget) YTicks(count int) Widget {
	if count < 2 {
		panic("flowui: bar chart Y tick count must be at least 2")
	}
	w.yTickCount = count
	return w
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
	return formatAxisNumber(value, interval)
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
