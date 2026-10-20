package main

import (
	"image"
	"image/color"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	lucide "github.com/qianniancn/flowui-icons-lucide"
	"github.com/qianniancn/flowui/animation"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/card"
	iconpkg "github.com/qianniancn/flowui/widget/icon"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
)

func motionDemoCards(model Model, send ui.Send[Msg]) []ui.Widget {
	playIcon := lucide.Pause
	playLabel := "Pause timeline"
	if !model.MotionPlaying {
		playIcon = lucide.Play
		playLabel = "Play timeline"
	}
	layoutIcon := lucide.Maximize2
	layoutLabel := "Expand layout"
	if model.MotionExpanded {
		layoutIcon = lucide.Minimize2
		layoutLabel = "Collapse layout"
	}

	return []ui.Widget{
		motionDemoCard("Tween and interpolation", nil, tweenTransformDemo{forward: model.MotionForward}),
		motionDemoCard("Spring presets", nil, springComparison(model.MotionForward)),
		motionDemoCard(
			"Keyframe timeline",
			flowlayout.Row(
				motionIconButton("catalog-timeline-play", playIcon, playLabel, func() {
					send(func(model *Model) { model.MotionPlaying = !model.MotionPlaying })
				}),
				motionIconButton("catalog-timeline-restart", lucide.RotateCcw, "Restart timeline", func() {
					send(func(model *Model) { model.MotionRun++ })
				}),
			).Gap(4),
			timelineDemo{playing: model.MotionPlaying, revision: model.MotionRun},
		),
		motionDemoCard(
			"Animated layout",
			motionIconButton("catalog-layout-toggle", layoutIcon, layoutLabel, func() {
				send(func(model *Model) { model.MotionExpanded = !model.MotionExpanded })
			}),
			fixedDemoStage{
				height: 132,
				child: animation.AnimateLayout("catalog-animated-layout", motionLayoutTarget(model.MotionExpanded)).
					Spring(animation.SpringGentle()),
			},
		),
		motionDemoCard(
			"Animated rectangle",
			motionIconButton("catalog-rect-toggle", lucide.RefreshCw, "Move rectangle", func() {
				send(func(model *Model) { model.MotionRectAlt = !model.MotionRectAlt })
			}),
			rectDemo{alternate: model.MotionRectAlt},
		),
		motionDemoCard("Style transitions", nil, styleTransitionDemo()),
	}
}

func motionDemoCard(title string, actions, content ui.Widget) ui.Widget {
	header := []ui.Widget{flowlayout.Expanded(text.New(title).Size(14))}
	if actions != nil {
		header = append(header, actions)
	}
	return (card.New(
		flowlayout.Row(header...).AlignMiddle(),
		flowlayout.Divider(),
		content,
	).
		Variant(card.Secondary)).StyleWithMetrics(style.New().MinHeight(220), 14, 12)
}

func motionIconButton(key string, icon lucide.Data, label string, onClick func()) ui.Widget {
	return button.New("").Key(key).Content(iconpkg.New(icon).Size(14)).
		Variant(button.Ghost).
		Size(button.Small).
		IconOnly().
		Label(label).
		OnClick(onClick)
}

type tweenTransformDemo struct {
	forward bool
}

func (demo tweenTransformDemo) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := motionDemoSize(gtx, 132)
	fillMotionBackground(gtx, size, ctx.Theme().Palette.SurfaceTertiary)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}

	progress := animation.Tween("catalog-transform-progress", boolFloat(demo.forward)).
		Initial(0).
		Duration(650*time.Millisecond).
		Easing(animation.EaseCubicInOut).
		Value(ctx, gtx)

	left := f32.Pt(float32(gtx.Dp(28)), float32(size.Y)/2)
	right := f32.Pt(float32(max(size.X-gtx.Dp(28), gtx.Dp(28))), float32(size.Y)/2)
	point := animation.LerpPoint(left, right, progress)
	palette := ctx.Theme().Palette
	drawMotionLine(gtx, left, right, 2, palette.Separator)
	drawMotionDot(gtx, left, gtx.Dp(4), palette.Separator)
	drawMotionDot(gtx, right, gtx.Dp(4), palette.Separator)

	blockSize := max(gtx.Dp(28), 1)
	half := blockSize / 2
	block := image.Rect(int(point.X)-half, int(point.Y)-half, int(point.X)+half, int(point.Y)+half)
	scale := animation.LerpFloat(.82, 1.18, progress)
	angle := float32(animation.LerpFloat64(0, math.Pi*1.5, progress))
	transform := f32.AffineId().Scale(point, f32.Pt(scale, scale)).Rotate(point, angle)
	stack := op.Affine(transform).Push(gtx.Ops)
	paint.FillShape(gtx.Ops, animation.LerpColor(palette.Accent, palette.Success, progress), clip.UniformRRect(block, gtx.Dp(5)).Op(gtx.Ops))
	stack.Pop()
	return layout.Dimensions{Size: size}
}

type springDemoSpec struct {
	key    string
	label  string
	config animation.SpringConfig
	mix    float32
}

var springDemoSpecs = [...]springDemoSpec{
	{key: "default", label: "Default", config: animation.DefaultSpring(), mix: 0},
	{key: "snappy", label: "Snappy", config: animation.SpringSnappy(), mix: .33},
	{key: "gentle", label: "Gentle", config: animation.SpringGentle(), mix: .66},
	{key: "bouncy", label: "Bouncy", config: animation.SpringBouncy(), mix: 1},
}

func springComparison(forward bool) ui.Widget {
	rows := make([]ui.Widget, 0, len(springDemoSpecs))
	for _, spec := range springDemoSpecs {
		rows = append(rows, flowlayout.Row(
			flowlayout.Box(text.New(spec.label).Size(12)).Style(style.New().Width(64)),
			flowlayout.Expanded(springLane{spec: spec, forward: forward}),
		).AlignMiddle().Gap(8))
	}
	return fixedDemoStage{height: 132, child: flowlayout.Column(rows...).Gap(4)}
}

type springLane struct {
	spec    springDemoSpec
	forward bool
}

func (lane springLane) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := motionDemoSize(gtx, 26)
	left := float32(gtx.Dp(18))
	right := float32(max(size.X-gtx.Dp(18), gtx.Dp(18)))
	target := left
	if lane.forward {
		target = right
	}
	x := animation.Tween("catalog-spring-"+lane.spec.key, target).
		Initial(left).
		Spring(lane.spec.config).
		Value(ctx, gtx)
	y := float32(size.Y) / 2
	palette := ctx.Theme().Palette
	drawMotionLine(gtx, f32.Pt(left, y), f32.Pt(right, y), 2, palette.Separator)
	drawMotionDot(gtx, f32.Pt(x, y), gtx.Dp(6), animation.LerpColor(palette.Accent, palette.Warning, lane.spec.mix))
	return layout.Dimensions{Size: size}
}

type timelineDemo struct {
	playing  bool
	revision uint64
}

func (demo timelineDemo) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := motionDemoSize(gtx, 132)
	palette := ctx.Theme().Palette
	fillMotionBackground(gtx, size, palette.SurfaceTertiary)
	value := animation.Timeline("catalog-keyframe-track").
		Keyframe(0, 0).
		Keyframe(.18, .9).
		Keyframe(.42, .25).
		Keyframe(.7, 1).
		Keyframe(1, 0).
		Duration(2400*time.Millisecond).
		Easing(animation.EaseCubicInOut).
		Loop(true).
		Playing(demo.playing).
		Revision(demo.revision).
		Value(ctx, gtx)
	left := float32(gtx.Dp(24))
	right := float32(max(size.X-gtx.Dp(24), gtx.Dp(24)))
	y := float32(size.Y) / 2
	drawMotionLine(gtx, f32.Pt(left, y), f32.Pt(right, y), 2, palette.Separator)
	for _, stop := range []float32{0, .25, .5, .75, 1} {
		x := animation.LerpFloat(left, right, stop)
		drawMotionLine(gtx, f32.Pt(x, y-float32(gtx.Dp(8))), f32.Pt(x, y+float32(gtx.Dp(8))), 1, palette.Separator)
	}
	drawMotionDot(gtx, f32.Pt(animation.LerpFloat(left, right, value), y), gtx.Dp(9), animation.LerpColor(palette.Accent, palette.Warning, value))
	return layout.Dimensions{Size: size}
}

func motionLayoutTarget(expanded bool) ui.Widget {
	width, height := 124, 52
	label := "Compact"
	if expanded {
		width, height = 220, 108
		label = "Expanded"
	}
	return (surface.New(flowlayout.Box(text.New(label).Size(13)).Style(style.New().Width(unit.Dp(width)).Height(unit.Dp(height)).Padding(12))).
		Variant(surface.Tertiary)).Style(style.New().Radius(6))
}

type rectDemo struct {
	alternate bool
}

func (demo rectDemo) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := motionDemoSize(gtx, 132)
	palette := ctx.Theme().Palette
	fillMotionBackground(gtx, size, palette.SurfaceTertiary)
	compact := image.Rect(gtx.Dp(18), gtx.Dp(18), gtx.Dp(76), gtx.Dp(62))
	expanded := image.Rect(max(size.X-gtx.Dp(118), gtx.Dp(18)), max(size.Y-gtx.Dp(84), gtx.Dp(18)), max(size.X-gtx.Dp(18), gtx.Dp(76)), max(size.Y-gtx.Dp(18), gtx.Dp(62)))
	target := compact
	if demo.alternate {
		target = expanded
	}
	current := animation.AnimateRect("catalog-moving-rect", target).Spring(animation.SpringGentle()).Value(ctx, gtx)
	colorProgress := animation.Tween("catalog-moving-rect-color", boolFloat(demo.alternate)).Duration(420*time.Millisecond).Easing(animation.EaseCubicInOut).Value(ctx, gtx)
	guide := palette.Separator
	guide.A = 0x70
	guideRect := animation.LerpRect(compact, expanded, colorProgress)
	drawMotionOutline(gtx, guideRect, gtx.Dp(7), 1, guide)
	paint.FillShape(gtx.Ops, animation.LerpColor(palette.Accent, palette.Success, colorProgress), clip.UniformRRect(current, gtx.Dp(7)).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

func styleTransitionDemo() ui.Widget {
	return flowlayout.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		active := ctx.Theme()
		primary := active.Semantic().Tones[theme.TonePrimary]
		lift := style.New().Width(92).Height(36).
			Transition(style.PropTransform, 180*time.Millisecond, style.TransitionEase(animation.EaseCubicOut)).
			Transition(style.PropBackgroundColor, 180*time.Millisecond, style.TransitionEase(animation.EaseCubicOut)).
			On(style.Hovered, style.New().Translate(0, -3).Scale(1.04).Background(style.Solid(primary.Hover))).
			On(style.Pressed, style.New().Translate(0, 0).Scale(.96))
		morph := style.New().Width(92).Height(36).
			Transition(style.PropBackgroundColor, 220*time.Millisecond, style.TransitionEase(animation.EaseCubicInOut)).
			Transition(style.PropRadius, 220*time.Millisecond, style.TransitionEase(animation.EaseCubicInOut)).
			On(style.Hovered, style.New().Background(style.Solid(primary.Fill)).TextColor(style.Solid(primary.OnFill)).Radius(18))
		return fixedDemoStage{height: 132, child: flowlayout.Row(
			(button.New("Lift").Variant(button.Primary)).Style(lift),
			(button.New("Morph").Variant(button.Secondary)).Style(morph),
		).Gap(12)}.Layout(ctx, gtx)
	})
}

type fixedDemoStage struct {
	height int
	child  ui.Widget
}

func (stage fixedDemoStage) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := motionDemoSize(gtx, stage.height)
	fillMotionBackground(gtx, size, ctx.Theme().Palette.SurfaceTertiary)
	if stage.child == nil {
		return layout.Dimensions{Size: size}
	}
	stageGtx := gtx
	stageGtx.Constraints = layout.Exact(size)
	layout.Center.Layout(stageGtx, func(gtx layout.Context) layout.Dimensions {
		gtx.Constraints.Min = image.Point{}
		return stage.child.Layout(ctx, gtx)
	})
	return layout.Dimensions{Size: size}
}

type easingCurveSpec struct {
	label  string
	key    string
	easing animation.Easing
}

var easingCurveSpecs = [...]easingCurveSpec{
	{"Linear", "linear", animation.EaseLinear},
	{"Quadratic in", "quadratic-in", animation.EaseQuadraticIn},
	{"Quadratic out", "quadratic-out", animation.EaseQuadraticOut},
	{"Quadratic in-out", "quadratic-in-out", animation.EaseQuadraticInOut},
	{"Cubic in", "cubic-in", animation.EaseCubicIn},
	{"Cubic out", "cubic-out", animation.EaseCubicOut},
	{"Cubic in-out", "cubic-in-out", animation.EaseCubicInOut},
	{"Quartic in", "quartic-in", animation.EaseQuarticIn},
	{"Quartic out", "quartic-out", animation.EaseQuarticOut},
	{"Quartic in-out", "quartic-in-out", animation.EaseQuarticInOut},
	{"Quintic in", "quintic-in", animation.EaseQuinticIn},
	{"Quintic out", "quintic-out", animation.EaseQuinticOut},
	{"Quintic in-out", "quintic-in-out", animation.EaseQuinticInOut},
	{"Sinusoidal in", "sinusoidal-in", animation.EaseSinusoidalIn},
	{"Sinusoidal out", "sinusoidal-out", animation.EaseSinusoidalOut},
	{"Sinusoidal in-out", "sinusoidal-in-out", animation.EaseSinusoidalInOut},
	{"Exponential in", "exponential-in", animation.EaseExponentialIn},
	{"Exponential out", "exponential-out", animation.EaseExponentialOut},
	{"Exponential in-out", "exponential-in-out", animation.EaseExponentialInOut},
	{"Circular in", "circular-in", animation.EaseCircularIn},
	{"Circular out", "circular-out", animation.EaseCircularOut},
	{"Circular in-out", "circular-in-out", animation.EaseCircularInOut},
	{"Elastic in", "elastic-in", animation.EaseElasticIn},
	{"Elastic out", "elastic-out", animation.EaseElasticOut},
	{"Elastic in-out", "elastic-in-out", animation.EaseElasticInOut},
	{"Back in", "back-in", animation.EaseBackIn},
	{"Back out", "back-out", animation.EaseBackOut},
	{"Back in-out", "back-in-out", animation.EaseBackInOut},
	{"Bounce in", "bounce-in", animation.EaseBounceIn},
	{"Bounce out", "bounce-out", animation.EaseBounceOut},
	{"Bounce in-out", "bounce-in-out", animation.EaseBounceInOut},
}

func easingCurveCards(forward bool) []ui.Widget {
	cards := make([]ui.Widget, 0, len(easingCurveSpecs))
	for _, spec := range easingCurveSpecs {
		cards = append(cards, flowlayout.Column(
			text.New(spec.label).Size(13),
			(surface.New(easingCurve{key: spec.key, forward: forward, easing: spec.easing})).Style(style.New().Radius(6)),
		).Gap(6))
	}
	return cards
}

type easingCurve struct {
	key     string
	forward bool
	easing  animation.Easing
}

func (curve easingCurve) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := motionDemoSize(gtx, 112)
	timeline := animation.Tween("catalog-curve-"+curve.key, boolFloat(curve.forward)).Initial(0).Duration(850*time.Millisecond).Easing(animation.EaseLinear).Value(ctx, gtx)
	paddingX := float32(gtx.Dp(16))
	paddingY := float32(gtx.Dp(10))
	plotLeft := paddingX
	plotRight := max(float32(size.X)-paddingX, plotLeft)
	plotTop := paddingY
	plotBottom := max(float32(size.Y)-paddingY, plotTop)
	mapPoint := func(progress, value float32) f32.Point {
		return f32.Pt(animation.LerpFloat(plotLeft, plotRight, progress), animation.LerpFloat(plotBottom, plotTop, (value-curveMin)/(curveMax-curveMin)))
	}
	palette := ctx.Theme().Palette
	drawMotionLine(gtx, mapPoint(0, 0), mapPoint(1, 0), 1, palette.Separator)
	drawMotionLine(gtx, mapPoint(0, 1), mapPoint(1, 1), 1, palette.Separator)
	var path clip.Path
	path.Begin(gtx.Ops)
	for step := 0; step <= curveSamples; step++ {
		progress := float32(step) / curveSamples
		point := mapPoint(progress, curve.easing(progress))
		if step == 0 {
			path.MoveTo(point)
		} else {
			path.LineTo(point)
		}
	}
	stroke := clip.Stroke{Path: path.End(), Width: 2}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, palette.Accent)
	stroke.Pop()
	position := mapPoint(timeline, curve.easing(timeline))
	guide := palette.Accent
	guide.A = 0x40
	drawMotionLine(gtx, f32.Pt(position.X, plotTop), f32.Pt(position.X, plotBottom), 1, guide)
	diameter := max(gtx.Dp(10), 2)
	radius := float32(diameter) / 2
	markerRect := image.Rect(int(position.X-radius), int(position.Y-radius), int(position.X+radius), int(position.Y+radius))
	paint.FillShape(gtx.Ops, animation.LerpColor(palette.Accent, palette.Warning, timeline), clip.Ellipse(markerRect).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

const (
	curveMin     = float32(-0.35)
	curveMax     = float32(1.35)
	curveSamples = 96
)

func motionDemoSize(gtx layout.Context, height int) image.Point {
	return gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(unit.Dp(height))))
}

func fillMotionBackground(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(6)).Op(gtx.Ops))
}

func drawMotionDot(gtx layout.Context, center f32.Point, radius int, col color.NRGBA) {
	radius = max(radius, 1)
	rect := image.Rect(int(center.X)-radius, int(center.Y)-radius, int(center.X)+radius, int(center.Y)+radius)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(rect).Op(gtx.Ops))
}

func drawMotionLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func drawMotionOutline(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	stroke := clip.Stroke{Path: clip.UniformRRect(rect, radius).Path(gtx.Ops), Width: float32(width)}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func boolFloat(value bool) float32 {
	if value {
		return 1
	}
	return 0
}
