package input

import (
	"image"
	"image/color"
	"math"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawInputFrame(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style inputStyle, ringWidthDp float32) {
	tokens := activeTheme.Components.Input
	drawFieldFrame(gtx, rect, radius, tokens.Radius, fieldShadow(activeTheme, tokens.ShadowColor, style.ShadowOpacity, tokens.ShadowStrength), style, ringWidthDp)
}

func drawTextAreaFrame(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style inputStyle, ringWidthDp float32) {
	tokens := activeTheme.Components.TextArea
	drawFieldFrame(gtx, rect, radius, tokens.Radius, fieldShadow(activeTheme, tokens.ShadowColor, style.ShadowOpacity, tokens.ShadowStrength), style, ringWidthDp)
}

func drawFieldFrame(gtx layout.Context, rect image.Rectangle, radius int, radiusDp unit.Dp, shadow render.BoxShadow, style inputStyle, ringWidthDp float32) {
	if rect.Empty() {
		return
	}
	if style.ShadowOpacity > 0 {
		render.DrawShadow(
			gtx,
			rect,
			render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
			shadow,
		)
	}
	drawInputRing(gtx, rect, radius, style.Ring, ringWidthDp)
	paint.FillShape(gtx.Ops, style.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func fieldShadow(activeTheme *theme.Theme, shadowColor color.NRGBA, opacity, strength float32) render.BoxShadow {
	shadowColor = theme.ColorOr(shadowColor, activeTheme.Palette.Shadow)
	shadow := render.ThemeShadow(activeTheme.Shadows.Control, shadowColor, opacity)
	if !(strength > 0) || math.IsInf(float64(strength), 0) {
		return render.BoxShadow{Blur: -1}
	}
	for index := range shadow.Layers {
		alpha := min(float32(shadow.Layers[index].Color.A)*strength, 255)
		shadow.Layers[index].Color.A = uint8(alpha + .5)
	}
	return shadow
}

func drawInputRing(gtx layout.Context, rect image.Rectangle, radius int, ring color.NRGBA, widthDp float32) {
	if ring.A == 0 || widthDp <= 0 {
		return
	}
	width := max(gtx.Dp(unit.Dp(widthDp)), 1)
	outer := rect.Inset(-width)
	outerRadius := min(radius+width, min(outer.Dx(), outer.Dy())/2)
	paint.FillShape(gtx.Ops, ring, clip.UniformRRect(outer, max(outerRadius, 0)).Op(gtx.Ops))
}
