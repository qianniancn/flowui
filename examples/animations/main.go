package main

import (
	"image"
	"image/color"
	"time"

	"gioui.org/f32"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"github.com/qianniancn/flowui/animation"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
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
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	action := "Reverse"
	if !model.forward {
		action = "Forward"
	}
	return flowlayout.Scroll(flowlayout.Box(
		flowlayout.Column(
			flowlayout.Row(
				flowlayout.Expanded(text.New("Motion").Size(24)),
				button.New(action).Key("toggle-direction").OnClick(func() { send(ToggleDirection{}) }),
			).AlignMiddle(),
			flowlayout.Divider(),
			text.New("Animation primitives").Size(18),
			flowlayout.AutoGrid(300, motionDemoCards(model, send)...).ColumnGap(16).RowGap(16),
			flowlayout.Divider(),
			text.New("Easing curves").Size(18),
			flowlayout.AutoGrid(200, easingCurveCards(model.forward)...).ColumnGap(16).RowGap(12),
		).Gap(16),
	).Style(style.New().FillWidth().MaxWidth(1040).Padding(24)),
	).Vertical()
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
		cards = append(cards, curveCard(spec.label, easingCurve{
			key:     spec.key,
			forward: forward,
			easing:  spec.easing,
		}))
	}
	return cards
}

func curveCard(label string, curve ui.Widget) ui.Widget {
	return flowlayout.Column(
		text.New(label).Size(13),
		(surface.New(curve)).Style(style.New().Radius(6)),
	).Gap(6)
}

type easingCurve struct {
	key     string
	forward bool
	easing  animation.Easing
}

func (curve easingCurve) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(112)))
	timeline := animation.Tween("curve-"+curve.key, boolFloat(curve.forward)).
		Initial(0).
		Duration(850*time.Millisecond).
		Easing(animation.EaseLinear).
		Value(ctx, gtx)

	paddingX := float32(gtx.Dp(16))
	paddingY := float32(gtx.Dp(10))
	plotLeft := paddingX
	plotRight := max(float32(size.X)-paddingX, plotLeft)
	plotTop := paddingY
	plotBottom := max(float32(size.Y)-paddingY, plotTop)

	mapPoint := func(progress, value float32) f32.Point {
		return f32.Pt(
			animation.LerpFloat(plotLeft, plotRight, progress),
			animation.LerpFloat(plotBottom, plotTop, (value-curveMin)/(curveMax-curveMin)),
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
	markerColor := animation.LerpColor(
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
	ui.Run(ui.NewProgram(Model{forward: true, timelinePlaying: true},
		Update, View), window.Title("FlowUI Motion"),
		window.Size(1100, 760),
		window.Theme(base.Light()),
	)
}
