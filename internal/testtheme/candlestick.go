package testtheme

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Candlestick(active *theme.Theme, request theme.CandlestickRequest) theme.CandlestickVisual {
	tokens := active.Components.CandlestickChart
	return theme.CandlestickVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(background)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.Opacity(active.DisabledOpacityValue())}),
		Override: theme.CandlestickOverrideStyle(request),
		Metrics: theme.CandlestickMetrics{
			Height: tokens.Height, PlotPaddingTop: tokens.PlotPaddingTop, PlotPaddingRight: tokens.PlotPaddingRight,
			PlotPaddingBottom: tokens.PlotPaddingBottom, PlotPaddingLeft: tokens.PlotPaddingLeft,
			AxisNameGap: tokens.AxisNameGap, TickLabelGap: tokens.TickLabelGap, AxisTextSize: tokens.AxisTextSize,
			GridWidth: tokens.GridWidth, AxisWidth: tokens.AxisWidth, CrosshairWidth: tokens.CrosshairWidth,
			CrosshairLabelPadding: tokens.CrosshairLabelPadding, CrosshairLabelRadius: active.Components.Tooltip.Radius,
			AnnotationLineWidth: tokens.AnnotationLineWidth, AnnotationPointSize: tokens.AnnotationPointSize,
			WickWidth: tokens.WickWidth, BorderWidth: tokens.BorderWidth, EmphasisBorderWidth: tokens.EmphasisBorderWidth,
			TooltipGap: tokens.TooltipGap, TooltipRowGap: tokens.TooltipRowGap, TooltipTextSize: active.Components.Tooltip.TextSize,
			TextWeight: font.Normal, EmphasisWeight: font.Medium,
		},
		Colors: theme.CandlestickColors{
			Axis: accent, AxisLabel: muted, Grid: muted, Crosshair: focus,
			CrosshairLabelBackground: muted, CrosshairLabelForeground: background, OverlayForeground: foreground,
			AnnotationArea: accent, AnnotationLine: focus, AnnotationPoint: accent,
			Up: tokens.UpColor, Down: tokens.DownColor, Doji: accent,
		},
		DisabledOpacity: active.DisabledOpacityValue(), Motion: active.Motion,
	}
}
