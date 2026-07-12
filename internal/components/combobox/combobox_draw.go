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

func drawComboBoxCheck(gtx layout.Context, theme *theme.Theme, size image.Point) {
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(lucide.Check, iconGtx, theme.Palette.Accent)
}
