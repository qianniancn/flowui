package theme

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SeriesColor resolves a default series color for behavior-owned data.
func (c BarChartColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// BarChartRequest contains the instance-level visual declaration for a bar
// chart. Values, stacking, scales, and pointer behavior remain widget-owned.
type BarChartRequest struct {
	Override *BarChartOverride
}

// BarChartOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type BarChartOverride struct {
	Declaration style.Style
}

// BarChartOverrideStyle extracts the provider-owned instance layer.
func BarChartOverrideStyle(request BarChartRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// BarChartMetrics contains geometry and typography consumed by bar-chart
// layout and drawing. It contains no theme or semantic objects.
type BarChartMetrics struct {
	AnimationDuration   time.Duration
	UpdateDuration      time.Duration
	Height              unit.Dp
	PlotPaddingTop      unit.Dp
	PlotPaddingRight    unit.Dp
	PlotPaddingBottom   unit.Dp
	PlotPaddingLeft     unit.Dp
	AxisNameGap         unit.Dp
	TickLabelGap        unit.Dp
	AxisTextSize        unit.Sp
	LegendTextSize      unit.Sp
	TextWeight          font.Weight
	EmphasisWeight      font.Weight
	LegendMarkerSize    unit.Dp
	LegendMarkerGap     unit.Dp
	LegendMarkerRadius  unit.Dp
	LegendItemGap       unit.Dp
	LegendLineGap       unit.Dp
	LegendGap           unit.Dp
	GridWidth           unit.Dp
	AxisWidth           unit.Dp
	AnnotationLineWidth unit.Dp
	AnnotationPointSize unit.Dp
	BarRadius           unit.Dp
	MinBarHeight        unit.Dp
	BackgroundRadius    unit.Dp
	TooltipGap          unit.Dp
	TooltipRowGap       unit.Dp
	TooltipMarkerSize   unit.Dp
	TooltipTextSize     unit.Sp
}

// BarChartLabelForeground resolves a readable foreground for a label drawn on
// top of a bar background supplied by the widget.
type BarChartLabelForeground func(color.NRGBA) color.NRGBA

// BarChartColors contains all palette decisions used by bar-chart behavior.
// Explicit series colors remain widget inputs; the provider owns the fallback
// palette and annotation/interaction colors.
type BarChartColors struct {
	Axis              color.NRGBA
	AxisLabel         color.NRGBA
	Grid              color.NRGBA
	CategoryHover     color.NRGBA
	BarBackground     color.NRGBA
	LegendHover       color.NRGBA
	AnnotationArea    color.NRGBA
	AnnotationLine    color.NRGBA
	AnnotationPoint   color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// BarChartVisual is the complete visual plan consumed by bar-chart behavior.
// The root declaration owns disabled opacity and instance styling; metrics
// and colors own the remaining visual recipe.
type BarChartVisual struct {
	Defaults            style.Style
	Override            style.Style
	Metrics             BarChartMetrics
	Colors              BarChartColors
	BarLabelForeground  BarChartLabelForeground
	DisabledOpacity     float32
	LegendHiddenOpacity float32
	Motion              MotionTheme
}

// BarChartProvider supplies bar-chart metrics and palette without taking
// ownership of values, scales, selection, or pointer behavior.
type BarChartProvider interface {
	BarChart(active *Theme, request BarChartRequest) BarChartVisual
}

// ValidateBarChartVisual checks the geometry, palette, and root declaration
// contract before a bar-chart visual enters its provider cache.
func ValidateBarChartVisual(visual BarChartVisual) {
	if visual.BarLabelForeground == nil {
		invalidVisual("BarChart", "BarLabelForeground", "must be provided")
	}

	for _, sample := range []color.NRGBA{{A: 0xff}, {R: 0xff, G: 0xff, B: 0xff, A: 0xff}} {
		if visual.BarLabelForeground(sample).A == 0 {
			invalidVisual("BarChart", "BarLabelForeground.A", "must be nonzero")
		}
	}
	m := visual.Metrics
	if !validPositiveDp(m.Height) {
		invalidVisual("BarChart", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.PlotPaddingTop) {
		invalidVisual("BarChart", "Metrics.PlotPaddingTop", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingRight) {
		invalidVisual("BarChart", "Metrics.PlotPaddingRight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingBottom) {
		invalidVisual("BarChart", "Metrics.PlotPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingLeft) {
		invalidVisual("BarChart", "Metrics.PlotPaddingLeft", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.AxisNameGap) {
		invalidVisual("BarChart", "Metrics.AxisNameGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TickLabelGap) {
		invalidVisual("BarChart", "Metrics.TickLabelGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.AxisTextSize) {
		invalidVisual("BarChart", "Metrics.AxisTextSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.LegendTextSize) {
		invalidVisual("BarChart", "Metrics.LegendTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LegendMarkerSize) {
		invalidVisual("BarChart", "Metrics.LegendMarkerSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.LegendMarkerGap) {
		invalidVisual("BarChart", "Metrics.LegendMarkerGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendMarkerRadius) {
		invalidVisual("BarChart", "Metrics.LegendMarkerRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendItemGap) {
		invalidVisual("BarChart", "Metrics.LegendItemGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendLineGap) {
		invalidVisual("BarChart", "Metrics.LegendLineGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendGap) {
		invalidVisual("BarChart", "Metrics.LegendGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.GridWidth) {
		invalidVisual("BarChart", "Metrics.GridWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AxisWidth) {
		invalidVisual("BarChart", "Metrics.AxisWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AnnotationLineWidth) {
		invalidVisual("BarChart", "Metrics.AnnotationLineWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AnnotationPointSize) {
		invalidVisual("BarChart", "Metrics.AnnotationPointSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.BarRadius) {
		invalidVisual("BarChart", "Metrics.BarRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.MinBarHeight) {
		invalidVisual("BarChart", "Metrics.MinBarHeight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.BackgroundRadius) {
		invalidVisual("BarChart", "Metrics.BackgroundRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipGap) {
		invalidVisual("BarChart", "Metrics.TooltipGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipRowGap) {
		invalidVisual("BarChart", "Metrics.TooltipRowGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.TooltipMarkerSize) {
		invalidVisual("BarChart", "Metrics.TooltipMarkerSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.TooltipTextSize) {
		invalidVisual("BarChart", "Metrics.TooltipTextSize", "must be finite and greater than 0")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("BarChart", "DisabledOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.LegendHiddenOpacity) {
		invalidVisual("BarChart", "LegendHiddenOpacity", "must be finite and between 0 and 1")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("BarChart", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("BarChart", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("BarChart", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
