package closebutton

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/frame"
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
			drawCloseIcon(gtx, iconSize, style.foreground)
			return layout.Dimensions{Size: iconSize}
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

func drawCloseIcon(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	transform := newCloseIconTransform(size)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(transform.point(3.46967, 3.46967))
	drawCloseIconArc(&path, transform, 4, 4, 225, 90)
	path.LineTo(transform.point(8, 6.93934))
	path.LineTo(transform.point(11.46967, 3.46967))
	drawCloseIconArc(&path, transform, 12, 4, 225, 180)
	path.LineTo(transform.point(9.06066, 8))
	path.LineTo(transform.point(12.53033, 11.46967))
	drawCloseIconArc(&path, transform, 12, 12, 315, 180)
	path.LineTo(transform.point(8, 9.06066))
	path.LineTo(transform.point(4.53033, 12.53033))
	drawCloseIconArc(&path, transform, 4, 12, 45, 180)
	path.LineTo(transform.point(6.93934, 8))
	path.LineTo(transform.point(3.46967, 4.53033))
	drawCloseIconArc(&path, transform, 4, 4, 135, 90)
	path.Close()
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

type closeIconTransform struct {
	scale   float32
	offsetX float32
	offsetY float32
}

func newCloseIconTransform(size image.Point) closeIconTransform {
	scale := float32(min(size.X, size.Y)) / 16
	return closeIconTransform{
		scale:   scale,
		offsetX: (float32(size.X) - 16*scale) / 2,
		offsetY: (float32(size.Y) - 16*scale) / 2,
	}
}

func (t closeIconTransform) point(x, y float64) f32.Point {
	return f32.Pt(t.offsetX+float32(x)*t.scale, t.offsetY+float32(y)*t.scale)
}

func drawCloseIconArc(path *clip.Path, transform closeIconTransform, centerX, centerY, startDegrees, sweepDegrees float64) {
	segments := int(math.Ceil(math.Abs(sweepDegrees) / 90))
	step := sweepDegrees / float64(segments) * math.Pi / 180
	angle := startDegrees * math.Pi / 180
	for range segments {
		next := angle + step
		alpha := 4.0 / 3.0 * math.Tan((next-angle)/4)
		startX := centerX + 0.75*math.Cos(angle)
		startY := centerY + 0.75*math.Sin(angle)
		endX := centerX + 0.75*math.Cos(next)
		endY := centerY + 0.75*math.Sin(next)
		control1X := startX - alpha*0.75*math.Sin(angle)
		control1Y := startY + alpha*0.75*math.Cos(angle)
		control2X := endX + alpha*0.75*math.Sin(next)
		control2Y := endY - alpha*0.75*math.Cos(next)
		path.CubeTo(
			transform.point(control1X, control1Y),
			transform.point(control2X, control2Y),
			transform.point(endX, endY),
		)
		angle = next
	}
}
