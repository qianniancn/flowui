package avatar

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

type avatarStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	diameter   unit.Dp
	radius     unit.Dp
	textSize   unit.Sp
	iconSize   unit.Dp
}

type avatarResolvedStyle struct {
	root  flowstyle.ResolvedStyle
	label flowstyle.ResolvedStyle
	icon  flowstyle.ResolvedStyle
}

func (a Widget) resolveStyle(ctx *frame.Context, gtx layout.Context) avatarResolvedStyle {
	activeTheme := frame.ActiveTheme(ctx)
	defaults := avatarDefaultDeclaration()
	variant := avatarVariantDeclaration(activeTheme, a.color, a.variant)
	size := avatarSizeDeclaration(activeTheme, a.size)
	state := flowstyle.StyleState{}
	resolved := avatarResolvedStyle{
		root:  styleruntime.ResolveStatic(ctx, state, defaults, variant, size, a.customStyle),
		label: styleruntime.ResolvePartStatic(ctx, flowstyle.PartLabel, state, defaults, variant, size, a.customStyle),
		icon:  styleruntime.ResolvePartStatic(ctx, flowstyle.PartIcon, state, defaults, variant, size, a.customStyle),
	}
	if !styleruntime.HasTransitions(resolved.root, resolved.label, resolved.icon) {
		return resolved
	}
	key := frame.ClaimKey(ctx, stateutil.KindStyle, "avatar")
	resolved.root = styleruntime.ApplyTransitions(ctx, gtx, key, resolved.root)
	resolved.label = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartLabel, resolved.label)
	resolved.icon = styleruntime.ApplyPartTransitions(ctx, gtx, key, flowstyle.PartIcon, resolved.icon)
	return resolved
}

func avatarDefaultDeclaration() flowstyle.Style {
	return flowstyle.Style{}.
		Overflow(flowstyle.OverflowHidden).
		Opacity(1).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontWeight(int(font.Medium)).MaxLines(1))

}

func avatarVariantDeclaration(activeTheme *theme.Theme, avatarColor Color, variant Variant) flowstyle.Style {
	resolved := avatarStyleFor(activeTheme, avatarColor, variant, SizeMedium)
	return flowstyle.Style{}.
		Background(flowstyle.SolidColor{Color: resolved.background}).
		TextColor(flowstyle.SolidColor{Color: resolved.foreground})

}

func avatarSizeDeclaration(activeTheme *theme.Theme, size Size) flowstyle.Style {
	resolved := avatarStyleFor(activeTheme, ColorDefault, VariantDefault, size)
	return flowstyle.Style{}.
		Width(resolved.diameter).
		AspectRatio(1).
		Radius(resolved.radius).
		Part(flowstyle.PartLabel, flowstyle.Style{}.FontSize(resolved.textSize)).
		Part(flowstyle.PartIcon, flowstyle.Style{}.Width(resolved.iconSize).Height(resolved.iconSize))

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

func avatarSoftBackground(palette theme.Palette, avatarColor Color) color.NRGBA {
	switch avatarColor {
	case ColorAccent:
		return palette.AccentSoft
	case ColorSuccess:
		return palette.SuccessSoft
	case ColorWarning:
		return palette.WarningSoft
	case ColorDanger:
		return palette.DangerSoft
	default:
		value := palette.DefaultColor()
		value.A = byte((uint16(value.A) + 1) / 2)
		return value
	}
}
