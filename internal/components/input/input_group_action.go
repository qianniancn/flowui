package input

import (
	"time"

	"gioui.org/io/event"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	layoutui "github.com/qianniancn/flowui/internal/components/layout"
	"github.com/qianniancn/flowui/internal/frame"
	"github.com/qianniancn/flowui/internal/host"
	flowstyle "github.com/qianniancn/flowui/internal/style"
)

const inputGroupActionSize = unit.Dp(24)

const inputGroupActionTransitionDuration = 100 * time.Millisecond

// InputGroupActionWidget is a compact clickable control for an input prefix
// or suffix. Its size and interaction style are stable across field heights.
type InputGroupActionWidget struct {
	box         host.BoxWidget
	onClick     func()
	focusTarget event.Tag
}

// InputGroupAction creates a compact labelled action around child content.
func InputGroupAction(key, label string, child frame.Widget) InputGroupActionWidget {
	return InputGroupActionWidget{
		box: host.Box(child).
			Key(key).
			Label(label).
			Align(layoutui.AlignCenter).
			Style(inputGroupActionStyle()),
	}
}

// OnClick sets the action callback.
func (a InputGroupActionWidget) OnClick(fn func()) InputGroupActionWidget {
	a.onClick = fn
	return a
}

// Disabled controls whether the action can be activated.
func (a InputGroupActionWidget) Disabled(disabled bool) InputGroupActionWidget {
	a.box = a.box.Disabled(disabled)
	return a
}

// Label changes the accessible action label.
func (a InputGroupActionWidget) Label(label string) InputGroupActionWidget {
	a.box = a.box.Label(label)
	return a
}

// Style adds an instance style override after the action defaults.
func (a InputGroupActionWidget) Style(value flowstyle.Style) InputGroupActionWidget {
	a.box = a.box.Style(value)
	return a
}

func (a InputGroupActionWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	return a.box.OnClick(func() {
		if a.focusTarget != nil {
			frame.RequestFocus(ctx, a.focusTarget)
		}
		if a.onClick != nil {
			a.onClick()
		}
	}).Layout(ctx, gtx)
}

func (a InputGroupActionWidget) withFocusTarget(target event.Tag) InputGroupActionWidget {
	a.focusTarget = target
	return a
}

func inputGroupActionWithFocus(widget frame.Widget, target event.Tag) frame.Widget {
	action, ok := widget.(InputGroupActionWidget)
	if !ok {
		return widget
	}
	return action.withFocusTarget(target)
}

func inputGroupActionStyle() flowstyle.Style {
	return flowstyle.Style{}.
		Width(inputGroupActionSize).
		Height(inputGroupActionSize).
		Radius(4).
		TextColor(flowstyle.TokenMutedForeground).
		Cursor(pointer.CursorPointer).
		Outline(unit.Dp(2), 0, flowstyle.WithAlpha(flowstyle.TokenFocus, 0)).
		Transition(flowstyle.PropBackgroundColor, inputGroupActionTransitionDuration).
		Transition(flowstyle.PropOutlineColor, inputGroupActionTransitionDuration).
		Transition(flowstyle.PropTextColor, inputGroupActionTransitionDuration).
		When(flowstyle.Hovered,
			flowstyle.Style{}.
				Background(flowstyle.TokenDefaultHover).
				TextColor(flowstyle.TokenForeground),
		).
		When(flowstyle.Pressed,
			flowstyle.Style{}.Background(flowstyle.TokenDefault),
		).
		When(flowstyle.All(flowstyle.FocusVisible, flowstyle.Not(flowstyle.Disabled)),
			flowstyle.Style{}.Outline(2, 0, flowstyle.TokenFocus),
		).
		When(flowstyle.Disabled,
			flowstyle.Style{}.
				Cursor(pointer.CursorDefault).
				Opacity(0.45),
		)
}
