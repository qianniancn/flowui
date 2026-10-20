// Package linechartstyle contains the visual recipe for line-chart providers.
// Line-chart behavior stays in widget/linechart; this package owns geometry,
// typography, palette, and the instance declaration boundary.
package linechartstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildLineChart returns the complete visual plan for the active theme.
func BuildLineChart(active *theme.Theme, request theme.LineChartRequest) theme.LineChartVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.LineChart
	semantic := active.Semantic()
	grid := withAlpha(semantic.Border, 0.8)
	crosshair := withAlpha(semantic.MutedForeground, 0.75)
	return theme.LineChartVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.LineChartOverrideStyle(request),
		Metrics: theme.LineChartMetrics{
			AnimationDuration:   1 * time.Second,
			UpdateDuration:      500 * time.Millisecond,
			Height:              tokens.Height,
			PlotPaddingTop:      tokens.PlotPaddingTop,
			PlotPaddingRight:    tokens.PlotPaddingRight,
			PlotPaddingBottom:   tokens.PlotPaddingBottom,
			PlotPaddingLeft:     tokens.PlotPaddingLeft,
			AxisNameGap:         tokens.AxisNameGap,
			TickLabelGap:        tokens.TickLabelGap,
			AxisTextSize:        tokens.AxisTextSize,
			TextWeight:          font.Normal,
			EmphasisWeight:      font.Medium,
			LegendTextSize:      tokens.LegendTextSize,
			LegendMarkerWidth:   tokens.LegendMarkerWidth,
			LegendMarkerSize:    tokens.LegendMarkerSize,
			LegendItemGap:       tokens.LegendItemGap,
			LegendLineGap:       tokens.LegendLineGap,
			LegendGap:           tokens.LegendGap,
			GridWidth:           tokens.GridWidth,
			AxisWidth:           tokens.AxisWidth,
			LineWidth:           tokens.LineWidth,
			PointSize:           tokens.PointSize,
			AnnotationLineWidth: tokens.AnnotationLineWidth,
			AnnotationPointSize: tokens.AnnotationPointSize,
			HoverPointSize:      tokens.HoverPointSize,
			CrosshairWidth:      tokens.CrosshairWidth,
			TooltipGap:          tokens.TooltipGap,
			TooltipRowGap:       tokens.TooltipRowGap,
			TooltipMarkerSize:   tokens.TooltipMarkerSize,
			TooltipTextSize:     active.Components.Tooltip.TextSize,
		},
		Colors: theme.LineChartColors{
			Axis:              semantic.Border,
			AxisLabel:         semantic.MutedForeground,
			Grid:              grid,
			Crosshair:         crosshair,
			Background:        semantic.Surface,
			LegendHover:       semantic.SurfaceHover,
			AnnotationArea:    withAlpha(semantic.Tone(theme.TonePrimary).Fill, 0x18/255.0),
			AnnotationLine:    withAlpha(semantic.Tone(theme.ToneDanger).Fill, 0xc0/255.0),
			AnnotationPoint:   semantic.Tone(theme.ToneWarning).Fill,
			OverlayForeground: semantic.OverlayForeground,
			Series:            tokens.SeriesColors,
		},
		DisabledOpacity:     active.DisabledOpacityValue(),
		LegendHiddenOpacity: 0.38,
		AreaOpacity:         0x30 / 255.0,
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
