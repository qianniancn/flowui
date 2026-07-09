package flowui

import (
	"image"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func drawCheckbox(gtx layout.Context, theme *Theme, style checkboxStyle) layout.Dimensions {
	size := gtx.Dp(theme.Components.Checkbox.Size)
	focusSpace := max(gtx.Dp(theme.Components.Checkbox.FocusSpace), 1)
	maxSize := min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y) - focusSpace*2
	size = min(size, max(maxSize, 0))
	bounds := image.Pt(size+focusSpace*2, size+focusSpace*2)
	dims := gtx.Constraints.Constrain(bounds)
	if size <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-size)/2, (dims.Y-size)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(image.Pt(size, size)),
	}
	radius := min(max(gtx.Dp(theme.Shape.CheckboxRadius), 1), size/2)

	drawCheckboxFocus(gtx, rect, radius, style)
	drawCheckboxFrame(gtx, theme, rect, radius, style)
	drawCheckboxFill(gtx, rect, radius, style)
	drawCheckboxCheck(gtx, theme, rect, style)

	return layout.Dimensions{Size: dims}
}

func drawCheckboxFrame(gtx layout.Context, theme *Theme, rect image.Rectangle, radius int, style checkboxStyle) {
	border := style.border
	border.A = byte(float32(border.A)*(1-style.selected) + 0.5)
	if border.A == 0 {
		paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
		return
	}

	width := max(gtx.Dp(theme.Components.Checkbox.BorderWidth), 1)
	paint.FillShape(gtx.Ops, border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(inner, max(radius-width, 0)).Op(gtx.Ops))
}

func drawCheckboxFocus(gtx layout.Context, rect image.Rectangle, radius int, style checkboxStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(unit.Dp(2)), 1)
	focusRect := rect.Inset(-max(width/2, 1))
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, radius+width).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawCheckboxFill(gtx layout.Context, rect image.Rectangle, radius int, style checkboxStyle) {
	if style.selected == 0 {
		return
	}
	scale := 0.7 + 0.3*style.selected
	size := rect.Size()
	center := f32.Pt(
		float32(rect.Min.X)+float32(size.X)/2,
		float32(rect.Min.Y)+float32(size.Y)/2,
	)
	stack := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
	col := style.accent
	col.A = byte(float32(col.A)*style.selected + 0.5)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	stack.Pop()
}

func drawCheckboxCheck(gtx layout.Context, theme *Theme, rect image.Rectangle, style checkboxStyle) {
	if style.selected == 0 {
		return
	}
	width := float32(rect.Dx())
	height := float32(rect.Dy())
	x := float32(rect.Min.X)
	y := float32(rect.Min.Y)
	points := [3]f32.Point{
		f32.Pt(x+width*0.27, y+height*0.52),
		f32.Pt(x+width*0.43, y+height*0.68),
		f32.Pt(x+width*0.75, y+height*0.31),
	}
	path := checkboxCheckPath(gtx.Ops, points, style.selected)
	col := style.accentFg
	col.A = byte(float32(col.A)*style.selected + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(dpFloat(gtx, theme.Components.Checkbox.CheckStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func checkboxCheckPath(ops *op.Ops, points [3]f32.Point, progress float32) clip.PathSpec {
	progress = min(max(progress, 0), 1)
	first := points[0]
	second := points[1]
	third := points[2]
	firstLen := checkboxDistance(first, second)
	secondLen := checkboxDistance(second, third)
	total := firstLen + secondLen
	drawLen := total * progress

	var path clip.Path
	path.Begin(ops)
	path.MoveTo(first)
	if drawLen <= firstLen {
		path.LineTo(checkboxPointOnLine(first, second, drawLen/firstLen))
		return path.End()
	}
	path.LineTo(second)
	path.LineTo(checkboxPointOnLine(second, third, (drawLen-firstLen)/secondLen))
	return path.End()
}

func checkboxPointOnLine(from, to f32.Point, progress float32) f32.Point {
	return f32.Pt(
		lerp(from.X, to.X, progress),
		lerp(from.Y, to.Y, progress),
	)
}

func checkboxDistance(from, to f32.Point) float32 {
	dx := to.X - from.X
	dy := to.Y - from.Y
	return float32(math.Hypot(float64(dx), float64(dy)))
}
