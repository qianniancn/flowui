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
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawTreeRoot(gtx layout.Context, activeTheme *theme.Theme, tokens theme.TreeTheme, size image.Point, radius int, style treeRootStyle) {
	rect := image.Rectangle{Max: size}
	if rect.Empty() || style.background.A == 0 {
		return
	}
	if style.shadow {
		shapeRadius := tokens.SurfaceRadius
		render.DrawShadow(gtx, rect, render.RoundedShadowCorners(shapeRadius, shapeRadius, shapeRadius, shapeRadius), render.SurfaceShadow(activeTheme.Palette.SurfaceShadow))
	}
	paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func treeRootRadius(gtx layout.Context, tokens theme.TreeTheme, variant Variant, size image.Point) int {
	if variant != VariantSurface {
		return 0
	}
	return min(max(gtx.Dp(tokens.SurfaceRadius), 0), min(size.X, size.Y)/2)
}

func drawTreeRow(gtx layout.Context, tokens theme.TreeTheme, size image.Point, style treeItemStyle, focus float32) {
	rect := image.Rectangle{Max: size}
	radius := min(max(gtx.Dp(tokens.RowRadius), 0), min(size.X, size.Y)/2)
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(tokens.FocusRingWidth), 1)
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

func drawTreeDropIndicator(gtx layout.Context, size image.Point, position DropPosition, depth int, tokens theme.TreeTheme, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	width := max(gtx.Dp(unit.Dp(2)), 1)
	if position == DropInside {
		inset := max(width/2+1, 1)
		rect := image.Rectangle{Max: size}.Inset(inset)
		if rect.Empty() {
			return
		}
		radius := min(max(gtx.Dp(tokens.RowRadius)-inset, 0), min(rect.Dx(), rect.Dy())/2)
		stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, col)
		stroke.Pop()
		return
	}
	y := 0
	if position == DropAfter {
		y = size.Y - width
	}
	x := max(gtx.Dp(tokens.RowPaddingX+unit.Dp(depth)*tokens.Indent), 0)
	rect := image.Rect(min(x, size.X), max(y, 0), size.X, min(max(y+width, 0), size.Y))
	if !rect.Empty() {
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(rect, width/2).Op(gtx.Ops))
	}
}

func drawTreeToggleIcon(gtx layout.Context, tokens theme.TreeTheme, size image.Point, expansion float32, connectors bool, col, background color.NRGBA) {
	if connectors {
		drawTreeConnectorToggle(gtx, tokens, size, expansion, col, background)
		return
	}
	diameter := min(gtx.Dp(tokens.ChevronIconSize), min(size.X, size.Y))
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

func drawTreeConnectorToggle(gtx layout.Context, tokens theme.TreeTheme, size image.Point, expansion float32, foreground, background color.NRGBA) {
	diameterDp := max(tokens.ChevronIconSize, unit.Dp(14))
	diameter := min(gtx.Dp(diameterDp), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	if background.A == 0 {
		background = foreground
		background.A = 32
	}
	position := image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)
	offset := op.Offset(position).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(image.Rect(0, 0, diameter, diameter), max(gtx.Dp(unit.Dp(3)), 1)).Op(gtx.Ops))

	thickness := max(gtx.Dp(unit.Dp(1)), 1)
	horizontal, vertical := treeConnectorToggleBars(diameter, thickness)
	paint.FillShape(gtx.Ops, foreground, clip.Rect(horizontal).Op())
	if expansion < 1 {
		opacity := paint.PushOpacity(gtx.Ops, max(1-expansion, 0))
		paint.FillShape(gtx.Ops, foreground, clip.Rect(vertical).Op())
		opacity.Pop()
	}
	offset.Pop()
}

func treeConnectorToggleBars(size, thickness int) (image.Rectangle, image.Rectangle) {
	margin := max(size/4, 1)
	center := size / 2
	horizontal := image.Rect(margin, center-thickness/2, size-margin, center+(thickness+1)/2)
	vertical := image.Rect(center-thickness/2, margin, center+(thickness+1)/2, size-margin)
	return horizontal, vertical
}

type treeGuideSegment struct {
	from   image.Point
	to     image.Point
	extend bool
}

func drawTreeGuides(gtx layout.Context, entry flatItem, tokens theme.TreeTheme, expanded, connectors bool, guideStyle GuideStyle, col color.NRGBA) {
	rowHeight := gtx.Constraints.Max.Y
	segments := treeGuideSegments(
		entry,
		gtx.Dp(tokens.RowPaddingX),
		gtx.Dp(tokens.Indent),
		gtx.Dp(tokens.ChevronSlotSize),
		gtx.Dp(tokens.ContentGap),
		rowHeight,
		expanded,
		connectors,
	)
	if len(segments) == 0 || col.A == 0 {
		return
	}
	col.A = byte(float32(col.A)*0.5 + 0.5)
	width := max(gtx.Dp(unit.Dp(1)), 1)
	for _, segment := range segments {
		to := segment.to
		if segment.extend && segment.from.X == to.X {
			to.Y += gtx.Dp(tokens.Gap)
		}
		drawTreeGuideSegment(gtx, segment.from, to, width, guideStyle, col)
	}
}

func drawTreeGuideSegment(gtx layout.Context, from, to image.Point, width int, style GuideStyle, col color.NRGBA) {
	if style == GuideSolid {
		var rect image.Rectangle
		switch {
		case from.X == to.X && from.Y < to.Y:
			rect = image.Rect(from.X-width/2, from.Y, from.X+(width+1)/2, to.Y)
		case from.Y == to.Y && from.X < to.X:
			rect = image.Rect(from.X, from.Y-width/2, to.X, from.Y+(width+1)/2)
		default:
			return
		}
		paint.FillShape(gtx.Ops, col, clip.Rect(rect).Op())
		return
	}

	dash := max(gtx.Dp(unit.Dp(1)), 1)
	period := max(gtx.Dp(unit.Dp(3)), dash+1)
	if from.X == to.X && from.Y < to.Y {
		for _, part := range treeGuideDashes(from.Y, to.Y, dash, period) {
			rect := image.Rect(from.X-width/2, part[0], from.X+(width+1)/2, part[1])
			paint.FillShape(gtx.Ops, col, clip.Rect(rect).Op())
		}
	} else if from.Y == to.Y && from.X < to.X {
		for _, part := range treeGuideDashes(from.X, to.X, dash, period) {
			rect := image.Rect(part[0], from.Y-width/2, part[1], from.Y+(width+1)/2)
			paint.FillShape(gtx.Ops, col, clip.Rect(rect).Op())
		}
	}
}

func treeGuideDashes(start, end, dash, period int) [][2]int {
	if start >= end || dash <= 0 || period <= 0 {
		return nil
	}
	parts := make([][2]int, 0, (end-start+period-1)/period)
	for position := start; position < end; position += period {
		parts = append(parts, [2]int{position, min(position+dash, end)})
	}
	return parts
}

func treeGuideSegments(entry flatItem, rowPadding, indent, toggleSlot, contentGap, rowHeight int, expanded, connectors bool) []treeGuideSegment {
	if indent <= 0 || toggleSlot <= 0 || rowHeight <= 0 {
		return nil
	}
	centerY := rowHeight / 2
	guideX := func(depth int) int {
		return rowPadding + depth*indent + toggleSlot/2
	}
	toggleCenterX := func(depth int) int {
		return guideX(depth)
	}
	segments := make([]treeGuideSegment, 0, entry.depth+1)
	for level := 0; level < entry.depth-1 && level+1 < len(entry.ancestorsLast); level++ {
		if !entry.ancestorsLast[level+1] {
			x := guideX(level)
			segments = append(segments, treeGuideSegment{from: image.Pt(x, 0), to: image.Pt(x, rowHeight), extend: true})
		}
	}
	if entry.depth > 0 {
		x := guideX(entry.depth - 1)
		y2 := rowHeight
		if entry.isLast && connectors {
			y2 = centerY
		}
		segments = append(segments, treeGuideSegment{from: image.Pt(x, 0), to: image.Pt(x, y2), extend: !entry.isLast})
		if connectors {
			endX := toggleCenterX(entry.depth)
			if len(entry.item.Children) == 0 {
				endX = rowPadding + entry.depth*indent + toggleSlot + contentGap
			}
			segments = append(segments, treeGuideSegment{from: image.Pt(x, centerY), to: image.Pt(endX, centerY)})
		}
	}
	if expanded && len(entry.item.Children) > 0 {
		x := guideX(entry.depth)
		y1 := min(centerY+toggleSlot/2, rowHeight)
		segments = append(segments, treeGuideSegment{from: image.Pt(x, y1), to: image.Pt(x, rowHeight), extend: true})
	}
	return segments
}
