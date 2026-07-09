package flowui

import (
	"image/color"

	"gioui.org/unit"
)

type switchStyle struct {
	trackOff    color.NRGBA
	trackOn     color.NRGBA
	thumb       color.NRGBA
	thumbOn     color.NRGBA
	thumbFgOff  color.NRGBA
	thumbFg     color.NRGBA
	label       color.NRGBA
	description color.NRGBA
	focusColor  color.NRGBA
	selected    float32
	focus       float32
}

type switchSizeStyle struct {
	trackWidth  unit.Dp
	trackHeight unit.Dp
	thumbWidth  unit.Dp
	thumbHeight unit.Dp
}

func switchStyleFor(theme *Theme, hovered, pressed, disabled, invalid bool) switchStyle {
	style := switchStyle{
		trackOff:    theme.Palette.SurfaceRaised,
		trackOn:     theme.Palette.Accent,
		thumb:       theme.Palette.Surface,
		thumbOn:     theme.Palette.AccentForeground,
		thumbFgOff:  theme.Palette.MutedForeground,
		thumbFg:     theme.Palette.Accent,
		label:       theme.Palette.Foreground,
		description: theme.Palette.MutedForeground,
		focusColor:  theme.Palette.Focus,
	}
	if hovered {
		style.trackOff = theme.Palette.SurfacePressed
		style.trackOn = theme.Palette.AccentHover
	}
	if pressed {
		style.trackOff = theme.Palette.SurfacePressed
		style.trackOn = theme.Palette.AccentHover
	}
	if invalid {
		style.trackOn = theme.Palette.Danger
		style.thumbFg = theme.Palette.Danger
		style.focusColor = theme.Palette.Danger
		style.focusColor.A = theme.Palette.Focus.A
		if hovered || pressed {
			style.trackOn = theme.Palette.DangerHover
		}
	}
	if disabled {
		style.trackOff = theme.DisabledColor(style.trackOff)
		style.trackOn = theme.DisabledColor(style.trackOn)
		style.thumb = theme.DisabledColor(style.thumb)
		style.thumbOn = theme.DisabledColor(style.thumbOn)
		style.thumbFgOff = theme.DisabledColor(style.thumbFgOff)
		style.thumbFg = theme.DisabledColor(style.thumbFg)
		style.label = theme.DisabledColor(style.label)
		style.description = theme.DisabledColor(style.description)
		style.focusColor = color.NRGBA{}
	}
	return style
}

func switchSizeStyleFor(theme *Theme, size SwitchSize) switchSizeStyle {
	switch size {
	case SwitchSmall:
		return switchSizeStyle{
			trackWidth:  theme.Components.Switch.SmallTrackWidth,
			trackHeight: theme.Components.Switch.SmallTrackHeight,
			thumbWidth:  theme.Components.Switch.SmallThumbWidth,
			thumbHeight: theme.Components.Switch.SmallThumbHeight,
		}
	case SwitchLarge:
		return switchSizeStyle{
			trackWidth:  theme.Components.Switch.LargeTrackWidth,
			trackHeight: theme.Components.Switch.LargeTrackHeight,
			thumbWidth:  theme.Components.Switch.LargeThumbWidth,
			thumbHeight: theme.Components.Switch.LargeThumbHeight,
		}
	default:
		return switchSizeStyle{
			trackWidth:  theme.Components.Switch.MediumTrackWidth,
			trackHeight: theme.Components.Switch.MediumTrackHeight,
			thumbWidth:  theme.Components.Switch.MediumThumbWidth,
			thumbHeight: theme.Components.Switch.MediumThumbHeight,
		}
	}
}
