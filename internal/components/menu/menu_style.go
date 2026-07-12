package menu

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

type panelStyle struct {
	background    color.NRGBA
	foreground    color.NRGBA
	border        color.NRGBA
	shadow        color.NRGBA
	shadowOpacity float32
}

type itemStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	description color.NRGBA
	shortcut    color.NRGBA
	indicator   color.NRGBA
	focusColor  color.NRGBA
	focus       float32
	opacity     float32
}

func menuPanelStyle(activeTheme *theme.Theme) panelStyle {
	return panelStyle{
		background:    activeTheme.Palette.OverlayColor(),
		foreground:    activeTheme.Palette.OverlayForegroundColor(),
		border:        activeTheme.Palette.Border,
		shadow:        activeTheme.Palette.OverlayShadowColor(),
		shadowOpacity: min(max(activeTheme.Components.Menu.ShadowOpacity, 0), 1),
	}
}

func menuItemStyle(activeTheme *theme.Theme, variant ItemVariant, hovered, disabled bool) itemStyle {
	foreground := activeTheme.Palette.OverlayForegroundColor()
	background := color.NRGBA{}
	if hovered {
		background = activeTheme.Palette.SurfaceTertiary
	}
	if variant == ItemDanger {
		foreground = activeTheme.Palette.Danger
	}
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	return itemStyle{
		background:  background,
		foreground:  foreground,
		description: activeTheme.Palette.MutedForeground,
		shortcut:    activeTheme.Palette.MutedForeground,
		indicator:   foreground,
		focusColor:  activeTheme.Palette.Focus,
		opacity:     opacity,
	}
}
