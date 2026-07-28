package popover

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/theme"
)

type popoverStyle struct {
	surface color.NRGBA
	text    color.NRGBA
	muted   color.NRGBA
}

func popoverStyleFor(theme *theme.Theme) popoverStyle {
	return popoverStyle{
		surface: theme.Palette.OverlayColor(),
		text:    theme.Palette.OverlayForegroundColor(),
		muted:   theme.Palette.MutedForeground,
	}
}
