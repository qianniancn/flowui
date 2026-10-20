package theme

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// PieChartRequest contains the instance-level visual declaration for a pie
// chart. Slice data, angles, selection, and pointer behavior remain widget-owned.
type PieChartRequest struct {
	Override *PieChartOverride
}

// PieChartOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type PieChartOverride struct {
	Declaration style.Style
}

// PieChartOverrideStyle extracts the provider-owned instance layer.
func PieChartOverrideStyle(request PieChartRequest) style.Style {
	if request.Override == nil {
		return style.Style{}
	}
	return request.Override.Declaration
}

// PieChartMetrics contains geometry and typography consumed by pie-chart
// layout and drawing. It contains no theme or semantic objects.
type PieChartMetrics struct {
	Height             unit.Dp
	PlotPaddingTop     unit.Dp
	PlotPaddingRight   unit.Dp
	PlotPaddingBottom  unit.Dp
	PlotPaddingLeft    unit.Dp
	LegendTextSize     unit.Sp
	LegendMarkerSize   unit.Dp
	LegendMarkerGap    unit.Dp
	LegendMarkerRadius unit.Dp
	LegendItemGap      unit.Dp
	LegendLineGap      unit.Dp
	LegendGap          unit.Dp
	LabelTextSize      unit.Sp
	LabelLineLength    unit.Dp
	LabelLineLength2   unit.Dp
	LabelLineWidth     unit.Dp
	LabelGap           unit.Dp
	EmphasisSize       unit.Dp
	TooltipGap         unit.Dp
	TooltipRowGap      unit.Dp
	TooltipMarkerSize  unit.Dp
	TooltipTextSize    unit.Sp
	TextWeight         font.Weight
	EmphasisWeight     font.Weight
}

// PieChartColors contains all palette decisions used by pie-chart behavior.
// Explicit slice colors remain widget inputs; the provider owns fallback colors.
type PieChartColors struct {
	Text              color.NRGBA
	MutedText         color.NRGBA
	Empty             color.NRGBA
	LegendHover       color.NRGBA
	OverlayForeground color.NRGBA
	Series            [9]color.NRGBA
}

// SeriesColor resolves a default slice color for behavior-owned data.
func (c PieChartColors) SeriesColor(index int) color.NRGBA {
	return c.Series[index%len(c.Series)]
}

// PieChartVisual is the complete visual plan consumed by pie-chart behavior.
type PieChartVisual struct {
	Defaults            style.Style
	Override            style.Style
	Metrics             PieChartMetrics
	Colors              PieChartColors
	DisabledOpacity     float32
	LegendHiddenOpacity float32
	Motion              MotionTheme
}

// PieChartProvider supplies pie-chart metrics and palette without taking
// ownership of slice data, angles, selection, or pointer behavior.
type PieChartProvider interface {
	PieChart(active *Theme, request PieChartRequest) PieChartVisual
}

// ValidatePieChartVisual checks the geometry, palette, and root declaration
// contract before a pie-chart visual enters its provider cache.
func ValidatePieChartVisual(visual PieChartVisual) {
	m := visual.Metrics
	if !validPositiveDp(m.Height) {
		invalidVisual("PieChart", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.PlotPaddingTop) {
		invalidVisual("PieChart", "Metrics.PlotPaddingTop", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingRight) {
		invalidVisual("PieChart", "Metrics.PlotPaddingRight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingBottom) {
		invalidVisual("PieChart", "Metrics.PlotPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.PlotPaddingLeft) {
		invalidVisual("PieChart", "Metrics.PlotPaddingLeft", "must be finite and non-negative")
	}
	if !validPositiveSp(m.LegendTextSize) {
		invalidVisual("PieChart", "Metrics.LegendTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LegendMarkerSize) {
		invalidVisual("PieChart", "Metrics.LegendMarkerSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.LegendMarkerGap) {
		invalidVisual("PieChart", "Metrics.LegendMarkerGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendMarkerRadius) {
		invalidVisual("PieChart", "Metrics.LegendMarkerRadius", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendItemGap) {
		invalidVisual("PieChart", "Metrics.LegendItemGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendLineGap) {
		invalidVisual("PieChart", "Metrics.LegendLineGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.LegendGap) {
		invalidVisual("PieChart", "Metrics.LegendGap", "must be finite and non-negative")
	}
	if !validPositiveSp(m.LabelTextSize) {
		invalidVisual("PieChart", "Metrics.LabelTextSize", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LabelLineLength) {
		invalidVisual("PieChart", "Metrics.LabelLineLength", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LabelLineLength2) {
		invalidVisual("PieChart", "Metrics.LabelLineLength2", "must be finite and greater than 0")
	}
	if !validPositiveDp(m.LabelLineWidth) {
		invalidVisual("PieChart", "Metrics.LabelLineWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.LabelGap) {
		invalidVisual("PieChart", "Metrics.LabelGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.EmphasisSize) {
		invalidVisual("PieChart", "Metrics.EmphasisSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(m.TooltipGap) {
		invalidVisual("PieChart", "Metrics.TooltipGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(m.TooltipRowGap) {
		invalidVisual("PieChart", "Metrics.TooltipRowGap", "must be finite and non-negative")
	}
	if !validPositiveDp(m.TooltipMarkerSize) {
		invalidVisual("PieChart", "Metrics.TooltipMarkerSize", "must be finite and greater than 0")
	}
	if !validPositiveSp(m.TooltipTextSize) {
		invalidVisual("PieChart", "Metrics.TooltipTextSize", "must be finite and greater than 0")
	}
	if !hasValidOpacityValue(visual.DisabledOpacity) {
		invalidVisual("PieChart", "DisabledOpacity", "must be finite and between 0 and 1")
	}
	if !hasValidOpacityValue(visual.LegendHiddenOpacity) {
		invalidVisual("PieChart", "LegendHiddenOpacity", "must be finite and between 0 and 1")
	}

	if visual.Colors.Text.A == 0 {
		invalidVisual("PieChart", "Colors.Text", "must have nonzero alpha")
	}
	if visual.Colors.MutedText.A == 0 {
		invalidVisual("PieChart", "Colors.MutedText", "must have nonzero alpha")
	}
	if visual.Colors.Empty.A == 0 {
		invalidVisual("PieChart", "Colors.Empty", "must have nonzero alpha")
	}
	if visual.Colors.LegendHover.A == 0 {
		invalidVisual("PieChart", "Colors.LegendHover", "must have nonzero alpha")
	}
	if visual.Colors.OverlayForeground.A == 0 {
		invalidVisual("PieChart", "Colors.OverlayForeground", "must have nonzero alpha")
	}
	validSeries := false
	for _, value := range visual.Colors.Series {
		if value.A != 0 {
			validSeries = true
			break
		}
	}
	if !validSeries {
		invalidVisual("PieChart", "Colors.Series", "must contain at least one color with nonzero alpha")
	}

	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("PieChart", "Root.Box", "must be provided")
	}
	if root.Paint == nil {
		invalidVisual("PieChart", "Root.Paint", "must be provided")
	}
	if !hasValidOpacity(root.Paint.Opacity) {
		invalidVisual("PieChart", "Root.Paint.Opacity", "must be finite and between 0 and 1")
	}
}
