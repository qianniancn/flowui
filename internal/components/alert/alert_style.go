package alert

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type alertStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	indicator   color.NRGBA
	title       color.NRGBA
	description color.NRGBA
}

func alertStyleFor(activeTheme *theme.Theme, status Status) alertStyle {
	foreground := theme.ColorOr(activeTheme.Palette.SurfaceForeground, activeTheme.Palette.Foreground)
	statusColor := foreground
	switch status {
	case StatusAccent:
		statusColor = theme.ColorOr(activeTheme.Palette.AccentSoftForeground, activeTheme.Palette.Accent)
	case StatusSuccess:
		statusColor = activeTheme.Palette.SuccessSoftForegroundColor()
	case StatusWarning:
		statusColor = activeTheme.Palette.WarningSoftForegroundColor()
	case StatusDanger:
		statusColor = theme.ColorOr(activeTheme.Palette.DangerSoftForeground, activeTheme.Palette.Danger)
	}
	return alertStyle{
		background:  activeTheme.Palette.Surface,
		foreground:  foreground,
		indicator:   statusColor,
		title:       statusColor,
		description: activeTheme.Palette.MutedForeground,
	}
}
