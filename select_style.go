package flowui

import "image/color"

type selectStyle struct {
	field       inputStyle
	description color.NRGBA
	error       color.NRGBA
}

func selectStyleFor(theme *Theme, variant SelectVariant, hovered, focusVisible, disabled, invalid bool) selectStyle {
	style := selectStyle{
		field:       inputStyleFor(theme, variant, hovered, focusVisible, disabled, invalid),
		description: theme.Palette.MutedForeground,
		error:       theme.Palette.Danger,
	}
	if disabled {
		style.description = theme.DisabledColor(style.description)
		style.error = theme.DisabledColor(style.error)
	}
	return style
}
