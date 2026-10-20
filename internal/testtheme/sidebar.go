package testtheme

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/theme"
)

// WithSidebarProvider installs the neutral sidebar capability on a copied
// theme. It is used by legacy sidebar tests that start from a theme without a
// complete provider.
func WithSidebarProvider(value *theme.Theme) *theme.Theme {
	if value == nil {
		return nil
	}
	clone := *value
	clone.VisualProvider = sidebarOnlyProvider{ComponentVisualProvider: BehaviorProvider()}
	return &clone
}

type sidebarOnlyProvider struct {
	theme.ComponentVisualProvider
}

var _ theme.VisualProvider = sidebarOnlyProvider{}

func (sidebarOnlyProvider) Sidebar(active *theme.Theme, request theme.SidebarRequest) theme.SidebarVisual {
	return behaviorProvider{}.Sidebar(active, request)
}

func (behaviorProvider) Sidebar(active *theme.Theme, request theme.SidebarRequest) theme.SidebarVisual {
	tokens := active.Components.Sidebar
	metrics := request.Apply(theme.SidebarMetrics(tokens))
	sem := active.Semantic()
	foreground, sidebarBackground, sidebarMuted := sem.SurfaceForeground, sem.Surface, sem.MutedForeground
	return theme.SidebarVisual{
		Metrics:    metrics,
		Foreground: foreground, Background: sidebarBackground, MutedForeground: sidebarMuted,
		Motion: active.Motion, ExpansionEasing: theme.EaseOut, CollapseEasing: theme.EaseInOut,
		Override:   theme.SidebarOverrideStyle(request),
		Root:       func(layout.Context, image.Point) {},
		RenderItem: func(layout.Context, image.Point, theme.SidebarItemVisual, float32) {},
		Switcher: func(gtx layout.Context, request theme.SidebarSwitcherRequest) layout.Dimensions {
			size := max(gtx.Dp(request.Size), 0)
			return layout.Dimensions{Size: image.Pt(size, size)}
		},
		Item: func(request theme.SidebarItemRequest) theme.SidebarItemVisual {
			value := theme.SidebarItemVisual{Foreground: foreground, Focus: sem.Ring, Opacity: 1, Radius: metrics.ItemRadius}
			if request.Hovered {
				value.Background = sem.SurfaceTertiary
			}
			if request.Pressed {
				value.Background = sem.SurfacePressed
			}
			if request.Selected {
				tone := sem.Tone(theme.TonePrimary)
				value.Background = tone.Soft
				value.Foreground = tone.OnSoft
			}
			if request.Disabled {
				value.Opacity = active.DisabledOpacityValue()
				value.Focus = color.NRGBA{}
			}
			return value
		},
		Text: func(gtx layout.Context, request theme.SidebarTextRequest) layout.Dimensions {
			value := request.Text
			if request.Initial {
				for _, runeValue := range value {
					value = string(runeValue)
					break
				}
			}
			label := material.Label(theme.MaterialOf(active), request.Size, value)
			label.Color = request.Color
			label.Font.Weight = font.Medium
			label.MaxLines = 1
			return label.Layout(gtx)
		},
		Section: func(gtx layout.Context, request theme.SidebarSectionRequest) layout.Dimensions {
			if request.Collapsed {
				return layout.Dimensions{Size: image.Pt(0, gtx.Dp(metrics.SectionHeight))}
			}
			label := material.Label(theme.MaterialOf(active), metrics.SectionTextSize, request.Title)
			label.Color = sidebarMuted
			label.Font.Weight = font.Medium
			label.MaxLines = 1
			return label.Layout(gtx)
		},
	}
}
