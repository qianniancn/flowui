package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/menu"
)

const defaultTextAreaRows = 3

type TextAreaWidget struct {
	override                   *theme.InputOverride
	key                        string
	value                      string
	hasValue                   bool
	defaultValue               string
	hasDefault                 bool
	hint                       string
	onChange                   func(string)
	variant                    TextAreaVariant
	disabled                   bool
	invalid                    bool
	fullWidth                  bool
	readOnly                   bool
	maxLength                  int
	rows                       int
	label                      string
	contextMenu                menu.Widget
	hasContextMenu             bool
	defaultContextMenuDisabled bool
}

type TextAreaVariant = field.Variant

const (
	TextAreaPrimary   = field.Primary
	TextAreaSecondary = field.Secondary
)

// NewTextArea creates a multi-line input holding value.
func NewTextArea(value string) TextAreaWidget {
	return TextAreaWidget{
		key:      core.AutoKey(),
		value:    value,
		hasValue: true,
	}
}

// Key sets an explicit identity for the text area state.
func (t TextAreaWidget) Key(key string) TextAreaWidget {
	t.key = key
	return t
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

// ContextMenu replaces the built-in cut/copy/paste menu shown on right-click.
// The custom menu is responsible for its own labels and actions.
func (t TextAreaWidget) ContextMenu(content menu.Widget) TextAreaWidget {
	t.contextMenu = content
	t.hasContextMenu = true
	return t
}

// DefaultContextMenu enables or disables the built-in edit menu shown on
// right-click. It is enabled by default; a custom menu set through ContextMenu
// takes precedence.
func (t TextAreaWidget) DefaultContextMenu(enabled bool) TextAreaWidget {
	t.defaultContextMenuDisabled = !enabled
	return t
}

func (t TextAreaWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return t.layoutWithOverride(ctx, gtx, t.override)
}

func (t TextAreaWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.InputOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	fullKey := core.FullKey(ctx, t.key)
	state := textAreaStateFor(ctx, fullKey)
	state.bind(t)
	currentValue := state.currentValue(t)

	gtx, key, editor, enabled := t.prepareEditor(ctx, gtx, state, currentValue, t.disabled)
	disabled := !enabled
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered:      state.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		ReadOnly:     t.readOnly,
		Invalid:      t.invalid,
	}
	activeTheme := core.ActiveTheme(ctx)
	visual := state.visualDeclarations(t, activeTheme, 0, instanceOverride)
	root := resolveInputEditorRootStyle(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      key,
		state:    styleState,
		defaults: visual.Defaults,
		custom:   visual.Override,
	})
	initialEditorStyle := visual.EditorStyle(root, style.ResolvedStyle{}, style.ResolvedStyle{})
	height := textAreaHeight(gtx, root, initialEditorStyle.LineHeight, t.resolvedRows())
	if height > 0 {
		visual = state.visualDeclarations(t, activeTheme, height, instanceOverride)
	}
	root, placeholder, selection := resolveInputEditorStyles(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      key,
		state:    styleState,
		defaults: visual.Defaults,
		custom:   visual.Override,
	})
	editorStyle := visual.EditorStyle(root, placeholder, selection)
	editorLayout := t.editorLayoutWithStyle(ctx, key, enabled, editor, editorStyle)
	frame := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return t.layoutFrame(ctx, gtx, state, root, enabled, editorLayout)
	})
	if !enabled || (!t.hasContextMenu && t.defaultContextMenuDisabled) {
		return frame(ctx, gtx)
	}
	snapshot := SnapshotEditor(editor, t.readOnly, false)
	return EditContextMenu(ctx, gtx, t.key, frame, editor, snapshot, t.contextMenu, t.hasContextMenu, state.queueEdit)
}

func (t TextAreaWidget) prepareEditor(ctx *core.Context, gtx layout.Context, state *textAreaState, currentValue string, disabled bool) (layout.Context, string, *widget.Editor, bool) {
	fieldKey, editor := core.InputEditor(ctx, t.key)
	enabled := gtx.Enabled() && !disabled
	core.RegisterFieldFocus(ctx, fieldKey, editor, enabled)

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
	// Context-menu edits must run before the update loop so their ChangeEvent
	// surfaces through editor.Update like an ordinary keystroke; material's
	// render pass runs its own Update and would swallow the flag otherwise.
	if enabled {
		state.takePendingEdit(gtx, editor, t.readOnly)
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

func (t TextAreaWidget) editorLayoutWithStyle(ctx *core.Context, key string, enabled bool, editor *widget.Editor, resolved theme.InputEditorStyle) layout.Widget {
	return t.withSemantics(ctx, key, enabled, editorLayoutFor(ctx, editor, t.hint, resolved))
}

func textAreaHeight(gtx layout.Context, resolved style.ResolvedStyle, lineHeight unit.Sp, rows int) unit.Dp {
	if rows <= 0 || lineHeight <= 0 {
		return 0
	}
	height := gtx.Sp(lineHeight) * rows
	if resolved.Box != nil && resolved.Box.Padding != nil {
		padding := resolved.Box.Padding
		height += gtx.Dp(padding.Top + padding.Bottom)
	}
	return gtx.Metric.PxToDp(height)
}

func (t TextAreaWidget) resolvedRows() int {
	if t.rows > 0 {
		return t.rows
	}
	return defaultTextAreaRows
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w TextAreaWidget) Style(values ...style.Style) TextAreaWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.InputOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
