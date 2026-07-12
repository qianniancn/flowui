package toast

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/components/icon"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
	"github.com/qianniancn/flowui-icons-lucide"
)

func drawToastSurface(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, surface color.NRGBA) {
	radiusDp := activeTheme.Components.Toast.Radius
	render.DrawShadow(
		gtx,
		rect,
		render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
		render.PopupShadow(activeTheme.Palette.OverlayShadow),
	)
	paint.FillShape(gtx.Ops, surface, clip.UniformRRect(rect, radius).Op(gtx.Ops))
}

func drawToastFocus(gtx layout.Context, rect image.Rectangle, radius int, col color.NRGBA, widthDp unit.Dp, opacity float32) {
	if opacity <= 0 || rect.Empty() {
		return
	}
	width := max(gtx.Dp(widthDp), 1)
	inset := max(width/2, 1)
	focusRect := rect.Inset(-inset)
	focusRadius := min(radius+inset, min(focusRect.Dx(), focusRect.Dy())/2)
	col.A = byte(float32(col.A)*opacity + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(focusRadius, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawToastCloseButton(gtx layout.Context, activeTheme *theme.Theme, size image.Point, style toastStyle, hovered bool) {
	rect := image.Rectangle{Max: size}
	radius := min(size.X, size.Y) / 2
	background := activeTheme.Palette.Overlay
	if hovered {
		background = activeTheme.Palette.SurfaceRaised
	}
	paint.FillShape(gtx.Ops, style.border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	inner := rect.Inset(max(gtx.Dp(unit.Dp(1)), 1))
	paint.FillShape(gtx.Ops, background, clip.UniformRRect(inner, max(radius-1, 0)).Op(gtx.Ops))
}

func drawToastIndicator(gtx layout.Context, size image.Point, foreground color.NRGBA, variant ToastVariant) {
	data := lucide.Info
	switch variant {
	case ToastSuccess:
		data = lucide.CircleCheck
	case ToastWarning:
		data = lucide.TriangleAlert
	case ToastDanger:
		data = lucide.CircleAlert
	}
	iconGtx := gtx
	iconGtx.Constraints = layout.Exact(size)
	icon.Layout(data, iconGtx, foreground)
}
