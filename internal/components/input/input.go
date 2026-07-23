package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
	styleruntime "github.com/qianniancn/FlowUI/internal/style/runtime"
)

type InputWidget struct {
	key         string
	value       string
	hint        string
	onChange    func(string)
	onSubmit    func(string)
	variant     InputVariant
	disabled    bool
	invalid     bool
	fullWidth   bool
	inputType   InputType
	readOnly    bool
	maxLength   int
	label       string
	customStyle flowstyle.Style
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

func (i InputWidget) Style(value flowstyle.Style) InputWidget {
	i.customStyle = value
	return i
}

func (i InputWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	gtx, key, editor, enabled := i.prepareEditor(ctx, gtx, i.disabled)
	disabled := !enabled
	state := inputStateFor(ctx, key)
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := frame.FocusVisible(ctx, editor, focused)
	styleState := flowstyle.StyleState{
		Hovered:      state.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		ReadOnly:     i.readOnly,
		Invalid:      i.invalid,
	}
	activeTheme := frame.ActiveTheme(ctx)
	defaults := inputDefaultDeclaration(activeTheme, i.variant, i.fullWidth)
	root := styleruntime.Resolve(ctx, gtx, key, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, i.customStyle)
	placeholder := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartPlaceholder, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, i.customStyle)
	selection := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartSelection, styleState, defaults, flowstyle.Style{}, flowstyle.Style{}, i.customStyle)
	style := resolvedInputStyle(root, placeholder, selection, activeTheme)
	textSize, lineHeight := resolvedTypography(root, activeTheme.Components.Input.TextSize, activeTheme.Components.Input.LineHeight)
	editorLayout := i.editorLayoutWithTypography(ctx, key, enabled, editor, style, textSize, lineHeight)
	return i.layoutFrame(ctx, gtx, state, root, enabled, editorLayout)
}

func (i InputWidget) prepareEditor(ctx *frame.Context, gtx layout.Context, disabled bool) (layout.Context, string, *widget.Editor, bool) {
	key, editor := frame.InputEditor(ctx, i.key)
	enabled := gtx.Enabled() && !disabled
	frame.RegisterFieldFocus(ctx, key, editor, enabled)

	editor.SingleLine = true
	editor.Submit = i.onSubmit != nil
	editor.ReadOnly = i.readOnly
	editor.MaxLen = i.maxLength
	editor.Mask, editor.InputHint, editor.Filter = inputTypeConfig(i.inputType)
	if !enabled {
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
	return gtx, key, editor, enabled
}

func (i InputWidget) editorLayout(ctx *frame.Context, key string, enabled bool, editor *widget.Editor, style inputStyle) layout.Widget {
	tokens := frame.ActiveTheme(ctx).Components.Input
	return i.editorLayoutWithTypography(ctx, key, enabled, editor, style, tokens.TextSize, tokens.LineHeight)
}

func (i InputWidget) editorLayoutWithTypography(ctx *frame.Context, key string, enabled bool, editor *widget.Editor, style inputStyle, textSize, lineHeight unit.Sp) layout.Widget {
	return i.withSemantics(ctx, key, enabled, editorLayoutFor(ctx, editor, i.hint, style, textSize, lineHeight))
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
