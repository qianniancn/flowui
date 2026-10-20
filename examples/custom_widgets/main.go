package main

import (
	"image"
	"math"
	"time"

	"gioui.org/f32"
	"gioui.org/io/semantic"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/animation"
	flowlayout "github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/portal"
	"github.com/qianniancn/flowui/widget/surface"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Open        bool
	Transformed bool
}

type messageKind uint8

const (
	setOpen messageKind = iota
	setTransformed
)

type Message struct {
	kind  messageKind
	value bool
}

func Update(model *Model, msg Message) ui.Cmd[Message] {
	switch msg.kind {
	case setOpen:
		model.Open = msg.value
	case setTransformed:
		model.Transformed = msg.value
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Message]) ui.Widget {
	trigger := customTrigger{
		key:     "custom-trigger",
		label:   "Inspect workspace",
		pressed: model.Open,
		onClick: func() { send(Message{kind: setOpen, value: !model.Open}) },
	}
	popup := portal.New("custom-portal", model.Open, trigger, func(anchor image.Rectangle, interactive bool) ui.Widget {
		return portalPanel(anchor, interactive, func() { send(Message{kind: setOpen}) })
	})
	transform := transformedButton{
		button: customTrigger{
			key:     "transform-button",
			label:   "Apply transform",
			pressed: model.Transformed,
			onClick: func() { send(Message{kind: setTransformed, value: !model.Transformed}) },
		},
		transformed: model.Transformed,
	}
	if model.Transformed {
		transform.button.label = "Reset transform"
	}
	return flowlayout.Center(
		flowlayout.Column(
			text.New("Workspace").Size(20),
			text.New("main.go  |  Go  |  UTF-8").Size(13),
			popup,
			transform,
		).Gap(14).AlignMiddle(),
	)
}

type customTriggerState struct {
	click widget.Clickable
}

// customTrigger 是手写组件：行为只持有状态，外观全部来自 style.Style。
type customTrigger struct {
	key     string
	label   string
	pressed bool
	onClick func()
	style   style.Style
}

func (button customTrigger) Style(value style.Style) customTrigger {
	button.style = value
	return button
}

func (button customTrigger) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	state := ui.UseState[customTriggerState](ctx, button.key)
	before := activePresses(state.click.History())
	for state.click.Clicked(gtx) {
		if button.onClick != nil {
			button.onClick()
		}
	}
	if activePresses(state.click.History()) > before {
		ctx.RequestFocusVisible(&state.click, false)
	}

	focused := gtx.Focused(&state.click)
	visual := style.State{
		Hovered:      state.click.Hovered(),
		Pressed:      state.click.Pressed(),
		Focused:      focused,
		FocusVisible: ctx.FocusVisible(&state.click, focused),
		Selected:     button.pressed,
	}
	active := ctx.Theme()
	sem := active.Semantic()
	base := style.New().
		Width(190).
		Height(36).
		PaddingX(12).
		Radius(6).
		Background(style.Solid(sem.SurfaceSecondary)).
		TextColor(style.Solid(sem.SurfaceSecondaryForeground)).
		Cursor(style.CursorPointer).
		Outline(2, 1, style.WithAlpha(sem.Ring, 0)).
		Part(style.PartLabel, style.New().FontSize(13).MaxLines(1)).
		Transition(style.PropBackgroundColor, 120*time.Millisecond).
		Transition(style.PropOutlineColor, 100*time.Millisecond).
		On(style.Hovered, style.New().Background(style.Solid(sem.SurfaceTertiary))).
		On(style.Pressed, style.New().Background(style.Solid(sem.SurfacePressed))).
		On(style.Selected, style.New().Background(style.Solid(sem.SurfaceTertiary))).
		On(style.FocusVisible, style.New().Outline(2, 1, style.Solid(sem.Ring)))

	resolved := style.Cascade(visual, base, button.style)
	label := style.CascadePart(visual, style.PartLabel, base, button.style)
	content := flowlayout.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		return flowlayout.Center(
			flowlayout.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
				return flowlayout.Resolved(ctx, gtx, label, text.New(button.label))
			}),
		).Layout(ctx, gtx)
	})
	return flowlayout.InteractiveResolved(ctx, gtx, resolved, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
		return state.click.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			semantic.Button.Add(gtx.Ops)
			semantic.LabelOp(button.label).Add(gtx.Ops)
			return visual(gtx)
		})
	})
}

type transformedButton struct {
	button      customTrigger
	transformed bool
}

func (button transformedButton) Layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	key := button.button.key
	button.button.key = "button"
	return flowlayout.Key(key, flowlayout.WidgetFunc(button.layout)).Layout(ctx, gtx)
}

func (button transformedButton) layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	areaSize := gtx.Constraints.Constrain(image.Pt(gtx.Dp(260), gtx.Dp(90)))
	buttonSize := image.Pt(min(gtx.Dp(190), areaSize.X), min(gtx.Dp(36), areaSize.Y))
	position := image.Pt((areaSize.X-buttonSize.X)/2, (areaSize.Y-buttonSize.Y)/2)
	target := float32(0)
	if button.transformed {
		target = 1
	}
	progress := animation.Tween("motion", target).
		Initial(0).
		Duration(240*time.Millisecond).
		Easing(animation.EaseCubicInOut).
		Value(ctx, gtx)
	center := f32.Pt(float32(buttonSize.X)/2, float32(buttonSize.Y)/2)
	scale := animation.LerpFloat(1, 1.08, progress)
	angle := animation.LerpFloat(0, float32(-4*math.Pi/180), progress)
	offset := animation.LerpFloat(0, float32(gtx.Dp(18)), progress)
	local := f32.AffineId().
		Scale(center, f32.Pt(scale, scale)).
		Rotate(center, angle).
		Offset(f32.Pt(offset, 0))
	transform := f32.AffineId().
		Offset(f32.Pt(float32(position.X), float32(position.Y))).
		Mul(local)

	childGtx := gtx
	childGtx.Constraints = layout.Exact(buttonSize)
	stack := op.Affine(transform).Push(gtx.Ops)
	button.button.Layout(ctx, childGtx)
	stack.Pop()
	return layout.Dimensions{Size: areaSize}
}

func portalPanel(anchor image.Rectangle, interactive bool, close func()) ui.Widget {
	return flowlayout.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		if !interactive {
			gtx = gtx.Disabled()
		}
		panelGtx := gtx
		panelGtx.Constraints.Min = image.Point{}
		panelGtx.Constraints.Max = image.Pt(min(gtx.Dp(300), gtx.Constraints.Max.X), min(gtx.Dp(180), gtx.Constraints.Max.Y))
		panel := flowlayout.Key("custom-panel", (surface.New(
			flowlayout.Box(
				flowlayout.Column(
					text.New("Workspace details").Size(15),
					text.New("Project: FlowUI").Size(13),
					text.New("Mode: Development").Size(13),
					button.New("Close").Size(button.Small).OnClick(close),
				).Gap(10),
			).Style(style.New().Width(280).Padding(14)),
		)).Style(style.New().Radius(8).Shadow(style.ShadowSurface)))

		macro := op.Record(gtx.Ops)
		_, placement := flowlayout.TrackOverlayPlacement(ctx, func() layout.Dimensions {
			return panel.Layout(ctx, panelGtx)
		})
		call := macro.Stop()
		position := image.Pt(anchor.Min.X, anchor.Max.Y+gtx.Dp(8))
		placement.PlaceOffset(position)
		offset := op.Offset(position).Push(gtx.Ops)
		call.Add(gtx.Ops)
		offset.Pop()
		return layout.Dimensions{Size: gtx.Constraints.Max}
	})
}

func activePresses(history []widget.Press) int {
	count := 0
	for _, press := range history {
		if press.End.IsZero() && !press.Cancelled {
			count++
		}
	}
	return count
}

func main() {
	ui.Run(ui.NewProgram(Model{Transformed: true},
		Update, View), window.Title("FlowUI Custom Widgets"), window.Size(760, 480), window.Theme(base.Light()))
}
