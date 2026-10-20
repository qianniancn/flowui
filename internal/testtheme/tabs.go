package testtheme

import (
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Tabs(active *theme.Theme, request theme.TabsRequest) theme.TabsVisual {
	tokens := active.Components.Tabs
	selected := accent
	indicator := accent
	if request.Variant == 1 {
		selected, indicator = foreground, focus
	}
	if request.Color == 1 {
		selected, indicator = accentFg, focus
	}
	metrics := theme.TabsMetrics{
		RootGap: tokens.RootGap, ListPadding: tokens.ListPadding,
		TabHeight: tokens.TabHeight, SmallTabHeight: tokens.SmallTabHeight, LargeTabHeight: tokens.LargeTabHeight,
		TabMinWidth: tokens.TabMinWidth, TabPaddingX: tokens.TabPaddingX, SmallTabPaddingX: tokens.SmallTabPaddingX,
		LargeTabPaddingX: tokens.LargeTabPaddingX, TabGap: tokens.TabGap,
		IconSize: tokens.IconSize, IconGap: tokens.IconGap,
		CloseButtonSize: tokens.CloseButtonSize, CloseButtonGap: tokens.CloseButtonGap, ExtraContentGap: tokens.ExtraContentGap,
		IndicatorLineWidth: tokens.IndicatorLineWidth, IndicatorWidth: tokens.IndicatorWidth,
		IndicatorMinWidth: tokens.IndicatorMinWidth, IndicatorInset: tokens.IndicatorInset,
		PanelGap:      tokens.PanelGap,
		ColorDuration: tokens.ColorDuration, IndicatorDuration: tokens.IndicatorDuration, PanelDuration: tokens.PanelDuration,
		ScrollButtonSize: tokens.ScrollButtonSize, ScrollButtonInset: tokens.ScrollButtonInset,
	}
	colors := theme.TabsColors{
		ListBackground: background, ScrollBackground: background, ListBorder: focus, Foreground: muted, SelectedForeground: selected,
		Indicator: indicator, Separator: muted, Focus: focus, PanelBackground: color.NRGBA{},
		PanelBorder: color.NRGBA{}, ScrollForeground: foreground, DisabledForeground: muted,
	}
	return theme.TabsVisual{
		Root:     style.Style{}.TextColor(style.Solid(foreground)),
		Panel:    style.Style{}.Part(style.PartPanel, style.Style{}.Padding(tokens.PanelPadding).Opacity(1)),
		Override: theme.TabsOverrideStyle(request), Metrics: metrics, Colors: colors, Motion: active.Motion,
		Indicator: neutralTabsIndicatorRenderer,
		Focus:     neutralTabsFocusRenderer, Separator: neutralTabsSeparatorRenderer,
		List: neutralTabsListRenderer, Scroll: neutralTabsScrollRenderer,
		Chevron: neutralTabsChevronRenderer, Close: neutralTabsCloseRenderer,
		Add: neutralTabsAddRenderer, Overflow: neutralTabsOverflowRenderer,
		Item: func(item theme.TabsItemRequest) theme.TabsItemVisual {
			itemSelected := selected
			itemIndicator, itemSeparator, itemFocus := indicator, muted, focus
			if item.State.Disabled {
				itemSelected = active.DisabledColor(itemSelected)
				itemIndicator = active.DisabledColor(itemIndicator)
				itemSeparator = active.DisabledColor(itemSeparator)
				itemFocus = color.NRGBA{}
			}
			textSize, weight := tokens.TextSize, font.Medium
			switch request.Size {
			case 1:
				textSize, weight = tokens.TextSize, font.Normal
			case 2:
				textSize = tokens.LargeTextSize
			}
			declaration := style.Style{}.Part(style.PartItem, style.Style{}.
				TextColor(style.Solid(item.Foreground)).FontSize(textSize).FontWeight(int(weight)).Opacity(1).
				Transition(style.PropOpacity, tokens.ColorDuration).
				OnNot(style.Hovered, style.Selected|style.Disabled, style.Style{}.Opacity(.7)).
				On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue())))
			return theme.TabsItemVisual{Declaration: declaration, Foreground: muted, SelectedForeground: itemSelected,
				Indicator: itemIndicator, Separator: itemSeparator, Focus: itemFocus}
		},
	}
}

// neutralTabsIndicatorRenderer keeps behavior fixtures independent of the
// shipped tabs recipe while satisfying the renderer contract.
func neutralTabsIndicatorRenderer(layout.Context, theme.TabsIndicatorRenderRequest) {}

func neutralTabsFocusRenderer(layout.Context, theme.TabsFocusRenderRequest) {}

func neutralTabsSeparatorRenderer(layout.Context, theme.TabsSeparatorRenderRequest) {}

func neutralTabsListRenderer(layout.Context, theme.TabsListRenderRequest) {}

func neutralTabsScrollRenderer(layout.Context, theme.TabsScrollRenderRequest) {}

func neutralTabsChevronRenderer(layout.Context, theme.TabsChevronRenderRequest) {}

func neutralTabsCloseRenderer(layout.Context, theme.TabsCloseRenderRequest) {}

func neutralTabsAddRenderer(layout.Context, theme.TabsAddRenderRequest) {}

func neutralTabsOverflowRenderer(layout.Context, theme.TabsOverflowRenderRequest) {}
