package modal

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawModalIconFrame(gtx layout.Context, theme *theme.Theme, size image.Point) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	paint.FillShape(gtx.Ops, theme.Palette.AccentSoft, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}
