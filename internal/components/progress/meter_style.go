package progress

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type meterStyle struct {
	track  color.NRGBA
	fill   color.NRGBA
	label  color.NRGBA
	output color.NRGBA
}

type meterSizeStyle struct {
	height unit.Dp
	radius unit.Dp
}

func meterStyleFor(activeTheme *theme.Theme, meterColor MeterColor, disabled bool) meterStyle {
	style := meterStyle{
		track:  activeTheme.Palette.DefaultColor(),
		fill:   activeTheme.Palette.Accent,
		label:  activeTheme.Palette.Foreground,
		output: activeTheme.Palette.Foreground,
	}
	switch meterColor {
	case MeterDefault:
		style.fill = activeTheme.Palette.DefaultForegroundColor()
	case MeterSuccess:
		style.fill = activeTheme.Palette.Success
	case MeterWarning:
		style.fill = activeTheme.Palette.Warning
	case MeterDanger:
		style.fill = activeTheme.Palette.Danger
	}
	if disabled {
		style.track = activeTheme.DisabledColor(style.track)
		style.fill = activeTheme.DisabledColor(style.fill)
		style.label = activeTheme.DisabledColor(style.label)
		style.output = activeTheme.DisabledColor(style.output)
	}
	return style
}

func meterSizeStyleFor(activeTheme *theme.Theme, size MeterSize) meterSizeStyle {
	tokens := activeTheme.Components.Meter
	switch size {
	case MeterSmall:
		return meterSizeStyle{height: tokens.SmallHeight, radius: tokens.SmallRadius}
	case MeterLarge:
		return meterSizeStyle{height: tokens.LargeHeight, radius: tokens.LargeRadius}
	default:
		return meterSizeStyle{height: tokens.MediumHeight, radius: tokens.MediumRadius}
	}
}
