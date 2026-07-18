package tree

import (
	"image/color"

	"github.com/qianniancn/FlowUI/internal/theme"
)

func treeTokensFor(activeTheme *theme.Theme, size Size) theme.TreeTheme {
	tokens := activeTheme.Components.Tree
	if size == SizeSmall {
		tokens.Padding = tokens.SmallPadding
		tokens.Gap = tokens.SmallGap
		tokens.RowHeight = tokens.SmallRowHeight
		tokens.DescriptionRowHeight = tokens.SmallDescriptionRowHeight
		tokens.RowRadius = tokens.SmallRowRadius
		tokens.RowPaddingX = tokens.SmallRowPaddingX
		tokens.RowPaddingY = tokens.SmallRowPaddingY
		tokens.Indent = tokens.SmallIndent
		tokens.ChevronSlotSize = tokens.SmallChevronSlotSize
		tokens.ChevronIconSize = tokens.SmallChevronIconSize
		tokens.ContentGap = tokens.SmallContentGap
		tokens.ItemTextSize = tokens.SmallItemTextSize
		tokens.ItemDescriptionSize = tokens.SmallItemDescriptionSize
	}
	return tokens
}

type treeRootStyle struct {
	background color.NRGBA
	foreground color.NRGBA
	shadow     bool
}

func treeRootStyleFor(activeTheme *theme.Theme, variant Variant) treeRootStyle {
	style := treeRootStyle{foreground: activeTheme.Palette.Foreground}
	if variant == VariantSurface {
		style.background = activeTheme.Palette.Surface
		style.foreground = theme.ColorOr(activeTheme.Palette.SurfaceForeground, activeTheme.Palette.Foreground)
		style.shadow = true
	}
	return style
}

type treeItemStyle struct {
	background  color.NRGBA
	foreground  color.NRGBA
	description color.NRGBA
	chevron     color.NRGBA
	focus       color.NRGBA
	opacity     float32
}

func treeItemStyleFor(activeTheme *theme.Theme, selected, hovered, disabled bool) treeItemStyle {
	style := treeItemStyle{
		foreground:  activeTheme.Palette.Foreground,
		description: activeTheme.Palette.MutedForeground,
		chevron:     activeTheme.Palette.MutedForeground,
		focus:       activeTheme.Palette.Focus,
		opacity:     1,
	}
	if hovered {
		style.background = theme.ColorOr(activeTheme.Palette.SurfaceTertiary, activeTheme.Palette.SurfaceRaised)
	}
	if selected {
		style.background = activeTheme.Palette.AccentSoft
		style.foreground = theme.ColorOr(activeTheme.Palette.AccentSoftForeground, activeTheme.Palette.Accent)
		style.description = style.foreground
		style.description.A = byte(float32(style.description.A)*0.78 + 0.5)
		style.chevron = style.foreground
		if hovered {
			style.background = theme.ColorOr(activeTheme.Palette.AccentSoftHover, activeTheme.Palette.AccentSoft)
		}
	}
	if disabled {
		style.opacity = activeTheme.DisabledOpacityValue()
		style.focus = color.NRGBA{}
	}
	return style
}
