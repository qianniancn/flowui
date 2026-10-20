// Package tabsstyle contains the theme-owned visual recipe for tabs.
// Selection, focus, keyboard navigation and panel lifetime remain in
// widget/tabs; this package owns chrome declarations, colors and geometry.
package tabsstyle

import (
	"gioui.org/font"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"image/color"
)

// BuildTabs builds the complete provider-owned tabs visual plan.
func BuildTabs(active *theme.Theme, request theme.TabsRequest) theme.TabsVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Tabs
	colors := tabsColors(active, request)
	metrics := tabsMetrics(tokens)
	renderers := newTabsRenderers(tokens)
	return theme.TabsVisual{
		Root:      style.Style{}.TextColor(style.Solid(colors.Foreground)),
		Panel:     style.Style{}.Part(style.PartPanel, panelDeclaration(tokens, colors)),
		Override:  theme.TabsOverrideStyle(request),
		Metrics:   metrics,
		Colors:    colors,
		Motion:    active.Motion,
		Indicator: renderers.indicator,
		Focus:     renderers.focus,
		Separator: renderers.separator,
		List:      renderers.list,
		Scroll:    renderers.scroll,
		Chevron:   renderers.chevron,
		Close:     renderers.close,
		Add:       renderers.add,
		Overflow:  renderers.overflow,
		Item: func(item theme.TabsItemRequest) theme.TabsItemVisual {
			return itemVisual(active, request, item, colors)
		},
	}
}

func tabsMetrics(tokens theme.TabsTheme) theme.TabsMetrics {
	return theme.TabsMetrics{
		RootGap: tokens.RootGap, ListPadding: tokens.ListPadding,
		TabHeight: tokens.TabHeight, SmallTabHeight: tokens.SmallTabHeight, LargeTabHeight: tokens.LargeTabHeight,
		TabMinWidth: tokens.TabMinWidth, TabPaddingX: tokens.TabPaddingX, SmallTabPaddingX: tokens.SmallTabPaddingX,
		LargeTabPaddingX: tokens.LargeTabPaddingX, TabGap: tokens.TabGap,
		IconSize: tokens.IconSize, IconGap: tokens.IconGap,
		CloseButtonSize: tokens.CloseButtonSize, CloseButtonGap: tokens.CloseButtonGap,
		ExtraContentGap: tokens.ExtraContentGap, IndicatorLineWidth: tokens.IndicatorLineWidth, IndicatorWidth: tokens.IndicatorWidth,
		IndicatorMinWidth: tokens.IndicatorMinWidth, IndicatorInset: tokens.IndicatorInset,
		PanelGap:      tokens.PanelGap,
		ColorDuration: tokens.ColorDuration, IndicatorDuration: tokens.IndicatorDuration,
		PanelDuration: tokens.PanelDuration, ScrollButtonSize: tokens.ScrollButtonSize,
		ScrollButtonInset: tokens.ScrollButtonInset,
	}
}

func tabsColors(active *theme.Theme, request theme.TabsRequest) theme.TabsColors {
	sem := active.Semantic()
	listBackground := sem.Tone(theme.ToneDefault).Fill
	listBorder := sem.Border
	selectedForeground := sem.SegmentForeground
	indicator := sem.Segment
	if request.Variant == 1 {
		listBackground = color.NRGBA{}
		selectedForeground = sem.Foreground
		indicator = sem.Tone(theme.TonePrimary).Fill
	}
	if request.Color == 1 {
		indicator = sem.Tone(theme.TonePrimary).Fill
		if request.Variant == 1 {
			selectedForeground = sem.Tone(theme.TonePrimary).Fill
		} else {
			selectedForeground = sem.Tone(theme.TonePrimary).OnFill
		}
	}
	return theme.TabsColors{
		ListBackground:     listBackground,
		ScrollBackground:   sem.Background,
		ListBorder:         listBorder,
		Foreground:         sem.MutedForeground,
		SelectedForeground: selectedForeground,
		Indicator:          indicator,
		Separator:          sem.MutedForeground,
		Focus:              sem.Ring,
		PanelBackground:    active.Components.Tabs.PanelBackground,
		PanelBorder:        active.Components.Tabs.PanelBorderColor,
		ScrollForeground:   sem.Foreground,
		DisabledForeground: active.DisabledColor(sem.Foreground),
	}
}

func itemVisual(active *theme.Theme, request theme.TabsRequest, item theme.TabsItemRequest, colors theme.TabsColors) theme.TabsItemVisual {
	selectedForeground := colors.SelectedForeground
	indicator := colors.Indicator
	separator := colors.Separator
	focus := colors.Focus
	if item.State.Disabled {
		indicator = active.DisabledColor(indicator)
		separator = active.DisabledColor(separator)
		focus = color.NRGBA{}
	}
	size, weight := active.Components.Tabs.TextSize, font.Medium
	switch request.Size {
	case 1:
		size, weight = active.Components.Tabs.TextSize, font.Normal
	case 2:
		size = active.Components.Tabs.LargeTextSize
	}
	declaration := style.Style{}.
		TextColor(style.Solid(item.Foreground)).
		FontSize(size).
		FontWeight(int(weight)).
		Opacity(1).
		Transition(style.PropOpacity, theme.ResolveMotionDuration(active.Motion, active.Components.Tabs.ColorDuration)).
		OnNot(style.Hovered, style.Selected|style.Disabled, style.Style{}.Opacity(.7)).
		On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue()))
	if item.State.Disabled {
		selectedForeground = active.DisabledColor(selectedForeground)
	}
	return theme.TabsItemVisual{
		Declaration:        style.Style{}.Part(style.PartItem, declaration),
		Foreground:         colors.Foreground,
		SelectedForeground: selectedForeground,
		Indicator:          indicator,
		Separator:          separator,
		Focus:              focus,
	}
}

func panelDeclaration(tokens theme.TabsTheme, colors theme.TabsColors) style.Style {
	declaration := style.Style{}.Padding(tokens.PanelPadding).Opacity(1)
	if colors.PanelBackground.A != 0 {
		declaration = declaration.Background(style.Solid(colors.PanelBackground)).Radius(tokens.PanelRadius)
		if colors.PanelBorder.A != 0 {
			declaration = declaration.BorderColor(style.Solid(colors.PanelBorder)).BorderWidth(1)
		}
	}
	return declaration
}
