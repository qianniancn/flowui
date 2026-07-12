package chip

import (
	"image/color"

	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type chipStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	radius     unit.Dp
}

type chipSizeStyle struct {
	height        unit.Dp
	paddingX      unit.Dp
	paddingY      unit.Dp
	labelPaddingX unit.Dp
	contentGap    unit.Dp
	textSize      unit.Sp
	lineHeight    unit.Sp
}

func chipStyleFor(activeTheme *theme.Theme, chipColor Color, variant Variant) chipStyle {
	palette := activeTheme.Palette
	defaultBackground := theme.ColorOr(palette.SurfaceTertiary, palette.SurfacePressed)
	defaultForeground := theme.ColorOr(palette.SurfaceTertiaryForeground, palette.Foreground)
	style := chipStyle{
		background: defaultBackground,
		foreground: defaultForeground,
		radius:     activeTheme.Components.Chip.Radius,
	}

	solidBackground, solidForeground := chipSolidColors(palette, chipColor, defaultBackground, defaultForeground)
	softBackground, softForeground := chipSoftColors(palette, chipColor, defaultForeground)
	switch variant {
	case VariantPrimary:
		style.background = solidBackground
		style.foreground = solidForeground
	case VariantTertiary:
		style.background = color.NRGBA{}
		style.foreground = softForeground
	case VariantSoft:
		style.background = softBackground
		style.foreground = softForeground
	default:
		style.foreground = softForeground
	}
	return style
}

func chipSolidColors(palette theme.Palette, chipColor Color, defaultBackground, defaultForeground color.NRGBA) (color.NRGBA, color.NRGBA) {
	switch chipColor {
	case ColorAccent:
		return palette.Accent, theme.ColorOr(palette.AccentForeground, defaultForeground)
	case ColorSuccess:
		return palette.Success, theme.ColorOr(palette.SuccessForeground, defaultForeground)
	case ColorWarning:
		return palette.Warning, theme.ColorOr(palette.WarningForeground, defaultForeground)
	case ColorDanger:
		return palette.Danger, theme.ColorOr(palette.DangerForeground, defaultForeground)
	default:
		return defaultBackground, defaultForeground
	}
}

func chipSoftColors(palette theme.Palette, chipColor Color, defaultForeground color.NRGBA) (color.NRGBA, color.NRGBA) {
	switch chipColor {
	case ColorAccent:
		return softColor(palette.AccentSoft, palette.Accent), theme.ColorOr(palette.AccentSoftForeground, palette.Accent)
	case ColorSuccess:
		return softColor(palette.SuccessSoft, palette.Success), palette.SuccessSoftForegroundColor()
	case ColorWarning:
		return softColor(palette.WarningSoft, palette.Warning), palette.WarningSoftForegroundColor()
	case ColorDanger:
		return softColor(palette.DangerSoft, palette.Danger), theme.ColorOr(palette.DangerSoftForeground, palette.Danger)
	default:
		return defaultSoftColor(theme.ColorOr(palette.SurfaceTertiary, palette.SurfacePressed)), defaultForeground
	}
}

func defaultSoftColor(value color.NRGBA) color.NRGBA {
	value.A = byte((uint16(value.A) + 1) / 2)
	return value
}

func chipSizeStyleFor(activeTheme *theme.Theme, size Size) chipSizeStyle {
	tokens := activeTheme.Components.Chip
	style := chipSizeStyle{
		height:        tokens.MediumHeight,
		paddingX:      tokens.MediumPaddingX,
		paddingY:      tokens.MediumPaddingY,
		labelPaddingX: tokens.LabelPaddingX,
		contentGap:    tokens.ContentGap,
		textSize:      tokens.MediumTextSize,
		lineHeight:    tokens.LineHeight,
	}
	switch size {
	case SizeSmall:
		style.height = tokens.SmallHeight
		style.paddingX = tokens.SmallPaddingX
		style.paddingY = tokens.SmallPaddingY
		style.textSize = tokens.SmallTextSize
	case SizeLarge:
		style.height = tokens.LargeHeight
		style.paddingX = tokens.LargePaddingX
		style.paddingY = tokens.LargePaddingY
		style.textSize = tokens.LargeTextSize
	}
	return style
}

func softColor(value, fallback color.NRGBA) color.NRGBA {
	if value.A != 0 {
		return value
	}
	if fallback.A != 0 {
		fallback.A = 0x26
	}
	return fallback
}
