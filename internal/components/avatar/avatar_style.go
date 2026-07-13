package avatar

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type avatarStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	diameter   unit.Dp
	radius     unit.Dp
	textSize   unit.Sp
	iconSize   unit.Dp
}

func avatarStyleFor(activeTheme *theme.Theme, avatarColor Color, variant Variant, size Size) avatarStyle {
	tokens := activeTheme.Components.Avatar
	style := avatarStyle{
		background: activeTheme.Palette.DefaultColor(),
		foreground: avatarForeground(activeTheme.Palette, avatarColor),
		diameter:   tokens.MediumSize,
		radius:     tokens.MediumRadius,
		textSize:   tokens.MediumTextSize,
		iconSize:   tokens.MediumIconSize,
	}
	if variant == VariantSoft {
		style.background = avatarSoftBackground(activeTheme.Palette, avatarColor)
	}
	switch size {
	case SizeSmall:
		style.diameter = tokens.SmallSize
		style.radius = tokens.SmallRadius
		style.textSize = tokens.SmallTextSize
		style.iconSize = tokens.SmallIconSize
	case SizeLarge:
		style.diameter = tokens.LargeSize
		style.radius = tokens.LargeRadius
		style.textSize = tokens.LargeTextSize
		style.iconSize = tokens.LargeIconSize
	}
	return style
}

func avatarForeground(palette theme.Palette, avatarColor Color) color.NRGBA {
	switch avatarColor {
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

func avatarSoftBackground(palette theme.Palette, avatarColor Color) color.NRGBA {
	var value color.NRGBA
	var fallback color.NRGBA
	switch avatarColor {
	case ColorAccent:
		value, fallback = palette.AccentSoft, palette.Accent
	case ColorSuccess:
		value, fallback = palette.SuccessSoft, palette.Success
	case ColorWarning:
		value, fallback = palette.WarningSoft, palette.Warning
	case ColorDanger:
		value, fallback = palette.DangerSoft, palette.Danger
	default:
		value, fallback = palette.DefaultColor(), palette.DefaultColor()
		value.A = byte((uint16(value.A) + 1) / 2)
		return value
	}
	if value.A != 0 {
		return value
	}
	fallback.A = 0x26
	return fallback
}
