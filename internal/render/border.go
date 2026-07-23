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
	radius = min(max(radius, 0), min(rect.Dx(), rect.Dy())/2)
	DrawRRectBorder(gtx, clip.UniformRRect(rect, radius), width, col)
}

// DrawRRectBorder paints a border inside a rounded rectangle.
func DrawRRectBorder(gtx layout.Context, shape clip.RRect, width int, col color.NRGBA) {
	rect := shape.Rect
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	width = min(width, min(rect.Dx(), rect.Dy())/2)

	var path clip.Path
	path.Begin(gtx.Ops)
	appendRoundedRectCorners(&path, rect, shape.NW, shape.NE, shape.SE, shape.SW, true)
	inner := rect.Inset(width)
	if !inner.Empty() {
		appendRoundedRectCorners(
			&path,
			inner,
			max(shape.NW-width, 0),
			max(shape.NE-width, 0),
			max(shape.SE-width, 0),
			max(shape.SW-width, 0),
			false,
		)
	}
	paint.FillShape(gtx.Ops, col, clip.Outline{Path: path.End()}.Op())
}

func appendRoundedRect(path *clip.Path, rect image.Rectangle, radius int, clockwise bool) {
	appendRoundedRectCorners(path, rect, radius, radius, radius, radius, clockwise)
}

func appendRoundedRectCorners(path *clip.Path, rect image.Rectangle, nw, ne, se, sw int, clockwise bool) {
	const iq = 1 - 4*(1.4142135623730951-1)/3

	l, t := float32(rect.Min.X), float32(rect.Min.Y)
	r, b := float32(rect.Max.X), float32(rect.Max.Y)
	limit := min(rect.Dx(), rect.Dy()) / 2
	nw = min(max(nw, 0), limit)
	ne = min(max(ne, 0), limit)
	se = min(max(se, 0), limit)
	sw = min(max(sw, 0), limit)
	nwf, nef, sef, swf := float32(nw), float32(ne), float32(se), float32(sw)
	nwi, nei, sei, swi := nwf*iq, nef*iq, sef*iq, swf*iq

	path.MoveTo(f32.Pt(l+nwf, t))
	if clockwise {
		path.LineTo(f32.Pt(r-nef, t))
		path.CubeTo(f32.Pt(r-nei, t), f32.Pt(r, t+nei), f32.Pt(r, t+nef))
		path.LineTo(f32.Pt(r, b-sef))
		path.CubeTo(f32.Pt(r, b-sei), f32.Pt(r-sei, b), f32.Pt(r-sef, b))
		path.LineTo(f32.Pt(l+swf, b))
		path.CubeTo(f32.Pt(l+swi, b), f32.Pt(l, b-swi), f32.Pt(l, b-swf))
		path.LineTo(f32.Pt(l, t+nwf))
		path.CubeTo(f32.Pt(l, t+nwi), f32.Pt(l+nwi, t), f32.Pt(l+nwf, t))
		return
	}
	path.CubeTo(f32.Pt(l+nwi, t), f32.Pt(l, t+nwi), f32.Pt(l, t+nwf))
	path.LineTo(f32.Pt(l, b-swf))
	path.CubeTo(f32.Pt(l, b-swi), f32.Pt(l+swi, b), f32.Pt(l+swf, b))
	path.LineTo(f32.Pt(r-sef, b))
	path.CubeTo(f32.Pt(r-sei, b), f32.Pt(r, b-sei), f32.Pt(r, b-sef))
	path.LineTo(f32.Pt(r, t+nef))
	path.CubeTo(f32.Pt(r, t+nei), f32.Pt(r-nei, t), f32.Pt(r-nef, t))
	path.LineTo(f32.Pt(l+nwf, t))
}
