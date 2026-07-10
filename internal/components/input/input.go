package input

import (
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
)

type InputWidget struct {
	key       string
	value     string
	hint      string
	onChange  func(string)
	onSubmit  func(string)
	variant   InputVariant
	disabled  bool
	invalid   bool
	fullWidth bool
}

type InputVariant = field.Variant

const (
	InputPrimary   = field.Primary
	InputSecondary = field.Secondary
)

const (
	inputHeight   = unit.Dp(40)
	inputRadius   = unit.Dp(12)
	inputTextSize = unit.Sp(14)
)

const stateSlotInput = "input"

func inputStateFor(ctx *frame.Context, key string) *field.State {
	return frame.UseState[field.State](ctx, key, stateSlotInput)
}

func Input(key, value string) InputWidget {
	return InputWidget{key: key, value: value}
}

func (i InputWidget) Hint(hint string) InputWidget {
	i.hint = hint
	return i
}

func (i InputWidget) OnChange(fn func(string)) InputWidget {
	i.onChange = fn
	return i
}

func (i InputWidget) OnSubmit(fn func(string)) InputWidget {
	i.onSubmit = fn
	return i
}

func (i InputWidget) Disabled(disabled bool) InputWidget {
	i.disabled = disabled
	return i
}

func (i InputWidget) Invalid(invalid bool) InputWidget {
	i.invalid = invalid
	return i
}

func (i InputWidget) Variant(variant InputVariant) InputWidget {
	i.variant = variant
	return i
}

func (i InputWidget) FullWidth() InputWidget {
	i.fullWidth = true
	return i
}

func (i InputWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, editor := frame.InputEditor(ctx, i.key)
	frame.RegisterFieldFocus(ctx, key, editor, gtx.Enabled() && !i.disabled)
	state := inputStateFor(ctx, key)
	state.Update(ctx, gtx, i.disabled, editor)

	editor.SingleLine = true
	editor.Submit = i.onSubmit != nil
	if i.disabled {
		gtx = gtx.Disabled()
	}

	var events inputEvents
	for {
		event, ok := editor.Update(gtx)
		if !ok {
			break
		}
		events.add(event)
	}
	if events.changed || events.submitted {
		i.dispatchEvents(editor, events)
	} else if editor.Text() != i.value {
		editor.SetText(i.value)
	}

	focused := gtx.Focused(editor)
	style := field.ResolveStyle(frame.ActiveTheme(ctx), i.variant, state.Hovered, focused, i.disabled, i.invalid)
	style.Background = state.Background(gtx, style.Background)
	style.Border = state.BorderColor(gtx, style.Border)
	editorStyle := material.Editor(frame.ActiveTheme(ctx).Material, editor, i.hint)
	editorStyle.TextSize = frame.ActiveTheme(ctx).Typography.ControlSize
	editorStyle.Color = style.Foreground
	editorStyle.HintColor = style.Placeholder
	editorStyle.SelectionColor = style.Selection

	return i.layoutFrame(ctx, gtx, state, style, frame.WithFieldSemantics(ctx, key, editorStyle.Layout))
}
