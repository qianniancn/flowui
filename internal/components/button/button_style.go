package button

import (
	"image/color"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type buttonStyle struct {
	height     unit.Dp
	inset      layout.Inset
	textSize   float32
	bg         color.NRGBA
	fg         color.NRGBA
	border     color.NRGBA
	hasBorder  bool
	focus      float32
	focusColor color.NRGBA
	focusWidth unit.Dp
}

type buttonPalette struct {
	bg        color.NRGBA
	hover     color.NRGBA
	pressed   color.NRGBA
	fg        color.NRGBA
	border    color.NRGBA
	hasBorder bool
}

func (b ButtonWidget) style(theme *theme.Theme, clickable *widget.Clickable) buttonStyle {
	style := buttonSizeStyle(theme, b.size, b.iconOnly)
	colors := buttonColors(theme, b.variant)

	bg := colors.bg
	switch {
	case b.disabled:
		bg = theme.DisabledColor(bg)
	case clickable.Pressed():
		bg = colors.pressed
	case clickable.Hovered():
		bg = colors.hover
	}
	style.bg = bg
	style.fg = colors.fg
	style.border = colors.border
	style.hasBorder = colors.hasBorder
	style.focusColor = theme.Palette.Focus
	style.focusWidth = theme.Components.Button.FocusRingWidth
	if b.disabled {
		style.fg = theme.DisabledColor(style.fg)
		style.border = theme.DisabledColor(style.border)
	}
	return style
}

func buttonSizeStyle(theme *theme.Theme, size ButtonSize, iconOnly bool) buttonStyle {
	style := buttonStyle{
		height:   theme.Spacing.ControlHeight,
		textSize: float32(theme.Typography.ControlSize),
		inset: layout.Inset{
			Left:  theme.Spacing.ControlPaddingX,
			Right: theme.Spacing.ControlPaddingX,
		},
	}
	switch size {
	case ButtonSmall:
		style.height = theme.Spacing.SmallControlHeight
		style.inset.Left = theme.Spacing.SmallControlPaddingX
		style.inset.Right = theme.Spacing.SmallControlPaddingX
	case ButtonLarge:
		style.height = theme.Spacing.LargeControlHeight
		style.textSize = float32(theme.Typography.ControlSize) + 2
		style.inset.Left = theme.Spacing.LargeControlPaddingX
		style.inset.Right = theme.Spacing.LargeControlPaddingX
	}
	if iconOnly {
		style.inset.Left = 0
		style.inset.Right = 0
	}
	return style
}

func buttonSpinnerSize(theme *theme.Theme, size ButtonSize) unit.Dp {
	switch size {
	case ButtonSmall:
		return theme.Components.Button.SpinnerSmall
	case ButtonLarge:
		return theme.Components.Button.SpinnerLarge
	default:
		return theme.Components.Button.SpinnerMedium
	}
}

func buttonColors(theme *theme.Theme, variant ButtonVariant) buttonPalette {
	transparent := color.NRGBA{}
	foreground := theme.Palette.Foreground
	defaultBg := theme.Palette.SurfaceRaised
	defaultHover := theme.Palette.SurfacePressed
	accent := theme.Palette.Accent
	accentHover := theme.Palette.AccentHover
	accentFg := theme.Palette.AccentForeground
	accentSoftFg := theme.Palette.AccentSoftForeground
	danger := theme.Palette.Danger
	dangerHover := theme.Palette.DangerHover
	dangerSoft := theme.Palette.DangerSoft
	dangerSoftHover := theme.Palette.DangerSoftHover
	dangerSoftFg := theme.Palette.DangerSoftForeground
	border := theme.Palette.Border
	outlineHover := theme.Palette.SurfaceRaised
	outlineHover.A = byte(uint16(outlineHover.A) * 0x99 / 0xff)

	switch variant {
	case ButtonSecondary:
		return buttonPalette{bg: defaultBg, hover: defaultHover, pressed: defaultHover, fg: accentSoftFg}
	case ButtonTertiary:
		return buttonPalette{bg: defaultBg, hover: defaultHover, pressed: defaultHover, fg: foreground}
	case ButtonGhost:
		return buttonPalette{bg: transparent, hover: defaultBg, pressed: defaultBg, fg: foreground}
	case ButtonOutline:
		return buttonPalette{
			bg:        transparent,
			hover:     outlineHover,
			pressed:   defaultBg,
			fg:        foreground,
			border:    border,
			hasBorder: true,
		}
	case ButtonDanger:
		return buttonPalette{bg: danger, hover: dangerHover, pressed: dangerHover, fg: accentFg}
	case ButtonDangerSoft:
		return buttonPalette{bg: dangerSoft, hover: dangerSoftHover, pressed: dangerSoftHover, fg: dangerSoftFg}
	default:
		return buttonPalette{bg: accent, hover: accentHover, pressed: accentHover, fg: accentFg}
	}
}
