package badge

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
	stateutil "github.com/qianniancn/FlowUI/internal/state"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
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

type badgeResolvedStyle struct {
	root  flowstyle.ResolvedStyle
	label flowstyle.ResolvedStyle
}

func (b Widget) resolveStyle(ctx *frame.Context, gtx layout.Context) badgeResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := badgeDefaultDeclaration(activeTheme, ctx.BackgroundColor())
	variant := badgeVariantDeclaration(activeTheme, b.color, b.variant)
	size := badgeSizeDeclaration(activeTheme, b.size)
	state := flowstyle.StyleState{}
	resolved := badgeResolvedStyle{
		root:  styleruntime.ResolveStatic(ctx, state, defaults, variant, size, b.customStyle),
		label: styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, defaults, variant, size, b.customStyle),
	}
	if !styleruntime.HasTransitions(resolved.root, resolved.label) {
		return resolved
	}
	key := frame.ClaimKey(ctx, stateutil.KindStyle, "badge")
	resolved.root = styleruntime.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, resolved.label)
	return resolved
}

func badgeDefaultDeclaration(activeTheme *theme.Theme, border color.NRGBA) flowstyle.Style {
	return flowstyle.Style{}.
		BorderColor(flowstyle.SolidColor{Color: border}).
		BorderWidth(activeTheme.Components.Badge.BorderWidth).
		Opacity(1).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontWeight(int(font.Medium)).MaxLines(1))

}

func badgeVariantDeclaration(activeTheme *theme.Theme, badgeColor Color, variant Variant) flowstyle.Style {
	resolved := badgeStyleFor(activeTheme, badgeColor, variant, SizeMedium, color.NRGBA{})
	return flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: resolved.background}).
		TextColor(flowstyle.SolidColor{Color: resolved.foreground})

}

func badgeSizeDeclaration(activeTheme *theme.Theme, size Size) flowstyle.Style {
	resolved := badgeStyleFor(activeTheme, ColorDefault, VariantPrimary, size, color.NRGBA{})
	return flowstyle.Style{}.
		MinWidth(resolved.minSize).
		MinHeight(resolved.minSize).
		PaddingX(resolved.paddingX).
		Radius(resolved.radius).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontSize(resolved.textSize).LineHeight(resolved.lineHeight))

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
		return palette.Accent, palette.AccentForeground
	case ColorSuccess:
		return palette.Success, palette.SuccessForeground
	case ColorWarning:
		return palette.Warning, palette.WarningForeground
	case ColorDanger:
		return palette.Danger, palette.DangerForeground
	default:
		return palette.DefaultColor(), palette.DefaultForegroundColor()
	}
}

func badgeSecondaryForeground(palette theme.Palette, badgeColor Color) color.NRGBA {
	switch badgeColor {
	case ColorAccent:
		return palette.AccentSoftForeground
	case ColorSuccess:
		return palette.SuccessSoftForegroundColor()
	case ColorWarning:
		return palette.WarningSoftForegroundColor()
	case ColorDanger:
		return palette.DangerSoftForeground
	default:
		return palette.DefaultForegroundColor()
	}
}

func badgeSoftColors(palette theme.Palette, badgeColor Color) (color.NRGBA, color.NRGBA) {
	switch badgeColor {
	case ColorAccent:
		return palette.AccentSoft, badgeSecondaryForeground(palette, badgeColor)
	case ColorSuccess:
		return palette.SuccessSoft, badgeSecondaryForeground(palette, badgeColor)
	case ColorWarning:
		return palette.WarningSoft, badgeSecondaryForeground(palette, badgeColor)
	case ColorDanger:
		return palette.DangerSoft, badgeSecondaryForeground(palette, badgeColor)
	default:
		background := palette.DefaultColor()
		background.A = byte((uint16(background.A) + 1) / 2)
		return background, palette.DefaultForegroundColor()
	}
}
