// Package spinnerstyle contains the theme-owned visual recipe for standalone
// spinner behavior. It maps semantic color and size axes to declarations;
// widget/spinner remains responsible for loading semantics and animation
// invalidation.
package spinnerstyle

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"github.com/qianniancn/flowui/appearance"
	"github.com/qianniancn/flowui/draw"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
)

// BuildSpinner builds the complete provider-owned spinner visual plan.
func BuildSpinner(active *theme.Theme, request theme.SpinnerRequest) theme.SpinnerVisual {
	if active == nil {
		panic(theme.ErrMissingVisualProvider)
	}
	tokens := active.Components.Spinner
	diameter := spinnerDiameter(tokens, request.Size)
	return theme.SpinnerVisual{
		Defaults: style.Style{}.
			Width(diameter).
			Height(diameter).
			AspectRatio(1).
			Part(style.PartIndicator, style.Style{}.
				TextColor(style.Solid(spinnerColor(active, request.Color)))),
		Override: theme.SpinnerOverrideStyle(request.Override),
		Color:    spinnerColorResolver,
		Renderer: spinnerRenderer,
		Metrics: theme.SpinnerMetrics{
			StrokeRatio:  tokens.StrokeRatio,
			InsetRatio:   tokens.InsetRatio,
			SpinDuration: tokens.SpinDuration,
		},
		Motion: active.Motion,
	}
}

// spinnerRenderer is the theme-owned indicator strategy. The standalone
// widget only schedules this renderer and supplies the resolved provider
// inputs; it does not own the arc geometry or gradient recipe.
func spinnerRenderer(gtx layout.Context, request theme.SpinnerRenderRequest) layout.Dimensions {
	diameter := max(gtx.Dp(request.Size), 1)
	if diameter <= 0 || request.Color.A == 0 {
		return layout.Dimensions{Size: image.Pt(diameter, diameter)}
	}
	geometry, ok := resolveSpinnerGeometry(diameter, request.StrokeRatio, request.InsetRatio)
	if !ok {
		return layout.Dimensions{Size: image.Pt(diameter, diameter)}
	}
	transform := op.Affine(f32.AffineId().Rotate(geometry.center, spinnerPhase(gtx.Now, request.Period))).Push(gtx.Ops)
	for _, arc := range spinnerArcs {
		drawSpinnerArc(gtx, diameter, geometry, arc, request.Color)
	}
	transform.Pop()
	return layout.Dimensions{Size: image.Pt(diameter, diameter)}
}

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
	{startAngle: -99.61 * math.Pi / 180, sweep: -170.39 * math.Pi / 180, gradientStart: 2.745 / 24, gradientEnd: 20.79 / 24, startAlpha: 1, endAlpha: .55},
	{startAngle: -48.11 * math.Pi / 180, sweep: 138.11 * math.Pi / 180, gradientStart: 6.974 / 24, gradientEnd: 20.145 / 24, startAlpha: 0, endAlpha: .55},
}

func resolveSpinnerGeometry(diameter int, strokeRatio, insetRatio float32) (spinnerGeometry, bool) {
	if diameter <= 0 || math.IsNaN(float64(strokeRatio)) || math.IsInf(float64(strokeRatio), 0) ||
		math.IsNaN(float64(insetRatio)) || math.IsInf(float64(insetRatio), 0) ||
		strokeRatio <= 0 || strokeRatio >= .5 || insetRatio < 0 || insetRatio >= .5 {
		return spinnerGeometry{}, false
	}
	strokeWidth := max(float32(diameter)*strokeRatio, 1)
	outerRadius := float32(diameter) * (.5 - insetRatio)
	if outerRadius <= strokeWidth {
		return spinnerGeometry{}, false
	}
	return spinnerGeometry{center: f32.Pt(float32(diameter)/2, float32(diameter)/2), radius: outerRadius - strokeWidth/2, strokeWidth: strokeWidth}, true
}

func drawSpinnerArc(gtx layout.Context, diameter int, geometry spinnerGeometry, arc spinnerArcSpec, col color.NRGBA) {
	path := spinnerArcPath(gtx, geometry, arc)
	shape := clip.Outline{Path: path}.Op().Push(gtx.Ops)
	draw.PaintLinearGradient(
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
	return f32.Pt(center.X+float32(math.Cos(float64(angle)))*radius, center.Y+float32(math.Sin(float64(angle)))*radius)
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

func spinnerDiameter(tokens theme.SpinnerTheme, size theme.SpinnerSize) unit.Dp {
	switch size {
	case theme.SpinnerSmall:
		return tokens.SmallSize
	case theme.SpinnerLarge:
		return tokens.LargeSize
	case theme.SpinnerExtraLarge:
		return tokens.ExtraLargeSize
	default:
		return tokens.MediumSize
	}
}

func spinnerColor(active *theme.Theme, value theme.SpinnerColor) color.NRGBA {
	if value == theme.SpinnerCurrent {
		return active.Semantic().Foreground
	}
	switch value {
	case theme.SpinnerSuccess:
		return appearance.Tone(active, theme.ToneSuccess).Fill
	case theme.SpinnerWarning:
		return appearance.Tone(active, theme.ToneWarning).Fill
	case theme.SpinnerDanger:
		return appearance.Tone(active, theme.ToneDanger).Fill
	default:
		return appearance.Tone(active, theme.TonePrimary).Fill
	}
}

func spinnerColorResolver(value style.ResolvedStyle) (color.NRGBA, bool) {
	if value.Text == nil {
		return color.NRGBA{}, false
	}
	resolved, ok := computed.Color(value.Text.Color)
	if !ok {
		return color.NRGBA{}, false
	}
	if value.Paint != nil {
		resolved = computed.ApplyOpacity(resolved, value.Paint.Opacity)
	}
	return resolved, true
}
