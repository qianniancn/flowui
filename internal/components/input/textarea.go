package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
)

const defaultTextAreaRows = 3

type TextAreaWidget struct {
	key       string
	value     string
	hint      string
	onChange  func(string)
	variant   TextAreaVariant
	disabled  bool
	invalid   bool
	fullWidth bool
	readOnly  bool
	maxLength int
	rows      int
	label     string
}

type TextAreaVariant = field.Variant

const (
	TextAreaPrimary   = field.Primary
	TextAreaSecondary = field.Secondary
)

func TextArea(key, value string) TextAreaWidget {
	return TextAreaWidget{key: key, value: value}
}

func (t TextAreaWidget) Hint(hint string) TextAreaWidget {
	t.hint = hint
	return t
}

func (t TextAreaWidget) Placeholder(placeholder string) TextAreaWidget {
	t.hint = placeholder
	return t
}

func (t TextAreaWidget) OnChange(fn func(string)) TextAreaWidget {
	t.onChange = fn
	return t
}

func (t TextAreaWidget) Disabled(disabled bool) TextAreaWidget {
	t.disabled = disabled
	return t
}

func (t TextAreaWidget) Invalid(invalid bool) TextAreaWidget {
	t.invalid = invalid
	return t
}

func (t TextAreaWidget) Variant(variant TextAreaVariant) TextAreaWidget {
	t.variant = variant
	return t
}

func (t TextAreaWidget) FullWidth() TextAreaWidget {
	t.fullWidth = true
	return t
}

func (t TextAreaWidget) ReadOnly(readOnly bool) TextAreaWidget {
	t.readOnly = readOnly
	return t
}

func (t TextAreaWidget) MaxLength(maxLength int) TextAreaWidget {
	t.maxLength = max(maxLength, 0)
	return t
}

func (t TextAreaWidget) Rows(rows int) TextAreaWidget {
	t.rows = max(rows, 1)
	return t
}

func (t TextAreaWidget) Label(label string) TextAreaWidget {
	t.label = label
	return t
}

func (t TextAreaWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	gtx, key, editor, enabled := t.prepareEditor(ctx, gtx, t.disabled)
	disabled := !enabled
	state := inputStateFor(ctx, key)
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	style := textAreaStyleFor(frame.ActiveTheme(ctx), t.variant, state.Hovered, focused, disabled, t.invalid)
	style.Background = state.Background(gtx, style.Background)
	style.Ring = state.BorderColor(gtx, style.Ring)

	return t.layoutFrame(ctx, gtx, state, style, enabled, t.editorLayout(ctx, key, enabled, editor, style))
}

func (t TextAreaWidget) prepareEditor(ctx *frame.Context, gtx layout.Context, disabled bool) (layout.Context, string, *widget.Editor, bool) {
	fieldKey, editor := frame.InputEditor(ctx, t.key)
	enabled := gtx.Enabled() && !disabled
	frame.RegisterFieldFocus(ctx, fieldKey, editor, enabled)

	editor.SingleLine = false
	editor.Submit = false
	editor.ReadOnly = t.readOnly
	editor.MaxLen = t.maxLength
	editor.Mask = 0
	editor.InputHint = key.HintText
	editor.Filter = ""
	if !enabled {
		gtx = gtx.Disabled()
	}

	changed := false
	for {
		event, ok := editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok := event.(widget.ChangeEvent); ok {
			changed = true
		}
	}
	if changed && t.onChange != nil {
		t.onChange(editor.Text())
	} else if !changed && editor.Text() != t.value {
		editor.SetText(t.value)
	}
	return gtx, fieldKey, editor, enabled
}

func (t TextAreaWidget) editorLayout(ctx *frame.Context, key string, enabled bool, editor *widget.Editor, style inputStyle) layout.Widget {
	tokens := frame.ActiveTheme(ctx).Components.TextArea
	return t.editorLayoutWithTypography(ctx, key, enabled, editor, style, tokens.TextSize, tokens.LineHeight)
}

func (t TextAreaWidget) editorLayoutWithTypography(ctx *frame.Context, key string, enabled bool, editor *widget.Editor, style inputStyle, textSize, lineHeight unit.Sp) layout.Widget {
	return t.withSemantics(ctx, key, enabled, editorLayoutFor(ctx, editor, t.hint, style, textSize, lineHeight))
}

func (t TextAreaWidget) resolvedRows() int {
	if t.rows > 0 {
		return t.rows
	}
	return defaultTextAreaRows
}
