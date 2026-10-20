// Package candlestickstyle contains the visual recipe for candlestick
// providers. Candlestick behavior stays in widget/candlestick; this package
// owns geometry, typography, palette, and the instance declaration boundary.
package candlestickstyle

import (
	"image/color"
	"time"

	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildCandlestick returns the complete visual plan for the active theme.
func BuildCandlestick(active *theme.Theme, request theme.CandlestickRequest) theme.CandlestickVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.CandlestickChart
	semantic := active.Semantic()
	background := tokens.CrosshairLabelBackground
	if background.A == 0 {
		background = semantic.MutedForeground
	}
	foreground := tokens.CrosshairLabelForeground
	if foreground.A == 0 {
		foreground = semantic.Background
	}
	return theme.CandlestickVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(semantic.Surface)).
			Opacity(1).
			Rules(style.Rule{When: style.Disabled, Set: style.Style{}.
				Opacity(active.DisabledOpacityValue())}),
		Override: theme.CandlestickOverrideStyle(request),
		Metrics: theme.CandlestickMetrics{
			AnimationDuration:     300 * time.Millisecond,
			UpdateDuration:        300 * time.Millisecond,
			Height:                tokens.Height,
			PlotPaddingTop:        tokens.PlotPaddingTop,
			PlotPaddingRight:      tokens.PlotPaddingRight,
			PlotPaddingBottom:     tokens.PlotPaddingBottom,
			PlotPaddingLeft:       tokens.PlotPaddingLeft,
			AxisNameGap:           tokens.AxisNameGap,
			TickLabelGap:          tokens.TickLabelGap,
			AxisTextSize:          tokens.AxisTextSize,
			GridWidth:             tokens.GridWidth,
			AxisWidth:             tokens.AxisWidth,
			AnnotationLineWidth:   tokens.AnnotationLineWidth,
			AnnotationPointSize:   tokens.AnnotationPointSize,
			CrosshairWidth:        tokens.CrosshairWidth,
			CrosshairLabelPadding: tokens.CrosshairLabelPadding,
			CrosshairLabelRadius:  active.Components.Tooltip.Radius,
			WickWidth:             tokens.WickWidth,
			BorderWidth:           tokens.BorderWidth,
			EmphasisBorderWidth:   tokens.EmphasisBorderWidth,
			TooltipGap:            tokens.TooltipGap,
			TooltipRowGap:         tokens.TooltipRowGap,
			TooltipTextSize:       active.Components.Tooltip.TextSize,
			TextWeight:            font.Normal,
			EmphasisWeight:        font.Medium,
		},
		Colors: theme.CandlestickColors{
			Axis:                     semantic.Border,
			AxisLabel:                semantic.MutedForeground,
			Grid:                     withAlpha(semantic.Border, .8),
			Crosshair:                withAlpha(semantic.MutedForeground, .75),
			CrosshairLabelBackground: background,
			CrosshairLabelForeground: foreground,
			OverlayForeground:        semantic.OverlayForeground,
			AnnotationArea:           withAlpha(semantic.Tone(theme.TonePrimary).Fill, 0x18/255.0),
			AnnotationLine:           withAlpha(semantic.Tone(theme.ToneDanger).Fill, 0xc0/255.0),
			AnnotationPoint:          semantic.Tone(theme.ToneWarning).Fill,
			Up:                       tokens.UpColor,
			Down:                     tokens.DownColor,
			Doji:                     tokens.DojiColor,
		},
		DisabledOpacity: active.DisabledOpacityValue(),
		Motion:          active.Motion,
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
