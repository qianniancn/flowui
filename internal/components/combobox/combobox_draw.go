package combobox

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/flowui/internal/components/icon"
	"github.com/qianniancn/flowui/internal/components/optionrow"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawComboBoxPanel(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int) {
	render.DrawSurface(gtx, rect, radius, theme.Palette.OverlayColor(), render.ThemeShadow(theme.Shadows.Overlay, theme.Palette.OverlayShadowColor(), 1))
}

func drawComboBoxItem(gtx layout.Context, theme *theme.Theme, size image.Point, style comboBoxItemStyle) {
	optionrow.DrawBackground(gtx, size, theme.Components.ComboBox.ItemRadius, style.bg)
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
	optionrow.DrawCheck(gtx, size, progress, theme.Components.ComboBox.ItemCheckInset, theme.Components.ComboBox.ItemCheckStroke, theme.Palette.Foreground)
}
