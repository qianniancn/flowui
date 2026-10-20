package testtheme

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Menubar(active *theme.Theme, request theme.MenubarRequest) theme.MenubarVisual {
	tokens := active.Components.Menubar
	if request.Compact {
		tokens.TriggerHeight = minDp(tokens.TriggerHeight, 28)
		tokens.TriggerPaddingX = minDp(tokens.TriggerPaddingX, 8)
		tokens.TriggerRadius = minDp(tokens.TriggerRadius, 4)
		tokens.TriggerTextSize = minSp(tokens.TriggerTextSize, 13)
	}
	return theme.MenubarVisual{
		Metrics: theme.MenubarMetrics{
			TriggerHeight:          tokens.TriggerHeight,
			TriggerPaddingX:        tokens.TriggerPaddingX,
			TriggerRadius:          tokens.TriggerRadius,
			TriggerFocusRingWidth:  tokens.TriggerFocusRingWidth,
			TriggerFocusRingOffset: tokens.TriggerFocusRingOffset,
			Gap:                    tokens.Gap,
			PanelGap:               tokens.PanelGap,
		},
		Motion:   active.Motion,
		Override: theme.MenubarOverrideStyle(request),
		Trigger: func(theme.MenubarTriggerRequest) theme.MenubarTriggerVisual {
			return theme.MenubarTriggerVisual{
				Foreground: foreground,
				Focus:      focus,
				Opacity:    1,
			}
		},
		RenderTrigger: func(layout.Context, image.Point, theme.MenubarMetrics, theme.MenubarTriggerVisual, float32) {},
		RenderLabel: func(gtx layout.Context, request theme.MenubarLabelRequest) layout.Dimensions {
			label := material.Label(theme.MaterialOf(active), tokens.TriggerTextSize, request.Text)
			label.Color = request.Foreground
			label.SelectionColor = active.Semantic().Selection
			label.Font.Weight = font.Normal
			return label.Layout(gtx)
		},
	}
}
