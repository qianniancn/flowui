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
	"github.com/qianniancn/flowui/ui"
)

func motionDemoCards(model Model, send ui.Send[Msg]) []ui.Widget {
	playIcon := lucide.Pause
	playLabel := "Pause timeline"
	if !model.timelinePlaying {
		playIcon = lucide.Play
		playLabel = "Play timeline"
	}
	layoutIcon := lucide.Maximize2
	layoutLabel := "Expand layout"
	if model.layoutExpanded {
		layoutIcon = lucide.Minimize2
		layoutLabel = "Collapse layout"
	}

	return []ui.Widget{
		motionDemoCard("Tween and interpolation", nil, tweenTransformDemo{forward: model.forward}),
		motionDemoCard("Spring presets", nil, springComparison(model.forward)),
		motionDemoCard(
			"Keyframe timeline",
			ui.Row(
				motionIconButton("timeline-play", playIcon, playLabel, func() { send(ToggleTimeline{}) }),
				motionIconButton("timeline-restart", lucide.RotateCcw, "Restart timeline", func() { send(RestartTimeline{}) }),
			).Gap(4),
			timelineDemo{playing: model.timelinePlaying, revision: model.timelineRun},
		),
		motionDemoCard(
			"Animated layout",
			motionIconButton("layout-toggle", layoutIcon, layoutLabel, func() { send(ToggleLayout{}) }),
			fixedDemoStage{
				height: 132,
				child: ui.AnimateLayout("layout-size", layoutTarget(model.layoutExpanded)).
					Spring(ui.SpringGentle()),
			},
		),
		motionDemoCard(
			"Animated rectangle",
			motionIconButton("rect-toggle", lucide.RefreshCw, "Move rectangle", func() { send(ToggleRect{}) }),
			rectDemo{alternate: model.rectAlternate},
		),
		motionDemoCard("Style transitions", nil, styleTransitionDemo()),
	}
}

func motionDemoCard(title string, actions, content ui.Widget) ui.Widget {
	header := []ui.Widget{ui.Expanded(ui.Text(title).Size(14))}
	if actions != nil {
		header = append(header, actions)
	}
	return ui.Card(
		ui.Row(header...).AlignMiddle(),
		ui.Divider(),
		content,
	).
		Variant(ui.CardSecondary).
		Padding(14).
		Gap(12).
		Style(ui.MinHeight(220))
}

func motionIconButton(key string, icon lucide.Data, label string, onClick func()) ui.Widget {
	return ui.Button(key, ui.Icon(icon).Size(14)).
		Variant(ui.ButtonGhost).
		Size(ui.ButtonSmall).
		IconOnly().
		Label(label).
		OnClick(onClick)
}

type tweenTransformDemo struct {
	forward bool
}

func (demo tweenTransformDemo) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := demoSize(gtx, 132)
	fillDemoBackground(gtx, size, ctx.Theme().Palette.SurfaceTertiary)
	if size.X <= 0 || size.Y <= 0 {
		return layout.Dimensions{Size: size}
	}

	progress := ui.Tween("transform-progress", boolFloat(demo.forward)).
		Initial(0).
		Duration(650*time.Millisecond).
		Easing(ui.EaseCubicInOut).
		Value(ctx, gtx)

	left := f32.Pt(float32(gtx.Dp(28)), float32(size.Y)/2)
	right := f32.Pt(float32(max(size.X-gtx.Dp(28), gtx.Dp(28))), float32(size.Y)/2)
	point := ui.LerpPoint(left, right, progress)
	palette := ctx.Theme().Palette
	drawCurveLine(gtx, left, right, 2, palette.Separator)
	drawDemoDot(gtx, left, gtx.Dp(4), palette.Separator)
	drawDemoDot(gtx, right, gtx.Dp(4), palette.Separator)

	blockSize := max(gtx.Dp(28), 1)
	half := blockSize / 2
	block := image.Rect(int(point.X)-half, int(point.Y)-half, int(point.X)+half, int(point.Y)+half)
	scale := ui.LerpFloat(.82, 1.18, progress)
	angle := ui.LerpFloat(0, float32(math.Pi*1.5), progress)
	transform := f32.AffineId().
		Scale(point, f32.Pt(scale, scale)).
		Rotate(point, angle)
	stack := op.Affine(transform).Push(gtx.Ops)
	paint.FillShape(
		gtx.Ops,
		ui.LerpColor(palette.Accent, palette.Success, progress),
		clip.UniformRRect(block, gtx.Dp(5)).Op(gtx.Ops),
	)
	stack.Pop()
	return layout.Dimensions{Size: size}
}

type springDemoSpec struct {
	key    string
	label  string
	config ui.SpringConfig
	mix    float32
}

var springDemoSpecs = [...]springDemoSpec{
	{key: "default", label: "Default", config: ui.DefaultSpring(), mix: 0},
	{key: "snappy", label: "Snappy", config: ui.SpringSnappy(), mix: .33},
	{key: "gentle", label: "Gentle", config: ui.SpringGentle(), mix: .66},
	{key: "bouncy", label: "Bouncy", config: ui.SpringBouncy(), mix: 1},
}

func springComparison(forward bool) ui.Widget {
	rows := make([]ui.Widget, 0, len(springDemoSpecs))
	for _, spec := range springDemoSpecs {
		rows = append(rows, ui.Row(
			ui.Box(ui.Text(spec.label).Size(12)).Style(ui.Width(64)),
			ui.Expanded(springLane{spec: spec, forward: forward}),
		).AlignMiddle().Gap(8))
	}
	return fixedDemoStage{height: 132, child: ui.Column(rows...).Gap(4)}
}

type springLane struct {
	spec    springDemoSpec
	forward bool
}

func (lane springLane) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := demoSize(gtx, 26)
	left := float32(gtx.Dp(18))
	right := float32(max(size.X-gtx.Dp(18), gtx.Dp(18)))
	target := left
	if lane.forward {
		target = right
	}
	x := ui.Tween("spring-"+lane.spec.key, target).
		Initial(left).
		Spring(lane.spec.config).
		Value(ctx, gtx)
	y := float32(size.Y) / 2
	palette := ctx.Theme().Palette
	drawCurveLine(gtx, f32.Pt(left, y), f32.Pt(right, y), 2, palette.Separator)
	drawDemoDot(gtx, f32.Pt(x, y), gtx.Dp(6), ui.LerpColor(palette.Accent, palette.Warning, lane.spec.mix))
	return layout.Dimensions{Size: size}
}

type timelineDemo struct {
	playing  bool
	revision uint64
}

func (demo timelineDemo) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := demoSize(gtx, 132)
	palette := ctx.Theme().Palette
	fillDemoBackground(gtx, size, palette.SurfaceTertiary)
	value := ui.Timeline("keyframe-track").
		Keyframe(0, 0).
		Keyframe(.18, .9).
		Keyframe(.42, .25).
		Keyframe(.7, 1).
		Keyframe(1, 0).
		Duration(2400*time.Millisecond).
		Easing(ui.EaseCubicInOut).
		Loop(true).
		Playing(demo.playing).
		Revision(demo.revision).
		Value(ctx, gtx)

	left := float32(gtx.Dp(24))
	right := float32(max(size.X-gtx.Dp(24), gtx.Dp(24)))
	y := float32(size.Y) / 2
	drawCurveLine(gtx, f32.Pt(left, y), f32.Pt(right, y), 2, palette.Separator)
	for _, stop := range []float32{0, .25, .5, .75, 1} {
		x := ui.LerpFloat(left, right, stop)
		drawCurveLine(gtx, f32.Pt(x, y-float32(gtx.Dp(8))), f32.Pt(x, y+float32(gtx.Dp(8))), 1, palette.Separator)
	}
	point := f32.Pt(ui.LerpFloat(left, right, value), y)
	drawDemoDot(gtx, point, gtx.Dp(9), ui.LerpColor(palette.Accent, palette.Warning, value))
	return layout.Dimensions{Size: size}
}

func layoutTarget(expanded bool) ui.Widget {
	width, height := 124, 52
	label := "Compact"
	if expanded {
		width, height = 220, 108
		label = "Expanded"
	}
	return ui.Surface(
		ui.Box(ui.Text(label).Size(13)).Style(ui.Width(unit.Dp(width)).Height(unit.Dp(height)).Padding(12)),
	).
		Variant(ui.SurfaceTertiary).
		Style(ui.Radius(6))
}

type rectDemo struct {
	alternate bool
}

func (demo rectDemo) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := demoSize(gtx, 132)
	palette := ctx.Theme().Palette
	fillDemoBackground(gtx, size, palette.SurfaceTertiary)
	compact := image.Rect(gtx.Dp(18), gtx.Dp(18), gtx.Dp(76), gtx.Dp(62))
	expanded := image.Rect(
		max(size.X-gtx.Dp(118), gtx.Dp(18)),
		max(size.Y-gtx.Dp(84), gtx.Dp(18)),
		max(size.X-gtx.Dp(18), gtx.Dp(76)),
		max(size.Y-gtx.Dp(18), gtx.Dp(62)),
	)
	target := compact
	if demo.alternate {
		target = expanded
	}
	current := ui.AnimateRect("moving-rect", target).
		Spring(ui.SpringGentle()).
		Value(ctx, gtx)
	colorProgress := ui.Tween("moving-rect-color", boolFloat(demo.alternate)).
		Duration(420*time.Millisecond).
		Easing(ui.EaseCubicInOut).
		Value(ctx, gtx)
	guide := palette.Separator
	guide.A = 0x70
	drawDemoOutline(gtx, target, gtx.Dp(7), 1, guide)
	paint.FillShape(
		gtx.Ops,
		ui.LerpColor(palette.Accent, palette.Success, colorProgress),
		clip.UniformRRect(current, gtx.Dp(7)).Op(gtx.Ops),
	)
	return layout.Dimensions{Size: size}
}

func styleTransitionDemo() ui.Widget {
	lift := ui.Width(92).Height(36).
		Transition(ui.PropTransform, 180*time.Millisecond, ui.TransitionEase(ui.EaseCubicOut)).
		Transition(ui.PropBackgroundColor, 180*time.Millisecond, ui.TransitionEase(ui.EaseCubicOut)).
		When(ui.Hovered, ui.Translate(0, -3).
			Scale(1.04, 1.04).
			Background(ui.TokenAccentHover)).
		When(ui.Pressed, ui.Translate(0, 0).Scale(.96, .96))
	morph := ui.Width(92).Height(36).
		Transition(ui.PropBackgroundColor, 220*time.Millisecond, ui.TransitionEase(ui.EaseCubicInOut)).
		Transition(ui.PropRadius, 220*time.Millisecond, ui.TransitionEase(ui.EaseCubicInOut)).
		When(ui.Hovered, ui.Background(ui.TokenAccent).
			TextColor(ui.TokenAccentForeground).
			Radius(18))
	return fixedDemoStage{
		height: 132,
		child: ui.Row(
			ui.Button("transition-lift", ui.Text("Lift")).
				Variant(ui.ButtonPrimary).
				Style(lift),
			ui.Button("transition-morph", ui.Text("Morph")).
				Variant(ui.ButtonSecondary).
				Style(morph),
		).Gap(12),
	}
}

type fixedDemoStage struct {
	height int
	child  ui.Widget
}

func (stage fixedDemoStage) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	size := demoSize(gtx, stage.height)
	fillDemoBackground(gtx, size, ctx.Theme().Palette.SurfaceTertiary)
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

func demoSize(gtx layout.Context, height int) image.Point {
	return gtx.Constraints.Constrain(image.Pt(gtx.Constraints.Max.X, gtx.Dp(unit.Dp(height))))
}

func fillDemoBackground(gtx layout.Context, size image.Point, col color.NRGBA) {
	if size.X <= 0 || size.Y <= 0 || col.A == 0 {
		return
	}
	paint.FillShape(gtx.Ops, col, clip.UniformRRect(image.Rectangle{Max: size}, gtx.Dp(6)).Op(gtx.Ops))
}

func drawDemoDot(gtx layout.Context, center f32.Point, radius int, col color.NRGBA) {
	radius = max(radius, 1)
	rect := image.Rect(int(center.X)-radius, int(center.Y)-radius, int(center.X)+radius, int(center.Y)+radius)
	paint.FillShape(gtx.Ops, col, clip.Ellipse(rect).Op(gtx.Ops))
}

func drawDemoOutline(gtx layout.Context, rect image.Rectangle, radius, width int, col color.NRGBA) {
	if rect.Empty() || width <= 0 || col.A == 0 {
		return
	}
	stroke := clip.Stroke{
		Path:  clip.UniformRRect(rect, radius).Path(gtx.Ops),
		Width: float32(width),
	}.Op().Push(gtx.Ops)
	paint.Fill(gtx.Ops, col)
	stroke.Pop()
}
