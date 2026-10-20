package testtheme

import (
	"time"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) Tooltip(active *theme.Theme, request theme.TooltipRequest) theme.TooltipVisual {
	tokens := active.Components.Tooltip
	sem := active.Semantic()
	return theme.TooltipVisual{
		Defaults: style.Style{}.
			Background(style.Solid(sem.Overlay)).
			TextColor(style.Solid(sem.OverlayForeground)).
			FontSize(tokens.TextSize).
			Padding(tokens.Padding).
			Radius(tokens.Radius).
			BorderColor(style.Solid(sem.Border)).
			BorderWidth(tokens.BorderWidth),
		Content: style.Style{}.
			FontSize(tokens.TextSize).
			TextColor(style.Solid(sem.OverlayForeground)),
		Override: theme.TooltipOverrideStyle(request),
		Metrics: theme.TooltipMetrics{
			Offset: tokens.Offset, ArrowOffset: tokens.ArrowOffset,
			Padding: tokens.Padding, Radius: tokens.Radius,
			BorderWidth: tokens.BorderWidth, MaxWidth: tokens.MaxWidth,
			ArrowSize: tokens.ArrowSize, TextSize: tokens.TextSize,
			AnimationScale: tokens.AnimationScale, ExitScale: tokens.ExitScale,
			AnimationDistance: tokens.AnimationDistance,
			Delay:             tokens.Delay, CloseDelay: tokens.CloseDelay,
			EnterDuration: 150 * time.Millisecond,
			ExitDuration:  100 * time.Millisecond,
			MoveDuration:  400 * time.Millisecond,
		},
		Motion: active.Motion,
		Arrow:  func(layout.Context, theme.TooltipArrowRequest) {},
		PanelColor: func(resolved style.ResolvedStyle) theme.TooltipPanelColors {
			result := theme.TooltipPanelColors{Surface: sem.Overlay, Border: sem.Border}
			if resolved.Paint == nil {
				return result
			}
			if value, ok := resolved.Paint.Background.(style.SolidColor); ok {
				result.Surface = value.Color
			}
			if resolved.Paint.Border != nil {
				if value, ok := resolved.Paint.Border.Color.(style.SolidColor); ok {
					result.Border = value.Color
				}
			}
			return result
		},
	}
}
