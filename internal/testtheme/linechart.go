package testtheme

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) LineChart(active *theme.Theme, request theme.LineChartRequest) theme.LineChartVisual {
	tokens := active.Components.LineChart
	return theme.LineChartVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.LineChartOverrideStyle(request),
		Metrics: theme.LineChartMetrics{
			Height: tokens.Height, PlotPaddingTop: tokens.PlotPaddingTop, PlotPaddingRight: tokens.PlotPaddingRight,
			PlotPaddingBottom: tokens.PlotPaddingBottom, PlotPaddingLeft: tokens.PlotPaddingLeft, AxisNameGap: tokens.AxisNameGap,
			TickLabelGap: tokens.TickLabelGap, AxisTextSize: tokens.AxisTextSize, TextWeight: font.Normal, EmphasisWeight: font.Medium, LegendTextSize: tokens.LegendTextSize,
			LegendMarkerWidth: tokens.LegendMarkerWidth, LegendMarkerSize: tokens.LegendMarkerSize, LegendItemGap: tokens.LegendItemGap,
			LegendLineGap: tokens.LegendLineGap, LegendGap: tokens.LegendGap, GridWidth: tokens.GridWidth, AxisWidth: tokens.AxisWidth,
			LineWidth: tokens.LineWidth, PointSize: tokens.PointSize, HoverPointSize: tokens.HoverPointSize, CrosshairWidth: tokens.CrosshairWidth,
			AnnotationLineWidth: tokens.AnnotationLineWidth, AnnotationPointSize: tokens.AnnotationPointSize,
			TooltipGap: tokens.TooltipGap, TooltipRowGap: tokens.TooltipRowGap, TooltipMarkerSize: tokens.TooltipMarkerSize,
			TooltipTextSize: active.Components.Tooltip.TextSize,
		},
		Colors: theme.LineChartColors{
			Axis: accent, AxisLabel: muted, Grid: muted, Crosshair: focus, Background: background, LegendHover: background,
			AnnotationArea: accent, AnnotationLine: focus, AnnotationPoint: accent, OverlayForeground: foreground, Series: tokens.SeriesColors,
		},
		DisabledOpacity: active.DisabledOpacityValue(), LegendHiddenOpacity: 0.38, AreaOpacity: 0x30 / 255.0, Motion: active.Motion,
	}
}
