package menubar

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/theme"
)

type menubarTriggerStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	focus      color.NRGBA
	opacity    float32
}

func menubarTriggerStyleFor(activeTheme *theme.Theme, open, hovered, pressed, disabled bool) menubarTriggerStyle {
	background := color.NRGBA{}
	if open || hovered {
		background = activeTheme.Palette.DefaultColor()
	}
	if pressed {
		background = activeTheme.Palette.DefaultHoverColor()
	}
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return menubarTriggerStyle{
		background: background,
		foreground: activeTheme.Palette.Foreground,
		focus:      activeTheme.Palette.Focus,
		opacity:    opacity,
	}
}
