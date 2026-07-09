package flowui

import (
	"image/color"

	"gioui.org/unit"
)

func inputStyleFor(theme *Theme, variant InputVariant, hovered, focused, disabled, invalid bool) inputStyle {
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

	style := inputStyle{
		bg:            fieldBg,
		border:        transparent,
		fg:            foreground,
		placeholder:   placeholder,
		selection:     theme.Palette.Selection,
		shadowOpacity: theme.Components.Input.ShadowOpacity,
	}
	if variant == InputSecondary {
		style.bg = defaultBg
		fieldHover = defaultHover
		fieldFocus = defaultBg
		style.shadowOpacity = 0
	}
	if hovered && !focused {
		style.bg = fieldHover
	}
	if focused {
		style.bg = fieldFocus
		style.border = accent
		style.borderWidth = unit.Dp(2)
	}
	if invalid {
		style.bg = fieldFocus
		style.border = danger
		if focused {
			style.borderWidth = unit.Dp(2)
		} else {
			style.borderWidth = unit.Dp(1)
		}
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.border = theme.DisabledColor(style.border)
		style.fg = theme.DisabledColor(style.fg)
		style.placeholder = theme.DisabledColor(style.placeholder)
		style.selection = theme.DisabledColor(style.selection)
		style.shadowOpacity *= theme.disabledOpacity()
	}
	return style
}

type inputStyle struct {
	bg            color.NRGBA
	border        color.NRGBA
	fg            color.NRGBA
	placeholder   color.NRGBA
	selection     color.NRGBA
	borderWidth   unit.Dp
	shadowOpacity float32
}
