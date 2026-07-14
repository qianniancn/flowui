package progress

import (
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
)

type progressCircleGeometry struct {
	center      f32.Point
	outerRadius float32
	innerRadius float32
}

func drawProgressCircle(gtx layout.Context, diameter int, strokeRatio float32, style progressCircleStyle, progress float32, indeterminate bool) {
	geometry, ok := resolveProgressCircleGeometry(diameter, strokeRatio)
	if !ok {
		return
	}
	drawProgressCircleRing(gtx, geometry, style.track)

	progress = min(max(progress, 0), 1)
	start := float32(-math.Pi / 2)
	if indeterminate {
		progress = .25
		start += progressCirclePhase(gtx.Now)
		gtx.Execute(op.InvalidateCmd{})
	}
	if progress > 0 {
		drawProgressCircleArc(gtx, geometry, start, progress*float32(2*math.Pi), style.fill)
	}
}

func resolveProgressCircleGeometry(diameter int, strokeRatio float32) (progressCircleGeometry, bool) {
	if diameter <= 0 {
		return progressCircleGeometry{}, false
	}
	if strokeRatio <= 0 || strokeRatio >= .5 {
		strokeRatio = 4.0 / 36.0
	}
	strokeWidth := max(float32(diameter)*strokeRatio, 1)
	outerRadius := float32(diameter) / 2
	innerRadius := outerRadius - strokeWidth
	if innerRadius <= 0 {
		return progressCircleGeometry{}, false
	}
	return progressCircleGeometry{
		center:      f32.Pt(float32(diameter)/2, float32(diameter)/2),
		outerRadius: outerRadius,
		innerRadius: innerRadius,
	}, true
}

func drawProgressCircleRing(gtx layout.Context, geometry progressCircleGeometry, col color.NRGBA) {
	if col.A == 0 {
		return
	}
	path := progressCircleRingPath(gtx, geometry)
	stack := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stack.Pop()
}

func drawProgressCircleArc(gtx layout.Context, geometry progressCircleGeometry, start, sweep float32, col color.NRGBA) {
	if col.A == 0 || sweep <= 0 {
		return
	}
	full := sweep >= float32(2*math.Pi)-1e-4
	if full {
		drawProgressCircleRing(gtx, geometry, col)
		return
	}
	path := progressCircleArcPath(gtx, geometry, start, sweep)
	stack := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stack.Pop()

}

func progressCircleRingPath(gtx layout.Context, geometry progressCircleGeometry) clip.PathSpec {
	start := float32(-math.Pi / 2)
	outerStart := progressCirclePoint(geometry.center, geometry.outerRadius, start)
	innerStart := progressCirclePoint(geometry.center, geometry.innerRadius, start)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(outerStart)
	path.ArcTo(geometry.center, geometry.center, math.Pi)
	path.ArcTo(geometry.center, geometry.center, math.Pi)
	path.LineTo(innerStart)
	path.ArcTo(geometry.center, geometry.center, -math.Pi)
	path.ArcTo(geometry.center, geometry.center, -math.Pi)
	path.Close()
	return path.End()
}

func progressCircleArcPath(gtx layout.Context, geometry progressCircleGeometry, start, sweep float32) clip.PathSpec {
	outerStart := progressCirclePoint(geometry.center, geometry.outerRadius, start)
	centerRadius := (geometry.outerRadius + geometry.innerRadius) / 2
	startCap := progressCirclePoint(geometry.center, centerRadius, start)
	endCap := progressCirclePoint(geometry.center, centerRadius, start+sweep)
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(outerStart)
	path.ArcTo(geometry.center, geometry.center, sweep)
	path.ArcTo(endCap, endCap, math.Pi)
	path.ArcTo(geometry.center, geometry.center, -sweep)
	path.ArcTo(startCap, startCap, math.Pi)
	path.Close()
	return path.End()
}

func progressCirclePoint(center f32.Point, radius, angle float32) f32.Point {
	return f32.Pt(
		center.X+float32(math.Cos(float64(angle)))*radius,
		center.Y+float32(math.Sin(float64(angle)))*radius,
	)
}

func progressCirclePhase(now time.Time) float32 {
	if now.IsZero() {
		return 0
	}
	elapsed := now.UnixNano() % int64(progressCircleSpinDuration)
	if elapsed < 0 {
		elapsed += int64(progressCircleSpinDuration)
	}
	return float32(elapsed) / float32(progressCircleSpinDuration) * float32(2*math.Pi)
}
