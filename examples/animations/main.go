package main

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	forward         bool
	timelinePlaying bool
	timelineRun     uint64
	layoutExpanded  bool
	rectAlternate   bool
}

type Msg interface{ animationMsg() }

type ToggleDirection struct{}
type ToggleTimeline struct{}
type RestartTimeline struct{}
type ToggleLayout struct{}
type ToggleRect struct{}

func (ToggleDirection) animationMsg() {}
func (ToggleTimeline) animationMsg()  {}
func (RestartTimeline) animationMsg() {}
func (ToggleLayout) animationMsg()    {}
func (ToggleRect) animationMsg()      {}

func Update(model *Model, msg Msg) {
	switch msg.(type) {
	case ToggleDirection:
		model.forward = !model.forward
	case ToggleTimeline:
		model.timelinePlaying = !model.timelinePlaying
	case RestartTimeline:
		model.timelineRun++
	case ToggleLayout:
		model.layoutExpanded = !model.layoutExpanded
	case ToggleRect:
		model.rectAlternate = !model.rectAlternate
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	action := "Reverse"
	if !model.forward {
		action = "Forward"
	}
	return ui.Scroll("animation-curves",
		ui.Box(
			ui.Column(
				ui.Row(
					ui.Expanded(ui.Text("Motion").Size(24)),
					ui.Button("toggle-direction", ui.Text(action)).OnClick(func() { send(ToggleDirection{}) }),
				).AlignMiddle(),
				ui.Divider(),
				ui.Text("Animation primitives").Size(18),
				ui.AutoGrid(300, motionDemoCards(model, send)...).ColumnGap(16).RowGap(16),
				ui.Divider(),
				ui.Text("Easing curves").Size(18),
				ui.AutoGrid(200, easingCurveCards(model.forward)...).ColumnGap(16).RowGap(12),
			).Gap(16),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(1040)).Style(ui.Padding(24)),
	).Vertical()
}

type easingCurveSpec struct {
	label  string
	key    string
	easing ui.Easing
}

var easingCurveSpecs = [...]easingCurveSpec{
	{"Linear", "linear", ui.EaseLinear},
	{"Quadratic in", "quadratic-in", ui.EaseQuadraticIn},
	{"Quadratic out", "quadratic-out", ui.EaseQuadraticOut},
	{"Quadratic in-out", "quadratic-in-out", ui.EaseQuadraticInOut},
	{"Cubic in", "cubic-in", ui.EaseCubicIn},
	{"Cubic out", "cubic-out", ui.EaseCubicOut},
	{"Cubic in-out", "cubic-in-out", ui.EaseCubicInOut},
	{"Quartic in", "quartic-in", ui.EaseQuarticIn},
	{"Quartic out", "quartic-out", ui.EaseQuarticOut},
	{"Quartic in-out", "quartic-in-out", ui.EaseQuarticInOut},
	{"Quintic in", "quintic-in", ui.EaseQuinticIn},
	{"Quintic out", "quintic-out", ui.EaseQuinticOut},
	{"Quintic in-out", "quintic-in-out", ui.EaseQuinticInOut},
	{"Sinusoidal in", "sinusoidal-in", ui.EaseSinusoidalIn},
	{"Sinusoidal out", "sinusoidal-out", ui.EaseSinusoidalOut},
	{"Sinusoidal in-out", "sinusoidal-in-out", ui.EaseSinusoidalInOut},
	{"Exponential in", "exponential-in", ui.EaseExponentialIn},
	{"Exponential out", "exponential-out", ui.EaseExponentialOut},
	{"Exponential in-out", "exponential-in-out", ui.EaseExponentialInOut},
	{"Circular in", "circular-in", ui.EaseCircularIn},
	{"Circular out", "circular-out", ui.EaseCircularOut},
	{"Circular in-out", "circular-in-out", ui.EaseCircularInOut},
	{"Elastic in", "elastic-in", ui.EaseElasticIn},
	{"Elastic out", "elastic-out", ui.EaseElasticOut},
	{"Elastic in-out", "elastic-in-out", ui.EaseElasticInOut},
	{"Back in", "back-in", ui.EaseBackIn},
	{"Back out", "back-out", ui.EaseBackOut},
	{"Back in-out", "back-in-out", ui.EaseBackInOut},
	{"Bounce in", "bounce-in", ui.EaseBounceIn},
	{"Bounce out", "bounce-out", ui.EaseBounceOut},
	{"Bounce in-out", "bounce-in-out", ui.EaseBounceInOut},
}

func easingCurveCards(forward bool) []ui.Widget {
	cards := make([]ui.Widget, 0, len(easingCurveSpecs))
	for _, spec := range easingCurveSpecs {
		cards = append(cards, curveCard(spec.label, easingCurve{
			key:     spec.key,
			forward: forward,
			easing:  spec.easing,
		}))
	}
	return cards
}

func curveCard(label string, curve ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(label).Size(13),
		ui.Surface(curve).Style(ui.Radius(6)),
	).Gap(6)
}

type easingCurve struct {
	key     string
	forward bool
	easing  ui.Easing
}

func (curve easingCurve) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(112)))
	timeline := ui.Tween("curve-"+curve.key, boolFloat(curve.forward)).
		Initial(0).
		Duration(850*time.Millisecond).
		Easing(ui.EaseLinear).
		Value(ctx, gtx)

	paddingX := float32(gtx.Dp(16))
	paddingY := float32(gtx.Dp(10))
	plotLeft := paddingX
	plotRight := max(float32(size.X)-paddingX, plotLeft)
	plotTop := paddingY
	plotBottom := max(float32(size.Y)-paddingY, plotTop)

	mapPoint := func(progress, value float32) f32.Point {
		return f32.Pt(
			ui.LerpFloat(plotLeft, plotRight, progress),
			ui.LerpFloat(plotBottom, plotTop, (value-curveMin)/(curveMax-curveMin)),
		)
	}

	palette := ctx.Theme().Palette
	gridColor := palette.Separator
	drawCurveLine(gtx, mapPoint(0, 0), mapPoint(1, 0), 1, gridColor)
	drawCurveLine(gtx, mapPoint(0, 1), mapPoint(1, 1), 1, gridColor)

	curveColor := palette.Accent
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
	paint.Fill(gtx.Ops, curveColor)
	stroke.Pop()

	position := mapPoint(timeline, curve.easing(timeline))
	guideColor := curveColor
	guideColor.A = 0x40
	drawCurveLine(gtx, f32.Pt(position.X, plotTop), f32.Pt(position.X, plotBottom), 1, guideColor)
	markerColor := ui.LerpColor(
		curveColor,
		palette.Warning,
		timeline,
	)
	diameter := max(gtx.Dp(10), 2)
	radius := float32(diameter) / 2
	markerRect := image.Rect(
		int(position.X-radius),
		int(position.Y-radius),
		int(position.X+radius),
		int(position.Y+radius),
	)
	paint.FillShape(gtx.Ops, markerColor, clip.Ellipse(markerRect).Op(gtx.Ops))
	return layout.Dimensions{Size: size}
}

const (
	curveMin     = float32(-0.35)
	curveMax     = float32(1.35)
	curveSamples = 96
)

func drawCurveLine(gtx layout.Context, from, to f32.Point, width float32, col color.NRGBA) {
	var path clip.Path
	path.Begin(gtx.Ops)
	path.MoveTo(from)
	path.LineTo(to)
	stroke := clip.Stroke{Path: path.End(), Width: width}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}

func boolFloat(value bool) float32 {
	if value {
		return 1
	}
	return 0
}

func main() {
	ui.Run(
		Model{forward: true, timelinePlaying: true},
		Update,
		View,
		ui.Title("FlowUI Motion"),
		ui.Size(1100, 760),
	)
}
