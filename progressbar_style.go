package flowui

import (
	"image/color"

	"gioui.org/unit"
)

type progressBarStyle struct {
	track  color.NRGBA
	fill   color.NRGBA
	label  color.NRGBA
	output color.NRGBA
}

type progressBarSizeStyle struct {
	height unit.Dp
	radius unit.Dp
}

func progressBarStyleFor(theme *Theme, barColor ProgressBarColor, disabled bool) progressBarStyle {
	style := progressBarStyle{
		track:  theme.Palette.SurfaceRaised,
		fill:   theme.Palette.Accent,
		label:  theme.Palette.Foreground,
		output: theme.Palette.Foreground,
	}
	switch barColor {
	case ProgressBarDefault:
		style.fill = theme.Palette.Foreground
	case ProgressBarSuccess:
		style.fill = theme.Palette.Success
	case ProgressBarWarning:
		style.fill = theme.Palette.Warning
	case ProgressBarDanger:
		style.fill = theme.Palette.Danger
	default:
		style.fill = theme.Palette.Accent
	}
	if disabled {
		style.track = theme.DisabledColor(style.track)
		style.fill = theme.DisabledColor(style.fill)
		style.label = theme.DisabledColor(style.label)
		style.output = theme.DisabledColor(style.output)
	}
	return style
}

func progressBarSizeStyleFor(theme *Theme, size ProgressBarSize) progressBarSizeStyle {
	progressTheme := theme.Components.ProgressBar
	switch size {
	case ProgressBarSmall:
		return progressBarSizeStyle{
			height: progressTheme.SmallHeight,
			radius: progressTheme.SmallRadius,
		}
	case ProgressBarLarge:
		return progressBarSizeStyle{
			height: progressTheme.LargeHeight,
			radius: progressTheme.LargeRadius,
		}
	default:
		return progressBarSizeStyle{
			height: progressTheme.MediumHeight,
			radius: progressTheme.MediumRadius,
		}
	}
}
