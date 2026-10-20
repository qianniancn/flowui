package testtheme

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Heatmap(active *theme.Theme, request theme.HeatmapRequest) theme.HeatmapVisual {
	tokens := active.Components.Heatmap
	colors := theme.HeatmapOverrideColors(request.Override, theme.HeatmapColors{
		Min:               tokens.MinColor,
		Max:               tokens.MaxColor,
		Empty:             background,
		AxisText:          muted,
		Selection:         foreground,
		Muted:             muted,
		OverlayForeground: foreground,
	})
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
		Scale: func(minValue, maxValue, value float64) color.NRGBA {
			if maxValue <= minValue {
				return colors.Min
			}
			t := float32((value - minValue) / (maxValue - minValue))
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			return color.NRGBA{
				R: byte(float32(colors.Min.R) + (float32(colors.Max.R)-float32(colors.Min.R))*t),
				G: byte(float32(colors.Min.G) + (float32(colors.Max.G)-float32(colors.Min.G))*t),
				B: byte(float32(colors.Min.B) + (float32(colors.Max.B)-float32(colors.Min.B))*t),
				A: byte(float32(colors.Min.A) + (float32(colors.Max.A)-float32(colors.Min.A))*t),
			}
		},
		Motion: active.Motion,
	}
}
