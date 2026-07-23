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
		iconBackground: palette.SurfaceSecondary,
		iconForeground: palette.OverlayForeground,
	}
	switch status {
	case StatusAccent:
		style.iconBackground = palette.AccentSoft
		style.iconForeground = palette.AccentSoftForeground
	case StatusSuccess:
		style.iconBackground = palette.SuccessSoft
		style.iconForeground = palette.SuccessSoftForegroundColor()
	case StatusWarning:
		style.iconBackground = palette.WarningSoft
		style.iconForeground = palette.WarningSoftForegroundColor()
	case StatusDanger:
		style.iconBackground = palette.DangerSoft
		style.iconForeground = palette.DangerSoftForeground
	}
	return style
}
