package chip

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/frame"
	stateutil "github.com/qianniancn/flowui/internal/state"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

type chipStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	radius     unit.Dp
}

type chipResolvedStyle struct {
	root          flowstyle.ResolvedStyle
	label         flowstyle.ResolvedStyle
	icon          flowstyle.ResolvedStyle
	labelPaddingX unit.Dp
	contentGap    unit.Dp
}

func (c Widget) resolveStyle(ctx *frame.Context, gtx layout.Context) chipResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := chipDefaultDeclaration(activeTheme)
	variant := chipVariantDeclaration(activeTheme, c.color, c.variant)
	size := chipSizeDeclaration(activeTheme, c.size)
	state := flowstyle.StyleState{}
	tokens := chipSizeStyleFor(activeTheme, c.size)
	resolved := chipResolvedStyle{
		root:          styleruntime.ResolveStatic(ctx, state, defaults, variant, size, c.customStyle),
		label:         styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, defaults, variant, size, c.customStyle),
		icon:          styleruntime.ResolvePartStatic(ctx, flowstyle.PartIcon, state, defaults, variant, size, c.customStyle),
		labelPaddingX: tokens.labelPaddingX,
		contentGap:    tokens.contentGap,
	}
	if !styleruntime.HasTransitions(resolved.root, resolved.label, resolved.icon) {
		return resolved
	}
	key := frame.ClaimKey(ctx, stateutil.KindStyle, "chip")
	resolved.root = styleruntime.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, resolved.label)
	resolved.icon = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartIcon, resolved.icon)
	return resolved
}

func chipDefaultDeclaration(activeTheme *theme.Theme) flowstyle.Style {
	return flowstyle.Style{}.
		Radius(activeTheme.Components.Chip.Radius).
		Opacity(1).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontWeight(int(font.Medium)).MaxLines(1))

}

func chipVariantDeclaration(activeTheme *theme.Theme, chipColor Color, variant Variant) flowstyle.Style {
	resolved := chipStyleFor(activeTheme, chipColor, variant)
	return flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: resolved.background}).
		TextColor(flowstyle.SolidColor{Color: resolved.foreground})

}

func chipSizeDeclaration(activeTheme *theme.Theme, size Size) flowstyle.Style {
	resolved := chipSizeStyleFor(activeTheme, size)
	return flowstyle.Style{}.
		Height(resolved.height).
		PaddingX(resolved.paddingX).
		PaddingY(resolved.paddingY).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontSize(resolved.textSize).LineHeight(resolved.lineHeight))

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
	defaultBackground := palette.SurfaceTertiary
	defaultForeground := palette.SurfaceTertiaryForeground
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
		return palette.Accent, palette.AccentForeground
	case ColorSuccess:
		return palette.Success, palette.SuccessForeground
	case ColorWarning:
		return palette.Warning, palette.WarningForeground
	case ColorDanger:
		return palette.Danger, palette.DangerForeground
	default:
		return defaultBackground, defaultForeground
	}
}

func chipSoftColors(palette theme.Palette, chipColor Color, defaultForeground color.NRGBA) (color.NRGBA, color.NRGBA) {
	switch chipColor {
	case ColorAccent:
		return palette.AccentSoft, palette.AccentSoftForeground
	case ColorSuccess:
		return palette.SuccessSoft, palette.SuccessSoftForegroundColor()
	case ColorWarning:
		return palette.WarningSoft, palette.WarningSoftForegroundColor()
	case ColorDanger:
		return palette.DangerSoft, palette.DangerSoftForeground
	default:
		return defaultSoftColor(palette.SurfaceTertiary), defaultForeground
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
