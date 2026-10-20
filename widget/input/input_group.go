package input

import (
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/menu"
)

const stateSlotInputGroup = "input-group"

// GroupWidget combines an Input or TextArea with optional prefix and suffix content
// inside one field shell. The look comes from the active theme.
type GroupWidget struct {
	override           *theme.InputOverride
	input              Widget
	textArea           TextAreaWidget
	multiline          bool
	prefix             core.Widget
	suffix             core.Widget
	prefixAction       bool
	suffixAction       bool
	focusOnActionPress bool
	variant            Variant
	disabled           bool
	invalid            bool
	fullWidth          bool
}

func Group(input Widget) GroupWidget {
	return GroupWidget{
		input:     input,
		variant:   input.variant,
		disabled:  input.disabled,
		invalid:   input.invalid,
		fullWidth: input.fullWidth,
	}
}

func GroupTextArea(textArea TextAreaWidget) GroupWidget {
	return GroupWidget{
		textArea:  textArea,
		multiline: true,
		variant:   textArea.variant,
		disabled:  textArea.disabled,
		invalid:   textArea.invalid,
		fullWidth: textArea.fullWidth,
	}
}

func (g GroupWidget) Prefix(prefix core.Widget) GroupWidget {
	g.prefix = prefix
	g.prefixAction = false
	return g
}

func (g GroupWidget) Suffix(suffix core.Widget) GroupWidget {
	g.suffix = suffix
	g.suffixAction = false
	return g
}

// PrefixAction adds a compact interactive control before the editor. The
// active InputAction provider supplies its slot padding; use StyledGroup with
// a PartPrefix declaration when an instance needs a different visual inset.
func (g GroupWidget) PrefixAction(action core.Widget) GroupWidget {
	g.prefix = action
	g.prefixAction = true
	return g
}

// SuffixAction adds a compact interactive control after the editor. The
// active InputAction provider supplies its slot padding; use StyledGroup with
// a PartSuffix declaration when an instance needs a different visual inset.
func (g GroupWidget) SuffixAction(action core.Widget) GroupWidget {
	g.suffix = action
	g.suffixAction = true
	return g
}

// FocusOnActionPress controls whether pressing a prefix or suffix action also
// focuses the editor. It is disabled by default for action slots.
func (g GroupWidget) FocusOnActionPress(enabled bool) GroupWidget {
	g.focusOnActionPress = enabled
	return g
}

func (g GroupWidget) Variant(variant Variant) GroupWidget {
	g.variant = variant
	return g
}

func (g GroupWidget) Disabled(disabled bool) GroupWidget {
	g.disabled = disabled
	return g
}

func (g GroupWidget) Invalid(invalid bool) GroupWidget {
	g.invalid = invalid
	return g
}

func (g GroupWidget) FullWidth() GroupWidget {
	g.fullWidth = true
	return g
}

func (g GroupWidget) Layout(ctx *core.Context, gtx layout.Context) layout.Dimensions {
	return g.layoutWithOverride(ctx, gtx, g.override)
}

func (g GroupWidget) layoutWithOverride(ctx *core.Context, gtx layout.Context, instanceOverride *theme.InputOverride) layout.Dimensions {
	core.RequireVisualProvider(ctx)
	var key string
	var editor *widget.Editor
	var enabled bool
	var state *inputGroupState
	var queueEdit func(EditAction)
	var readOnly, secret, hasCustomMenu, menuDisabled bool
	var customMenu menu.Widget

	if g.multiline {
		fullKey := core.FullKey(ctx, g.textArea.key)
		textAreaState := textAreaStateFor(ctx, fullKey)
		textAreaState.bind(g.textArea)
		currentValue := textAreaState.currentValue(g.textArea)
		gtx, key, editor, enabled = g.textArea.prepareEditor(ctx, gtx, textAreaState, currentValue, g.disabled)
		queueEdit = textAreaState.queueEdit
		readOnly = g.textArea.readOnly
		customMenu, hasCustomMenu = g.textArea.contextMenu, g.textArea.hasContextMenu
		menuDisabled = !g.textArea.hasContextMenu && g.textArea.defaultContextMenuDisabled
	} else {
		// Create temporary input state for disclosure
		fullKey := core.FullKey(ctx, g.input.key)
		inputState := inputStateFor(ctx, fullKey)
		inputState.bind(g.input)
		currentValue := inputState.currentValue(g.input)
		gtx, key, editor, enabled = g.input.prepareEditor(ctx, gtx, inputState, currentValue, g.disabled)
		queueEdit = inputState.queueEdit
		readOnly = g.input.readOnly
		secret = g.input.inputType == Password
		customMenu, hasCustomMenu = g.input.contextMenu, g.input.hasContextMenu
		menuDisabled = !g.input.hasContextMenu && g.input.defaultContextMenuDisabled
	}
	disabled := !enabled
	state = inputGroupStateFor(ctx, key)
	var shouldFocus func(pointer.Event) bool
	if !g.focusOnActionPress && (g.prefixAction || g.suffixAction) {
		shouldFocus = state.shouldFocusPress
	}
	state.State.UpdateWithFocus(ctx, gtx, disabled, editor, shouldFocus)

	focused := gtx.Focused(editor)
	focusVisible := core.FocusVisible(ctx, editor, focused)
	styleState := style.State{
		Hovered: state.Hovered, Focused: focused, FocusVisible: focusVisible,
		Disabled: disabled, ReadOnly: editor.ReadOnly, Invalid: g.invalid,
	}
	activeTheme := core.ActiveTheme(ctx)
	var actionVisual theme.InputActionVisual
	if g.prefixAction || g.suffixAction {
		actionVisual = inputGroupActionVisualCached(ctx, key, theme.InputActionRequest{})
	}
	visual := state.visualDeclarations(g, activeTheme, 0, instanceOverride)
	resolved := resolveInputGroupStyle(ctx, gtx, key, styleState, visual.Defaults, visual.Override)
	editorStyle := visual.EditorStyle(resolved.content, resolved.placeholder, resolved.selection)
	if g.multiline {
		minHeight := inputGroupTextAreaHeight(gtx, resolved.content, editorStyle.LineHeight, g.textArea.resolvedRows())
		if minHeight > 0 {
			visual = state.visualDeclarations(g, activeTheme, minHeight, instanceOverride)
			resolved = resolveInputGroupStyle(ctx, gtx, key, styleState, visual.Defaults, visual.Override)
			editorStyle = visual.EditorStyle(resolved.content, resolved.placeholder, resolved.selection)
		}
	}
	resolved.editor = editorStyle
	var editorLayout layout.Widget
	if g.multiline {
		editorLayout = g.textArea.editorLayoutWithStyle(ctx, key, enabled, editor, editorStyle)
	} else {
		editorLayout = g.input.editorLayoutWithStyle(ctx, key, enabled, editor, editorStyle)
	}
	if g.focusOnActionPress {
		g.prefix = inputGroupActionWithFocus(g.prefix, editor)
		g.suffix = inputGroupActionWithFocus(g.suffix, editor)
	}
	frame := core.WidgetFunc(func(ctx *core.Context, gtx layout.Context) layout.Dimensions {
		return g.layoutFrame(ctx, gtx, state, resolved, enabled, editorLayout, actionVisual)
	})
	if !enabled || menuDisabled {
		return frame(ctx, gtx)
	}
	snapshot := SnapshotEditor(editor, readOnly, secret)
	return EditContextMenu(ctx, gtx, key, frame, editor, snapshot, customMenu, hasCustomMenu, queueEdit)
}

var inputGroupStyleParts = [...]style.Part{
	style.PartRoot,
	style.PartContent,
	style.PartPrefix,
	style.PartSuffix,
	style.PartIndicator,
	style.PartPlaceholder,
	style.PartSelection,
}

func resolveInputGroupStyle(ctx *core.Context, gtx layout.Context, key string, state style.State, defaults, custom style.Style) inputGroupResolvedStyle {
	var resolved [len(inputGroupStyleParts)]style.ResolvedStyle
	computed.ResolvePartsInto(ctx, gtx, key, state, inputGroupStyleParts[:], resolved[:], defaults, style.Style{}, style.Style{}, custom)
	return inputGroupResolvedStyle{
		root:        resolved[0],
		content:     resolved[1],
		prefix:      resolved[2],
		suffix:      resolved[3],
		divider:     resolved[4],
		placeholder: resolved[5],
		selection:   resolved[6],
	}
}

func inputGroupTextAreaHeight(gtx layout.Context, resolved style.ResolvedStyle, lineHeight unit.Sp, rows int) unit.Dp {
	if rows <= 0 || lineHeight <= 0 {
		return 0
	}
	height := gtx.Sp(lineHeight) * rows
	if resolved.Box != nil {
		if resolved.Box.Padding != nil {
			padding := resolved.Box.Padding
			height += gtx.Dp(padding.Top + padding.Bottom)
		}
		if resolved.Box.MinHeight != nil {
			height = max(height, gtx.Dp(*resolved.Box.MinHeight))
		}
	}
	return gtx.Metric.PxToDp(height)
}

// Style merges instance declarations into the widget. The declaration
// reaches the provider at layout time, so Style can sit anywhere in the chain.
func (w GroupWidget) Style(values ...style.Style) GroupWidget {
	declaration := style.Join(values...)
	if w.override == nil {
		w.override = &theme.InputOverride{}
	}
	w.override.Declaration = style.Join(w.override.Declaration, declaration)
	return w
}
