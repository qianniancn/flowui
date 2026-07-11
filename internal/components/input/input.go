package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
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
	inputType InputType
	readOnly  bool
	maxLength int
	label     string
}

type InputVariant = field.Variant

const (
	InputPrimary   = field.Primary
	InputSecondary = field.Secondary
)

type InputType uint8

const (
	InputText InputType = iota
	InputEmail
	InputNumber
	InputPassword
)

const stateSlotInput = "input"

func Input(key, value string) InputWidget {
	return InputWidget{key: key, value: value}
}

func (i InputWidget) Hint(hint string) InputWidget {
	i.hint = hint
	return i
}

func (i InputWidget) Placeholder(placeholder string) InputWidget {
	i.hint = placeholder
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

func (i InputWidget) Type(inputType InputType) InputWidget {
	i.inputType = inputType
	return i
}

func (i InputWidget) ReadOnly(readOnly bool) InputWidget {
	i.readOnly = readOnly
	return i
}

func (i InputWidget) MaxLength(maxLength int) InputWidget {
	i.maxLength = max(maxLength, 0)
	return i
}

func (i InputWidget) Label(label string) InputWidget {
	i.label = label
	return i
}

func (i InputWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	key, editor := frame.InputEditor(ctx, i.key)
	enabled := gtx.Enabled() && !i.disabled
	disabled := !enabled
	frame.RegisterFieldFocus(ctx, key, editor, enabled)
	state := inputStateFor(ctx, key)
	state.State.Update(ctx, gtx, disabled, editor)

	editor.SingleLine = true
	editor.Submit = i.onSubmit != nil
	editor.ReadOnly = i.readOnly
	editor.MaxLen = i.maxLength
	editor.Mask, editor.InputHint, editor.Filter = inputTypeConfig(i.inputType)
	if disabled {
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
	style := inputStyleFor(frame.ActiveTheme(ctx), i.variant, state.Hovered, focused, disabled, i.invalid)
	style.Background = state.Background(gtx, style.Background)
	style.Ring = state.BorderColor(gtx, style.Ring)
	editorStyle := material.Editor(frame.ActiveTheme(ctx).Material, editor, i.hint)
	editorStyle.TextSize = frame.ActiveTheme(ctx).Components.Input.TextSize
	editorStyle.LineHeight = frame.ActiveTheme(ctx).Components.Input.LineHeight
	editorStyle.Color = style.Foreground
	editorStyle.HintColor = style.Placeholder
	editorStyle.SelectionColor = style.Selection

	return i.layoutFrame(ctx, gtx, state, style, enabled, i.withSemantics(ctx, key, enabled, editorStyle.Layout))
}

func inputTypeConfig(inputType InputType) (rune, key.InputHint, string) {
	switch inputType {
	case InputEmail:
		return 0, key.HintEmail, ""
	case InputNumber:
		return 0, key.HintNumeric, "0123456789+-.eE"
	case InputPassword:
		return '\u2022', key.HintPassword, ""
	default:
		return 0, key.HintText, ""
	}
}
