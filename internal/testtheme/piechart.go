package testtheme

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) PieChart(active *theme.Theme, request theme.PieChartRequest) theme.PieChartVisual {
	tokens := active.Components.PieChart
	return theme.PieChartVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.PieChartOverrideStyle(request),
		Metrics: theme.PieChartMetrics{
			Height: tokens.Height, PlotPaddingTop: tokens.PlotPaddingTop, PlotPaddingRight: tokens.PlotPaddingRight,
			PlotPaddingBottom: tokens.PlotPaddingBottom, PlotPaddingLeft: tokens.PlotPaddingLeft,
			LegendTextSize: tokens.LegendTextSize, LegendMarkerSize: tokens.LegendMarkerSize,
			LegendMarkerGap: tokens.LegendMarkerGap, LegendMarkerRadius: tokens.LegendMarkerRadius,
			LegendItemGap: tokens.LegendItemGap, LegendLineGap: tokens.LegendLineGap, LegendGap: tokens.LegendGap,
			LabelTextSize: tokens.LabelTextSize, LabelLineLength: tokens.LabelLineLength, LabelLineLength2: tokens.LabelLineLength2,
			LabelLineWidth: tokens.LabelLineWidth, LabelGap: tokens.LabelGap, EmphasisSize: tokens.EmphasisSize,
			TooltipGap: tokens.TooltipGap, TooltipRowGap: tokens.TooltipRowGap, TooltipMarkerSize: tokens.TooltipMarkerSize,
			TooltipTextSize: active.Components.Tooltip.TextSize, TextWeight: font.Normal, EmphasisWeight: font.Medium,
		},
		Colors: theme.PieChartColors{
			Text: foreground, MutedText: muted, Empty: muted, LegendHover: background,
			OverlayForeground: foreground, Series: tokens.SeriesColors,
		},
		DisabledOpacity: active.DisabledOpacityValue(), LegendHiddenOpacity: .38, Motion: active.Motion,
	}
}
