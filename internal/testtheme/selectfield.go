package testtheme

import (
	"image"
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Select(active *theme.Theme, request theme.SelectRequest) theme.SelectVisual {
	content := style.Style{}.
		Background(style.Solid(background)).
		TextColor(style.Solid(foreground)).
		Radius(4).
		MinHeight(40).
		PaddingX(8).
		PaddingY(8).
		FontSize(14)
	defaults := style.Style{}.
		Part(theme.SelectPartTrigger, content).
		Part(theme.SelectPartPlaceholder, style.Style{}.
			TextColor(style.Solid(muted)).
			FontSize(14)).
		Part(theme.SelectPartSelection, style.Style{}.
			Background(style.Solid(accent))).
		Part(theme.SelectPartIndicator, style.Style{}.
			Width(16).
			Height(16).
			TextColor(style.Solid(muted))).
		Part(theme.SelectPartIcon, style.Style{}.
			Width(16).
			Height(16).
			TextColor(style.Solid(muted))).
		Part(theme.SelectPartPanel, style.Style{}.
			Background(style.Solid(background)).
			TextColor(style.Solid(foreground)).
			Radius(4).
			Padding(8)).
		Part(theme.SelectPartDescription, style.Style{}.
			TextColor(style.Solid(muted)).
			FontSize(12))
	return theme.SelectVisual{
		Defaults: defaults,
		Override: theme.SelectOverrideStyle(request),
		Metrics: theme.SelectMetrics{
			ContentGap:        8,
			PanelGap:          4,
			PanelMaxHeight:    240,
			AnimationScale:    .96,
			AnimationDistance: 4,
			OpenDuration:      150 * time.Millisecond,
			CloseDuration:     100 * time.Millisecond,
			IndicatorDuration: 150 * time.Millisecond,
		},
		Motion: active.Motion,
		Indicator: func(layout.Context, image.Point, float32, style.ResolvedStyle) layout.Dimensions {
			return layout.Dimensions{Size: image.Pt(16, 16)}
		},
		Text:           behaviorSelectText(active),
		IndicatorWidth: behaviorPartWidth,
		IconSize:       behaviorPartSize,
	}
}

func behaviorSelectText(active *theme.Theme) theme.SelectTextRenderer {
	return func(gtx layout.Context, request theme.SelectTextRequest) layout.Dimensions {
		label := material.Label(theme.MaterialOf(active), 14, request.Text)
		label.Color = foreground
		return label.Layout(gtx)
	}
}
