// Package titlebarstyle contains the theme-owned visual recipe for title bars.
// Title-bar behavior keeps window lifecycle and slot ordering; this package
// owns chrome declarations and client-side control drawing.
package titlebarstyle

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

type controlMetrics struct {
	controlWidth   unit.Dp
	iconSize       unit.Dp
	iconStroke     unit.Dp
	focusRingWidth unit.Dp
}

// BuildTitleBar builds a complete provider-owned title-bar visual plan.
func BuildTitleBar(active *theme.Theme, request theme.TitleBarRequest) theme.TitleBarVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.TitleBar
	sem := active.Semantic()
	metrics := controlMetrics{
		controlWidth:   tokens.ControlWidth,
		iconSize:       tokens.IconSize,
		iconStroke:     tokens.IconStrokeWidth,
		focusRingWidth: tokens.FocusRingWidth,
	}
	controlStyle := ControlDeclaration(active, false)
	closeControlStyle := ControlDeclaration(active, true)
	defaults := style.Style{}.
		Height(tokens.Height).
		PaddingLeft(tokens.PaddingX).
		Background(style.Solid(sem.SurfaceSecondary)).
		TextColor(style.Solid(sem.SurfaceSecondaryForeground)).
		Part(style.PartPrefix, style.Style{}.PaddingRight(tokens.LeadingGap)).
		Part(style.PartLabel, style.Style{}.FontSize(tokens.TitleTextSize)).
		Part(style.PartIndicator, style.Style{}.
			Height(tokens.BorderWidth).
			Background(style.Solid(sem.Separator)))
	return theme.TitleBarVisual{
		Defaults:          defaults,
		Override:          theme.TitleBarOverrideStyle(request),
		Focus:             sem.Ring,
		ControlStyle:      controlStyle,
		CloseControlStyle: closeControlStyle,
		Control: func(gtx layout.Context, control theme.TitleBarControlRenderRequest) {
			renderControl(gtx, control, metrics)
		},
		ForegroundColor: func(resolved style.ResolvedStyle) (color.NRGBA, bool) {
			if resolved.Text == nil {
				return color.NRGBA{}, false
			}
			return computed.Color(resolved.Text.Color)
		},
	}
}

// ControlDeclaration returns one provider-owned window-control declaration.
func ControlDeclaration(active *theme.Theme, closeAction bool) style.Style {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.TitleBar
	control := style.Style{}.
		Width(tokens.ControlWidth).
		FillHeight()
	sem := active.Semantic()
	if closeAction {
		return style.Style{}.Part(style.PartItem, control).
			Rules(
				style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem, style.Style{}.
					Background(style.Solid(tokens.CloseHover)).
					TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnFill)))},
				style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem, style.Style{}.
					Background(style.Solid(tokens.ClosePressed)).
					TextColor(style.Solid(sem.Tone(theme.TonePrimary).OnFill)))},
			)
	}
	return style.Style{}.Part(style.PartItem, control).
		Rules(
			style.Rule{When: style.Hovered, Set: style.Style{}.Part(style.PartItem,
				style.Style{}.Background(style.Solid(sem.SurfaceTertiary)))},
			style.Rule{When: style.Pressed, Set: style.Style{}.Part(style.PartItem,
				style.Style{}.Background(style.Solid(theme.ColorOr(tokens.ControlPressed, sem.SurfacePressed))))},
		)
}

func renderControl(gtx layout.Context, request theme.TitleBarControlRenderRequest, metrics controlMetrics) {
	drawControlGlyph(gtx, request.Size, request.Glyph, request.Foreground, metrics)
	if !request.FocusVisible || request.Focus.A == 0 {
		return
	}
	width := max(gtx.Dp(metrics.focusRingWidth), 1)
	ring := image.Rectangle{Max: request.Size}.Inset(width)
	if ring.Empty() {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(ring, 0).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, request.Focus)
	stroke.Pop()
}

func drawControlGlyph(gtx layout.Context, size image.Point, glyph theme.TitleBarControlGlyph, foreground color.NRGBA, metrics controlMetrics) {
	if glyph == theme.TitleBarGlyphNone || foreground.A == 0 || size.X <= 0 || size.Y <= 0 {
		return
	}
	glyphSize := min(max(gtx.Dp(metrics.iconSize), 1), min(size.X, size.Y))
	stroke := controlGlyphStroke(gtx.Metric, metrics.iconStroke, glyphSize)
	left := (float32(size.X)-float32(glyphSize))/2 + stroke
	top := (float32(size.Y)-float32(glyphSize))/2 + stroke
	inner := controlGlyphBounds{left: left, top: top, right: left + float32(glyphSize) - stroke*2, bottom: top + float32(glyphSize) - stroke*2}
	if inner.empty() {
		return
	}
	switch glyph {
	case theme.TitleBarGlyphMinimize:
		y := (inner.top + inner.bottom) / 2
		drawControlLine(gtx.Ops, f32.Pt(inner.left, y), f32.Pt(inner.right, y), stroke, foreground)
	case theme.TitleBarGlyphMaximize:
		drawControlOutline(gtx.Ops, inner, stroke, foreground)
	case theme.TitleBarGlyphRestore:
		drawRestoreGlyph(gtx.Ops, inner, stroke, foreground)
	case theme.TitleBarGlyphClose:
		drawControlLine(gtx.Ops, f32.Pt(inner.left, inner.top), f32.Pt(inner.right, inner.bottom), stroke, foreground)
		drawControlLine(gtx.Ops, f32.Pt(inner.right, inner.top), f32.Pt(inner.left, inner.bottom), stroke, foreground)
	}
}

type controlGlyphBounds struct{ left, top, right, bottom float32 }

func (b controlGlyphBounds) empty() bool { return b.right <= b.left || b.bottom <= b.top }

func controlGlyphStroke(metric unit.Metric, width unit.Dp, glyphSize int) float32 {
	if width <= 0 {
		width = 1
	}
	scale := metric.PxPerDp
	if scale <= 0 {
		scale = 1
	}
	stroke := max(float32(width)*scale, 1)
	return min(stroke, max(float32(glyphSize)/3, 1))
}

func drawRestoreGlyph(ops *op.Ops, bounds controlGlyphBounds, stroke float32, foreground color.NRGBA) {
	offset := stroke * 2
	side := min(bounds.right-bounds.left, bounds.bottom-bounds.top) - offset
	if side <= stroke*2 {
		drawControlOutline(ops, bounds, stroke, foreground)
		return
	}
	back := controlGlyphBounds{left: bounds.left + offset, top: bounds.top, right: bounds.left + offset + side, bottom: bounds.top + side}
	front := controlGlyphBounds{left: bounds.left, top: bounds.top + offset, right: bounds.left + side, bottom: bounds.top + offset + side}
	drawControlLine(ops, f32.Pt(back.left, back.top), f32.Pt(back.right, back.top), stroke, foreground)
	drawControlLine(ops, f32.Pt(back.right, back.top), f32.Pt(back.right, back.bottom), stroke, foreground)
	drawControlLine(ops, f32.Pt(back.left, back.top), f32.Pt(back.left, front.top), stroke, foreground)
	drawControlLine(ops, f32.Pt(front.right, back.bottom), f32.Pt(back.right, back.bottom), stroke, foreground)
	drawControlOutline(ops, front, stroke, foreground)
}

func drawControlOutline(ops *op.Ops, bounds controlGlyphBounds, stroke float32, foreground color.NRGBA) {
	if bounds.empty() || stroke <= 0 {
		return
	}
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(f32.Pt(bounds.left, bounds.top))
	path.LineTo(f32.Pt(bounds.right, bounds.top))
	path.LineTo(f32.Pt(bounds.right, bounds.bottom))
	path.LineTo(f32.Pt(bounds.left, bounds.bottom))
	path.Close()
	paint.FillShape(ops, foreground, clip.Stroke{Path: path.End(), Width: stroke}.Op())
}

func drawControlLine(ops *op.Ops, from, to f32.Point, stroke float32, foreground color.NRGBA) {
	var path clip.Path
	path.Begin(ops)
	path.MoveTo(from)
	path.LineTo(to)
	paint.FillShape(ops, foreground, clip.Stroke{Path: path.End(), Width: stroke}.Op())
}
