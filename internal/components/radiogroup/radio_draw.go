package radiogroup

import (
	"image"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
	"github.com/qianniancn/FlowUI/internal/theme"
)

func drawRadio(gtx layout.Context, theme *theme.Theme, style radioStyle, scale float32) layout.Dimensions {
	size := gtx.Dp(theme.Components.RadioGroup.Size)
	focusSpace := max(gtx.Dp(theme.Components.RadioGroup.FocusSpace), 1)
	maxSize := min(gtx.Constraints.Max.X, gtx.Constraints.Max.Y) - focusSpace*2
	size = min(size, max(maxSize, 0))
	bounds := image.Pt(size+focusSpace*2, size+focusSpace*2)
	dims := gtx.Constraints.Constrain(bounds)
	if size <= 0 {
		return layout.Dimensions{Size: dims}
	}

	origin := image.Pt((dims.X-size)/2, (dims.Y-size)/2)
	rect := image.Rectangle{
		Min: origin,
		Max: origin.Add(image.Pt(size, size)),
	}
	center := f32.Pt(
		float32(rect.Min.X)+float32(size)/2,
		float32(rect.Min.Y)+float32(size)/2,
	)
	stack := op.Affine(f32.AffineId().Scale(center, f32.Pt(scale, scale))).Push(gtx.Ops)
	drawRadioFocus(gtx, rect, style)
	drawRadioControl(gtx, theme, rect, style)
	drawRadioDot(gtx, theme, rect, style)
	stack.Pop()

	return layout.Dimensions{Size: dims}
}

func drawRadioFocus(gtx layout.Context, rect image.Rectangle, style radioStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(unit.Dp(2)), 1)
	focusRect := rect.Inset(-max(width/2, 1))
	col := style.focusColor
	col.A = byte(float32(col.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.Ellipse(focusRect).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawRadioControl(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, style radioStyle) {
	bg := render.LerpColor(style.bg, style.selectedBg, style.selected)
	border := style.border
	border.A = byte(float32(border.A)*(1-style.selected) + 0.5)
	if border.A == 0 {
		paint.FillShape(gtx.Ops, bg, clip.Ellipse(rect).Op(gtx.Ops))
		return
	}

	width := max(gtx.Dp(theme.Components.RadioGroup.BorderWidth), 1)
	paint.FillShape(gtx.Ops, border, clip.Ellipse(rect).Op(gtx.Ops))
	inner := rect.Inset(width)
	if inner.Empty() {
		return
	}
	paint.FillShape(gtx.Ops, bg, clip.Ellipse(inner).Op(gtx.Ops))
}

func drawRadioDot(gtx layout.Context, theme *theme.Theme, rect image.Rectangle, style radioStyle) {
	if style.selected == 0 {
		return
	}
	targetScale := theme.Components.RadioGroup.DotScale
	if style.pressed {
		targetScale = theme.Components.RadioGroup.DotPressedScale
	}
	size := max(int(float32(rect.Dx())*targetScale*style.selected+0.5), 1)
	center := rect.Min.Add(rect.Size().Div(2))
	dot := image.Rectangle{
		Min: center.Sub(image.Pt(size/2, size/2)),
		Max: center.Sub(image.Pt(size/2, size/2)).Add(image.Pt(size, size)),
	}
	col := style.dot
	col.A = byte(float32(col.A)*style.selected + 0.5)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(dot).Op(gtx.Ops))
}
