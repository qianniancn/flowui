// Package piechartstyle contains the visual recipe for pie-chart providers.
// Pie-chart behavior stays in widget/piechart; this package owns geometry,
// typography, palette, and the instance declaration boundary.
package piechartstyle

import (
	"image/color"

	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildPieChart returns the complete visual plan for the active theme.
func BuildPieChart(active *theme.Theme, request theme.PieChartRequest) theme.PieChartVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.PieChart
	semantic := active.Semantic()
	return theme.PieChartVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.PieChartOverrideStyle(request),
		Metrics: theme.PieChartMetrics{
			Height:             tokens.Height,
			PlotPaddingTop:     tokens.PlotPaddingTop,
			PlotPaddingRight:   tokens.PlotPaddingRight,
			PlotPaddingBottom:  tokens.PlotPaddingBottom,
			PlotPaddingLeft:    tokens.PlotPaddingLeft,
			LegendTextSize:     tokens.LegendTextSize,
			LegendMarkerSize:   tokens.LegendMarkerSize,
			LegendMarkerGap:    tokens.LegendMarkerGap,
			LegendMarkerRadius: tokens.LegendMarkerRadius,
			LegendItemGap:      tokens.LegendItemGap,
			LegendLineGap:      tokens.LegendLineGap,
			LegendGap:          tokens.LegendGap,
			LabelTextSize:      tokens.LabelTextSize,
			LabelLineLength:    tokens.LabelLineLength,
			LabelLineLength2:   tokens.LabelLineLength2,
			LabelLineWidth:     tokens.LabelLineWidth,
			LabelGap:           tokens.LabelGap,
			EmphasisSize:       tokens.EmphasisSize,
			TooltipGap:         tokens.TooltipGap,
			TooltipRowGap:      tokens.TooltipRowGap,
			TooltipMarkerSize:  tokens.TooltipMarkerSize,
			TooltipTextSize:    active.Components.Tooltip.TextSize,
			TextWeight:         font.Normal,
			EmphasisWeight:     font.Medium,
		},
		Colors: theme.PieChartColors{
			Text:              semantic.Foreground,
			MutedText:         semantic.MutedForeground,
			Empty:             withAlpha(semantic.SurfaceSecondary, .72),
			LegendHover:       semantic.SurfaceHover,
			OverlayForeground: semantic.OverlayForeground,
			Series:            tokens.SeriesColors,
		},
		DisabledOpacity:     active.DisabledOpacityValue(),
		LegendHiddenOpacity: 0.38,
		Motion:              active.Motion,
	}
}

func withAlpha(value color.NRGBA, factor float32) color.NRGBA {
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	value.A = uint8(float32(value.A) * factor)
	return value
}
