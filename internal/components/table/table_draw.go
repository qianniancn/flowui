package table

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
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawTableRoot(gtx layout.Context, size image.Point, radius int, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func drawTableBorder(gtx layout.Context, size image.Point, radius, width int, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || width <= 0 || col.A == 0 {
		return
	}
	inset := max((width+1)/2, 1)
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawTableHeader(gtx layout.Context, activeTheme *theme.Theme, size image.Point, radius int, col, separator color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	headerClip := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	if separator.A != 0 {
		height := max(gtx.Dp(activeTheme.Components.Table.SeparatorWidth), 1)
		paint.FillShape(gtx.Ops, separator, clip.Rect(image.Rect(0, max(size.Y-height, 0), size.X, size.Y)).Op())
	}
	headerClip.Pop()
}

func drawTableBody(gtx layout.Context, size image.Point, radius int, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
}

func drawTableRow(gtx layout.Context, activeTheme *theme.Theme, size image.Point, style tableRowStyle, separator color.NRGBA, showSeparator bool, focus float32) {
	rect := image.Rectangle{Max: size}
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.Rect(rect).Op())
	}
	if showSeparator && separator.A != 0 && size.Y > 0 {
		height := max(gtx.Dp(activeTheme.Components.Table.SeparatorWidth), 1)
		paint.FillShape(gtx.Ops, separator, clip.Rect(image.Rect(0, max(size.Y-height, 0), size.X, size.Y)).Op())
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Table.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(activeTheme.Components.Table.FocusRadius), 1), min(focusRect.Dx(), focusRect.Dy())/2)
	col := style.focus
	col.A = byte(float32(col.A)*focus + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(focusRect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawTableHeaderSeparator(gtx layout.Context, activeTheme *theme.Theme, x, height int, col color.NRGBA, full bool) {
	lineHeight := min(gtx.Dp(activeTheme.Components.Table.ColumnSeparatorHeight), height)
	if full {
		lineHeight = height
	}
	width := max(gtx.Dp(activeTheme.Components.Table.SeparatorWidth), 1)
	y := max((height-lineHeight)/2, 0)
	paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, y, x+width, y+lineHeight)).Op())
}

func drawTableRowSeparators(gtx layout.Context, activeTheme *theme.Theme, columns tableColumns, height int, col color.NRGBA) {
	if height <= 0 || col.A == 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Table.SeparatorWidth), 1)
	x := columns.selection
	if x > 0 && len(columns.widths) > 0 {
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, 0, x+width, height)).Op())
	}
	for index, columnWidth := range columns.widths {
		x += columnWidth
		if index < len(columns.widths)-1 {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, 0, x+width, height)).Op())
		}
	}
}

func drawTableColumnResizer(gtx layout.Context, x, height, baseHeight, activeWidth int, base, accent color.NRGBA, active bool, focus float32) {
	if height <= 0 {
		return
	}
	width := 1
	lineHeight := height
	colorValue := accent
	if !active && focus <= 0 {
		colorValue = base
		lineHeight = min(height, baseHeight)
	} else {
		width = activeWidth
		if !active {
			colorValue.A = byte(float32(colorValue.A)*focus + 0.5)
		}
	}
	y := max((height-lineHeight)/2, 0)
	left := x - width/2
	paint.FillShape(gtx.Ops, colorValue, clip.Rect(image.Rect(left, y, left+width, y+lineHeight)).Op())
}

func drawTableCellFocus(gtx layout.Context, activeTheme *theme.Theme, size image.Point, opacity float32, col color.NRGBA) {
	if opacity <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Table.FocusRingWidth), 1)
	rect := image.Rectangle{Max: size}.Inset(max(width/2+1, 1))
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(activeTheme.Components.Table.FocusRadius), 1), min(rect.Dx(), rect.Dy())/2)
	col.A = byte(float32(col.A)*opacity + 0.5)
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawTableSortIndicator(gtx layout.Context, activeTheme *theme.Theme, size image.Point, direction SortDirection, col color.NRGBA) {
	diameter := min(gtx.Dp(activeTheme.Components.Table.SortIconSize), min(size.X, size.Y))
	if diameter <= 0 {
		return
	}
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	angle := float32(0)
	if direction == SortDescending {
		angle = float32(math.Pi)
	}
	rotation := op.Affine(f32.AffineId().Rotate(center, angle)).Push(gtx.Ops)
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.ChevronUp, iconGtx, col)
	offset.Pop()
	rotation.Pop()
}

func tableRootRadius(gtx layout.Context, activeTheme *theme.Theme, size image.Point, variant Variant, unified bool) int {
	radius := activeTheme.Components.Table.RootRadius
	if variant == VariantSecondary {
		if !unified {
			return 0
		}
		radius = activeTheme.Components.Table.HeaderRadius
	}
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func tableHeaderRadius(gtx layout.Context, activeTheme *theme.Theme, size image.Point, variant Variant, unified bool) int {
	if variant != VariantSecondary || unified {
		return 0
	}
	return min(max(gtx.Dp(activeTheme.Components.Table.HeaderRadius), 0), min(size.X, size.Y)/2)
}

func tableBodyRadius(gtx layout.Context, activeTheme *theme.Theme, size image.Point, variant Variant) int {
	if variant != VariantPrimary {
		return 0
	}
	return min(max(gtx.Dp(activeTheme.Components.Table.BodyRadius), 0), min(size.X, size.Y)/2)
}
