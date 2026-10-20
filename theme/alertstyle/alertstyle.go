// Package alertstyle contains the theme-owned visual recipe for alert
// behavior. It maps semantic status to declarations and the indicator glyph;
// widget/alert only lays out the returned plan.
package alertstyle

import (
	"image/color"

	"gioui.org/font"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildAlert builds the provider-owned alert visual plan.
func BuildAlert(active *theme.Theme, request theme.AlertRequest) theme.AlertVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Alert
	sem := active.Semantic()
	statusColor := alertStatusColor(active, request.Status)
	return theme.AlertVisual{
		Defaults: style.Style{}.
			FillWidth().
			Background(style.Solid(sem.Surface)).
			TextColor(style.Solid(sem.SurfaceForeground)).
			PaddingX(tokens.PaddingX).
			PaddingY(tokens.PaddingY).
			Radius(tokens.Radius).
			Overflow(style.OverflowHidden).
			Shadow(style.ShadowSurface).
			Opacity(1),
		Title: style.Style{}.
			TextColor(style.Solid(statusColor)).
			FontSize(tokens.TitleSize).
			FontWeight(int(font.Medium)).
			LineHeight(tokens.TitleLineHeight),
		Description: style.Style{}.
			TextColor(style.Solid(sem.MutedForeground)).
			FontSize(tokens.DescriptionSize).
			LineHeight(tokens.DescriptionLineHeight),
		Content: style.Style{}.
			TextColor(style.Solid(sem.MutedForeground)).
			FontSize(tokens.DescriptionSize).
			LineHeight(tokens.DescriptionLineHeight),
		Indicator: style.Style{}.
			TextColor(style.Solid(statusColor)),
		Override: theme.AlertOverrideStyle(request),
		Metrics: theme.AlertMetrics{
			Gap:              tokens.Gap,
			IndicatorPadding: tokens.IndicatorPadding,
			IconSize:         tokens.IconSize,
		},
		Icon: statusIcon(request.Status),
	}
}

func alertStatusColor(active *theme.Theme, status theme.AlertStatus) color.NRGBA {
	value := active.Semantic().SurfaceForeground
	switch status {
	case theme.AlertAccent:
		value = appearance.Tone(active, theme.TonePrimary).OnSoft
	case theme.AlertSuccess:
		value = appearance.Tone(active, theme.ToneSuccess).OnSoft
	case theme.AlertWarning:
		value = appearance.Tone(active, theme.ToneWarning).OnSoft
	case theme.AlertDanger:
		value = appearance.Tone(active, theme.ToneDanger).OnSoft
	}
	return value
}

func statusIcon(status theme.AlertStatus) []byte {
	switch status {
	case theme.AlertSuccess:
		return lucide.CircleCheck
	case theme.AlertWarning:
		return lucide.TriangleAlert
	case theme.AlertDanger:
		return lucide.CircleAlert
	default:
		return lucide.Info
	}
}
