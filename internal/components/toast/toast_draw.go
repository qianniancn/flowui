package toast

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawToastSurface(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, surface color.NRGBA) {
	radiusDp := activeTheme.Components.Toast.Radius
	render.DrawShadow(
		gtx,
		rect,
		render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
		render.PopupShadow(activeTheme.Palette.OverlayShadow),
	)
	paint.FillShape(gtx.Ops, surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawToastFocus(gtx layout.Context, rect image.Rectangle, radius int, col color.NRGBA, widthDp unit.Dp, opacity float32) {
	if opacity <= 0 || rect.Empty() {
		return
	}
	width := max(gtx.Dp(widthDp), 1)
	inset := max(width/2, 1)
	focusRect := rect.Inset(-inset)
	focusRadius := min(radius+inset, min(focusRect.Dx(), focusRect.Dy())/2)
	col.A = byte(float32(col.A)*opacity + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(focusRadius, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawToastCloseButton(gtx layout.Context, activeTheme *theme.Theme, size image.Point, style toastStyle, hovered bool) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	background := activeTheme.Palette.Overlay
	if hovered {
		background = activeTheme.Palette.SurfaceRaised
	}
	paint.FillShape(gtx.Ops, style.border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	inner := rect.Inset(max(gtx.Dp(unit.Dp(1)), 1))
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(inner, max(radius-1, 0)).Op(gtx.Ops))
}

func drawToastIndicator(gtx layout.Context, size image.Point, background, foreground color.NRGBA, variant ToastVariant) {
	switch variant {
	case ToastSuccess:
		drawToastCircleIndicator(gtx, size, background, foreground, true, false)
	case ToastWarning:
		drawToastWarningIndicator(gtx, size, background, foreground)
	case ToastDanger:
		drawToastCircleIndicator(gtx, size, background, foreground, false, true)
	default:
		drawToastCircleIndicator(gtx, size, background, foreground, false, false)
	}
}

func drawToastCircleIndicator(gtx layout.Context, size image.Point, background, foreground color.NRGBA, success, danger bool) {
	rect := image.Rectangle{Max: size}.Inset(1)
	paint.FillShape(gtx.Ops, foreground, clip.Ellipse(rect).Op(gtx.Ops))
	inner := rect.Inset(max(gtx.Dp(unit.Dp(1.5)), 1))
	paint.FillShape(gtx.Ops, background, clip.Ellipse(inner).Op(gtx.Ops))
	center := image.Pt(size.X/2, size.Y/2)
	if success {
		var path clip.Path
		path.Begin(gtx.Ops)
		path.MoveTo(f32.Pt(float32(center.X)-3.2, float32(center.Y)))
		path.LineTo(f32.Pt(float32(center.X)-0.8, float32(center.Y)+2.2))
		path.LineTo(f32.Pt(float32(center.X)+3.7, float32(center.Y)-3))
		stroke := clip.Stroke{Path: path.End(), Width: float32(max(gtx.Dp(unit.Dp(1.5)), 1))}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, foreground)
		stroke.Pop()
		return
	}
	lineHeight := 4
	lineY := center.Y
	dotY := center.Y - 3
	if danger {
		lineHeight = 3
		lineY = center.Y - 3
		dotY = center.Y + 3
	}
	paint.FillShape(gtx.Ops, foreground, clip.UniformRRect(image.Rect(center.X-1, lineY, center.X+1, lineY+lineHeight), 1).Op(gtx.Ops))
	paint.FillShape(gtx.Ops, foreground, clip.Ellipse(image.Rect(center.X-1, dotY-1, center.X+1, dotY+1)).Op(gtx.Ops))
}

func drawToastWarningIndicator(gtx layout.Context, size image.Point, background, foreground color.NRGBA) {
	outer := toastTrianglePath(gtx, f32.Pt(float32(size.X)/2, 1), f32.Pt(1, float32(size.Y)-2), f32.Pt(float32(size.X)-1, float32(size.Y)-2))
	paint.FillShape(gtx.Ops, foreground, outer)
	inner := toastTrianglePath(gtx, f32.Pt(float32(size.X)/2, 4), f32.Pt(4, float32(size.Y)-4), f32.Pt(float32(size.X)-4, float32(size.Y)-4))
	paint.FillShape(gtx.Ops, background, inner)
	centerX := size.X / 2
	paint.FillShape(gtx.Ops, foreground, clip.UniformRRect(image.Rect(centerX-1, 6, centerX+1, 10), 1).Op(gtx.Ops))
	paint.FillShape(gtx.Ops, foreground, clip.Ellipse(image.Rect(centerX-1, 11, centerX+1, 13)).Op(gtx.Ops))
}

func toastTrianglePath(gtx layout.Context, top, left, right f32.Point) clip.Op {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(top)
	path.LineTo(right)
	path.LineTo(left)
	path.Close()
	return clip.Outline{Path: path.End()}.Op()
}
