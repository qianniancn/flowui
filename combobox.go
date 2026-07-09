package flowui

import (
	"time"

	"gioui.org/layout"
	"gioui.org/widget/material"
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
	hint             string
	inputValue       string
	emptyText        string
	onChange         func(string)
	onInputChange    func(string)
	variant          InputVariant
	disabled         bool
	invalid          bool
	fullWidth        bool
	hasInputValue    bool
	allowCustomValue bool
}

const (
	comboBoxAnimationDuration   = 150 * time.Millisecond
	comboBoxItemColorDuration   = 100 * time.Millisecond
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

func (c ComboBoxWidget) Variant(variant InputVariant) ComboBoxWidget {
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

func (c ComboBoxWidget) Layout(ctx *Context, gtx layout.Context) layout.Dimensions {
	state := ctx.comboBoxState(c.key)
	editor := &state.editor
	editor.SingleLine = true
	editor.Submit = true
	state.beginFrame()
	state.checkItems(c.items)
	state.syncEditor(editor, c)
	state.input.update(ctx, gtx, c.disabled, editor)

	if c.disabled {
		gtx = gtx.Disabled()
	}

	query := editor.Text()
	selectedLabel, _ := c.selectedLabel()
	visible := comboBoxVisibleItems(c.items, query, selectedLabel)
	state.updateFocus(gtx.Focused(editor), c.disabled)
	if !c.disabled {
		state.clampHighlight(c.items, visible)
		if index, ok := state.updateKeys(gtx, editor, c.items, visible); ok {
			c.selectItem(editor, state, c.items[visible[index]])
		}
		c.updateEditor(editor, state, gtx)
		query = editor.Text()
		visible = comboBoxVisibleItems(c.items, query, selectedLabel)
		state.clampHighlight(c.items, visible)
	}

	focused := gtx.Focused(editor)
	style := inputStyleFor(ctx.Theme, c.variant, state.input.hovered, focused, c.disabled, c.invalid)
	style.bg = state.input.background(gtx, style.bg)
	style.border = state.input.borderColor(gtx, style.border)

	editorStyle := material.Editor(ctx.Theme.Material, editor, c.hint)
	editorStyle.TextSize = ctx.Theme.Components.ComboBox.TextSize
	editorStyle.Color = style.fg
	editorStyle.HintColor = style.placeholder
	editorStyle.SelectionColor = style.selection

	dims := c.layoutInput(ctx, gtx, state, editor, style, editorStyle.Layout)
	progress := state.popoverProgress(gtx, state.open && !c.disabled)
	if progress == 0 {
		state.endFrame()
		return dims
	}

	return c.layoutOpen(ctx, gtx, state, editor, dims, visible, progress)
}
