package testtheme

import (
	"image/color"

	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Alert(active *theme.Theme, request theme.AlertRequest) theme.AlertVisual {
	tokens := active.Components.Alert
	statusColor := foreground
	switch request.Status {
	case theme.AlertAccent:
		statusColor = focus
	case theme.AlertSuccess:
		statusColor = color.NRGBA{R: 0x20, G: 0x90, B: 0x50, A: 0xff}
	case theme.AlertWarning:
		statusColor = color.NRGBA{R: 0xa0, G: 0x70, B: 0x10, A: 0xff}
	case theme.AlertDanger:
		statusColor = color.NRGBA{R: 0xb0, G: 0x30, B: 0x30, A: 0xff}
	}
	defaults := style.Style{}.
		FillWidth().
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		PaddingX(tokens.PaddingX).
		PaddingY(tokens.PaddingY).
		Radius(tokens.Radius).
		Overflow(style.OverflowHidden)
	return theme.AlertVisual{
		Defaults: defaults,
		Title: style.Style{}.
			TextColor(style.Solid(statusColor)).
			FontSize(tokens.TitleSize).
			FontWeight(500).
			LineHeight(tokens.TitleLineHeight),
		Description: style.Style{}.
			TextColor(style.Solid(muted)).
			FontSize(tokens.DescriptionSize).
			LineHeight(tokens.DescriptionLineHeight),
		Content: style.Style{}.
			TextColor(style.Solid(muted)).
			FontSize(tokens.DescriptionSize).
			LineHeight(tokens.DescriptionLineHeight),
		Indicator: style.Style{}.TextColor(style.Solid(statusColor)),
		Override:  theme.AlertOverrideStyle(request),
		Metrics: theme.AlertMetrics{
			Gap:              tokens.Gap,
			IndicatorPadding: tokens.IndicatorPadding,
			IconSize:         tokens.IconSize,
		},
	}
}
