package toast

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/internal/components/icon"
	"github.com/qianniancn/flowui/internal/render"
	"github.com/qianniancn/flowui/internal/theme"
)

func drawToastSurface(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, surface color.NRGBA) {
	radiusDp := activeTheme.Components.Toast.Radius
	render.DrawShadow(
		gtx,
		rect,
		render.RoundedShadowCorners(radiusDp, radiusDp, radiusDp, radiusDp),
		render.ThemeShadow(activeTheme.Shadows.Overlay, activeTheme.Palette.OverlayShadow, 1),
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
