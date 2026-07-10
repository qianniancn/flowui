package button

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
)

func drawButton(gtx layout.Context, size image.Point, style buttonStyle) {
	radius := min(size.X, size.Y) / 2
	rect := image.Rectangle{Max: size}
	rr := clip.UniformRRect(rect, radius)

	if style.bg.A != 0 {
		paint.FillShape(gtx.Ops, style.bg, rr.Op(gtx.Ops))
	}
	if style.hasBorder {
		stroke := clip.Stroke{
			Path:  rr.Path(gtx.Ops),
			Width: float32(max(gtx.Dp(unit.Dp(1)), 1)),
		}.Op().Push(gtx.Ops)
		paint.Fill(gtx.Ops, style.border)
		stroke.Pop()
	}
	drawButtonFocus(gtx, size, radius, style)
}

func drawButtonSpinner(gtx layout.Context, size unit.Dp, col color.NRGBA) layout.Dimensions {
	d := max(gtx.Dp(size), 1)
	strokeWidth := float32(max(gtx.Dp(unit.Dp(2)), 1))
	radius := float32(d)/2 - strokeWidth/2
	if radius <= 0 {
		return layout.Dimensions{Size: image.Pt(d, d)}
	}

	const segments = 18
	start := buttonSpinnerPhase(gtx.Now)
	sweep := float32(math.Pi * 1.45)
	center := f32.Pt(float32(d)/2, float32(d)/2)

	var path clip.Path
	path.Begin(gtx.Ops)
	for i := range segments + 1 {
		angle := start + sweep*float32(i)/segments
		pt := f32.Pt(
			center.X+float32(math.Cos(float64(angle)))*radius,
			center.Y+float32(math.Sin(float64(angle)))*radius,
		)
		if i == 0 {
			path.MoveTo(pt)
			continue
		}
		path.LineTo(pt)
	}
	stroke := clip.Stroke{
		Path:  path.End(),
		Width: strokeWidth,
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()

	return layout.Dimensions{Size: image.Pt(d, d)}
}

func drawButtonFocus(gtx layout.Context, size image.Point, radius int, style buttonStyle) {
	if style.focus == 0 {
		return
	}
	width := max(gtx.Dp(style.focusWidth), 1)
	inset := width + 1
	rect := image.Rectangle{
		Min: image.Pt(inset, inset),
		Max: image.Pt(size.X-inset, size.Y-inset),
	}
	if rect.Empty() {
		return
	}
	ring := style.focusColor
	ring.A = byte(float32(ring.A)*style.focus + 0.5)
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, max(radius-inset, 0)).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, ring)
	stroke.Pop()
}

func buttonSpinnerPhase(now time.Time) float32 {
	if now.IsZero() {
		return 0
	}
	elapsed := now.UnixNano() % int64(buttonSpinnerPeriod)
	if elapsed < 0 {
		elapsed += int64(buttonSpinnerPeriod)
	}
	return float32(elapsed) / float32(buttonSpinnerPeriod) * float32(math.Pi*2)
}
