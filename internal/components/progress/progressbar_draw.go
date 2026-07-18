package progress

import (
	"image"
	"image/color"
	"time"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"github.com/qianniancn/FlowUI/internal/render"
)

type linearTrackStyle struct {
	track color.NRGBA
	fill  color.NRGBA
}

func drawProgressBar(gtx layout.Context, size image.Point, radius unit.Dp, style progressBarStyle, progress float32, indeterminate bool, period time.Duration) {
	drawLinearTrack(gtx, size, radius, linearTrackStyle{track: style.track, fill: style.fill}, progress, indeterminate, period)
}

func drawLinearTrack(gtx layout.Context, size image.Point, radius unit.Dp, style linearTrackStyle, progress float32, indeterminate bool, period time.Duration) {
	if size.X <= 0 || size.Y <= 0 {
		return
	}
	r := progressBarRadius(gtx, size, radius)
	trackRect := image.Rectangle{Max: size}
	track := clip.UniformRRect(trackRect, r)
	paint.FillShape(gtx.Ops, style.track, track.Op(gtx.Ops))

	clipStack := track.Push(gtx.Ops)
	if indeterminate {
		drawProgressBarIndeterminate(gtx, size, r, style, period)
	} else {
		drawProgressBarFill(gtx, size, r, style, progress)
	}
	clipStack.Pop()
}

func drawProgressBarFill(gtx layout.Context, size image.Point, radius int, style linearTrackStyle, progress float32) {
	if progress <= 0 {
		return
	}
	progress = min(max(progress, 0), 1)
	width := int(float32(size.X)*progress + 0.5)
	if width <= 0 {
		width = 1
	}
	rect := image.Rectangle{Max: image.Pt(min(width, size.X), size.Y)}
	paint.FillShape(gtx.Ops, style.fill, clip.UniformRRect(rect, min(radius, rect.Dx()/2)).Op(gtx.Ops))
}

func drawProgressBarIndeterminate(gtx layout.Context, size image.Point, radius int, style linearTrackStyle, period time.Duration) {
	width := max(int(float32(size.X)*progressBarIndeterminateFillRate+0.5), 1)
	x := progressBarIndeterminatePosition(gtx.Now, width, period)
	if period > 0 {
		gtx.Execute(op.InvalidateCmd{})
	}
	rect := image.Rect(x, 0, x+width, size.Y)
	paint.FillShape(gtx.Ops, style.fill, clip.UniformRRect(rect, min(radius, rect.Dx()/2)).Op(gtx.Ops))
}

func progressBarRadius(gtx layout.Context, size image.Point, radius unit.Dp) int {
	r := gtx.Dp(radius)
	if r <= 0 {
		r = size.Y / 2
	}
	return min(r, min(size.X, size.Y)/2)
}

func progressBarIndeterminateOffset(now time.Time, fillWidth int, period time.Duration) int {
	if fillWidth <= 0 || now.IsZero() || period <= 0 {
		return -fillWidth
	}
	elapsed := now.UnixNano() % int64(period)
	if elapsed < 0 {
		elapsed += int64(period)
	}
	progress := render.Ease(float32(elapsed) / float32(period))
	return int(render.Lerp(float32(-fillWidth), float32(fillWidth)*3.5, progress) + 0.5)
}

func progressBarIndeterminatePosition(now time.Time, fillWidth int, period time.Duration) int {
	if period <= 0 {
		return 0
	}
	return progressBarIndeterminateOffset(now, fillWidth, period)
}
