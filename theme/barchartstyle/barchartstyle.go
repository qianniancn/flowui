// Package barchartstyle contains the visual recipe for bar-chart providers.
// Bar-chart behavior stays in widget/barchart; this package owns geometry,
// typography, palette, and the instance declaration boundary.
package barchartstyle

import (
	"image/color"
	"time"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildBarChart returns the complete visual plan for the active theme.
func BuildBarChart(active *theme.Theme, request theme.BarChartRequest) theme.BarChartVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.BarChart
	semantic := active.Semantic()
	return theme.BarChartVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.BarChartOverrideStyle(request),
		Metrics: theme.BarChartMetrics{
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
			LegendTextSize:      tokens.LegendTextSize,
			TextWeight:          tokens.TextWeight,
			EmphasisWeight:      tokens.EmphasisWeight,
			LegendMarkerSize:    tokens.LegendMarkerSize,
			LegendMarkerGap:     tokens.LegendMarkerGap,
			LegendMarkerRadius:  tokens.LegendMarkerRadius,
			LegendItemGap:       tokens.LegendItemGap,
			LegendLineGap:       tokens.LegendLineGap,
			LegendGap:           tokens.LegendGap,
			GridWidth:           tokens.GridWidth,
			AxisWidth:           tokens.AxisWidth,
			AnnotationLineWidth: tokens.AnnotationLineWidth,
			AnnotationPointSize: tokens.AnnotationPointSize,
			BarRadius:           tokens.BarRadius,
			MinBarHeight:        tokens.MinBarHeight,
			BackgroundRadius:    tokens.BackgroundRadius,
			TooltipGap:          tokens.TooltipGap,
			TooltipRowGap:       tokens.TooltipRowGap,
			TooltipMarkerSize:   tokens.TooltipMarkerSize,
			TooltipTextSize:     active.Components.Tooltip.TextSize,
		},
		Colors: theme.BarChartColors{
			Axis:              semantic.Border,
			AxisLabel:         semantic.MutedForeground,
			Grid:              withAlpha(semantic.Border, 0.8),
			CategoryHover:     withAlpha(semantic.Tone(theme.TonePrimary).Fill, 18.0/255.0),
			BarBackground:     withAlpha(semantic.MutedForeground, 28.0/255.0),
			LegendHover:       semantic.SurfaceHover,
			AnnotationArea:    withAlpha(semantic.Tone(theme.TonePrimary).Fill, 0x18/255.0),
			AnnotationLine:    withAlpha(semantic.Tone(theme.ToneDanger).Fill, 0xc0/255.0),
			AnnotationPoint:   semantic.Tone(theme.ToneWarning).Fill,
			OverlayForeground: semantic.OverlayForeground,
			Series:            tokens.SeriesColors,
		},
		BarLabelForeground:  barLabelForeground,
		DisabledOpacity:     active.DisabledOpacityValue(),
		LegendHiddenOpacity: 0.38,
		Motion:              active.Motion,
	}
}

func barLabelForeground(background color.NRGBA) color.NRGBA {
	luminance := (299*int(background.R) + 587*int(background.G) + 114*int(background.B)) / 1000
	if luminance > 150 {
		return color.NRGBA{R: 0x18, G: 0x18, B: 0x1b, A: 0xff}
	}
	return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
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
