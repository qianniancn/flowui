package flowui

import "image/color"

type popoverStyle struct {
	surface color.NRGBA
	text    color.NRGBA
	muted   color.NRGBA
}

func popoverStyleFor(theme *Theme) popoverStyle {
	return popoverStyle{
		surface: theme.Palette.Surface,
		text:    theme.Palette.Foreground,
		muted:   theme.Palette.MutedForeground,
	}
}
