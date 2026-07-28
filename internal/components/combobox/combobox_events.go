package combobox

import (
	"unicode/utf8"

	"gioui.org/layout"
	"gioui.org/widget"
)

func (c ComboBoxWidget) updateEditor(editor *widget.Editor, state *comboBoxState, gtx layout.Context) {
	for {
		event, ok := editor.Update(gtx)
		if !ok {
			break
		}
		if _, ok := event.(widget.ChangeEvent); ok {
			if state.consumeEditorChange() {
				continue
			}
			state.open = true
			state.highlight = 0
			if c.onInputChange != nil {
				c.onInputChange(editor.Text())
			}
		}
	}
}

func (c ComboBoxWidget) selectItem(editor *widget.Editor, state *comboBoxState, item ComboBoxItem) {
	state.setEditorText(editor, item.Label)
	end := utf8.RuneCountInString(item.Label)
	editor.SetCaret(end, end)
	state.syncedSelectedKey = item.Key
	state.syncedInputValue = item.Label
	state.open = false
	// Update disclosure state and trigger onChange callback
	state.requestSelectedKey(c, item.Key)
}
