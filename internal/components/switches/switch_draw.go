package switches

import (
	"image"
	"image/color"

	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

type switchDrawResult struct {
	dims      layout.Dimensions
	trackRect image.Rectangle
	thumbRect image.Rectangle
}

func drawSwitch(gtx layout.Context, theme *theme.Theme, style switchStyle, size switchSizeStyle) switchDrawResult {
	rects := switchRects(gtx, theme, style, size)
	if rects.trackRect.Empty() {
		return rects
	}
	drawSwitchFocus(gtx, theme, rects.trackRect, style)
	drawSwitchTrack(gtx, rects.trackRect, style)
	drawSwitchThumb(gtx, theme, rects.thumbRect, style, size)
	return rects
}

func switchRects(gtx layout.Context, theme *theme.Theme, style switchStyle, size switchSizeStyle) switchDrawResult {
	focusSpace := max(gtx.Dp(theme.Components.Switch.FocusSpace), 1)
	trackWidth := min(gtx.Dp(size.trackWidth), max(gtx.Constraints.Max.X-focusSpace*2, 0))
	trackHeight := min(gtx.Dp(size.trackHeight), max(gtx.Constraints.Max.Y-focusSpace*2, 0))
	bounds := image.Pt(trackWidth+focusSpace*2, trackHeight+focusSpace*2)
	dims := layout.Dimensions{Size: gtx.Constraints.Constrain(bounds)}
	if trackWidth <= 0 || trackHeight <= 0 {
		return switchDrawResult{dims: dims}
	}

	trackOrigin := image.Pt((dims.Size.X-trackWidth)/2, (dims.Size.Y-trackHeight)/2)
	track := image.Rectangle{
		Min: trackOrigin,
		Max: trackOrigin.Add(image.Pt(trackWidth, trackHeight)),
	}
	thumbWidth := min(gtx.Dp(size.thumbWidth), trackWidth)
	thumbHeight := min(gtx.Dp(size.thumbHeight), trackHeight)
	padding := max((trackHeight-thumbHeight)/2, 0)
	travel := max(trackWidth-padding*2-thumbWidth, 0)
	thumbX := track.Min.X + padding + int(float32(travel)*style.selected+0.5)
	thumbY := track.Min.Y + (trackHeight-thumbHeight)/2
	thumb := image.Rectangle{
		Min: image.Pt(thumbX, thumbY),
		Max: image.Pt(thumbX+thumbWidth, thumbY+thumbHeight),
	}
	return switchDrawResult{
		dims:      dims,
		trackRect: track,
		thumbRect: thumb,
	}
}

func drawSwitchFocus(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, style switchStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(theme.Components.Switch.FocusRingWidth), 1)
	focusRect := rect.Inset(-max(width/2, 1))
	radius := min(rect.Dy()/2+width, focusRect.Dy()/2)
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(focusRect, radius).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawSwitchTrack(gtx layout.Context, rect image.Rectangle, style switchStyle) {
	bg := render.LerpColor(style.trackOff, style.trackOn, style.selected)
	paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, rect.Dy()/2).Op(gtx.Ops))
}

func drawSwitchThumb(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, style switchStyle, size switchSizeStyle) {
	if rect.Empty() {
		return
	}
	drawSwitchThumbShadow(gtx, theme, rect, size)
	bg := render.LerpColor(style.thumb, style.thumbOn, style.selected)
	paint.FillShape(gtx.Ops, bg, clip.UniformRRect(rect, rect.Dy()/2).Op(gtx.Ops))
}

func drawSwitchThumbShadow(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, size switchSizeStyle) {
	shadow := theme.Palette.Shadow
	if shadow.A == 0 {
		return
	}
	render.DrawShadow(gtx, rect, switchThumbShadowShape(size), render.ThemeShadow(theme.Shadows.SwitchThumb, shadow, 1))

}

func switchThumbShadowShape(size switchSizeStyle) render.ShadowShape {
	radius := size.thumbHeight / 2
	return render.RoundedShadowCorners(radius, radius, radius, radius)
}

func switchThumbContentColor(style switchStyle) color.NRGBA {
	return render.LerpColor(style.thumbFgOff, style.thumbFg, style.selected)
}
