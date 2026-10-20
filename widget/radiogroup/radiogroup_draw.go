package radiogroup

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	layoutui "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

func drawRadio(ctx *core.Context, gtx layout.Context, visual theme.RadioGroupVisual, style radioStyle, scale float32) layout.Dimensions {
	controlSize := radioIndicatorSize(gtx, style)
	focusSpace := max(gtx.Dp(visual.Metrics.FocusSpace), 1)
	controlSize.X = min(controlSize.X, max(gtx.Constraints.Max.X-focusSpace*2, 0))
	controlSize.Y = min(controlSize.Y, max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := controlSize.Add(image.Pt(focusSpace*2, focusSpace*2))
	dims := gtx.Constraints.Constrain(bounds)
	if controlSize.X <= 0 || controlSize.Y <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-controlSize.X)/2, (dims.Y-controlSize.Y)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(controlSize),
	}
	center := f32.Pt(
		float32(rect.Min.X)+float32(controlSize.X)/2,
		float32(rect.Min.Y)+float32(controlSize.Y)/2,
	)
	stack := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
	if visual.IndicatorColor == nil {
		panic(theme.ErrIncompleteVisualProvider)
	}
	foreground, ok := visual.IndicatorColor(style.indicatorOn)
	if !ok {
		panic(theme.ErrIncompleteVisualProvider)
	}
	dot := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return visual.IndicatorRenderer(gtx, theme.RadioIndicatorRenderRequest{
			Size:            gtx.Constraints.Min,
			Foreground:      foreground,
			Selection:       style.selected,
			Pressed:         style.pressed,
			DotScale:        visual.Metrics.DotScale,
			DotPressedScale: visual.Metrics.DotPressedScale,
		})
	})
	layoutRadioLayer(ctx, gtx, rect, style.indicatorOff, 1-style.selected, style.focus, nil)
	layoutRadioLayer(ctx, gtx, rect, style.indicatorOn, style.selected, style.focus, dot)
	stack.Pop()

	return layout.Dimensions{Size: dims}
}

func radioIndicatorSize(gtx layout.Context, resolved radioStyle) image.Point {
	var size image.Point
	for _, endpoint := range [...]style.ResolvedStyle{resolved.indicatorOff, resolved.indicatorOn} {
		if endpoint.Box == nil {
			panic(theme.ErrIncompleteVisualProvider)
		}
		width := endpoint.Box.Width
		if width == nil {
			width = endpoint.Box.MinWidth
		}
		height := endpoint.Box.Height
		if height == nil {
			height = endpoint.Box.MinHeight
		}
		if width == nil || height == nil || *width <= 0 || *height <= 0 {
			panic(theme.ErrIncompleteVisualProvider)
		}
		size.X = max(size.X, gtx.Dp(*width))
		size.Y = max(size.Y, gtx.Dp(*height))
	}
	if size.X <= 0 || size.Y <= 0 {
		panic(theme.ErrIncompleteVisualProvider)
	}
	return size
}

func layoutRadioLayer(ctx *core.Context, gtx layout.Context, rect image.Rectangle, resolved style.ResolvedStyle, opacity, focus float32, child core.Widget) {
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.Resolved(ctx, layerGtx, computed.ApplyOutlineOpacity(resolved, focus), child)
	fade.Pop()
	offset.Pop()
}
