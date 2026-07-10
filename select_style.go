package flowui

import "image/color"

type selectStyle struct {
	field       inputStyle
	label       color.NRGBA
	description color.NRGBA
	error       color.NRGBA
}

func selectStyleFor(theme *Theme, variant SelectVariant, hovered, focusVisible, disabled, invalid bool) selectStyle {
	style := selectStyle{
		field:       inputStyleFor(theme, variant, hovered, focusVisible, disabled, invalid),
		label:       theme.Palette.Foreground,
		description: theme.Palette.MutedForeground,
		error:       theme.Palette.Danger,
	}
	if disabled {
		style.label = theme.DisabledColor(style.label)
		style.description = theme.DisabledColor(style.description)
		style.error = theme.DisabledColor(style.error)
	}
	return style
}
