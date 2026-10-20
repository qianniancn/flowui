package theme

import (
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// SeriesColor resolves a default series color for behavior-owned data.
func (c LineChartColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// LineChartRequest contains the instance-level visual declaration for a line
// chart. Series data, scales, and interaction remain widget-owned.
type LineChartRequest struct {
	Override *LineChartOverride
}

// LineChartOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type LineChartOverride struct {
	Declaration style.Style
}

// LineChartOverrideStyle extracts the provider-owned instance layer.
func LineChartOverrideStyle(request LineChartRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// LineChartMetrics contains geometry and typography consumed by line-chart
// layout and drawing. It contains no theme or semantic objects.
type LineChartMetrics struct {
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
	TextWeight          font.Weight
	EmphasisWeight      font.Weight
	LegendTextSize      unit.Sp
	LegendMarkerWidth   unit.Dp
	LegendMarkerSize    unit.Dp
	LegendItemGap       unit.Dp
	LegendLineGap       unit.Dp
	LegendGap           unit.Dp
	GridWidth           unit.Dp
	AxisWidth           unit.Dp
	LineWidth           unit.Dp
	PointSize           unit.Dp
	AnnotationLineWidth unit.Dp
	AnnotationPointSize unit.Dp
	HoverPointSize      unit.Dp
	CrosshairWidth      unit.Dp
	TooltipGap          unit.Dp
	TooltipRowGap       unit.Dp
	TooltipMarkerSize   unit.Dp
	TooltipTextSize     unit.Sp
}

// LineChartColors contains all palette decisions used by line-chart
// behavior. The widget may apply explicit series colors, but never chooses a
// fallback palette or semantic fallback color itself.
type LineChartColors struct {
	Axis              color.NRGBA
	AxisLabel         color.NRGBA
	Grid              color.NRGBA
	Crosshair         color.NRGBA
	Background        color.NRGBA
	LegendHover       color.NRGBA
	AnnotationArea    color.NRGBA
	AnnotationLine    color.NRGBA
	AnnotationPoint   color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// LineChartVisual is the complete visual plan consumed by line-chart
// behavior. The root declaration owns disabled opacity and instance styling;
// metrics and colors own the remaining visual recipe.
type LineChartVisual struct {
	Defaults            style.Style
	Override            style.Style
	Metrics             LineChartMetrics
	Colors              LineChartColors
	DisabledOpacity     float32
	LegendHiddenOpacity float32
	AreaOpacity         float32
	Motion              MotionTheme
}

// LineChartProvider supplies line-chart metrics and palette without taking
// ownership of data, scales, selection, or pointer behavior.
type LineChartProvider interface {
	LineChart(active *Theme, request LineChartRequest) LineChartVisual
}

// ValidateLineChartVisual checks the geometry, palette, and root declaration
// contract before a line-chart visual enters its provider cache.
func ValidateLineChartVisual(visual LineChartVisual) {
	m := visual.Metrics
	if !validPositiveDp(m.Height) {
		invalidVisual("LineChart", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.PlotPaddingTop) {
		invalidVisual("LineChart", "Metrics.PlotPaddingTop", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingRight) {
		invalidVisual("LineChart", "Metrics.PlotPaddingRight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingBottom) {
		invalidVisual("LineChart", "Metrics.PlotPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingLeft) {
		invalidVisual("LineChart", "Metrics.PlotPaddingLeft", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.AxisNameGap) {
		invalidVisual("LineChart", "Metrics.AxisNameGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TickLabelGap) {
		invalidVisual("LineChart", "Metrics.TickLabelGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.AxisTextSize) {
		invalidVisual("LineChart", "Metrics.AxisTextSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.LegendTextSize) {
		invalidVisual("LineChart", "Metrics.LegendTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LegendMarkerWidth) {
		invalidVisual("LineChart", "Metrics.LegendMarkerWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LegendMarkerSize) {
		invalidVisual("LineChart", "Metrics.LegendMarkerSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.LegendItemGap) {
		invalidVisual("LineChart", "Metrics.LegendItemGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendLineGap) {
		invalidVisual("LineChart", "Metrics.LegendLineGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendGap) {
		invalidVisual("LineChart", "Metrics.LegendGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.GridWidth) {
		invalidVisual("LineChart", "Metrics.GridWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AxisWidth) {
		invalidVisual("LineChart", "Metrics.AxisWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LineWidth) {
		invalidVisual("LineChart", "Metrics.LineWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.PointSize) {
		invalidVisual("LineChart", "Metrics.PointSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AnnotationLineWidth) {
		invalidVisual("LineChart", "Metrics.AnnotationLineWidth", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.AnnotationPointSize) {
		invalidVisual("LineChart", "Metrics.AnnotationPointSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.HoverPointSize) {
		invalidVisual("LineChart", "Metrics.HoverPointSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.CrosshairWidth) {
		invalidVisual("LineChart", "Metrics.CrosshairWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.TooltipGap) {
		invalidVisual("LineChart", "Metrics.TooltipGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipRowGap) {
		invalidVisual("LineChart", "Metrics.TooltipRowGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.TooltipMarkerSize) {
		invalidVisual("LineChart", "Metrics.TooltipMarkerSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.TooltipTextSize) {
		invalidVisual("LineChart", "Metrics.TooltipTextSize", "must be finite and greater than 0")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("LineChart", "DisabledOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.LegendHiddenOpacity) {
		invalidVisual("LineChart", "LegendHiddenOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.AreaOpacity) {
		invalidVisual("LineChart", "AreaOpacity", "must be finite and between 0 and 1")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("LineChart", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("LineChart", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("LineChart", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
