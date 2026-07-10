package field

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func DrawFrame(gtx layout.Context, rect image.Rectangle, radius int, style Style) {
	if style.ShadowOpacity > 0 {
		drawShadow(gtx, rect, radius, style.ShadowOpacity)
	}
	if style.Border.A != 0 && style.BorderWidth != 0 {
		drawRing(gtx, rect, radius, style)
		return
	}
	paint.FillShape(gtx.Ops, style.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawShadow(gtx layout.Context, rect image.Rectangle, radius int, opacity float32) {
	drawShadowLayer(gtx, rect, radius, image.Pt(0, gtx.Dp(unit.Dp(2))), gtx.Dp(unit.Dp(2)), 0x0a, opacity)
	drawShadowLayer(gtx, rect, radius, image.Pt(0, gtx.Dp(unit.Dp(1))), gtx.Dp(unit.Dp(1)), 0x0f, opacity)
}

func drawShadowLayer(gtx layout.Context, rect image.Rectangle, radius int, offset image.Point, spread int, alpha byte, opacity float32) {
	col := color.NRGBA{A: alphaValue(alpha, opacity)}
	if col.A == 0 {
		return
	}
	shadowRect := rect.Add(offset).Inset(-spread)
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(shadowRect, radius+spread).Op(gtx.Ops))
}

func drawRing(gtx layout.Context, rect image.Rectangle, radius int, style Style) {
	width := max(gtx.Dp(style.BorderWidth), 1)
	paint.FillShape(gtx.Ops, style.Border, clip.UniformRRect(rect, radius).Op(gtx.Ops))

	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, style.Background, clip.UniformRRect(inner, max(radius-width, 0)).Op(gtx.Ops))
}

func alphaValue(alpha byte, opacity float32) byte {
	return byte(float32(alpha)*opacity + 0.5)
}
