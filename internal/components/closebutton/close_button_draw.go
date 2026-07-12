package closebutton

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/frame"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawCloseButton(gtx layout.Context, size image.Point, style closeButtonStyle) {
	rect := image.Rectangle{Max: size}
	radius := closeButtonRadius(gtx, size, style.radius)
	paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func closeButtonRadius(gtx layout.Context, size image.Point, radius unit.Dp) int {
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func drawCloseButtonFocus(gtx layout.Context, rect image.Rectangle, radius int, style closeButtonStyle) {
	if style.focusOpacity <= 0 {
		return
	}
	width := max(gtx.Dp(style.focusWidth), 1)
	focusRect, focusRadius := closeButtonFocusGeometry(rect, radius, width)
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focusOpacity + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, focusRadius).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func closeButtonFocusGeometry(rect image.Rectangle, radius, width int) (image.Rectangle, int) {
	inset := max(width/2, 1)
	focusRect := rect.Inset(-inset)
	focusRadius := min(radius+inset, min(focusRect.Dx(), focusRect.Dy())/2)
	return focusRect, max(focusRadius, 0)
}

func (b CloseButtonWidget) layoutIcon(ctx *frame.Context, gtx layout.Context, buttonSize image.Point, style closeButtonStyle, disabled bool) {
	padding := max(gtx.Dp(style.padding), 0)
	available := max(min(buttonSize.X, buttonSize.Y)-padding*2, 0)
	diameter := min(max(gtx.Dp(style.iconSize), 0), available)
	iconSize := image.Pt(diameter, diameter)

	layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints = layout.Exact(iconSize)
		if b.icon == nil {
			return icon.Layout(lucide.X, gtx, style.foreground)
		}
		if disabled {
			gtx = gtx.Disabled()
		}
		restore := frame.PushColors(ctx, style.foreground, style.background)
		defer restore()
		b.icon.Layout(ctx, gtx)
		return layout.Dimensions{Size: iconSize}
	})
}
