package input

import (
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/menu"
)

type Widget struct {
	override                   *theme.InputOverride
	key                        string
	value                      string
	hasValue                   bool
	defaultValue               string
	hasDefault                 bool
	hint                       string
	onChange                   func(string)
	onSubmit                   func(string)
	variant                    Variant
	disabled                   bool
	invalid                    bool
	warning                    bool
	size                       Size
	fullWidth                  bool
	inputType                  Type
	readOnly                   bool
	maxLength                  int
	label                      string
	contextMenu                menu.Widget
	hasContextMenu             bool
	defaultContextMenuDisabled bool
}

type Variant = field.Variant

// Variants other than Primary and Secondary are exported through the field
// alias, for example input.Filled, input.Borderless, and input.Underlined.
const (
	Primary    = field.Primary
	Secondary  = field.Secondary
	Filled     = field.Filled
	Borderless = field.Borderless
	Underlined = field.Underlined
)

// Size selects the field's visual geometry tier; the active Style Provider
// owns the resulting height, typography, and spacing.
type Size uint8

const (
	Medium Size = iota
	Small
	Large
)

type Type uint8

const (
	Text Type = iota
	Email
	Number
	Password
)

const stateSlotInput = "input"

// New creates a text input holding value.
func New(value string) Widget {
	return Widget{
		key:      core.AutoKey(),
		value:    value,
		hasValue: true,
	}
}

// Key sets an explicit identity for the input state.
func (i Widget) Key(key string) Widget {
	i.key = key
	return i
}

// Uncontrolled changes the initial value into an internal state seed.
// It is useful when the initial value is empty and should not imply control.
func (i Widget) Uncontrolled() Widget {
	i.defaultValue = i.value
	i.hasDefault = true
	i.hasValue = false
	return i
}

func (i Widget) Hint(hint string) Widget {
	i.hint = hint
	return i
}

func (i Widget) Placeholder(placeholder string) Widget {
	i.hint = placeholder
	return i
}

func (i Widget) OnChange(fn func(string)) Widget {
	i.onChange = fn
	return i
}

func (i Widget) Value(value string) Widget {
	i.value = value
	i.hasValue = true
	return i
}

func (i Widget) DefaultValue(value string) Widget {
	i.defaultValue = value
	i.hasDefault = true
	i.hasValue = false
	return i
}

func (i Widget) OnSubmit(fn func(string)) Widget {
	i.onSubmit = fn
	return i
}

func (i Widget) Disabled(disabled bool) Widget {
	i.disabled = disabled
	return i
}

func (i Widget) Invalid(invalid bool) Widget {
	i.invalid = invalid
	return i
}

// Warning marks the field with the warning state: an amber outline instead of
// the danger one. Pair with the theme's Warning palette.
func (i Widget) Warning(warning bool) Widget {
	i.warning = warning
	return i
}

// Size selects the visual geometry tier (Small/Medium/Large). The active Style
// Provider resolves its concrete dimensions; Medium is the provider default.
func (i Widget) Size(size Size) Widget {
	i.size = size
	return i
}

func (i Widget) Variant(variant Variant) Widget {
	i.variant = variant
	return i
}

func (i Widget) FullWidth() Widget {
	i.fullWidth = true
	return i
}

func (i Widget) Type(inputType Type) Widget {
	i.inputType = inputType
	return i
}

func (i Widget) ReadOnly(readOnly bool) Widget {
	i.readOnly = readOnly
	return i
}

func (i Widget) MaxLength(maxLength int) Widget {
	i.maxLength = max(maxLength, 0)
	return i
}

func (i Widget) Label(label string) Widget {
	i.label = label
	return i
}

// ContextMenu replaces the built-in cut/copy/paste menu shown on right-click.
// The custom menu is responsible for its own labels and actions.
func (i Widget) ContextMenu(content menu.Widget) Widget {
	i.contextMenu = content
	i.hasContextMenu = true
	return i
}

// DefaultContextMenu enables or disables the built-in edit menu shown on
// right-click. It is enabled by default; a custom menu set through ContextMenu
// takes precedence.
func (i Widget) DefaultContextMenu(enabled bool) Widget {
	i.defaultContextMenuDisabled = !enabled
	return i
}

func (i Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return i.layoutWithOverride(ctx, gtx, i.override)
}

func (i Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.InputOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	state := inputStateFor(ctx, core.FullKey(ctx, i.key))

	// Bind disclosure and get current value
	state.bind(i)
	currentValue := state.currentValue(i)

	gtx, key, editor, enabled := i.prepareEditor(ctx, gtx, state, currentValue, i.disabled)
	disabled := !enabled
	state.State.Update(ctx, gtx, disabled, editor)

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered:      state.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     disabled,
		ReadOnly:     i.readOnly,
		Invalid:      i.invalid,
		Warning:      i.warning,
	}
	activeTheme := core.ActiveTheme(ctx)
	visual := state.visualDeclarations(i, activeTheme, instanceOverride)
	root, placeholder, selection := resolveInputEditorStyles(inputEditorStyleResolution{
		ctx:      ctx,
		gtx:      gtx,
		key:      key,
		state:    styleState,
		defaults: visual.Defaults,
		custom:   visual.Override,
	})
	editorStyle := visual.EditorStyle(root, placeholder, selection)
	editorLayout := i.editorLayoutWithStyle(ctx, key, enabled, editor, editorStyle)
	frame := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return i.layoutFrame(ctx, gtx, state, root, enabled, editorLayout)
	})
	if !enabled || (!i.hasContextMenu && i.defaultContextMenuDisabled) {
		return frame(ctx, gtx)
	}
	snapshot := SnapshotEditor(editor, i.readOnly, i.inputType == Password)
	return EditContextMenu(ctx, gtx, i.key, frame, editor, snapshot, i.contextMenu, i.hasContextMenu, state.queueEdit)
}

func (i Widget) prepareEditor(ctx *core.Context, gtx layout.Context, state *inputState, currentValue string, disabled bool) (layout.Context, string, *widget.Editor, bool) {
	key, editor := core.InputEditor(ctx, i.key)
	enabled := gtx.Enabled() && !disabled
	core.RegisterFieldFocus(ctx, key, editor, enabled)

	editor.SingleLine = true
	// Single-line editors must consume carriage returns even when callers do
	// not need a submit callback; otherwise Gio replaces them with spaces.
	editor.Submit = true
	editor.ReadOnly = i.readOnly
	editor.MaxLen = i.maxLength
	editor.Mask, editor.InputHint, editor.Filter = inputTypeConfig(i.inputType)
	if !enabled {
		gtx = gtx.Disabled()
	}
	// Context-menu edits must run before the update loop so their ChangeEvent
	// surfaces through editor.Update like an ordinary keystroke; material's
	// render pass runs its own Update and would swallow the flag otherwise.
	if enabled {
		state.takePendingEdit(gtx, editor, i.readOnly)
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
		i.dispatchEvents(state, editor, events)
	} else if editor.Text() != currentValue {
		editor.SetText(currentValue)
	}
	return gtx, key, editor, enabled
}

func (i Widget) editorLayoutWithStyle(ctx *core.Context, key string, enabled bool, editor *widget.Editor, resolved theme.InputEditorStyle) layout.Widget {
	return i.withSemantics(ctx, key, enabled, editorLayoutFor(ctx, editor, i.hint, resolved))
}

func inputTypeConfig(inputType Type) (rune, key.InputHint, string) {
	switch inputType {
	case Email:
		return 0, key.HintEmail, ""
	case Number:
		return 0, key.HintNumeric, "0123456789+-.eE"
	case Password:
		return '\u2022', key.HintPassword, ""
	default:
		return 0, key.HintText, ""
	}
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.InputOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
