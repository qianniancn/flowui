package piechart

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

// Data describes one pie slice.
type Data struct {
	key      string
	label    string
	value    float64
	color    color.NRGBA
	hasColor bool
	hidden   bool
}

// RoseType controls ECharts-compatible Nightingale rose layout.
type RoseType uint8

const (
	RoseNone RoseType = iota
	RoseRadius
	RoseArea
)

func Slice(key, label string, value float64) Data {
	return Data{key: key, label: label, value: value}
}

func (d Data) Color(value color.NRGBA) Data {
	d.color = value
	d.hasColor = true
	return d
}

func (d Data) Hidden(hidden bool) Data {
	d.hidden = hidden
	return d
}

type Widget struct {
	key                     string
	data                    []Data
	height                  unit.Dp
	innerRadius             float32
	outerRadius             float32
	clockwise               bool
	startAngle              float32
	padAngle                float32
	minAngle                float32
	roseType                RoseType
	stillShowZeroSum        bool
	showLabels              bool
	showLegend              bool
	hasShowLegend           bool
	showTooltip             bool
	animation               bool
	animationDuration       time.Duration
	animationEasing         animation.Easing
	updateAnimationDuration time.Duration
	updateAnimationEasing   animation.Easing
	onLegendChange          func(string, bool)
	onDataClick             func(chart.Selection)
	tooltipContent          func(chart.Selection) frame.Widget
	label                   string
	emptyText               string
	disabled                bool
}

func New(key string, data []Data) Widget {
	return Widget{
		key:                     key,
		data:                    append([]Data(nil), data...),
		outerRadius:             .5,
		clockwise:               true,
		startAngle:              90,
		stillShowZeroSum:        true,
		showLabels:              true,
		showTooltip:             true,
		animation:               true,
		animationDuration:       time.Second,
		animationEasing:         animation.EaseCubicInOut,
		updateAnimationDuration: 500 * time.Millisecond,
		updateAnimationEasing:   animation.EaseCubicInOut,
	}
}

func (w Widget) Height(dp int) Widget {
	if dp <= 0 {
		panic("flowui: pie chart height must be positive")
	}
	w.height = unit.Dp(dp)
	return w
}

// InnerRadius sets the donut hole as a ratio of the available chart radius.
func (w Widget) InnerRadius(ratio float32) Widget {
	validateRadius("inner radius", ratio, true)
	w.innerRadius = ratio
	return w
}

// OuterRadius sets the pie radius as a ratio of the available chart radius.
func (w Widget) OuterRadius(ratio float32) Widget {
	validateRadius("outer radius", ratio, false)
	w.outerRadius = ratio
	return w
}

func (w Widget) Clockwise(clockwise bool) Widget {
	w.clockwise = clockwise
	return w
}

// StartAngle sets the first slice angle in degrees. Ninety starts at the top.
func (w Widget) StartAngle(degrees float32) Widget {
	validateFinite("start angle", degrees)
	w.startAngle = degrees
	return w
}

func (w Widget) PadAngle(degrees float32) Widget {
	validateNonnegativeAngle("pad angle", degrees)
	w.padAngle = degrees
	return w
}

func (w Widget) MinAngle(degrees float32) Widget {
	validateNonnegativeAngle("minimum angle", degrees)
	w.minAngle = degrees
	return w
}

// RoseType enables an ECharts-compatible Nightingale rose layout. Radius
// keeps value-proportional angles; Area gives every slice an equal angle.
func (w Widget) RoseType(value RoseType) Widget {
	if value > RoseArea {
		panic("flowui: invalid pie chart rose type")
	}
	w.roseType = value
	return w
}

func (w Widget) StillShowZeroSum(show bool) Widget {
	w.stillShowZeroSum = show
	return w
}

func (w Widget) Labels(show bool) Widget {
	w.showLabels = show
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

func (w Widget) Animation(enabled bool) Widget {
	w.animation = enabled
	return w
}

func (w Widget) AnimationDuration(duration time.Duration) Widget {
	validateAnimationDuration(duration)
	w.animationDuration = duration
	return w
}

func (w Widget) AnimationEasing(easing animation.Easing) Widget {
	validateAnimationEasing(easing)
	w.animationEasing = easing
	return w
}

func (w Widget) UpdateAnimationDuration(duration time.Duration) Widget {
	validateAnimationDuration(duration)
	w.updateAnimationDuration = duration
	return w
}

func (w Widget) UpdateAnimationEasing(easing animation.Easing) Widget {
	validateAnimationEasing(easing)
	w.updateAnimationEasing = easing
	return w
}

// OnLegendChange registers a controlled slice visibility request. Hidden is
// the next value to pass to Data.Hidden.
func (w Widget) OnLegendChange(fn func(dataKey string, hidden bool)) Widget {
	w.onLegendChange = fn
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

func (w Widget) validateRadii() {
	if w.innerRadius >= w.outerRadius {
		panic("flowui: pie chart inner radius must be smaller than outer radius")
	}
}

func validateRadius(name string, value float32, allowZero bool) {
	validateFinite(name, value)
	if value < 0 || value > 1 || (!allowZero && value == 0) {
		panic(fmt.Sprintf("flowui: pie chart %s must be between 0 and 1", name))
	}
}

func validateNonnegativeAngle(name string, value float32) {
	validateFinite(name, value)
	if value < 0 {
		panic(fmt.Sprintf("flowui: pie chart %s must not be negative", name))
	}
}

func validateFinite(name string, value float32) {
	if math.IsNaN(float64(value)) || math.IsInf(float64(value), 0) {
		panic(fmt.Sprintf("flowui: pie chart %s must be finite", name))
	}
}

func validateAnimationDuration(duration time.Duration) {
	if duration < 0 {
		panic("flowui: pie chart animation duration must not be negative")
	}
}

func validateAnimationEasing(easing animation.Easing) {
	if easing == nil {
		panic("flowui: pie chart animation easing must not be nil")
	}
}
