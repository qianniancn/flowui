package alertdialog

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type alertDialogStyle struct {
	iconBackground color.NRGBA
	iconForeground color.NRGBA
}

func alertDialogStyleFor(activeTheme *theme.Theme, status Status) alertDialogStyle {
	palette := activeTheme.Palette
	style := alertDialogStyle{
		iconBackground: theme.ColorOr(palette.SurfaceSecondary, palette.Surface),
		iconForeground: theme.ColorOr(palette.OverlayForeground, palette.Foreground),
	}
	switch status {
	case StatusAccent:
		style.iconBackground = softColor(palette.AccentSoft, palette.Accent)
		style.iconForeground = theme.ColorOr(palette.AccentSoftForeground, palette.Accent)
	case StatusSuccess:
		style.iconBackground = softColor(palette.SuccessSoft, palette.Success)
		style.iconForeground = palette.SuccessSoftForegroundColor()
	case StatusWarning:
		style.iconBackground = softColor(palette.WarningSoft, palette.Warning)
		style.iconForeground = palette.WarningSoftForegroundColor()
	case StatusDanger:
		style.iconBackground = softColor(palette.DangerSoft, palette.Danger)
		style.iconForeground = theme.ColorOr(palette.DangerSoftForeground, palette.Danger)
	}
	return style
}

func softColor(value, fallback color.NRGBA) color.NRGBA {
	if value.A != 0 {
		return value
	}
	if fallback.A == 0 {
		return value
	}
	fallback.A = 0x26
	return fallback
}
