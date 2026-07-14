package combobox

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawComboBoxPanel(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int) {
	render.DrawSurface(gtx, rect, radius, theme.Palette.OverlayColor(), render.PopupShadow(theme.Palette.OverlayShadowColor()))
}

func drawComboBoxItem(gtx layout.Context, theme *theme.Theme, size image.Point, style comboBoxItemStyle) {
	if style.bg.A == 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(theme.Components.ComboBox.ItemRadius), 1), min(size.X, size.Y)/2)
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawComboBoxChevron(gtx layout.Context, theme *theme.Theme, size image.Point, progress float32, col color.NRGBA) {
	col.A = byte(float32(col.A) * 0.9)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	angle := progress * float32(math.Pi)
	stack := op.Affine(f32.AffineId().Rotate(center, angle)).Push(gtx.Ops)
	diameter := min(icon.LucideSizeForStroke(gtx, theme.Components.ComboBox.ChevronStroke), min(size.X, size.Y))
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.ChevronDown, iconGtx, col)
	offset.Pop()
	stack.Pop()
}

func drawComboBoxCheck(gtx layout.Context, theme *theme.Theme, size image.Point, progress float32) {
	progress = min(max(progress, 0), 1)
	if progress == 0 {
		return
	}
	rect := image.Rectangle{Max: size}.Inset(max(gtx.Dp(theme.Components.ComboBox.ItemCheckInset), 0))
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
	path := render.CheckPath(gtx.Ops, points, progress)
	col := theme.Palette.Foreground
	col.A = byte(float32(col.A)*progress + 0.5)
	stroke := clip.Stroke{
		Path:  path,
		Width: max(render.DpFloat(gtx, theme.Components.ComboBox.ItemCheckStroke), 1),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}
