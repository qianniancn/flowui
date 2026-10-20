package testtheme

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Popover(active *theme.Theme, request theme.PopoverRequest) theme.PopoverVisual {
	tokens := active.Components.Popover
	sem := active.Semantic()
	return theme.PopoverVisual{
		Defaults: style.Style{}.
			Background(style.Solid(sem.Overlay)).
			TextColor(style.Solid(sem.OverlayForeground)).
			Padding(tokens.Padding).
			Radius(tokens.Radius),
		Heading: style.Style{}.
			FontSize(tokens.HeadingSize).
			FontWeight(500).
			TextColor(style.Solid(sem.OverlayForeground)),
		Content: style.Style{}.
			FontSize(tokens.BodyTextSize).
			TextColor(style.Solid(sem.MutedForeground)),
		Override: theme.PopoverOverrideStyle(request),
		Metrics:  theme.PopoverMetrics(tokens),
		Motion:   active.Motion,
		Arrow:    func(layout.Context, theme.PopoverArrowRequest) {},
		PanelColor: func(resolved style.ResolvedStyle) color.NRGBA {
			if resolved.Paint == nil {
				return sem.Overlay
			}
			switch value := resolved.Paint.Background.(type) {
			case style.SolidColor:
				return value.Color
			case *style.SolidColor:
				if value != nil {
					return value.Color
				}
			}
			return sem.Overlay
		},
	}
}
