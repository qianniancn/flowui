package piechart

import (
	"fmt"
	"image/color"
	"math"
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/animation"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/chart"
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
	override                *theme.PieChartOverride
	key                     string
	data                    []Data
	dataVersion             uint64
	hasDataVersion          bool
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
	tooltipContent          func(chart.Selection) core.Widget
	visualCache             *pieChartVisualCache
	label                   string
	emptyText               string
	hasEmptyText            bool
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
		visualCache:             &pieChartVisualCache{},
	}
}

// DataVersion enables resolved-data reuse. Increase version whenever the data
// or angle, radius, and visibility options change.
func (w Widget) DataVersion(version uint64) Widget {
	w.dataVersion = version
	w.hasDataVersion = true
	return w
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
	chart.ValidateAnimationDuration(duration, "pie chart")
	w.animationDuration = duration
	return w
}

func (w Widget) AnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "pie chart")
	w.animationEasing = easing
	return w
}

func (w Widget) UpdateAnimationDuration(duration time.Duration) Widget {
	chart.ValidateAnimationDuration(duration, "pie chart")
	w.updateAnimationDuration = duration
	return w
}

func (w Widget) UpdateAnimationEasing(easing animation.Easing) Widget {
	chart.ValidateAnimationEasing(easing, "pie chart")
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

func (w Widget) TooltipContent(fn func(chart.Selection) core.Widget) Widget {
	w.tooltipContent = fn
	return w
}

func (w Widget) Label(label string) Widget {
	w.label = label
	return w
}

func (w Widget) EmptyText(text string) Widget {
	w.emptyText = text
	w.hasEmptyText = true
	return w
}

func (w Widget) Disabled(disabled bool) Widget {
	w.disabled = disabled
	return w
}

func (w Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return w.layoutWithOverride(ctx, gtx, w.override)
}

func (w Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, override *theme.PieChartOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	active := core.ActiveTheme(ctx)
	visual := w.resolveVisual(active, override)
	resolved := computed.ResolveStatic(ctx, style.State{Disabled: w.disabled}, visual.Defaults, style.Style{}, style.Style{}, visual.Override)
	return layoutui.Resolved(ctx, gtx, resolved, pieChartVisualWidget{value: w, visual: visual})
}

type pieChartVisualWidget struct {
	value  Widget
	visual *theme.PieChartVisual
}

func (v pieChartVisualWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return v.value.layout(ctx, gtx, *v.visual)
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

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.PieChartOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
