// Package popoverstyle contains the theme-owned visual recipe for popover
// behavior. It does not import widget/popover: the widget supplies placement
// facts and executes the returned declarations and renderer.
package popoverstyle

import (
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildPopover returns the complete popover visual protocol for the active
// theme. All values are copied so a provider call cannot retain mutable theme
// state beyond the current visual plan.
func BuildPopover(active *theme.Theme, request theme.PopoverRequest) theme.PopoverVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Popover
	sem := active.Semantic()
	return theme.PopoverVisual{
		Defaults: style.Style{}.
			Background(style.Solid(sem.Overlay)).
			TextColor(style.Solid(sem.OverlayForeground)).
			Padding(tokens.Padding).
			Radius(tokens.Radius).
			Shadow(style.ShadowOverlay).
			Overflow(style.OverflowHidden),
		Heading: style.Style{}.
			FontSize(tokens.HeadingSize).
			FontWeight(500).
			TextColor(style.Solid(sem.OverlayForeground)),
		Content: style.Style{}.
			FontSize(tokens.BodyTextSize).
			TextColor(style.Solid(sem.MutedForeground)),
		Override: theme.PopoverOverrideStyle(request),
		Metrics: theme.PopoverMetrics{
			Offset: tokens.Offset, Padding: tokens.Padding, Radius: tokens.Radius,
			MaxWidth: tokens.MaxWidth, ArrowWidth: tokens.ArrowWidth,
			ArrowHeight: tokens.ArrowHeight, HeadingSize: tokens.HeadingSize,
			BodyTextSize: tokens.BodyTextSize, SectionGap: tokens.SectionGap,
			AnimationScale:    animationScale(tokens.AnimationScale),
			EnterDuration:     tokens.EnterDuration,
			ExitDuration:      tokens.ExitDuration,
			AnimationDistance: tokens.AnimationDistance,
		},
		Motion: active.Motion,
		Arrow:  arrowRenderer,
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

func animationScale(value float32) float32 {
	if value <= 0 || value > 1 {
		return .90
	}
	return value
}

func arrowRenderer(gtx layout.Context, request theme.PopoverArrowRequest) {
	if request.Arrow.X <= 0 || request.Arrow.Y <= 0 || request.Color.A == 0 {
		return
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	switch request.Side {
	case theme.PopoverSideTop:
		x := float32(request.Panel.X) / 2
		y := float32(request.Panel.Y)
		path.MoveTo(f32.Pt(x-float32(request.Arrow.X)/2, y))
		path.LineTo(f32.Pt(x+float32(request.Arrow.X)/2, y))
		path.LineTo(f32.Pt(x, y+float32(request.Arrow.Y)))
	case theme.PopoverSideLeft:
		x := float32(request.Panel.X)
		y := float32(request.Panel.Y) / 2
		path.MoveTo(f32.Pt(x, y-float32(request.Arrow.X)/2))
		path.LineTo(f32.Pt(x, y+float32(request.Arrow.X)/2))
		path.LineTo(f32.Pt(x+float32(request.Arrow.Y), y))
	case theme.PopoverSideRight:
		y := float32(request.Panel.Y) / 2
		path.MoveTo(f32.Pt(0, y-float32(request.Arrow.X)/2))
		path.LineTo(f32.Pt(0, y+float32(request.Arrow.X)/2))
		path.LineTo(f32.Pt(-float32(request.Arrow.Y), y))
	default:
		x := float32(request.Panel.X) / 2
		path.MoveTo(f32.Pt(x-float32(request.Arrow.X)/2, 0))
		path.LineTo(f32.Pt(x+float32(request.Arrow.X)/2, 0))
		path.LineTo(f32.Pt(x, -float32(request.Arrow.Y)))
	}
	paint.FillShape(gtx.Ops, request.Color, clip.Outline{Path: path.End()}.Op())
}
