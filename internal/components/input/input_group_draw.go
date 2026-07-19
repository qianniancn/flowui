package input

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawInputGroupFrame(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style inputGroupStyle, ringWidth float32) {
	if rect.Empty() {
		return
	}
	if style.ShadowOpacity > 0 {
		radiusDp := activeTheme.Components.InputGroup.Radius
		render.DrawShadow(
			gtx,
			rect,
			render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
			render.ThemeShadow(activeTheme.Shadows.Control, activeTheme.Palette.Shadow, style.ShadowOpacity),
		)
	}
	drawInputRing(gtx, rect, radius, style.Ring, ringWidth)
	paint.FillShape(gtx.Ops, style.Background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawInputGroupDivider(gtx layout.Context, x, height, width int, style inputGroupStyle) {
	if width <= 0 || height <= 0 {
		return
	}
	paint.FillShape(gtx.Ops, style.Divider, clip.Rect(image.Rect(x, 0, x+width, height)).Op())
}
