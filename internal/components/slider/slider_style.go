package slider

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type sliderStyle struct {
	track      color.NRGBA
	fill       color.NRGBA
	thumb      color.NRGBA
	thumbInner color.NRGBA
	focus      color.NRGBA
	label      color.NRGBA
	output     color.NRGBA
}

func sliderStyleFor(activeTheme *theme.Theme, disabled bool) sliderStyle {
	style := sliderStyle{
		track:      activeTheme.Palette.SurfaceRaised,
		fill:       activeTheme.Palette.Accent,
		thumb:      activeTheme.Palette.Accent,
		thumbInner: activeTheme.Palette.AccentForeground,
		focus:      activeTheme.Palette.Focus,
		label:      activeTheme.Palette.Foreground,
		output:     activeTheme.Palette.Foreground,
	}
	if disabled {
		style.track = activeTheme.DisabledColor(style.track)
		style.fill = activeTheme.DisabledColor(style.fill)
		style.thumb = activeTheme.DisabledColor(style.thumb)
		style.thumbInner = activeTheme.DisabledColor(style.thumbInner)
		style.focus = color.NRGBA{}
		style.label = activeTheme.DisabledColor(style.label)
		style.output = activeTheme.DisabledColor(style.output)
	}
	return style
}
