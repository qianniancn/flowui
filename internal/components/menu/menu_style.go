package menu

import (
	"image/color"

	flowstyle "github.com/qianniancn/FlowUI/internal/style"
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
	foreground  color.NRGBA
	description color.NRGBA
	shortcut    color.NRGBA
	indicator   color.NRGBA
}

func menuRootDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	panel := menuPanelStyle(activeTheme)
	tokens := activeTheme.Components.Menu
	return flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: panel.background}).
		TextColor(flowstyle.SolidColor{Color: panel.foreground}).
		BorderColor(flowstyle.SolidColor{Color: panel.border}).
		BorderWidth(tokens.BorderWidth).
		Radius(tokens.Radius).
		Shadow(flowstyle.ShadowMenu).
		Overflow(flowstyle.OverflowHidden)

}

func menuItemDefaultDeclaration(activeTheme *theme.Theme, tokens theme.MenuTheme) flowstyle.Style {
	pressedScale := tokens.PressedScale
	if pressedScale <= 0 || pressedScale > 1 {
		pressedScale = 0.98
	}
	item := flowstyle.Style{}.
		MinHeight(tokens.ItemMinHeight).
		PaddingX(tokens.ItemPaddingX).
		PaddingY(tokens.ItemPaddingY).
		Radius(tokens.ItemRadius).
		Background(flowstyle.RGBA(0x00000000)).
		TextColor(flowstyle.SolidColor{Color: menuForegroundColor(activeTheme)}).
		BorderWidth(tokens.FocusRingWidth).
		BorderColor(flowstyle.WithAlpha(flowstyle.TokenFocus, 0)).
		Opacity(1).
		Scale(1, 1).
		Transition(flowstyle.PropBackgroundColor, menuItemColorDuration).
		Transition(flowstyle.PropBorderColor, menuItemFocusDuration).
		Transition(flowstyle.PropTransform, menuItemPressDuration).
		When(flowstyle.Hovered, flowstyle.Style{}.Background(flowstyle.TokenDefault)).
		When(flowstyle.Pressed, flowstyle.Style{}.Scale(pressedScale, pressedScale)).
		When(flowstyle.FocusVisible, flowstyle.Style{}.BorderColor(flowstyle.SolidColor{Color: menuFocusColor(activeTheme)})).
		When(flowstyle.Disabled, flowstyle.Style{}.Opacity(activeTheme.DisabledOpacityValue()))
	return flowstyle.Style{}.Part(flowstyle.PartItem, item)
}

func menuItemVariantDeclaration(activeTheme *theme.Theme, variant ItemVariant) flowstyle.Style {
	if variant != ItemDanger {
		return flowstyle.Style{}
	}
	item := flowstyle.Style{}.
		TextColor(flowstyle.SolidColor{Color: menuDangerColor(activeTheme)}).
		When(flowstyle.FocusVisible, flowstyle.Style{}.BorderColor(
			flowstyle.WithAlpha(flowstyle.SolidColor{Color: menuDangerColor(activeTheme)}, float32(menuFocusColor(activeTheme).A)/255),
		))
	return flowstyle.Style{}.Part(flowstyle.PartItem, item)
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

func menuItemStyle(activeTheme *theme.Theme, variant ItemVariant) itemStyle {
	foreground := menuForegroundColor(activeTheme)
	if variant == ItemDanger {
		foreground = menuDangerColor(activeTheme)
	}
	style := itemStyle{
		foreground:  foreground,
		description: menuMutedColor(activeTheme),
		shortcut:    menuMutedColor(activeTheme),
		indicator:   menuMutedColor(activeTheme),
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
