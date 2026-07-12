package menu

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawMenuPanel(gtx layout.Context, activeTheme *theme.Theme, rect image.Rectangle, radius int, style panelStyle) {
	if rect.Empty() {
		return
	}
	render.DrawSurface(gtx, rect, radius, style.background, heroMenuShadow(style.shadow, style.shadowOpacity))
	width := max(gtx.Dp(activeTheme.Components.Menu.BorderWidth), 0)
	if width <= 0 || style.border.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, style.border, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(inner, max(radius-width, 0)).Op(gtx.Ops))
}

func heroMenuShadow(col color.NRGBA, opacity float32) render.BoxShadow {
	if opacity <= 0 || col.A == 0 {
		return render.BoxShadow{Blur: -1}
	}
	layer := func(offsetY, blur, alpha float32) render.ShadowLayer {
		layerColor := col
		layerColor.A = byte(255*alpha*opacity + 0.5)
		return render.ShadowLayer{OffsetY: offsetY, Blur: blur, Color: layerColor}
	}
	return render.BoxShadow{Layers: []render.ShadowLayer{
		layer(2, 8, 0.06),
		layer(-6, 12, 0.03),
		layer(14, 28, 0.08),
	}}
}

func drawMenuItem(gtx layout.Context, activeTheme *theme.Theme, size image.Point, radius int, style itemStyle) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	rect := image.Rectangle{Max: size}
	if style.background.A != 0 {
		paint.FillShape(gtx.Ops, style.background, clip.UniformRRect(rect, radius).Op(gtx.Ops))
	}
	if style.focus <= 0 {
		return
	}
	width := max(gtx.Dp(activeTheme.Components.Menu.FocusRingWidth), 1)
	inset := max(width/2+1, 1)
	focusRect := rect.Inset(inset)
	if focusRect.Empty() {
		return
	}
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawMenuSeparator(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.Rect{Max: size}.Op())
}

func drawRadioIndicator(gtx layout.Context, size image.Point, dotSize int, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	dot := min(max(dotSize, 1), min(size.X, size.Y))
	rect := image.Rect((size.X-dot)/2, (size.Y-dot)/2, (size.X+dot)/2, (size.Y+dot)/2)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(rect).Op(gtx.Ops))
}
