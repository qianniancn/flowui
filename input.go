package flowui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget/material"
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

type InputVariant int

const (
	InputPrimary InputVariant = iota
	InputSecondary
)

const (
	inputColorDuration = 150 * time.Millisecond
	inputHeight        = unit.Dp(40)
	inputRadius        = unit.Dp(12)
	inputTextSize      = unit.Sp(14)
)

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

func (i InputWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	key, editor := ctx.inputEditor(i.key)
	state := ctx.inputState(key)
	state.update(ctx, gtx, i.disabled, editor)

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
	style := inputStyleFor(ctx.Theme, i.variant, state.hovered, focused, i.disabled, i.invalid)
	style.bg = state.background(gtx, style.bg)
	style.border = state.borderColor(gtx, style.border)
	editorStyle := material.Editor(ctx.Theme.Material, editor, i.hint)
	editorStyle.TextSize = ctx.Theme.Typography.ControlSize
	editorStyle.Color = style.fg
	editorStyle.HintColor = style.placeholder
	editorStyle.SelectionColor = style.selection

	return i.layoutFrame(ctx, gtx, state, style, editorStyle.Layout)
}
