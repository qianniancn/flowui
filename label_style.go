package flowui

import "image/color"

type labelStyle struct {
	text     color.NRGBA
	required color.NRGBA
}

func labelStyleFor(theme *Theme, foreground color.NRGBA, disabled, invalid bool) labelStyle {
	style := labelStyle{
		text:     foreground,
		required: theme.Palette.Danger,
	}
	if invalid {
		style.text = theme.Palette.Danger
	}
	if disabled {
		style.text = theme.DisabledColor(style.text)
		style.required = theme.DisabledColor(style.required)
	}
	return style
}
