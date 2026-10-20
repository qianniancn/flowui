package testtheme

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	giotext "gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Menu(active *theme.Theme, request theme.MenuRequest) theme.MenuVisual {
	tokens := active.Components.Menu
	if request.Compact {
		tokens.Padding = minDp(tokens.Padding, 4)
		tokens.Radius = minDp(tokens.Radius, 8)
		tokens.ItemGap = minDp(tokens.ItemGap, 1)
		tokens.ItemMinHeight = minDp(tokens.ItemMinHeight, 28)
		tokens.ItemRadius = minDp(tokens.ItemRadius, 4)
		tokens.ItemPaddingX = minDp(tokens.ItemPaddingX, 8)
		tokens.ItemPaddingY = minDp(tokens.ItemPaddingY, 3)
		tokens.ItemTextSize = minSp(tokens.ItemTextSize, 13)
	}
	sem := active.Semantic()
	panelBackground := tokens.BackgroundColor
	if panelBackground.A == 0 {
		panelBackground = sem.Overlay
	}
	panelForeground := tokens.ForegroundColor
	if panelForeground.A == 0 {
		panelForeground = sem.OverlayForeground
	}
	panelBorder := tokens.BorderColor
	if panelBorder.A == 0 {
		panelBorder = sem.Border
	}
	muted := tokens.MutedColor
	if muted.A == 0 {
		muted = sem.MutedForeground
	}
	metrics := theme.MenuMetrics{
		Width: tokens.Width, MaxHeight: tokens.MaxHeight, MaxWidthFraction: tokens.MaxWidthFraction,
		Padding: tokens.Padding, Radius: tokens.Radius, BorderWidth: tokens.BorderWidth,
		ItemGap: tokens.ItemGap, ItemMinHeight: tokens.ItemMinHeight, ItemRadius: tokens.ItemRadius,
		ItemPaddingX: tokens.ItemPaddingX, ItemPaddingY: tokens.ItemPaddingY, ItemContentGap: tokens.ItemContentGap,
		ShortcutHeight:   tokens.ShortcutHeight,
		ShortcutPaddingX: tokens.ShortcutPaddingX, IndicatorSize: tokens.IndicatorSize,
		IndicatorContentGap: tokens.IndicatorContentGap, CheckmarkSize: tokens.CheckmarkSize,
		RadioDotSize: tokens.RadioDotSize, IndicatorOffsetY: tokens.IndicatorOffsetY,
		SubmenuIndicatorSize: tokens.SubmenuIndicatorSize, FocusRingWidth: tokens.FocusRingWidth,
		FocusRingOffset: tokens.FocusRingOffset, PressedScale: neutralPressedScale(tokens.PressedScale),
		SectionPaddingX:   tokens.SectionPaddingX,
		SectionPaddingTop: tokens.SectionPaddingTop, SectionPaddingBottom: tokens.SectionPaddingBottom,
		SeparatorMarginX: tokens.SeparatorMarginX, SeparatorMarginY: tokens.SeparatorMarginY,
		SeparatorWidth: tokens.SeparatorWidth, DescriptionLeadingHeight: tokens.DescriptionLeadingHeight,
		DescriptionLeadingInsetTop: tokens.DescriptionLeadingInsetTop, SubmenuGap: tokens.SubmenuGap,
		ContextMenuOffset: tokens.ContextMenuOffset, EnterScale: tokens.EnterScale, ExitScale: tokens.ExitScale,
		EnterDuration: tokens.EnterDuration, ExitDuration: tokens.ExitDuration, SubmenuOpenDelay: tokens.SubmenuOpenDelay,
		AnimationDistance: tokens.AnimationDistance, ShadowOpacity: tokens.ShadowOpacity,
	}
	return theme.MenuVisual{
		Metrics: metrics,
		Colors: theme.MenuColors{
			PanelBackground: panelBackground, PanelForeground: panelForeground, PanelBorder: panelBorder,
			Separator: sem.Separator, Foreground: panelForeground, Description: muted, Shortcut: muted,
			Indicator: accent, Danger: sem.Tone(theme.ToneDanger).Fill, Focus: sem.Ring,
		},
		Motion: active.Motion,
		Root: style.Style{}.
			Background(style.Solid(panelBackground)).
			TextColor(style.Solid(panelForeground)).
			BorderColor(style.Solid(panelBorder)).
			BorderWidth(tokens.BorderWidth).
			Radius(tokens.Radius),
		Override: theme.MenuOverrideStyle(request),
		PanelColors: func(resolved style.ResolvedStyle) theme.MenuPanelColors {
			result := theme.MenuPanelColors{Background: panelBackground, Border: panelBorder, Valid: true}
			if resolved.Paint == nil {
				return result
			}
			if resolved.Paint.Background != nil {
				value, ok := resolved.Paint.Background.(style.SolidColor)
				if !ok {
					result.Valid = false
				} else {
					result.Background = value.Color
				}
			}
			if resolved.Paint.Border != nil {
				value, ok := resolved.Paint.Border.Color.(style.SolidColor)
				if !ok {
					result.Valid = false
				} else {
					result.Border = value.Color
				}
			}
			return result
		},
		Text:      behaviorMenuText(active, tokens),
		TextWidth: behaviorMenuTextWidth(active, tokens),
		Item: func(item theme.MenuItemRequest) theme.MenuItemVisual {
			itemStyle := style.Style{}.
				MinHeight(tokens.ItemMinHeight).
				PaddingX(tokens.ItemPaddingX).
				PaddingY(tokens.ItemPaddingY).
				Radius(tokens.ItemRadius).
				TextColor(style.Solid(panelForeground)).
				On(style.Disabled, style.Style{}.Opacity(active.DisabledOpacityValue()))
			return theme.MenuItemVisual{
				Defaults: style.Style{}.Part(style.PartItem, itemStyle),
				Colors: func(resolved style.ResolvedStyle) theme.MenuItemColors {
					result := theme.MenuItemColors{Foreground: panelForeground, Description: muted, Shortcut: muted, Indicator: accent}
					if resolved.Text != nil {
						if value, ok := behaviorColor(resolved.Text.Color); ok {
							result.Foreground = value
							result.Description = value
							result.Shortcut = value
						}
					}
					return result
				},
			}
		},
		Separator: func(layout.Context, image.Point, color.NRGBA) {},
		Indicator: func(layout.Context, theme.MenuIndicatorRequest) {},
		SubmenuIndicator: func(_ layout.Context, size image.Point, _ color.NRGBA) layout.Dimensions {
			return layout.Dimensions{Size: size}
		},
	}
}

func behaviorMenuText(active *theme.Theme, tokens theme.MenuTheme) theme.MenuTextRenderer {
	return func(gtx layout.Context, request theme.MenuTextRequest) layout.Dimensions {
		return behaviorMenuLabel(active, tokens, request).Layout(gtx)
	}
}

func behaviorMenuTextWidth(active *theme.Theme, tokens theme.MenuTheme) theme.MenuTextWidthResolver {
	var cache behaviorMenuTextWidthCache
	return func(gtx layout.Context, request theme.MenuTextRequest) int {
		if request.Text == "" {
			return 0
		}
		label := behaviorMenuLabel(active, tokens, request)
		key := behaviorMenuTextWidthKey{
			Text: request.Text, Role: request.Role, Constraints: gtx.Constraints,
			Metric: gtx.Metric, Locale: gtx.Locale, Font: label.Font,
			Alignment: label.Alignment, MaxLines: label.MaxLines,
			WrapPolicy: label.WrapPolicy, Truncator: label.Truncator,
			TextSize: label.TextSize, LineHeight: label.LineHeight,
			LineHeightScale: label.LineHeightScale,
		}
		if width, ok := cache.lookup(key); ok {
			return width
		}
		measureGtx := gtx
		measureGtx.Constraints = layout.Constraints{
			Max: image.Pt(gtx.Constraints.Max.X, max(gtx.Constraints.Max.Y, gtx.Dp(1024))),
		}
		measureGtx.Ops = new(op.Ops)
		width := label.Layout(measureGtx).Size.X
		cache.store(key, width)
		return width
	}
}

const behaviorMenuTextWidthCacheCapacity = 32

type behaviorMenuTextWidthKey struct {
	Text            string
	Role            theme.MenuTextRole
	Constraints     layout.Constraints
	Metric          unit.Metric
	Locale          system.Locale
	Font            font.Font
	Alignment       giotext.Alignment
	MaxLines        int
	WrapPolicy      giotext.WrapPolicy
	Truncator       string
	TextSize        unit.Sp
	LineHeight      unit.Sp
	LineHeightScale float32
}

type behaviorMenuTextWidthEntry struct {
	key   behaviorMenuTextWidthKey
	width int
	valid bool
}

type behaviorMenuTextWidthCache struct {
	entries [behaviorMenuTextWidthCacheCapacity]behaviorMenuTextWidthEntry
	next    uint8
}

func (c *behaviorMenuTextWidthCache) lookup(key behaviorMenuTextWidthKey) (int, bool) {
	for _, entry := range c.entries {
		if entry.valid && entry.key == key {
			return entry.width, true
		}
	}
	return 0, false
}

func (c *behaviorMenuTextWidthCache) store(key behaviorMenuTextWidthKey, width int) {
	index := int(c.next) % len(c.entries)
	c.entries[index] = behaviorMenuTextWidthEntry{key: key, width: width, valid: true}
	c.next = uint8((index + 1) % len(c.entries))
}

func behaviorMenuLabel(active *theme.Theme, tokens theme.MenuTheme, request theme.MenuTextRequest) material.LabelStyle {
	size := tokens.ItemTextSize
	weight := font.Normal
	switch request.Role {
	case theme.MenuTextEmpty:
	case theme.MenuTextSection:
		size = tokens.SectionTextSize
		weight = font.Medium
	case theme.MenuTextItem:
		weight = font.Medium
	case theme.MenuTextDescription:
		size = tokens.ItemDescriptionSize
	case theme.MenuTextShortcut:
		size = tokens.ShortcutTextSize
		weight = font.Medium
	default:
		panic(theme.ErrIncompleteVisualProvider)
	}
	if request.Style.Text != nil {
		textStyle := request.Style.Text
		if textStyle.FontSize != nil && *textStyle.FontSize > 0 {
			size = *textStyle.FontSize
		}
		if request.Role == theme.MenuTextItem && textStyle.FontWeight != nil {
			weight = font.Weight(*textStyle.FontWeight)
		}
	}
	label := material.Label(theme.MaterialOf(active), size, request.Text)
	label.Color = request.Color
	if request.Style.Text != nil {
		if value, ok := behaviorColor(request.Style.Text.Color); ok {
			label.Color = value
		}
	}
	label.SelectionColor = active.Semantic().Selection
	label.Font.Weight = weight
	return label
}

func minDp(value, limit unit.Dp) unit.Dp {
	if value < limit {
		return value
	}
	return limit
}

func minSp(value, limit unit.Sp) unit.Sp {
	if value < limit {
		return value
	}
	return limit
}

func neutralPressedScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .98
	}
	return value
}
