package flowui

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func drawInputFrame(gtx layout.Context, rect image.Rectangle, radius int, style inputStyle) {
	if style.shadowOpacity > 0 {
		drawInputShadow(gtx, rect, radius, style.shadowOpacity)
	}
	if style.border.A != 0 && style.borderWidth != 0 {
		drawInputRing(gtx, rect, radius, style)
		return
	}
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawInputShadow(gtx layout.Context, rect image.Rectangle, radius int, opacity float32) {
	drawInputShadowLayer(gtx, rect, radius, image.Pt(0, gtx.Dp(unit.Dp(2))), gtx.Dp(unit.Dp(2)), 0x0a, opacity)
	drawInputShadowLayer(gtx, rect, radius, image.Pt(0, gtx.Dp(unit.Dp(1))), gtx.Dp(unit.Dp(1)), 0x0f, opacity)
}

func drawInputShadowLayer(gtx layout.Context, rect image.Rectangle, radius int, offset image.Point, spread int, alpha byte, opacity float32) {
	col := color.NRGBA{A: inputAlpha(alpha, opacity)}
	if col.A == 0 {
		return
	}
	shadowRect := rect.Add(offset).Inset(-spread)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(shadowRect, radius+spread).Op(gtx.Ops))
}

func drawInputRing(gtx layout.Context, rect image.Rectangle, radius int, style inputStyle) {
	width := max(gtx.Dp(style.borderWidth), 1)
	paint.FillShape(gtx.Ops, style.border, clip.UniformRRect(rect, radius).Op(gtx.Ops))

	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, style.bg, clip.UniformRRect(inner, max(radius-width, 0)).Op(gtx.Ops))
}

func inputAlpha(alpha byte, opacity float32) byte {
	return byte(float32(alpha)*opacity + 0.5)
}
