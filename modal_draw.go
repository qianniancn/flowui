package flowui

import (
	"image"
	"image/color"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func drawModalBackdrop(gtx layout.Context, size image.Point, style modalStyle, progress float32) {
	if style.backdrop.A == 0 || progress <= 0 {
		return
	}
	col := style.backdrop
	col.A = byte(float32(col.A)*progress + 0.5)
	paint.FillShape(gtx.Ops, col, clip.Rect{Max: size}.Op())
}

func drawModalSurface(gtx layout.Context, theme *Theme, rect image.Rectangle, radius int, size ModalSize) {
	if size != ModalFull {
		shadowRadius := theme.Components.Modal.Radius
		DrawShadow(gtx, rect, RoundedShadowCorners(shadowRadius, shadowRadius, shadowRadius, shadowRadius), PopupShadow(theme.Palette.Shadow))
	}
	paint.FillShape(gtx.Ops, theme.Palette.Surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func modalDialogRadius(gtx layout.Context, theme *Theme, size ModalSize, dialogSize image.Point) int {
	if size == ModalFull {
		return 0
	}
	radius := gtx.Dp(theme.Components.Modal.Radius)
	return min(max(radius, 0), min(dialogSize.X, dialogSize.Y)/2)
}

func drawModalIconFrame(gtx layout.Context, theme *Theme, size image.Point) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	paint.FillShape(gtx.Ops, theme.Palette.AccentSoft, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawModalCloseButton(gtx layout.Context, theme *Theme, size image.Point, hovered, pressed, focused bool) {
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
	drawModalCloseX(gtx, theme.Palette.MutedForeground, size)
}

func drawModalCloseFocus(gtx layout.Context, theme *Theme, rect image.Rectangle, radius int) {
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

func drawModalCloseX(gtx layout.Context, col color.NRGBA, size image.Point) {
	strokeWidth := float32(max(gtx.Dp(unit.Dp(1.8)), 1))
	center := f32.Pt(float32(size.X)/2, float32(size.Y)/2)
	half := float32(min(size.X, size.Y)) * 0.18

	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(f32.Pt(center.X-half, center.Y-half))
	path.LineTo(f32.Pt(center.X+half, center.Y+half))
	path.MoveTo(f32.Pt(center.X+half, center.Y-half))
	path.LineTo(f32.Pt(center.X-half, center.Y+half))
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: strokeWidth,
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}
