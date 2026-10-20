package testtheme

import (
	"image/color"
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) RadioGroup(active *theme.Theme, request theme.RadioGroupRequest) theme.RadioGroupVisual {
	tokens := active.Components.RadioGroup
	indicator := style.Style{}.Width(tokens.Size).Height(tokens.Size).
		Background(style.Solid(background)).BorderWidth(tokens.BorderWidth).
		BorderColor(style.Solid(foreground)).Radius(tokens.Size/2).
		Outline(2, 0, style.Solid(focus)).TextColor(style.Solid(accentFg)).
		On(style.Checked, style.Style{}.Background(style.Solid(accent)).BorderColor(style.Solid(accent)))
	return theme.RadioGroupVisual{
		Defaults: style.Style{}.
			Part(style.PartIndicator, indicator).
			Part(style.PartLabel, style.Style{}.TextColor(style.Solid(foreground)).FontSize(tokens.TextSize)).
			Part(style.PartDescription, style.Style{}.TextColor(style.Solid(muted)).FontSize(tokens.DescriptionSize)),
		Override: theme.RadioGroupOverrideStyle(request),
		Metrics: theme.RadioGroupMetrics{FocusSpace: tokens.FocusSpace, ContentGap: tokens.ContentGap,
			VerticalGap: tokens.VerticalGap, HorizontalGap: tokens.HorizontalGap, DescriptionGap: tokens.DescriptionGap,
			DotScale: tokens.DotScale, DotPressedScale: tokens.DotPressedScale, PressedScale: tokens.PressedScale},
		IndicatorColor: neutralRadioIndicatorColor, IndicatorRenderer: neutralRadioIndicatorRenderer, Motion: active.Motion,
		SelectionDuration: 200 * time.Millisecond, FocusDuration: 100 * time.Millisecond,
		PressInDuration: 80 * time.Millisecond, PressOutDuration: 140 * time.Millisecond,
	}
}

// neutralRadioIndicatorRenderer keeps behavior fixtures independent of the
// shipped radio recipe while satisfying the renderer contract.
func neutralRadioIndicatorRenderer(_ layout.Context, request theme.RadioIndicatorRenderRequest) layout.Dimensions {
	return layout.Dimensions{Size: request.Size}
}

func neutralRadioIndicatorColor(resolved style.ResolvedStyle) (color.NRGBA, bool) {
	if resolved.Text == nil {
		return color.NRGBA{}, false
	}
	switch value := resolved.Text.Color.(type) {
	case style.SolidColor:
		return value.Color, true
	case *style.SolidColor:
		if value != nil {
			return value.Color, true
		}
	}
	return color.NRGBA{}, false
}
