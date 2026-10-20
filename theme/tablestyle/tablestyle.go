// Package tablestyle contains the visual recipe for table providers.
// Table behavior stays in widget/table; this package owns colors, geometry,
// typography, motion-facing state visuals, and all painting primitives.
package tablestyle

import (
	"image"
	"image/color"
	"math"

	"gioui.org/f32"
	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/theme"
)

// BuildTable returns the complete table visual protocol for the active theme.
func BuildTable(active *theme.Theme, request theme.TableRequest) theme.TableVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Table
	sem := active.Semantic()
	columnSeparator := sem.Separator
	headerSeparator := columnSeparator
	headerSeparator.A = byte(float32(headerSeparator.A)*0.5 + 0.5)
	rowSeparator := lerpColor(sem.Surface, sem.Foreground, 0.19)
	rowSeparator.A = byte(float32(rowSeparator.A)*0.5 + 0.5)
	return theme.TableVisual{
		Metrics: theme.TableMetrics{
			RootPadding: tokens.RootPadding, RootRadius: tokens.RootRadius,
			HeaderRadius: tokens.HeaderRadius, BodyRadius: tokens.BodyRadius,
			HeaderHeight: tokens.HeaderHeight, RowMinHeight: tokens.RowMinHeight,
			EmptyHeight: tokens.EmptyHeight, MaxHeight: tokens.MaxHeight,
			MinColumnWidth: tokens.MinColumnWidth, CellPaddingX: tokens.CellPaddingX,
			CellPaddingY: tokens.CellPaddingY, HeaderTextSize: tokens.HeaderTextSize, HeaderTextWeight: font.Medium,
			CellTextSize: tokens.CellTextSize, SeparatorWidth: tokens.SeparatorWidth,
			ColumnSeparatorHeight: tokens.ColumnSeparatorHeight,
			ColumnResizerHitSize:  tokens.ColumnResizerHitSize,
			ColumnResizerWidth:    tokens.ColumnResizerWidth, ColumnResizeStep: tokens.ColumnResizeStep,
			FocusRingWidth: tokens.FocusRingWidth, FocusRadius: tokens.FocusRadius,
			SelectionColumnWidth: tokens.SelectionColumnWidth, SortIconSize: tokens.SortIconSize,
			SortGap: tokens.SortGap, FooterPaddingX: tokens.FooterPaddingX,
			FooterPaddingY: tokens.FooterPaddingY, LoadMoreHeight: tokens.LoadMoreHeight,
		},
		Colors: theme.TableColors{
			Root: sem.SurfaceTertiary, Header: sem.SurfaceTertiary, Body: sem.Surface,
			Foreground: sem.Foreground, Muted: sem.MutedForeground,
			ColumnSeparator: columnSeparator, HeaderSeparator: headerSeparator,
			RowSeparator: rowSeparator, Border: sem.Border, Focus: sem.Ring,
			Stripe: tokens.StripeBackground, ResizeAccent: sem.Tone(theme.TonePrimary).Fill,
		},
		Motion:   active.Motion,
		Override: theme.TableOverrideStyle(request),
		Row: func(row theme.TableRowRequest) theme.TableRowVisual {
			return rowVisual(active, request.Variant, row)
		},
		Root: func(gtx layout.Context, size image.Point, radius unit.Dp, col color.NRGBA) {
			drawRoot(gtx, size, clampRadius(gtx, size, radius), col)
		},
		Border: func(gtx layout.Context, size image.Point, radius, width unit.Dp, col color.NRGBA) {
			drawBorder(gtx, size, clampRadius(gtx, size, radius), max(gtx.Dp(width), 1), col)
		},
		Header: func(gtx layout.Context, size image.Point, radius unit.Dp, col, separator color.NRGBA) {
			drawHeader(gtx, size, clampRadius(gtx, size, radius), max(gtx.Dp(tokens.SeparatorWidth), 1), col, separator)
		},
		Body: func(gtx layout.Context, size image.Point, radius unit.Dp, col color.NRGBA) {
			drawBody(gtx, size, clampRadius(gtx, size, radius), col)
		},
		RowRender: func(gtx layout.Context, size image.Point, row theme.TableRowVisual, separator color.NRGBA, show bool, focus float32) {
			drawRow(gtx, size, row, max(gtx.Dp(tokens.SeparatorWidth), 1), tokens.FocusRingWidth, tokens.FocusRadius, separator, show, focus)
		},
		HeaderSep: func(gtx layout.Context, x, height int, col color.NRGBA, full bool) {
			drawHeaderSeparator(gtx, x, height, max(gtx.Dp(tokens.SeparatorWidth), 1), max(gtx.Dp(tokens.ColumnSeparatorHeight), 1), col, full)
		},
		RowSep: func(gtx layout.Context, selection int, widths []int, height int, col color.NRGBA) {
			drawRowSeparators(gtx, selection, widths, height, max(gtx.Dp(tokens.SeparatorWidth), 1), col)
		},
		Resizer: func(gtx layout.Context, x, height, baseHeight, activeWidth int, base, accent color.NRGBA, active bool, focus float32) {
			drawResizer(gtx, x, height, baseHeight, activeWidth, base, accent, active, focus)
		},
		CellFocus: func(gtx layout.Context, size image.Point, opacity float32, col color.NRGBA) {
			drawCellFocus(gtx, size, opacity, max(gtx.Dp(tokens.FocusRingWidth), 1), tokens.FocusRadius, col)
		},
		Sort: sortRenderer(tokens.SortIconSize),
		Text: textRenderer(active),
	}
}

func rowVisual(active *theme.Theme, variant theme.TableVariant, request theme.TableRowRequest) theme.TableRowVisual {
	sem := active.Semantic()
	value := theme.TableRowVisual{Background: request.Stripe, Foreground: sem.Foreground, Focus: sem.Ring, Opacity: 1}
	base := sem.Background
	hover := lerpColor(base, sem.Tone(theme.ToneDefault).Fill, .5)
	selected := lerpColor(base, sem.Surface, .1)
	if variant == theme.TablePrimary {
		hover = lerpColor(sem.SurfaceSecondary, sem.Surface, .4)
		selected = lerpColor(sem.SurfaceSecondary, sem.Surface, .1)
	}
	if request.Hovered || request.Pressed {
		value.Background = hover
	}
	if request.Selected {
		value.Background = selected
	}
	if request.Disabled {
		value.Focus = color.NRGBA{}
		value.Opacity = active.DisabledOpacityValue()
	}
	return value
}

func clampRadius(gtx layout.Context, size image.Point, radius unit.Dp) int {
	return min(max(gtx.Dp(radius), 0), min(size.X, size.Y)/2)
}

func drawRoot(gtx layout.Context, size image.Point, radius int, col color.NRGBA) {
	if size.X > 0 && size.Y > 0 && col.A != 0 {
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
	}
}

func drawBorder(gtx layout.Context, size image.Point, radius, width int, col color.NRGBA) {
	if size.X > 0 && size.Y > 0 && width > 0 && col.A != 0 {
		inset := max((width+1)/2, 1)
		draw.RoundedInsetStroke(gtx, image.Rectangle{Max: size}, max(radius-inset, 0), width, inset, col)
	}
}

func drawHeader(gtx layout.Context, size image.Point, radius, separatorWidth int, col, separator color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	clipOp := clip.UniformRRect(image.Rectangle{Max: size}, radius).Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	if separator.A != 0 {
		paint.FillShape(gtx.Ops, separator, clip.Rect(image.Rect(0, max(size.Y-separatorWidth, 0), size.X, size.Y)).Op())
	}
	clipOp.Pop()
}

func drawBody(gtx layout.Context, size image.Point, radius int, col color.NRGBA) {
	if size.X > 0 && size.Y > 0 && col.A != 0 {
		paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, radius).Op(gtx.Ops))
	}
}

func drawRow(gtx layout.Context, size image.Point, row theme.TableRowVisual, separatorWidth int, focusWidth, focusRadius unit.Dp, separator color.NRGBA, show bool, focus float32) {
	if row.Background.A != 0 {
		paint.FillShape(gtx.Ops, row.Background, clip.Rect(image.Rectangle{Max: size}).Op())
	}
	if show && separator.A != 0 && size.Y > 0 {
		paint.FillShape(gtx.Ops, separator, clip.Rect(image.Rect(0, max(size.Y-separatorWidth, 0), size.X, size.Y)).Op())
	}
	if focus <= 0 {
		return
	}
	width := max(gtx.Dp(focusWidth), 1)
	inset := max(width/2+1, 1)
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(focusRadius), 1), min(rect.Dx(), rect.Dy())/2)
	col := row.Focus
	col.A = byte(float32(col.A)*focus + .5)
	draw.RoundedInsetStroke(gtx, image.Rectangle{Max: size}, radius, width, inset, col)
}

func drawHeaderSeparator(gtx layout.Context, x, height, width, lineHeight int, col color.NRGBA, full bool) {
	if full {
		lineHeight = height
	}
	lineHeight = min(lineHeight, height)
	y := max((height-lineHeight)/2, 0)
	paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, y, x+width, y+lineHeight)).Op())
}

func drawRowSeparators(gtx layout.Context, selection int, widths []int, height, width int, col color.NRGBA) {
	if height <= 0 || col.A == 0 {
		return
	}
	x := selection
	if x > 0 && len(widths) > 0 {
		paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, 0, x+width, height)).Op())
	}
	for index, columnWidth := range widths {
		x += columnWidth
		if index < len(widths)-1 {
			paint.FillShape(gtx.Ops, col, clip.Rect(image.Rect(x, 0, x+width, height)).Op())
		}
	}
}

func drawResizer(gtx layout.Context, x, height, baseHeight, activeWidth int, base, accent color.NRGBA, active bool, focus float32) {
	if height <= 0 {
		return
	}
	lineHeight, width, value := height, 1, accent
	if !active && focus <= 0 {
		value, lineHeight = base, min(height, baseHeight)
	} else {
		width = activeWidth
		if !active {
			value.A = byte(float32(value.A)*focus + .5)
		}
	}
	y := max((height-lineHeight)/2, 0)
	paint.FillShape(gtx.Ops, value, clip.Rect(image.Rect(x-width/2, y, x-width/2+width, y+lineHeight)).Op())
}

func drawCellFocus(gtx layout.Context, size image.Point, opacity float32, width int, focusRadius unit.Dp, col color.NRGBA) {
	if opacity <= 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	inset := max(width/2+1, 1)
	rect := image.Rectangle{Max: size}.Inset(inset)
	if rect.Empty() {
		return
	}
	radius := min(max(gtx.Dp(focusRadius), 1), min(rect.Dx(), rect.Dy())/2)
	col.A = byte(float32(col.A)*opacity + .5)
	draw.RoundedInsetStroke(gtx, image.Rectangle{Max: size}, radius, width, inset, col)
}

func sortRenderer(size unit.Dp) theme.TableSortRenderer {
	icon, _ := widget.NewIcon(lucide.ChevronUp)
	return func(gtx layout.Context, bounds image.Point, direction theme.TableSortDirection, col color.NRGBA) {
		diameter := min(gtx.Dp(size), min(bounds.X, bounds.Y))
		if diameter <= 0 || icon == nil {
			return
		}
		center := f32.Pt(float32(bounds.X)/2, float32(bounds.Y)/2)
		angle := float32(0)
		if direction == theme.TableSortDescending {
			angle = float32(math.Pi)
		}
		rotation := op.Affine(f32.AffineId().Rotate(center, angle)).Push(gtx.Ops)
		offset := op.Offset(image.Pt((bounds.X-diameter)/2, (bounds.Y-diameter)/2)).Push(gtx.Ops)
		iconGtx := gtx
		iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
		icon.Layout(iconGtx, col)
		offset.Pop()
		rotation.Pop()
	}
}

func lerpColor(from, to color.NRGBA, progress float32) color.NRGBA {
	if from.A == 0 {
		from.R, from.G, from.B = to.R, to.G, to.B
	}
	if to.A == 0 {
		to.R, to.G, to.B = from.R, from.G, from.B
	}
	return color.NRGBA{
		R: lerpByte(from.R, to.R, progress),
		G: lerpByte(from.G, to.G, progress),
		B: lerpByte(from.B, to.B, progress),
		A: lerpByte(from.A, to.A, progress),
	}
}

func lerpByte(from, to byte, progress float32) byte {
	value := float32(from)*(1-progress) + float32(to)*progress
	return byte(min(max(value+0.5, 0), 255))
}

func textRenderer(active *theme.Theme) theme.TableTextRenderer {
	return func(gtx layout.Context, request theme.TableTextRequest) layout.Dimensions {
		label := material.Label(theme.MaterialOf(active), request.Size, request.Text)
		label.Color = request.Color
		label.Font.Weight = request.Weight
		label.MaxLines = 1
		label.Truncator = "..."
		return label.Layout(gtx)
	}
}
