package theme

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/style"
)

// HeatmapRequest contains the instance-level visual declaration for a
// heatmap. Data ranges, cell values, and pointer state remain widget-owned.
type HeatmapRequest struct {
	Override *HeatmapOverride
}

// HeatmapOverride carries caller customization across the neutral provider
// boundary. The provider owns the returned instance layer.
type HeatmapOverride struct {
	Declaration style.Style
	MinColor    color.NRGBA
	MaxColor    color.NRGBA
	HasMinColor bool
	HasMaxColor bool
}

// HeatmapOverrideStyle extracts the provider-owned instance layer.
func HeatmapOverrideStyle(override *HeatmapOverride) style.Style {
	if override != nil {
		return override.Declaration
	}
	return style.Style{}
}

// HeatmapOverrideColors applies provider-owned instance color overrides to a
// theme's base palette. The input palette is copied so callers can safely
// reuse their base visual plan.
func HeatmapOverrideColors(override *HeatmapOverride, colors HeatmapColors) HeatmapColors {
	if override == nil {
		return colors
	}
	if override.HasMinColor {
		colors.Min = override.MinColor
	}
	if override.HasMaxColor {
		colors.Max = override.MaxColor
	}
	return colors
}

// HeatmapMetrics contains geometry and typography consumed by heatmap layout.
// The widget owns cell placement and interaction geometry.
type HeatmapMetrics struct {
	Height               unit.Dp
	PlotPaddingTop       unit.Dp
	PlotPaddingRight     unit.Dp
	PlotPaddingBottom    unit.Dp
	PlotPaddingLeft      unit.Dp
	AxisTextSize         unit.Sp
	AxisFontWeight       font.Weight
	TickLabelGap         unit.Dp
	CellSize             unit.Dp
	CellGap              unit.Dp
	CellRadius           unit.Dp
	SelectionBorderWidth unit.Dp
	TooltipGap           unit.Dp
	TooltipTextSize      unit.Sp
	TooltipMarkerSize    unit.Dp
	TooltipRowGap        unit.Dp
	EmptyTextSize        unit.Sp
	EmptyFontWeight      font.Weight
}

// HeatmapColors contains provider-owned colors consumed by heatmap drawing.
type HeatmapColors struct {
	Min               color.NRGBA
	Max               color.NRGBA
	Empty             color.NRGBA
	AxisText          color.NRGBA
	Selection         color.NRGBA
	Muted             color.NRGBA
	OverlayForeground color.NRGBA
}

// HeatmapColorScale maps a data value in the supplied range to a provider
// owned display color. The widget supplies data bounds but does not interpret
// the visual scale.
type HeatmapColorScale func(minValue, maxValue, value float64) color.NRGBA

// HeatmapVisual is the complete visual plan consumed by heatmap behavior.
type HeatmapVisual struct {
	Defaults style.Style
	Override style.Style
	Metrics  HeatmapMetrics
	Colors   HeatmapColors
	Scale    HeatmapColorScale
	Motion   MotionTheme
}

// HeatmapProvider supplies heatmap geometry, colors, and motion policy
// without taking ownership of values, selection, or pointer behavior.
type HeatmapProvider interface {
	Heatmap(active *Theme, request HeatmapRequest) HeatmapVisual
}

// ValidateHeatmapVisual checks the geometry, scale and motion contract before
// a heatmap visual enters its provider cache. Heatmap behavior owns calendar
// values and pointer selection; it must not recover missing visual metrics in
// the draw path.
func ValidateHeatmapVisual(visual HeatmapVisual) {
	metrics := visual.Metrics
	if !validPositiveDp(metrics.Height) {
		invalidVisual("Heatmap", "Metrics.Height", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.PlotPaddingTop) {
		invalidVisual("Heatmap", "Metrics.PlotPaddingTop", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.PlotPaddingRight) {
		invalidVisual("Heatmap", "Metrics.PlotPaddingRight", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.PlotPaddingBottom) {
		invalidVisual("Heatmap", "Metrics.PlotPaddingBottom", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.PlotPaddingLeft) {
		invalidVisual("Heatmap", "Metrics.PlotPaddingLeft", "must be finite and non-negative")
	}
	if !validPositiveSp(metrics.AxisTextSize) {
		invalidVisual("Heatmap", "Metrics.AxisTextSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.TickLabelGap) {
		invalidVisual("Heatmap", "Metrics.TickLabelGap", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.CellSize) {
		invalidVisual("Heatmap", "Metrics.CellSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.CellGap) {
		invalidVisual("Heatmap", "Metrics.CellGap", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.CellRadius) {
		invalidVisual("Heatmap", "Metrics.CellRadius", "must be finite and non-negative")
	}
	if !validPositiveDp(metrics.SelectionBorderWidth) {
		invalidVisual("Heatmap", "Metrics.SelectionBorderWidth", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.TooltipGap) {
		invalidVisual("Heatmap", "Metrics.TooltipGap", "must be finite and non-negative")
	}
	if !validPositiveSp(metrics.TooltipTextSize) {
		invalidVisual("Heatmap", "Metrics.TooltipTextSize", "must be finite and greater than 0")
	}
	if !validNonNegativeDp(metrics.TooltipMarkerSize) {
		invalidVisual("Heatmap", "Metrics.TooltipMarkerSize", "must be finite and non-negative")
	}
	if !validNonNegativeDp(metrics.TooltipRowGap) {
		invalidVisual("Heatmap", "Metrics.TooltipRowGap", "must be finite and non-negative")
	}
	if !validPositiveSp(metrics.EmptyTextSize) {
		invalidVisual("Heatmap", "Metrics.EmptyTextSize", "must be finite and greater than 0")
	}
	if visual.Scale == nil {
		invalidVisual("Heatmap", "Scale", "must be provided")
	}

	validateMotion("Heatmap", "Motion", visual.Motion)
	root := style.Cascade(style.State{}, visual.Defaults, visual.Override)
	if root.Box == nil {
		invalidVisual("Heatmap", "Root.Box", "must be provided")
	}
}
