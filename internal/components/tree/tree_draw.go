package tree

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

func drawTreeRoot(gtx layout.Context, activeTheme *theme.Theme, size image.Point, radius int, style treeRootStyle) {
	rect := image.Rectangle{Max: size}
	if rect.Empty() || style.background.A == 0 {
		return
	}
	if style.shadow {
		shapeRadius := activeTheme.Components.Tree.SurfaceRadius
		render.DrawShadow(gtx, rect, render.RoundedShadowCorners(shapeRadius, shapeRadius, shapeRadius, shapeRadius), render.SurfaceShadow(activeTheme.Palette.SurfaceShadow))
	}
	paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func treeRootRadius(gtx layout.Context, activeTheme *theme.Theme, variant Variant, size image.Point) int {
	if variant != VariantSurface {
		return 0
	}
	return min(max(gtx.Dp(activeTheme.Components.Tree.SurfaceRadius), 0), min(size.X, size.Y)/2)
}

func drawTreeRow(gtx layout.Context, activeTheme *theme.Theme, size image.Point, style treeItemStyle, focus float32) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(activeTheme.Components.Tree.RowRadius), 0), min(size.X, size.Y)/2)
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Tree.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	col := style.focus
	col.A = byte(float32(col.A)*focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawTreeChevron(gtx layout.Context, activeTheme *theme.Theme, size image.Point, expansion float32, col color.NRGBA) {
	diameter := min(gtx.Dp(activeTheme.Components.Tree.ChevronIconSize), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	position := image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	rotation := op.Affine(f32.AffineId().Rotate(center, expansion*float32(math.Pi/2))).Push(gtx.Ops)
	offset := op.Offset(position).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.ChevronRight, iconGtx, col)
	offset.Pop()
	rotation.Pop()
}
