package render

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

// DrawRoundedBorder paints a rounded border inside rect without stroked-path joins.
func DrawRoundedBorder(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	radius = min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)
	width = min(width, min(rect.Dx(), rect.Dy())/2)

	var path clip.Path
	path.Begin(gtx.Ops)
	appendRoundedRect(&path, rect, radius, true)
	inner := rect.Inset(width)
	if !inner.Empty() {
		appendRoundedRect(&path, inner, max(radius-width, 0), false)
	}
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

func appendRoundedRect(path *clip.Path, rect image.Rectangle, radius int, clockwise bool) {
	const iq = 1 - 4*(1.4142135623730951-1)/3

	l, t := float32(rect.Min.X), float32(rect.Min.Y)
	r, b := float32(rect.Max.X), float32(rect.Max.Y)
	rad := float32(min(max(radius, 0), min(rect.Dx(), rect.Dy())/2))
	i := rad * iq

	path.MoveTo(f32.Pt(l+rad, t))
	if clockwise {
		path.LineTo(f32.Pt(r-rad, t))
		path.CubeTo(f32.Pt(r-i, t), f32.Pt(r, t+i), f32.Pt(r, t+rad))
		path.LineTo(f32.Pt(r, b-rad))
		path.CubeTo(f32.Pt(r, b-i), f32.Pt(r-i, b), f32.Pt(r-rad, b))
		path.LineTo(f32.Pt(l+rad, b))
		path.CubeTo(f32.Pt(l+i, b), f32.Pt(l, b-i), f32.Pt(l, b-rad))
		path.LineTo(f32.Pt(l, t+rad))
		path.CubeTo(f32.Pt(l, t+i), f32.Pt(l+i, t), f32.Pt(l+rad, t))
		return
	}
	path.CubeTo(f32.Pt(l+i, t), f32.Pt(l, t+i), f32.Pt(l, t+rad))
	path.LineTo(f32.Pt(l, b-rad))
	path.CubeTo(f32.Pt(l, b-i), f32.Pt(l+i, b), f32.Pt(l+rad, b))
	path.LineTo(f32.Pt(r-rad, b))
	path.CubeTo(f32.Pt(r-i, b), f32.Pt(r, b-i), f32.Pt(r, b-rad))
	path.LineTo(f32.Pt(r, t+rad))
	path.CubeTo(f32.Pt(r, t+i), f32.Pt(r-i, t), f32.Pt(r-rad, t))
	path.LineTo(f32.Pt(l+rad, t))
}
