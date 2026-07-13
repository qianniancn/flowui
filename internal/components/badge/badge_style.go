package badge

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type badgeStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	border      color.NRGBA
	minSize     unit.Dp
	radius      unit.Dp
	textSize    unit.Sp
	lineHeight  unit.Sp
	paddingX    unit.Dp
	borderWidth unit.Dp
}

func badgeStyleFor(activeTheme *theme.Theme, badgeColor Color, variant Variant, size Size, border color.NRGBA) badgeStyle {
	tokens := activeTheme.Components.Badge
	primaryBackground, primaryForeground := badgePrimaryColors(activeTheme.Palette, badgeColor)
	secondaryForeground := badgeSecondaryForeground(activeTheme.Palette, badgeColor)
	style := badgeStyle{
		background:  primaryBackground,
		foreground:  primaryForeground,
		border:      border,
		minSize:     tokens.MediumMinSize,
		radius:      tokens.MediumRadius,
		textSize:    tokens.MediumTextSize,
		lineHeight:  tokens.MediumLineHeight,
		paddingX:    tokens.LabelPaddingX,
		borderWidth: tokens.BorderWidth,
	}
	switch variant {
	case VariantSecondary:
		style.background = activeTheme.Palette.DefaultColor()
		style.foreground = secondaryForeground
	case VariantSoft:
		style.background, style.foreground = badgeSoftColors(activeTheme.Palette, badgeColor)
	}
	switch size {
	case SizeSmall:
		style.minSize = tokens.SmallMinSize
		style.radius = tokens.SmallRadius
		style.textSize = tokens.SmallTextSize
		style.lineHeight = tokens.SmallLineHeight
	case SizeLarge:
		style.minSize = tokens.LargeMinSize
		style.radius = tokens.LargeRadius
		style.textSize = tokens.LargeTextSize
		style.lineHeight = tokens.LargeLineHeight
	}
	return style
}

func badgePrimaryColors(palette theme.Palette, badgeColor Color) (color.NRGBA, color.NRGBA) {
	switch badgeColor {
	case ColorAccent:
		return palette.Accent, theme.ColorOr(palette.AccentForeground, palette.Foreground)
	case ColorSuccess:
		return palette.Success, theme.ColorOr(palette.SuccessForeground, palette.Foreground)
	case ColorWarning:
		return palette.Warning, theme.ColorOr(palette.WarningForeground, palette.Foreground)
	case ColorDanger:
		return palette.Danger, theme.ColorOr(palette.DangerForeground, palette.Foreground)
	default:
		return palette.DefaultColor(), palette.DefaultForegroundColor()
	}
}

func badgeSecondaryForeground(palette theme.Palette, badgeColor Color) color.NRGBA {
	switch badgeColor {
	case ColorAccent:
		return theme.ColorOr(palette.AccentSoftForeground, palette.Accent)
	case ColorSuccess:
		return palette.SuccessSoftForegroundColor()
	case ColorWarning:
		return palette.WarningSoftForegroundColor()
	case ColorDanger:
		return theme.ColorOr(palette.DangerSoftForeground, palette.Danger)
	default:
		return palette.DefaultForegroundColor()
	}
}

func badgeSoftColors(palette theme.Palette, badgeColor Color) (color.NRGBA, color.NRGBA) {
	var background color.NRGBA
	var fallback color.NRGBA
	switch badgeColor {
	case ColorAccent:
		background, fallback = palette.AccentSoft, palette.Accent
	case ColorSuccess:
		background, fallback = palette.SuccessSoft, palette.Success
	case ColorWarning:
		background, fallback = palette.WarningSoft, palette.Warning
	case ColorDanger:
		background, fallback = palette.DangerSoft, palette.Danger
	default:
		background = palette.DefaultColor()
		background.A = byte((uint16(background.A) + 1) / 2)
		return background, palette.DefaultForegroundColor()
	}
	if background.A == 0 {
		background = fallback
		background.A = 0x26
	}
	return background, badgeSecondaryForeground(palette, badgeColor)
}
