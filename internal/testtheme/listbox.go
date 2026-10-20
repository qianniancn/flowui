package testtheme

import (
	"image"
	"image/color"
	"time"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) ListBox(active *theme.Theme, _ theme.ListBoxRequest) theme.ListBoxVisual {
	return theme.ListBoxVisual{
		Metrics: theme.ListBoxMetrics{
			Padding: 0, Gap: 4, ItemMinHeight: 40,
			ItemContentGap:        8,
			SectionHeaderPaddingX: 8, SectionHeaderPaddingY: 4,
			ItemIndicatorSize: 16, SelectionDuration: 100 * time.Millisecond,
		},
		Colors: theme.ListBoxColors{
			MutedForeground: muted,
		},
		Motion: theme.MotionTheme{Enabled: true, DefaultDuration: 100 * time.Millisecond, ControlDuration: 100 * time.Millisecond, DurationScale: 1},
		Text:   behaviorListBoxText(active),
		Item: func(request theme.ListBoxItemRequest) theme.ListBoxItemVisual {
			item := style.Style{}.
				MinHeight(40).
				PaddingX(12).
				PaddingY(8).
				Radius(4).
				Background(style.RGBA(0x00000000)).
				BorderWidth(2).
				BorderColor(style.WithAlpha(focus, 0)).
				Opacity(1).
				Scale(1).
				TextColor(style.Solid(foreground)).
				On(style.Hovered, style.Style{}.Background(style.Solid(background))).
				On(style.Pressed, style.Style{}.Background(style.Solid(accent)).Scale(.98)).
				On(style.FocusVisible, style.Style{}.BorderColor(style.Solid(focus))).
				On(style.Disabled, style.Style{}.Opacity(.5))
			return theme.ListBoxItemVisual{
				Defaults: style.Style{}.Part(style.PartItem, item),
				Colors: func(resolved style.ResolvedStyle) theme.ListBoxItemColors {
					if resolved.Text != nil {
						if value, ok := behaviorColor(resolved.Text.Color); ok {
							return theme.ListBoxItemColors{Foreground: value, Description: value, Indicator: value}
						}
					}
					return theme.ListBoxItemColors{Foreground: foreground, Description: muted, Indicator: accent}
				},
			}
		},
		Indicator: func(layout.Context, image.Point, float32, color.NRGBA) {},
	}
}

func behaviorListBoxText(active *theme.Theme) theme.ListBoxTextRenderer {
	return func(gtx layout.Context, request theme.ListBoxTextRequest) layout.Dimensions {
		size := unit.Sp(14)
		weight := font.Normal
		switch request.Role {
		case theme.ListBoxTextSectionHeader, theme.ListBoxTextDescription:
			size = unit.Sp(12)
		case theme.ListBoxTextItem:
			weight = font.Medium
		case theme.ListBoxTextEmpty:
		default:
			panic(theme.ErrIncompleteVisualProvider)
		}
		label := material.Label(theme.MaterialOf(active), size, request.Text)
		label.Color = request.Color
		label.SelectionColor = active.Semantic().Selection
		label.Font.Weight = weight
		return label.Layout(gtx)
	}
}
