package candlestick

import (
	"fmt"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/animation"
	"github.com/qianniancn/FlowUI/internal/components/chart"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

// Candle stores values in ECharts order: open, close, lowest, highest.
type Candle struct {
	open  float64
	close float64
	low   float64
	high  float64
}

func OHLC(open, close, lowest, highest float64) Candle {
	return Candle{open: open, close: close, low: lowest, high: highest}
}

type Widget struct {
	key                     string
	data                    []Candle
	categories              []string
	times                   []time.Time
	dataVersion             uint64
	hasDataVersion          bool
	height                  unit.Dp
	showGrid                bool
	showTooltip             bool
	showCrosshair           bool
	disabled                bool
	label                   string
	emptyText               string
	xAxisLabel              string
	yAxisLabel              string
	yTickCount              int
	yMin                    float64
	yMax                    float64
	hasYRange               bool
	formatY                 func(float64) string
	formatTime              func(time.Time) string
	width                   unit.Dp
	maxWidth                unit.Dp
	minWidth                unit.Dp
	upColor                 color.NRGBA
	hasUpColor              bool
	downColor               color.NRGBA
	hasDownColor            bool
	dojiColor               color.NRGBA
	hasDojiColor            bool
	animation               bool
	animationDuration       time.Duration
	animationEasing         animation.Easing
	updateAnimationDuration time.Duration
	updateAnimationEasing   animation.Easing
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

func New(key string, data []Candle) Widget {
	return Widget{
		key:                     key,
		data:                    append([]Candle(nil), data...),
		showGrid:                true,
		showTooltip:             true,
		showCrosshair:           true,
		yTickCount:              5,
		animation:               true,
		animationDuration:       300 * time.Millisecond,
		animationEasing:         animation.EaseLinear,
		updateAnimationDuration: 300 * time.Millisecond,
		updateAnimationEasing:   animation.EaseLinear,
	}
}

// DataVersion enables resolved-data reuse. Increase version whenever the
// candles or color options change.
func (w Widget) DataVersion(version uint64) Widget {
	w.dataVersion = version
	w.hasDataVersion = true
	return w
}

func (w Widget) Categories(values []string) Widget {
	w.categories = append([]string(nil), values...)
	w.times = nil
	return w
}

// Times sets time-based category labels and replaces Categories.
func (w Widget) Times(values []time.Time) Widget {
	w.times = append([]time.Time(nil), values...)
	w.categories = nil
	return w
}

// FormatTime overrides automatic axis and selection time formatting.
func (w Widget) FormatTime(format func(time.Time) string) Widget {
	w.formatTime = format
	return w
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: candlestick chart height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}

func (w Widget) Grid(show bool) Widget {
	w.showGrid = show
	return w
}

func (w Widget) Tooltip(show bool) Widget {
	w.showTooltip = show
	return w
}

func (w Widget) Crosshair(show bool) Widget {
	w.showCrosshair = show
	return w
}

func (w Widget) YRange(minimum, maximum float64) Widget {
	if !chart.Finite(minimum) || !chart.Finite(maximum) || maximum <= minimum {
		panic("flowui: candlestick chart Y maximum must be greater than minimum")
	}
	w.yMin, w.yMax, w.hasYRange = minimum, maximum, true
	return w
}

func (w Widget) YTicks(count int) Widget {
	if count < 2 {
		panic("flowui: candlestick chart Y tick count must be at least 2")
	}
	w.yTickCount = count
	return w
}

func (w Widget) FormatY(format func(float64) string) Widget {
	w.formatY = format
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

func (w Widget) Width(dp int) Widget {
	if dp <= 0 {
		panic("flowui: candlestick chart width must be positive")
	}
	w.width = unit.Dp(dp)
	return w
}

func (w Widget) MaxWidth(dp int) Widget {
	if dp <= 0 {
		panic("flowui: candlestick chart maximum width must be positive")
	}
	w.maxWidth = unit.Dp(dp)
	return w
}

func (w Widget) MinWidth(dp int) Widget {
	if dp <= 0 {
		panic("flowui: candlestick chart minimum width must be positive")
	}
	w.minWidth = unit.Dp(dp)
	return w
}

func (w Widget) UpColor(value color.NRGBA) Widget {
	w.upColor, w.hasUpColor = value, true
	return w
}

func (w Widget) DownColor(value color.NRGBA) Widget {
	w.downColor, w.hasDownColor = value, true
	return w
}

func (w Widget) DojiColor(value color.NRGBA) Widget {
	w.dojiColor, w.hasDojiColor = value, true
	return w
}

func (w Widget) Animation(enabled bool) Widget {
	w.animation = enabled
	return w
}

func (w Widget) AnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "candlestick chart")
	w.animationDuration = duration
	return w
}

func (w Widget) AnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "candlestick chart")
	w.animationEasing = easing
	return w
}

func (w Widget) UpdateAnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "candlestick chart")
	w.updateAnimationDuration = duration
	return w
}

func (w Widget) UpdateAnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "candlestick chart")
	w.updateAnimationEasing = easing
	return w
}

func (w Widget) OnDataClick(fn func(chart.Selection)) Widget {
	w.onDataClick = fn
	return w
}

func (w Widget) TooltipContent(fn func(chart.Selection) frame.Widget) Widget {
	w.tooltipContent = fn
	return w
}

func (w Widget) DataWindow(start, end float32) Widget {
	w.dataWindow = chart.NewDataWindow(float64(start), float64(end))
	w.hasDataWindow = true
	return w
}

func (w Widget) OnDataWindowChange(fn func(chart.DataWindow)) Widget {
	w.onDataWindowChange = fn
	return w
}

func (w Widget) MarkLines(values []chart.MarkLine) Widget {
	chart.ValidateAnnotations(values, nil, nil)
	w.markLines = append([]chart.MarkLine(nil), values...)
	return w
}

func (w Widget) MarkAreas(values []chart.MarkArea) Widget {
	chart.ValidateAnnotations(nil, values, nil)
	w.markAreas = append([]chart.MarkArea(nil), values...)
	return w
}

// MarkPoints sets controlled markers. X is a candle index.
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

func (w Widget) categoryLabel(index int) string {
	if value, ok := w.timeAt(index); ok {
		if w.formatTime != nil {
			return w.formatTime(value)
		}
		return value.Format("2006-01-02 15:04:05")
	}
	if index >= 0 && index < len(w.categories) && w.categories[index] != "" {
		return w.categories[index]
	}
	return fmt.Sprint(index + 1)
}

func (w Widget) axisCategoryLabel(index int, format string) string {
	if value, ok := w.timeAt(index); ok {
		if w.formatTime != nil {
			return w.formatTime(value)
		}
		return value.Format(format)
	}
	return w.categoryLabel(index)
}

func (w Widget) timeAt(index int) (time.Time, bool) {
	if index < 0 || index >= len(w.times) || w.times[index].IsZero() {
		return time.Time{}, false
	}
	return w.times[index], true
}

func (w Widget) timeAxisFormat(start, end int) string {
	var earliest, latest, previous time.Time
	interval := time.Duration(0)
	for index := start; index < end; index++ {
		value, ok := w.timeAt(index)
		if !ok {
			continue
		}
		if earliest.IsZero() || value.Before(earliest) {
			earliest = value
		}
		if latest.IsZero() || value.After(latest) {
			latest = value
		}
		if !previous.IsZero() {
			delta := value.Sub(previous)
			if delta < 0 {
				delta = -delta
			}
			if delta > 0 && (interval == 0 || delta < interval) {
				interval = delta
			}
		}
		previous = value
	}
	span := latest.Sub(earliest)
	switch {
	case interval > 0 && interval < time.Minute:
		if span < 24*time.Hour {
			return "15:04:05"
		}
		return "01-02 15:04:05"
	case interval > 0 && interval < 18*time.Hour:
		if span < 24*time.Hour {
			return "15:04"
		}
		return "01-02 15:04"
	case interval > 0 && interval < 28*24*time.Hour:
		if span < 365*24*time.Hour {
			return "01-02"
		}
		return "2006-01-02"
	case interval > 0 && interval < 300*24*time.Hour:
		return "2006-01"
	case interval >= 300*24*time.Hour:
		return "2006"
	default:
		return "2006-01-02 15:04"
	}
}

func (w Widget) effectiveDataWindow() chart.DataWindow {
	if w.hasDataWindow {
		return w.dataWindow
	}
	return chart.FullDataWindow()
}

func (w Widget) yLabel(value, interval float64) string {
	if w.formatY != nil {
		return w.formatY(value)
	}
	return chart.FormatAxisNumber(value, interval)
}
