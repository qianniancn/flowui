package toast

import (
	"image/color"

	"github.com/qianniancn/flowui/internal/theme"
)

type toastStyle struct {
	surface     color.NRGBA
	foreground  color.NRGBA
	title       color.NRGBA
	description color.NRGBA
	indicator   color.NRGBA
	border      color.NRGBA
	focus       color.NRGBA
}

func toastStyleFor(activeTheme *theme.Theme, variant ToastVariant) toastStyle {
	style := toastStyle{
		surface:     activeTheme.Palette.Surface,
		foreground:  activeTheme.Palette.OverlayForeground,
		title:       activeTheme.Palette.OverlayForeground,
		description: activeTheme.Palette.MutedForeground,
		indicator:   activeTheme.Palette.OverlayForeground,
		border:      activeTheme.Palette.Border,
		focus:       activeTheme.Palette.Focus,
	}
	switch variant {
	case ToastAccent:
		style.title = activeTheme.Palette.AccentSoftForeground
	case ToastSuccess:
		style.title = activeTheme.Palette.SuccessSoftForegroundColor()
		style.indicator = style.title
	case ToastWarning:
		style.title = activeTheme.Palette.WarningSoftForegroundColor()
		style.indicator = style.title
	case ToastDanger:
		style.title = activeTheme.Palette.DangerSoftForeground
		style.indicator = style.title
	}
	return style
}
