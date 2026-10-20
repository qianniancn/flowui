package testtheme

import (
	"image/color"

	"gioui.org/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func (behaviorProvider) TitleBar(active *theme.Theme, request theme.TitleBarRequest) theme.TitleBarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.TitleBar
	sem := active.Semantic()
	controlStyle := func(closeAction bool) style.Style {
		control := style.Style{}.Width(tokens.ControlWidth).FillHeight()
		if closeAction {
			return style.Style{}.Part(style.PartItem, control).
				Rules(
					style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem, style.Style{}.
						Background(style.Solid(tokens.CloseHover)).TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnFill)))},
					style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem, style.Style{}.
						Background(style.Solid(tokens.ClosePressed)).TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnFill)))},
				)
		}
		return style.Style{}.Part(style.PartItem, control).
			Rules(
				style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem, style.Style{}.
					Background(style.Solid(sem.SurfaceTertiary)))},
				style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem, style.Style{}.
					Background(style.Solid(theme.ColorOr(tokens.ControlPressed, sem.SurfacePressed))))},
			)
	}
	return theme.TitleBarVisual{
		Defaults: style.Style{}.
			Height(tokens.Height).
			PaddingLeft(tokens.PaddingX).
			Background(style.Solid(sem.SurfaceSecondary)).
			TextColor(style.Solid(sem.SurfaceSecondaryForeground)).
			Part(style.PartPrefix, style.Style{}.PaddingRight(tokens.LeadingGap)).
			Part(style.PartLabel, style.Style{}.FontSize(tokens.TitleTextSize)).
			Part(style.PartIndicator, style.Style{}.Height(tokens.BorderWidth).Background(style.Solid(sem.Separator))),
		Override:          theme.TitleBarOverrideStyle(request),
		Focus:             sem.Ring,
		ControlStyle:      controlStyle(false),
		CloseControlStyle: controlStyle(true),
		Control:           func(layout.Context, theme.TitleBarControlRenderRequest) {},
		ForegroundColor: func(resolved style.ResolvedStyle) (color.NRGBA, bool) {
			if resolved.Text == nil {
				return color.NRGBA{}, false
			}
			value, ok := resolved.Text.Color.(style.SolidColor)
			if !ok {
				return color.NRGBA{}, false
			}
			return value.Color, true
		},
	}
}
