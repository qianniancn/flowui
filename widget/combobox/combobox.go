package combobox

import (
	"gioui.org/layout"
	"gioui.org/widget/material"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/menu"
)

type Item struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
}

type Widget struct {
	override                   *theme.ComboOverride
	key                        string
	selectedKey                string
	hasSelectedKey             bool
	defaultSelectedKey         string
	hasDefaultSelected         bool
	items                      []Item
	dataVersion                uint64
	hasDataVersion             bool
	hint                       string
	inputValue                 string
	emptyText                  string
	hasEmptyText               bool
	onChange                   func(string)
	onInputChange              func(string)
	variant                    field.Variant
	disabled                   bool
	invalid                    bool
	fullWidth                  bool
	hasInputValue              bool
	allowCustomValue           bool
	contextMenu                menu.Widget
	hasContextMenu             bool
	defaultContextMenuDisabled bool
}

// New creates a combo box. It starts uncontrolled with nothing selected;
// SelectedKey sets the controlled value and DefaultSelectedKey the initial one.
func New(items []Item) Widget {
	return Widget{
		key:   core.AutoKey(),
		items: items,
	}
}

// Key sets an explicit identity for the combo box state.
func (c Widget) Key(key string) Widget {
	c.key = key
	return c
}

func (c Widget) Hint(hint string) Widget {
	c.hint = hint
	return c
}

// DataVersion enables item validation and filtered-index reuse. Increase
// version whenever the item data changes.
func (c Widget) DataVersion(version uint64) Widget {
	c.dataVersion = version
	c.hasDataVersion = true
	return c
}

func (c Widget) InputValue(value string) Widget {
	c.inputValue = value
	c.hasInputValue = true
	return c
}

func (c Widget) EmptyText(text string) Widget {
	c.emptyText = text
	c.hasEmptyText = true
	return c
}

func (c Widget) OnChange(fn func(string)) Widget {
	c.onChange = fn
	return c
}

func (c Widget) SelectedKey(key string) Widget {
	c.selectedKey = key
	c.hasSelectedKey = true
	return c
}

func (c Widget) DefaultSelectedKey(key string) Widget {
	c.defaultSelectedKey = key
	c.hasDefaultSelected = true
	return c
}

func (c Widget) OnInputChange(fn func(string)) Widget {
	c.onInputChange = fn
	return c
}

func (c Widget) Disabled(disabled bool) Widget {
	c.disabled = disabled
	return c
}

func (c Widget) Invalid(invalid bool) Widget {
	c.invalid = invalid
	return c
}

func (c Widget) Variant(variant field.Variant) Widget {
	c.variant = variant
	return c
}

func (c Widget) FullWidth() Widget {
	c.fullWidth = true
	return c
}

func (c Widget) AllowCustomValue() Widget {
	c.allowCustomValue = true
	return c
}

// ContextMenu replaces the built-in edit menu shown on right click.
func (c Widget) ContextMenu(content menu.Widget) Widget {
	c.contextMenu = content
	c.hasContextMenu = true
	return c
}

// DefaultContextMenu toggles the built-in cut/copy/paste menu on right click.
func (c Widget) DefaultContextMenu(enabled bool) Widget {
	c.defaultContextMenuDisabled = !enabled
	return c
}

func (c Widget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return c.layoutWithOverride(ctx, gtx, c.override)
}

func (c Widget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.ComboOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	state := comboBoxStateFor(ctx, c.key)
	key := core.FullKey(ctx, c.key)
	naturallyDisabled := core.OverlayNaturallyDisabled(gtx)
	eventGtx := gtx
	editor := &state.editor
	core.RegisterFieldFocus(ctx, key, editor, eventGtx.Enabled() && !c.disabled)
	editor.SingleLine = true
	editor.Submit = true
	state.beginFrame()
	state.checkItems(c.items, c.hasDataVersion, c.dataVersion)

	// Bind disclosure state
	state.bind(c)
	selectedKey := state.currentSelectedKey(c)
	// Create a modified widget with the current selectedKey for internal use
	activeWidget := c
	activeWidget.selectedKey = selectedKey

	state.syncEditor(editor, activeWidget)
	state.takePendingEdit(eventGtx, editor)
	state.input.Update(ctx, eventGtx, c.disabled, editor)

	if c.disabled {
		eventGtx = eventGtx.Disabled()
	}

	query := editor.Text()
	selectedLabel, _ := state.selectedLabel(activeWidget)
	visible := state.visibleItems(activeWidget, query, selectedLabel)
	processMainEvents := !state.open
	state.updateFocus(gtx.Focused(editor), c.disabled)
	if !c.disabled && processMainEvents {
		state.highlight = -1
		if index, ok := state.updateKeys(gtx, editor, c.items, visible, false); ok {
			activeWidget.selectItem(editor, state, c.items[visible[index]])
			selectedKey = state.selectedKey
			activeWidget.selectedKey = selectedKey
		}
		activeWidget.updateEditor(editor, state, gtx)
		query = editor.Text()
		visible = state.visibleItems(activeWidget, query, selectedLabel)
		state.clampHighlight(c.items, visible)
	}
	// Register navigation filters before laying out the editor semantics. When
	// the popup is already open, the editor's semantic handler would otherwise
	// consume Arrow/Enter events before the overlay can observe them.
	if !c.disabled && state.open {
		if index, ok := state.updateKeys(eventGtx, editor, c.items, visible, false); ok {
			activeWidget.selectItem(editor, state, c.items[visible[index]])
			selectedKey = state.selectedKey
			activeWidget.selectedKey = selectedKey
		}
	}

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered:      state.input.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     c.disabled || !gtx.Enabled(),
		Selected:     selectedKey != "",
		Invalid:      c.invalid,
		Open:         state.open,
	}
	activeTheme := core.ActiveTheme(ctx)
	var declarations comboDeclarationBundle
	if instanceOverride == nil {
		declarations = comboStyleDeclarationsCached(activeTheme, c.variant, &state.styleCache)
	} else {
		declarations = comboVisualBundle(activeTheme, c.variant, instanceOverride)
	}
	override := declarations.visual.Override
	resolved := c.resolveStyle(ctx, gtx, state, key, styleState, override, declarations)
	editorStyle := material.Editor(core.ActiveMaterial(ctx), editor, c.hint)
	providerEditorStyle := resolved.visual.EditorStyle(resolved.content, resolved.placeholder, resolved.selection)
	editorStyle.TextSize = providerEditorStyle.TextSize
	editorStyle.LineHeight = providerEditorStyle.LineHeight
	editorStyle.Color = providerEditorStyle.Foreground
	editorStyle.HintColor = providerEditorStyle.Placeholder
	editorStyle.SelectionColor = providerEditorStyle.Selection

	field := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return c.layoutInput(ctx, gtx, state, editor, resolved, core.WithFieldSemantics(ctx, key, editorStyle.Layout))
	})
	var dims layout.Dimensions
	if !c.disabled && (c.hasContextMenu || !c.defaultContextMenuDisabled) {
		snapshot := input.SnapshotEditor(editor, false, false)
		dims = input.EditContextMenu(ctx, gtx, key, field, editor, snapshot, c.contextMenu, c.hasContextMenu, state.queueEdit)
	} else {
		dims = field(ctx, gtx)
	}
	progress := state.popoverProgress(gtx, state.open && !c.disabled, resolved.visual.Metrics.AnimationDuration, resolved.visual.Motion)
	if progress == 0 {
		state.endFrame()
		return dims
	}

	return c.layoutOpenWithBundle(ctx, eventGtx, state, editor, dims, visible, progress, naturallyDisabled, override, declarations)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w Widget) Style(values ...style.Style) Widget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.ComboOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
