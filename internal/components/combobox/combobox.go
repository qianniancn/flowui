package combobox

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
	layoutui "github.com/qianniancn/FlowUI/internal/components/layout"
	"github.com/qianniancn/FlowUI/internal/field"
	"github.com/qianniancn/FlowUI/internal/frame"
	flowstyle "github.com/qianniancn/FlowUI/internal/style"
)

type ComboBoxItem struct {
	Key         string
	Label       string
	Description string
	Disabled    bool
}

type ComboBoxWidget struct {
	key              string
	selectedKey      string
	items            []ComboBoxItem
	dataVersion      uint64
	hasDataVersion   bool
	hint             string
	inputValue       string
	emptyText        string
	onChange         func(string)
	onInputChange    func(string)
	variant          field.Variant
	disabled         bool
	invalid          bool
	fullWidth        bool
	hasInputValue    bool
	allowCustomValue bool
	customStyle      flowstyle.Style
}

const (
	comboBoxAnimationDuration   = 150 * time.Millisecond
	comboBoxItemColorDuration   = 100 * time.Millisecond
	comboBoxItemSelectDuration  = 250 * time.Millisecond
	comboBoxItemPressInDuration = 80 * time.Millisecond
	comboBoxItemPressDuration   = 140 * time.Millisecond
)

func ComboBox(key, selectedKey string, items []ComboBoxItem) ComboBoxWidget {
	return ComboBoxWidget{
		key:         key,
		selectedKey: selectedKey,
		items:       items,
		emptyText:   "No results found",
	}
}

func (c ComboBoxWidget) Hint(hint string) ComboBoxWidget {
	c.hint = hint
	return c
}

// DataVersion enables item validation and filtered-index reuse. Increase
// version whenever the item data changes.
func (c ComboBoxWidget) DataVersion(version uint64) ComboBoxWidget {
	c.dataVersion = version
	c.hasDataVersion = true
	return c
}

func (c ComboBoxWidget) InputValue(value string) ComboBoxWidget {
	c.inputValue = value
	c.hasInputValue = true
	return c
}

func (c ComboBoxWidget) EmptyText(text string) ComboBoxWidget {
	c.emptyText = text
	return c
}

func (c ComboBoxWidget) OnChange(fn func(string)) ComboBoxWidget {
	c.onChange = fn
	return c
}

func (c ComboBoxWidget) OnInputChange(fn func(string)) ComboBoxWidget {
	c.onInputChange = fn
	return c
}

func (c ComboBoxWidget) Disabled(disabled bool) ComboBoxWidget {
	c.disabled = disabled
	return c
}

func (c ComboBoxWidget) Invalid(invalid bool) ComboBoxWidget {
	c.invalid = invalid
	return c
}

func (c ComboBoxWidget) Variant(variant field.Variant) ComboBoxWidget {
	c.variant = variant
	return c
}

func (c ComboBoxWidget) FullWidth() ComboBoxWidget {
	c.fullWidth = true
	return c
}

func (c ComboBoxWidget) AllowCustomValue() ComboBoxWidget {
	c.allowCustomValue = true
	return c
}

func (c ComboBoxWidget) Style(value flowstyle.Style) ComboBoxWidget {
	c.customStyle = value
	return c
}

func (c ComboBoxWidget) Layout(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
	state := comboBoxStateFor(ctx, c.key)
	key := frame.FullKey(ctx, c.key)
	naturallyDisabled := frame.OverlayNaturallyDisabled(gtx)
	eventGtx := gtx
	editor := &state.editor
	frame.RegisterFieldFocus(ctx, key, editor, eventGtx.Enabled() && !c.disabled)
	editor.SingleLine = true
	editor.Submit = true
	state.beginFrame()
	state.checkItems(c.items, c.hasDataVersion, c.dataVersion)
	state.syncEditor(editor, c)
	state.input.Update(ctx, eventGtx, c.disabled, editor)

	if c.disabled {
		eventGtx = eventGtx.Disabled()
	}

	query := editor.Text()
	selectedLabel, _ := state.selectedLabel(c)
	visible := state.visibleItems(c, query, selectedLabel)
	processMainEvents := !state.open
	state.updateFocus(gtx.Focused(editor), c.disabled)
	if !c.disabled && processMainEvents {
		state.highlight = -1
		if index, ok := state.updateKeys(gtx, editor, c.items, visible); ok {
			c.selectItem(editor, state, c.items[visible[index]])
		}
		c.updateEditor(editor, state, gtx)
		query = editor.Text()
		visible = state.visibleItems(c, query, selectedLabel)
		state.clampHighlight(c.items, visible)
	}

	focused := gtx.Focused(editor)
	focusVisible := frame.FocusVisible(ctx, editor, focused)
	styleState := flowstyle.StyleState{
		Hovered:      state.input.Hovered,
		Focused:      focused,
		FocusVisible: focusVisible,
		Disabled:     c.disabled || !gtx.Enabled(),
		Selected:     c.selectedKey != "",
		Invalid:      c.invalid,
		Open:         state.open,
	}
	tokens := frame.ActiveTheme(ctx).Components
	resolved := field.Resolve(ctx, gtx, key, styleState, c.variant, field.DeclarationOptions{
		Radius:         tokens.ComboBox.Radius,
		FocusRingWidth: tokens.Input.FocusRingWidth, InvalidOutlineWidth: tokens.Input.InvalidOutlineWidth,
		ShadowColor: tokens.Input.ShadowColor, ShadowOpacity: tokens.Input.ShadowOpacity,
		ShadowStrength: tokens.Input.ShadowStrength,
	}, c.customStyle)

	editorStyle := material.Editor(frame.ActiveTheme(ctx).Material, editor, c.hint)
	editorStyle.TextSize = frame.ActiveTheme(ctx).Components.ComboBox.TextSize
	editorStyle.Color = resolved.Colors.Foreground
	editorStyle.HintColor = resolved.Colors.Placeholder
	editorStyle.SelectionColor = resolved.Colors.Selection

	dims := layoutui.LayoutStyled(ctx, eventGtx, key, styleState, c.customStyle, frame.WidgetFunc(func(ctx *frame.Context, gtx layout.Context) layout.Dimensions {
		return c.layoutInput(ctx, gtx, state, editor, resolved, frame.WithFieldSemantics(ctx, key, editorStyle.Layout))
	}))
	progress := state.popoverProgress(gtx, state.open && !c.disabled, frame.ActiveTheme(ctx).Motion)
	if progress == 0 {
		state.endFrame()
		return dims
	}

	return c.layoutOpen(ctx, eventGtx, state, editor, dims, visible, progress, naturallyDisabled)
}
