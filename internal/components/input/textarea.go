package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/frame"
	flowstyle "github.com/qianniancn/flowui/internal/style"
	styleruntime "github.com/qianniancn/flowui/internal/style/runtime"
)

const defaultTextAreaRows = 3

type TextAreaWidget struct {
	key          string
	value        string
	hasValue     bool
	defaultValue string
	hasDefault   bool
	hint         string
	onChange     func(string)
	variant      TextAreaVariant
	disabled     bool
	invalid      bool
	fullWidth    bool
	readOnly     bool
	maxLength    int
	rows         int
	label        string
	customStyle  flowstyle.Style
}

type TextAreaVariant = field.Variant

const (
	TextAreaPrimary   = field.Primary
	TextAreaSecondary = field.Secondary
)

func TextArea(key, value string) TextAreaWidget {
	return TextAreaWidget{
		key:      key,
		value:    value,
		hasValue: true,
	}
}

func (t TextAreaWidget) Value(value string) TextAreaWidget {
	t.value = value
	t.hasValue = true
	return t
}

func (t TextAreaWidget) DefaultValue(value string) TextAreaWidget {
	t.defaultValue = value
	t.hasDefault = true
	t.hasValue = false
	return t
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

func (t TextAreaWidget) Style(value flowstyle.Style) TextAreaWidget {
	t.customStyle = value
	return t
}

func (t TextAreaWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	fullKey := frame.FullKey(ctx, t.key)
	state := textAreaStateFor(ctx, fullKey)
	state.bind(t)
	currentValue := state.currentValue(t)

	gtx, key, editor, enabled := t.prepareEditor(ctx, gtx, state, currentValue, t.disabled)
	disabled := !enabled
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := frame.FocusVisible(ctx, editor, focused)
	styleState := flowstyle.StyleState{
		Hovered:      state.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		ReadOnly:     t.readOnly,
		Invalid:      t.invalid,
	}
	activeTheme := frame.ActiveTheme(ctx)
	tokens := activeTheme.Components.TextArea
	contentHeight := gtx.Sp(tokens.LineHeight)*t.resolvedRows() + gtx.Dp(tokens.PaddingY)*2
	height := gtx.Metric.PxToDp(contentHeight)
	defaults, variant, size := t.styleDeclarations(activeTheme, height)
	root := styleruntime.Resolve(ctx, gtx, key, styleState, defaults, variant, size, t.customStyle)
	placeholder := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartPlaceholder, styleState, defaults, variant, size, t.customStyle)
	selection := styleruntime.ResolvePart(ctx, gtx, key, flowstyle.PartSelection, styleState, defaults, variant, size, t.customStyle)
	style := resolvedInputStyle(root, placeholder, selection, activeTheme)
	textSize, lineHeight := resolvedTypography(root, tokens.TextSize, tokens.LineHeight)
	editorLayout := t.editorLayoutWithTypography(ctx, key, enabled, editor, style, textSize, lineHeight)
	return t.layoutFrame(ctx, gtx, state, root, enabled, editorLayout)
}

func (t TextAreaWidget) prepareEditor(ctx *frame.Context, gtx layout.Context, state *textAreaState, currentValue string, disabled bool) (layout.Context, string, *widget.Editor, bool) {
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
	if changed {
		state.requestValue(t, editor.Text())
	} else if !changed && editor.Text() != currentValue {
		editor.SetText(currentValue)
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
