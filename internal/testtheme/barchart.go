package testtheme

import (
	"image/color"

	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) BarChart(active *theme.Theme, request theme.BarChartRequest) theme.BarChartVisual {
	tokens := active.Components.BarChart
	return theme.BarChartVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.BarChartOverrideStyle(request),
		Metrics: theme.BarChartMetrics{
			Height: tokens.Height, PlotPaddingTop: tokens.PlotPaddingTop, PlotPaddingRight: tokens.PlotPaddingRight,
			PlotPaddingBottom: tokens.PlotPaddingBottom, PlotPaddingLeft: tokens.PlotPaddingLeft, AxisNameGap: tokens.AxisNameGap,
			TickLabelGap: tokens.TickLabelGap, AxisTextSize: tokens.AxisTextSize, LegendTextSize: tokens.LegendTextSize,
			TextWeight: font.Normal, EmphasisWeight: font.Medium,
			LegendMarkerSize: tokens.LegendMarkerSize, LegendMarkerGap: tokens.LegendMarkerGap, LegendMarkerRadius: tokens.LegendMarkerRadius,
			LegendItemGap: tokens.LegendItemGap, LegendLineGap: tokens.LegendLineGap, LegendGap: tokens.LegendGap,
			GridWidth: tokens.GridWidth, AxisWidth: tokens.AxisWidth, BarRadius: tokens.BarRadius, MinBarHeight: tokens.MinBarHeight,
			AnnotationLineWidth: tokens.AnnotationLineWidth, AnnotationPointSize: tokens.AnnotationPointSize,
			BackgroundRadius: tokens.BackgroundRadius, TooltipGap: tokens.TooltipGap, TooltipRowGap: tokens.TooltipRowGap,
			TooltipMarkerSize: tokens.TooltipMarkerSize, TooltipTextSize: active.Components.Tooltip.TextSize,
		},
		Colors: theme.BarChartColors{
			Axis: accent, AxisLabel: muted, Grid: muted, CategoryHover: accent, BarBackground: muted,
			LegendHover: background, AnnotationArea: accent, AnnotationLine: focus, AnnotationPoint: accent,
			OverlayForeground: foreground, Series: tokens.SeriesColors,
		},
		BarLabelForeground: func(color.NRGBA) color.NRGBA { return foreground },
		DisabledOpacity:    active.DisabledOpacityValue(), LegendHiddenOpacity: 0.38, Motion: active.Motion,
	}
}
