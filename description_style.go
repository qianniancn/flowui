package flowui

import "image/color"

type descriptionStyle struct {
	text color.NRGBA
}

func descriptionStyleFor(theme *Theme, disabled bool) descriptionStyle {
	style := descriptionStyle{text: theme.Palette.MutedForeground}
	if disabled {
		style.text = theme.DisabledColor(style.text)
	}
	return style
}
