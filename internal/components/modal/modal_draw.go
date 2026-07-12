package modal

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
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

func drawModalCloseButton(gtx layout.Context, theme *theme.Theme, size image.Point, hovered, pressed, focused bool) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	bg := color.NRGBA{}
	switch {
	case pressed:
		bg = theme.Palette.SurfacePressed
	case hovered:
		bg = theme.Palette.SurfaceRaised
	}
	if bg.A != 0 {
		paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if focused {
		drawModalCloseFocus(gtx, theme, rect, radius)
	}
	diameter := min(gtx.Dp(theme.Components.Modal.CloseIconSize), min(size.X, size.Y))
	if diameter <= 0 {
		diameter = min(size.X, size.Y) * 3 / 4
	}
	offset := op.Offset(image.Pt((size.X-diameter)/2, (size.Y-diameter)/2)).Push(gtx.Ops)
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(image.Pt(diameter, diameter))
	icon.Layout(lucide.X, iconGtx, theme.Palette.MutedForeground)
	offset.Pop()
}

func drawModalCloseFocus(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, radius int) {
	width := max(gtx.Dp(theme.Components.Button.FocusRingWidth), 1)
	inset := width + 1
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, theme.Palette.Focus)
	stroke.Pop()
}
