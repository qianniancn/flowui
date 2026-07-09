package flowui

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

func drawComboBoxPanel(gtx layout.Context, theme *Theme, rect image.Rectangle, radius int) {
	drawPopupSurface(gtx, theme, rect, radius)
}

func drawComboBoxItem(gtx layout.Context, theme *Theme, size image.Point, style comboBoxItemStyle) {
	if style.bg.A == 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Components.ComboBox.ItemRadius), 1), min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawComboBoxChevron(gtx layout.Context, theme *Theme, size image.Point, progress float32, col color.NRGBA) {
	col.A = byte(float32(col.A) * 0.9)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	angle := progress * float32(math.Pi)
	stack := op.Affine(f32.AffineId().Rotate(center, angle)).Push(gtx.Ops)
	width := dpFloat(gtx, theme.Components.ComboBox.ChevronStroke)
	if width < 1 {
		width = 1
	}
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X-5, center.Y-2))
	path.LineTo(f32.Pt(center.X, center.Y+3))
	path.LineTo(f32.Pt(center.X+5, center.Y-2))
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: width,
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
	stack.Pop()
}

func drawComboBoxCheck(gtx layout.Context, theme *Theme, size image.Point) {
	rect := image.Rectangle{Max: size}.Inset(gtx.Dp(theme.Components.ComboBox.ItemCheckInset))
	points := [3]f32.Point{
		f32.Pt(float32(rect.Min.X), float32(rect.Min.Y+rect.Dy()/2)),
		f32.Pt(float32(rect.Min.X+rect.Dx()/3), float32(rect.Max.Y)),
		f32.Pt(float32(rect.Max.X), float32(rect.Min.Y)),
	}
	path := checkboxCheckPath(gtx.Ops, points, 1)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(dpFloat(gtx, theme.Components.ComboBox.ItemCheckStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.Palette.Accent)
	stroke.Pop()
}
