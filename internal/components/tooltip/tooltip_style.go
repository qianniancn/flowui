package tooltip

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/theme"
)

type tooltipStyle struct {
	surface color.NRGBA
	text    color.NRGBA
	border  color.NRGBA
}

func tooltipStyleFor(theme *theme.Theme) tooltipStyle {
	return tooltipStyle{
		surface: theme.Palette.OverlayColor(),
		text:    theme.Palette.OverlayForegroundColor(),
		border:  theme.Palette.Border,
	}
}
