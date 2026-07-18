package spinner

import (
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"github.com/qianniancn/FlowUI/internal/render"
)

type spinnerGeometry struct {
	center      f32.Point
	radius      float32
	strokeWidth float32
}

type spinnerArcSpec struct {
	startAngle    float32
	sweep         float32
	gradientStart float32
	gradientEnd   float32
	startAlpha    float32
	endAlpha      float32
}

var spinnerArcs = [...]spinnerArcSpec{
	{
		startAngle:    -99.61 * math.Pi / 180,
		sweep:         -170.39 * math.Pi / 180,
		gradientStart: 2.745 / 24,
		gradientEnd:   20.79 / 24,
		startAlpha:    1,
		endAlpha:      .55,
	},
	{
		startAngle:    -48.11 * math.Pi / 180,
		sweep:         138.11 * math.Pi / 180,
		gradientStart: 6.974 / 24,
		gradientEnd:   20.145 / 24,
		startAlpha:    0,
		endAlpha:      .55,
	},
}

func drawSpinner(gtx layout.Context, diameter int, strokeRatio, insetRatio float32, col color.NRGBA, period time.Duration) {
	if diameter <= 0 || col.A == 0 {
		return
	}
	geometry, ok := resolveSpinnerGeometry(diameter, strokeRatio, insetRatio)
	if !ok {
		return
	}

	transform := op.Affine(f32.AffineId().Rotate(geometry.center, spinnerPhase(gtx.Now, period))).Push(gtx.Ops)
	for _, arc := range spinnerArcs {
		drawSpinnerArc(gtx, diameter, geometry, arc, col)
	}
	transform.Pop()
	if period > 0 {
		gtx.Execute(op.InvalidateCmd{})
	}
}

func resolveSpinnerGeometry(diameter int, strokeRatio, insetRatio float32) (spinnerGeometry, bool) {
	if diameter <= 0 {
		return spinnerGeometry{}, false
	}
	if strokeRatio <= 0 || strokeRatio >= .5 {
		strokeRatio = .125
	}
	if insetRatio < 0 || insetRatio >= .5 {
		insetRatio = .0625
	}
	strokeWidth := max(float32(diameter)*strokeRatio, 1)
	outerRadius := float32(diameter) * (.5 - insetRatio)
	if outerRadius <= strokeWidth {
		strokeWidth = max(float32(diameter)*.125, 1)
		outerRadius = float32(diameter) * (.5 - .0625)
	}
	if outerRadius <= strokeWidth {
		return spinnerGeometry{}, false
	}
	radius := outerRadius - strokeWidth/2
	return spinnerGeometry{
		center:      f32.Pt(float32(diameter)/2, float32(diameter)/2),
		radius:      radius,
		strokeWidth: strokeWidth,
	}, true
}

func drawSpinnerArc(gtx layout.Context, diameter int, geometry spinnerGeometry, arc spinnerArcSpec, col color.NRGBA) {
	path := spinnerArcPath(gtx, geometry, arc)
	shape := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	render.PaintLinearGradient(
		gtx,
		f32.Pt(geometry.center.X, float32(diameter)*arc.gradientStart),
		spinnerAlpha(col, arc.startAlpha),
		f32.Pt(geometry.center.X, float32(diameter)*arc.gradientEnd),
		spinnerAlpha(col, arc.endAlpha),
	)
	shape.Pop()
}

func spinnerArcPath(gtx layout.Context, geometry spinnerGeometry, arc spinnerArcSpec) clip.PathSpec {
	outerRadius := geometry.radius + geometry.strokeWidth/2
	innerRadius := geometry.radius - geometry.strokeWidth/2
	outerStart := spinnerPoint(geometry.center, outerRadius, arc.startAngle)
	outerEnd := spinnerPoint(geometry.center, outerRadius, arc.startAngle+arc.sweep)
	innerEnd := spinnerPoint(geometry.center, innerRadius, arc.startAngle+arc.sweep)
	capCenter := spinnerPoint(geometry.center, geometry.radius, arc.startAngle)

	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(outerStart)
	if arc.sweep < 0 {
		path.ArcTo(capCenter, capCenter, float32(math.Pi))
		path.ArcTo(geometry.center, geometry.center, arc.sweep)
		path.LineTo(outerEnd)
		path.ArcTo(geometry.center, geometry.center, -arc.sweep)
	} else {
		path.ArcTo(geometry.center, geometry.center, arc.sweep)
		path.LineTo(innerEnd)
		path.ArcTo(geometry.center, geometry.center, -arc.sweep)
		path.ArcTo(capCenter, capCenter, float32(math.Pi))
	}
	path.Close()
	return path.End()
}

func spinnerPoint(center f32.Point, radius, angle float32) f32.Point {
	return f32.Pt(
		center.X+float32(math.Cos(float64(angle)))*radius,
		center.Y+float32(math.Sin(float64(angle)))*radius,
	)
}

func spinnerAlpha(col color.NRGBA, opacity float32) color.NRGBA {
	opacity = min(max(opacity, 0), 1)
	col.A = byte(float32(col.A)*opacity + .5)
	return col
}

func spinnerPhase(now time.Time, period time.Duration) float32 {
	if now.IsZero() || period <= 0 {
		return 0
	}
	elapsed := now.UnixNano() % int64(period)
	if elapsed < 0 {
		elapsed += int64(period)
	}
	return float32(elapsed) / float32(period) * float32(math.Pi*2)
}
