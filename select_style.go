package flowui

import "image/color"

type selectStyle struct {
	field inputStyle
	error color.NRGBA
}

func selectStyleFor(theme *Theme, variant SelectVariant, hovered, focusVisible, disabled, invalid bool) selectStyle {
	style := selectStyle{
		field: inputStyleFor(theme, variant, hovered, focusVisible, disabled, invalid),
		error: theme.Palette.Danger,
	}
	if disabled {
		style.error = theme.DisabledColor(style.error)
	}
	return style
}
