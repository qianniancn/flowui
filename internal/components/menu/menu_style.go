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
	tokens := activeTheme.Components.Menu
	shadow := tokens.ShadowColor
	if shadow.A == 0 && tokens.ShadowOpacity > 0 {
		shadow = activeTheme.Palette.OverlayShadowColor()
	}
	border := tokens.BorderColor
	if border.A == 0 {
		border = activeTheme.Palette.Border
	}
	return panelStyle{
		background:    menuBackgroundColor(activeTheme),
		foreground:    menuForegroundColor(activeTheme),
		border:        border,
		shadow:        shadow,
		shadowOpacity: min(max(tokens.ShadowOpacity, 0), 1),
	}
}

func menuItemStyle(activeTheme *theme.Theme, variant ItemVariant, hovered, disabled bool) itemStyle {
	foreground := menuForegroundColor(activeTheme)
	background := color.NRGBA{}
	if hovered {
		background = activeTheme.Palette.DefaultColor()
	}
	if variant == ItemDanger {
		foreground = menuDangerColor(activeTheme)
	}
	opacity := float32(1)
	if disabled {
		opacity = activeTheme.DisabledOpacityValue()
	}
	style := itemStyle{
		background:  background,
		foreground:  foreground,
		description: menuMutedColor(activeTheme),
		shortcut:    menuMutedColor(activeTheme),
		indicator:   menuMutedColor(activeTheme),
		focusColor:  menuFocusColor(activeTheme),
		opacity:     opacity,
	}
	if variant == ItemDanger {
		style.indicator = menuDangerColor(activeTheme)
	}
	return style
}

func menuBackgroundColor(activeTheme *theme.Theme) color.NRGBA {
	return theme.ColorOr(activeTheme.Components.Menu.BackgroundColor, activeTheme.Palette.OverlayColor())
}

func menuForegroundColor(activeTheme *theme.Theme) color.NRGBA {
	return theme.ColorOr(activeTheme.Components.Menu.ForegroundColor, activeTheme.Palette.OverlayForegroundColor())
}

func menuMutedColor(activeTheme *theme.Theme) color.NRGBA {
	return theme.ColorOr(activeTheme.Components.Menu.MutedColor, activeTheme.Palette.MutedForeground)
}

func menuDangerColor(activeTheme *theme.Theme) color.NRGBA {
	return theme.ColorOr(activeTheme.Components.Menu.DangerColor, activeTheme.Palette.Danger)
}

func menuFocusColor(activeTheme *theme.Theme) color.NRGBA {
	return theme.ColorOr(activeTheme.Components.Menu.FocusColor, activeTheme.Palette.Focus)
}
