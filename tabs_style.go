package flowui

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
)

type tabsListStyle struct {
	background color.NRGBA
	border     color.NRGBA
}

type tabsItemStyle struct {
	foreground         color.NRGBA
	selectedForeground color.NRGBA
	indicator          color.NRGBA
	separator          color.NRGBA
	focus              color.NRGBA
}

type tabsSizeStyle struct {
	height   unit.Dp
	paddingX unit.Dp
	textSize unit.Sp
	weight   font.Weight
}

func tabsSizeStyleFor(theme *Theme, size TabsSize) tabsSizeStyle {
	style := tabsSizeStyle{
		height:   theme.Components.Tabs.TabHeight,
		paddingX: theme.Components.Tabs.TabPaddingX,
		textSize: theme.Components.Tabs.TextSize,
		weight:   font.Medium,
	}
	if size == TabsSmall {
		style.height = theme.Components.Tabs.SmallTabHeight
		style.paddingX = theme.Components.Tabs.SmallTabPaddingX
		style.weight = font.Normal
	}
	return style
}

func tabsListStyleFor(theme *Theme, variant TabsVariant) tabsListStyle {
	if variant == TabsSecondary {
		return tabsListStyle{border: theme.Palette.Border}
	}
	return tabsListStyle{background: theme.Palette.SurfaceRaised}
}

func tabsItemStyleFor(theme *Theme, variant TabsVariant, tabsColor TabsColor, hovered, disabled bool) tabsItemStyle {
	selectedForeground := paletteColor(theme.Palette.SegmentForeground, theme.Palette.Foreground)
	indicator := paletteColor(theme.Palette.Segment, theme.Palette.Surface)
	if variant == TabsSecondary {
		selectedForeground = theme.Palette.Foreground
		indicator = theme.Palette.Accent
	}
	if tabsColor == TabsColorAccent {
		indicator = theme.Palette.Accent
		if variant == TabsSecondary {
			selectedForeground = theme.Palette.Accent
		} else {
			selectedForeground = theme.Palette.AccentForeground
		}
	}
	style := tabsItemStyle{
		foreground:         theme.Palette.MutedForeground,
		selectedForeground: selectedForeground,
		indicator:          indicator,
		separator:          theme.Palette.MutedForeground,
		focus:              theme.Palette.Focus,
	}
	style.separator.A = byte(float32(style.separator.A)*0.25 + 0.5)
	if hovered {
		style.foreground.A = byte(float32(style.foreground.A)*0.7 + 0.5)
		style.separator.A = byte(float32(style.separator.A)*0.7 + 0.5)
	}
	if disabled {
		style.foreground = theme.DisabledColor(style.foreground)
		style.selectedForeground = theme.DisabledColor(style.selectedForeground)
		style.indicator = theme.DisabledColor(style.indicator)
		style.separator = theme.DisabledColor(style.separator)
		style.focus = color.NRGBA{}
	}
	return style
}
