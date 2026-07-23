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

func drawButtonSpinner(gtx layout.Context, size, strokeWidthDp unit.Dp, col color.NRGBA, period time.Duration) layout.Dimensions {
	d := max(gtx.Dp(size), 1)
	strokeWidth := float32(max(gtx.Dp(strokeWidthDp), 1))
	radius := float32(d)/2 - strokeWidth/2
	if radius <= 0 {
		return layout.Dimensions{Size: image.Pt(d, d)}
	}

	const segments = 18
	start := buttonSpinnerPhase(gtx.Now, period)
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

func buttonSpinnerPhase(now time.Time, period time.Duration) float32 {
	if now.IsZero() || period <= 0 {
		return 0
	}
	elapsed := now.UnixNano() % int64(period)
	if elapsed < 0 {
		elapsed += int64(period)
	}
	return float32(elapsed) / float32(period) * float32(math.Pi*2)
}
