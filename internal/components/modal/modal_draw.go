package modal

import (
	"image"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawModalBackdrop(gtx layout.Context, size image.Point, style modalStyle, progress float32) {
	if style.backdrop.A == 0 || progress <= 0 {
		return
	}
	col := style.backdrop
	col.A = byte(float32(col.A)*progress + 0.5)
	paint.FillShape(gtx.Ops, col, clip.Rect{Max: size}.Op())
}

func drawModalSurface(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int, size ModalSize) {
	if size != ModalFull {
		shadowRadius := theme.Components.Modal.Radius
		render.DrawShadow(gtx, rect, render.RoundedShadowCorners(shadowRadius, shadowRadius, shadowRadius, shadowRadius), render.PopupShadow(theme.Palette.OverlayShadowColor()))
	}
	paint.FillShape(gtx.Ops, theme.Palette.OverlayColor(), clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func modalDialogRadius(gtx layout.Context, theme *theme.Theme, size ModalSize, dialogSize image.Point) int {
	if size == ModalFull {
		return 0
	}
	radius := gtx.Dp(theme.Components.Modal.Radius)
	return min(max(radius, 0), min(dialogSize.X, dialogSize.Y)/2)
}

func drawModalIconFrame(gtx layout.Context, theme *theme.Theme, size image.Point) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	paint.FillShape(gtx.Ops, theme.Palette.AccentSoft, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}
