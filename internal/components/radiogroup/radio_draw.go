package radiogroup

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawRadio(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, style radioStyle, scale float32) layout.Dimensions {
	controlSize := radioIndicatorSize(gtx, style, activeTheme.Components.RadioGroup.Size)
	focusSpace := max(gtx.Dp(activeTheme.Components.RadioGroup.FocusSpace), 1)
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
	dot := frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		drawRadioDot(ctx, gtx, activeTheme, image.Rectangle{Max: gtx.Constraints.Min}, style)
		return layout.Dimensions{Size: gtx.Constraints.Min}
	})
	layoutRadioLayer(ctx, gtx, rect, style.indicatorOff, 1-style.selected, style.focus, nil)
	layoutRadioLayer(ctx, gtx, rect, style.indicatorOn, style.selected, style.focus, dot)
	stack.Pop()

	return layout.Dimensions{Size: dims}
}

func radioIndicatorSize(gtx layout.Context, style radioStyle, fallbackDp unit.Dp) image.Point {
	var size image.Point
	var hasWidth, hasHeight bool
	for _, endpoint := range [...]flowstyle.ResolvedStyle{style.indicatorOff, style.indicatorOn} {
		if endpoint.Box == nil {
			continue
		}
		if endpoint.Box.Width != nil {
			size.X = max(size.X, gtx.Dp(*endpoint.Box.Width))
			hasWidth = true
		}
		if endpoint.Box.Height != nil {
			size.Y = max(size.Y, gtx.Dp(*endpoint.Box.Height))
			hasHeight = true
		}
	}
	if !hasWidth {
		size.X = gtx.Dp(fallbackDp)
	}
	if !hasHeight {
		size.Y = gtx.Dp(fallbackDp)
	}
	return size
}

func layoutRadioLayer(ctx *frame.Context, gtx layout.Context, rect image.Rectangle, style flowstyle.ResolvedStyle, opacity, focus float32, child frame.Widget) {
	layerGtx := gtx
	layerGtx.Constraints = layout.Exact(rect.Size())
	offset := op.Offset(rect.Min).Push(gtx.Ops)
	fade := paint.PushOpacity(gtx.Ops, opacity)
	layoutui.LayoutResolved(ctx, layerGtx, styleruntime.ApplyOutlineOpacity(style, focus), child)
	fade.Pop()
	offset.Pop()
}

func drawRadioDot(ctx *frame.Context, gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, style radioStyle) {
	if style.selected == 0 {
		return
	}
	targetScale := activeTheme.Components.RadioGroup.DotScale
	if style.pressed {
		targetScale = activeTheme.Components.RadioGroup.DotPressedScale
	}
	size := max(int(float32(min(rect.Dx(), rect.Dy()))*targetScale*style.selected+0.5), 1)
	center := rect.Min.Add(rect.Size().Div(2))
	dot := image.Rectangle{
		Min: center.Sub(image.Pt(size/2, size/2)),
		Max: center.Sub(image.Pt(size/2, size/2)).Add(image.Pt(size, size)),
	}
	paint.FillShape(gtx.Ops, ctx.ForegroundColor(), clip.Ellipse(dot).Op(gtx.Ops))
}
