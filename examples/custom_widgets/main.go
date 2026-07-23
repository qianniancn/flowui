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
	"github.com/qianniancn/FlowUI/ui"
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

func Update(model *Model, msg Message) {
	switch msg.kind {
	case setOpen:
		model.Open = msg.value
	case setTransformed:
		model.Transformed = msg.value
	}
}

func View(_ *ui.Context, model Model, send ui.Send[Message]) ui.Widget {
	trigger := customTrigger{
		key:     "custom-trigger",
		label:   "Inspect workspace",
		pressed: model.Open,
		onClick: func() { send(Message{kind: setOpen, value: !model.Open}) },
	}
	portal := ui.Portal("custom-portal", model.Open, trigger, func(anchor image.Rectangle, interactive bool) ui.Widget {
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
	return ui.Center(
		ui.Column(
			ui.Text("Workspace").Size(20),
			ui.Text("main.go  |  Go  |  UTF-8").Size(13),
			portal,
			transform,
		).Gap(14).AlignMiddle(),
	)
}

type customTriggerState struct {
	click widget.Clickable
}

type customTrigger struct {
	key     string
	label   string
	pressed bool
	onClick func()
	style   ui.Style
}

func (button customTrigger) Style(value ui.Style) customTrigger {
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
	styleState := ui.StyleState{
		Hovered:      state.click.Hovered(),
		Pressed:      state.click.Pressed(),
		Focused:      focused,
		FocusVisible: ctx.FocusVisible(&state.click, focused),
		Selected:     button.pressed,
	}
	base := ui.Width(190).
		Height(36).
		PaddingX(12).
		Radius(6).
		Background(ui.TokenSurfaceSecondary).
		TextColor(ui.TokenSurfaceSecondaryForeground).
		Cursor(ui.CursorPointer).
		Outline(2, 1, ui.WithAlpha(ui.TokenFocus, 0)).
		Part(ui.PartLabel, ui.FontSize(13).MaxLines(1)).
		Transition(ui.PropBackgroundColor, 120*time.Millisecond).
		Transition(ui.PropOutlineColor, 100*time.Millisecond).
		When(ui.Hovered, ui.Background(ui.TokenSurfaceTertiary)).
		When(ui.Pressed, ui.Background(ui.TokenSurfacePressed)).
		When(ui.If(button.pressed), ui.Background(ui.TokenSurfaceTertiary)).
		When(ui.FocusVisible, ui.Outline(2, 1, ui.TokenFocus))

	resolved := ui.ResolveStyle(ctx, gtx, button.key, styleState, base, button.style)
	label := ui.ResolveStylePart(ctx, gtx, button.key, ui.PartLabel, styleState, base, button.style)
	content := ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			return ui.LayoutResolvedStyle(ctx, gtx, label, ui.Text(button.label))
		})
	})
	return ui.LayoutInteractiveResolvedStyle(ctx, gtx, resolved, content, func(gtx layout.Context, visual layout.Widget) layout.Dimensions {
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
	return ui.Key(key, ui.WidgetFunc(button.layout)).Layout(ctx, gtx)
}

func (button transformedButton) layout(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
	areaSize := gtx.Constraints.Constrain(image.Pt(gtx.Dp(260), gtx.Dp(90)))
	buttonSize := image.Pt(min(gtx.Dp(190), areaSize.X), min(gtx.Dp(36), areaSize.Y))
	position := image.Pt((areaSize.X-buttonSize.X)/2, (areaSize.Y-buttonSize.Y)/2)
	target := float32(0)
	if button.transformed {
		target = 1
	}
	progress := ui.Tween("motion", target).
		Initial(0).
		Duration(240*time.Millisecond).
		Easing(ui.EaseCubicInOut).
		Value(ctx, gtx)
	center := f32.Pt(float32(buttonSize.X)/2, float32(buttonSize.Y)/2)
	scale := ui.LerpFloat(1, 1.08, progress)
	angle := ui.LerpFloat(0, float32(-4*math.Pi/180), progress)
	offset := ui.LerpFloat(0, float32(gtx.Dp(18)), progress)
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
	return ui.WidgetFunc(func(ctx *ui.Context, gtx layout.Context) layout.Dimensions {
		if !interactive {
			gtx = gtx.Disabled()
		}
		panelGtx := gtx
		panelGtx.Constraints.Min = image.Point{}
		panelGtx.Constraints.Max = image.Pt(min(gtx.Dp(300), gtx.Constraints.Max.X), min(gtx.Dp(180), gtx.Constraints.Max.Y))
		panel := ui.Key("custom-panel", ui.Surface(
			ui.Box(
				ui.Column(
					ui.Text("Workspace details").Size(15),
					ui.Text("Project: FlowUI").Size(13),
					ui.Text("Mode: Development").Size(13),
					ui.Button("close", ui.Text("Close")).Size(ui.ButtonSmall).OnClick(close),
				).Gap(10),
			).Width(280).Padding(14),
		).Style(ui.Radius(8).Shadow(ui.ShadowSurface)))

		macro := op.Record(gtx.Ops)
		_, placement := ui.TrackOverlayPlacement(ctx, func() layout.Dimensions {
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
	ui.Run(Model{Transformed: true}, Update, View, ui.Title("FlowUI Custom Widgets"), ui.Size(760, 480))
}
