package alert

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawAlertSurface(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, background color.NRGBA) {
	if rect.Empty() {
		return
	}
	radiusDp := activeTheme.Components.Alert.Radius
	render.DrawShadow(
		gtx,
		rect,
		render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
		render.ThemeShadow(activeTheme.Shadows.Surface, activeTheme.Palette.SurfaceShadow, 1),
	)
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}
