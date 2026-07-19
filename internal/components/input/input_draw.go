package input

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawInputFrame(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style inputStyle, ringWidthDp float32) {
	drawFieldFrame(gtx, activeTheme, rect, radius, activeTheme.Components.Input.Radius, style, ringWidthDp)
}

func drawTextAreaFrame(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style inputStyle, ringWidthDp float32) {
	drawFieldFrame(gtx, activeTheme, rect, radius, activeTheme.Components.TextArea.Radius, style, ringWidthDp)
}

func drawFieldFrame(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, radiusDp unit.Dp, style inputStyle, ringWidthDp float32) {
	if rect.Empty() {
		return
	}
	if style.ShadowOpacity > 0 {
		render.DrawShadow(
			gtx,
			rect,
			render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
			render.ThemeShadow(activeTheme.Shadows.Control, activeTheme.Palette.Shadow, style.ShadowOpacity),
		)
	}
	drawInputRing(gtx, rect, radius, style.Ring, ringWidthDp)
	paint.FillShape(gtx.Ops, style.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
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
