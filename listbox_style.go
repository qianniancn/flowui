package flowui

import "image/color"

type listBoxItemStyle struct {
	bg          color.NRGBA
	fg          color.NRGBA
	description color.NRGBA
	indicator   color.NRGBA
	focusColor  color.NRGBA
	selected    float32
	focus       float32
}

func listBoxItemStyleFor(theme *Theme, variant ListBoxItemVariant, hovered, pressed, disabled bool) listBoxItemStyle {
	style := listBoxItemStyle{
		fg:          theme.Palette.Foreground,
		description: theme.Palette.MutedForeground,
		indicator:   theme.Palette.Foreground,
		focusColor:  theme.Palette.Focus,
	}
	if hovered {
		style.bg = theme.Palette.SurfaceRaised
	}
	if pressed {
		style.bg = theme.Palette.SurfacePressed
	}
	if variant == ListBoxItemDanger {
		style.fg = theme.Palette.Danger
		style.indicator = theme.Palette.Danger
		style.focusColor = theme.Palette.Danger
		style.focusColor.A = theme.Palette.Focus.A
	}
	if disabled {
		style.bg = theme.DisabledColor(style.bg)
		style.fg = theme.DisabledColor(style.fg)
		style.description = theme.DisabledColor(style.description)
		style.indicator = theme.DisabledColor(style.indicator)
		style.focusColor = color.NRGBA{}
	}
	return style
}
