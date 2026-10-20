// Package heatmapstyle contains the visual recipe for heatmap providers.
// Heatmap behavior stays in widget/heatmap; this package owns geometry,
// typography, palette, and the instance declaration boundary.
package heatmapstyle

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildHeatmap returns the complete visual plan for the active theme.
func BuildHeatmap(active *theme.Theme, request theme.HeatmapRequest) theme.HeatmapVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Heatmap
	semantic := active.Semantic()
	empty := tokens.EmptyColor
	if empty.A == 0 {
		empty = semantic.SurfaceSecondary
	}
	colors := theme.HeatmapColors{
		Min:               tokens.MinColor,
		Max:               tokens.MaxColor,
		Empty:             empty,
		AxisText:          semantic.MutedForeground,
		Selection:         semantic.Foreground,
		Muted:             semantic.MutedForeground,
		OverlayForeground: semantic.OverlayForeground,
	}
	colors = theme.HeatmapOverrideColors(request.Override, colors)
	return theme.HeatmapVisual{
		Defaults: style.Style{}.
			FillWidth().
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.HeatmapOverrideStyle(request.Override),
		Metrics: theme.HeatmapMetrics{
			Height:               tokens.Height,
			PlotPaddingTop:       tokens.PlotPaddingTop,
			PlotPaddingRight:     tokens.PlotPaddingRight,
			PlotPaddingBottom:    tokens.PlotPaddingBottom,
			PlotPaddingLeft:      tokens.PlotPaddingLeft,
			AxisTextSize:         tokens.AxisTextSize,
			AxisFontWeight:       tokens.AxisFontWeight,
			TickLabelGap:         tokens.TickLabelGap,
			CellSize:             tokens.CellSize,
			CellGap:              tokens.CellGap,
			CellRadius:           tokens.CellRadius,
			SelectionBorderWidth: tokens.SelectionBorderWidth,
			TooltipGap:           tokens.TooltipGap,
			TooltipTextSize:      active.Components.Tooltip.TextSize,
			TooltipMarkerSize:    tokens.TooltipMarkerSize,
			TooltipRowGap:        tokens.TooltipRowGap,
			EmptyTextSize:        tokens.AxisTextSize,
			EmptyFontWeight:      tokens.EmptyFontWeight,
		},
		Colors: colors,
		Scale:  linearColorScale(colors.Min, colors.Max),
		Motion: active.Motion,
	}
}

func linearColorScale(minColor, maxColor color.NRGBA) theme.HeatmapColorScale {
	return func(minValue, maxValue, value float64) color.NRGBA {
		if maxValue <= minValue {
			return minColor
		}
		t := float32((value - minValue) / (maxValue - minValue))
		if t < 0 {
			t = 0
		}
		if t > 1 {
			t = 1
		}
		return color.NRGBA{
			R: byte(float32(minColor.R) + (float32(maxColor.R)-float32(minColor.R))*t),
			G: byte(float32(minColor.G) + (float32(maxColor.G)-float32(minColor.G))*t),
			B: byte(float32(minColor.B) + (float32(maxColor.B)-float32(minColor.B))*t),
			A: byte(float32(minColor.A) + (float32(maxColor.A)-float32(minColor.A))*t),
		}
	}
}
