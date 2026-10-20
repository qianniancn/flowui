package testtheme

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Collapsible(active *theme.Theme, request theme.CollapsibleRequest) theme.CollapsibleVisual {
	tokens := active.Components.Collapsible
	return theme.CollapsibleVisual{
		Label: style.Style{}.
			TextColor(style.Solid(foreground)).
			FontSize(active.Typography.ControlSize).
			FontWeight(int(font.Medium)),
		Trigger: theme.CollapsibleTriggerOverrideStyle(request),
		Metrics: theme.CollapsibleMetrics{
			BodyPadding:       tokens.BodyPadding,
			ContentGap:        tokens.ContentGap,
			ContentDuration:   tokens.ContentDuration,
			IndicatorDuration: tokens.IndicatorDuration,
		},
		Motion: active.Motion,
		Indicator: func(gtx layout.Context, _ theme.CollapsibleIndicatorRequest) layout.Dimensions {
			size := gtx.Dp(tokens.IndicatorSize)
			if size <= 0 {
				return layout.Dimensions{}
			}
			return layout.Dimensions{Size: image.Pt(size, size)}
		},
		Override: theme.CollapsibleOverrideStyle(request),
	}
}
