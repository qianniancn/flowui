package tabs

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/theme"
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

func tabsSizeStyleFor(theme *theme.Theme, size TabsSize) tabsSizeStyle {
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

func tabsListStyleFor(theme *theme.Theme, variant TabsVariant) tabsListStyle {
	if variant == TabsSecondary {
		return tabsListStyle{border: theme.Palette.Border}
	}
	return tabsListStyle{background: theme.Palette.Default}
}

func tabsItemStyleFor(activeTheme *theme.Theme, variant TabsVariant, tabsColor TabsColor, hovered, disabled bool) tabsItemStyle {
	selectedForeground := activeTheme.Palette.SegmentForeground
	indicator := activeTheme.Palette.Segment
	if variant == TabsSecondary {
		selectedForeground = activeTheme.Palette.Foreground
		indicator = activeTheme.Palette.Accent
	}
	if tabsColor == TabsColorAccent {
		indicator = activeTheme.Palette.Accent
		if variant == TabsSecondary {
			selectedForeground = activeTheme.Palette.Accent
		} else {
			selectedForeground = activeTheme.Palette.AccentForeground
		}
	}
	style := tabsItemStyle{
		foreground:         activeTheme.Palette.MutedForeground,
		selectedForeground: selectedForeground,
		indicator:          indicator,
		separator:          activeTheme.Palette.MutedForeground,
		focus:              activeTheme.Palette.Focus,
	}
	style.separator.A = byte(float32(style.separator.A)*0.25 + 0.5)
	if hovered {
		style.separator.A = byte(float32(style.separator.A)*0.7 + 0.5)
	}
	if disabled {
		style.indicator = activeTheme.DisabledColor(style.indicator)
		style.separator = activeTheme.DisabledColor(style.separator)
		style.focus = color.NRGBA{}
	}
	return style
}
