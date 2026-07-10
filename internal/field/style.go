package field

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type Variant uint8

const (
	Primary Variant = iota
	Secondary
)

func ResolveStyle(theme *theme.Theme, variant Variant, hovered, focused, disabled, invalid bool) Style {
	transparent := color.NRGBA{}
	foreground := theme.Palette.Foreground
	placeholder := theme.Palette.MutedForeground
	fieldBg := theme.Palette.Surface
	fieldHover := theme.Palette.SurfaceHover
	fieldFocus := theme.Palette.Surface
	defaultBg := theme.Palette.SurfaceRaised
	defaultHover := theme.Palette.SurfacePressed
	accent := theme.Palette.Accent
	danger := theme.Palette.Danger

	style := Style{
		Background:    fieldBg,
		Border:        transparent,
		Foreground:    foreground,
		Placeholder:   placeholder,
		Selection:     theme.Palette.Selection,
		ShadowOpacity: theme.Components.Input.ShadowOpacity,
	}
	if variant == Secondary {
		style.Background = defaultBg
		fieldHover = defaultHover
		fieldFocus = defaultBg
		style.ShadowOpacity = 0
	}
	if hovered && !focused {
		style.Background = fieldHover
	}
	if focused {
		style.Background = fieldFocus
		style.Border = accent
		style.BorderWidth = unit.Dp(2)
	}
	if invalid {
		style.Background = fieldFocus
		style.Border = danger
		if focused {
			style.BorderWidth = unit.Dp(2)
		} else {
			style.BorderWidth = unit.Dp(1)
		}
	}
	if disabled {
		style.Background = theme.DisabledColor(style.Background)
		style.Border = theme.DisabledColor(style.Border)
		style.Foreground = theme.DisabledColor(style.Foreground)
		style.Placeholder = theme.DisabledColor(style.Placeholder)
		style.Selection = theme.DisabledColor(style.Selection)
		style.ShadowOpacity *= theme.DisabledOpacityValue()
	}
	return style
}

type Style struct {
	Background    color.NRGBA
	Border        color.NRGBA
	Foreground    color.NRGBA
	Placeholder   color.NRGBA
	Selection     color.NRGBA
	BorderWidth   unit.Dp
	ShadowOpacity float32
}
