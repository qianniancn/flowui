package selects

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawSelectIndicator(gtx layout.Context, theme *theme.Theme, size image.Point, progress float32, col color.NRGBA) {
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	diameter := max(
		gtx.Dp(theme.Components.Select.IndicatorSize),
		icon.LucideSizeForStroke(gtx, theme.Components.Select.IndicatorStroke),
	)
	diameter = min(diameter, min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	stack := op.Affine(f32.AffineId().Rotate(center, progress*float32(math.Pi))).Push(gtx.Ops)
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.ChevronDown, iconGtx, col)
	offset.Pop()
	stack.Pop()
}
