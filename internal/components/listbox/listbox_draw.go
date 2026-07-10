package listbox

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawListBoxItem(gtx layout.Context, theme *theme.Theme, size image.Point, style listBoxItemStyle) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Components.ListBox.ItemRadius), 1), min(size.X, size.Y)/2)
	if style.bg.A != 0 {
		paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	drawListBoxFocus(gtx, theme, rect, radius, style)
}

func drawListBoxFocus(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int, style listBoxItemStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(theme.Components.ListBox.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawListBoxIndicator(gtx layout.Context, theme *theme.Theme, size image.Point, style listBoxItemStyle) {
	if style.selected == 0 {
		return
	}
	rect := image.Rectangle{Max: size}.Inset(gtx.Dp(theme.Components.ListBox.ItemIndicatorInset))
	if rect.Empty() {
		return
	}
	x := float32(rect.Min.X)
	y := float32(rect.Min.Y)
	width := float32(rect.Dx())
	height := float32(rect.Dy())
	points := [3]f32.Point{
		f32.Pt(x+width*0.05, y+height*0.56),
		f32.Pt(x+width*0.40, y+height*0.86),
		f32.Pt(x+width*0.95, y+height*0.14),
	}
	path := render.CheckPath(gtx.Ops, points, style.selected)
	col := style.indicator
	col.A = byte(float32(col.A)*style.selected + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(render.DpFloat(gtx, theme.Components.ListBox.ItemIndicatorStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}
