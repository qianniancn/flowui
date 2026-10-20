package text

import (
	"io"
	"strings"

	"gioui.org/io/clipboard"
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/core"
)

// takePendingEdit runs a queued edit menu action: copy writes the selection to
// the clipboard, select-all only moves the caret.
func (s *selectableTextState) takePendingEdit(gtx layout.Context, selectable *widget.Selectable) {
	if !s.hasPendingEdit {
		return
	}
	s.hasPendingEdit = false
	switch s.pendingEdit {
	case core.SelectableEditCopy:
		if text := selectable.SelectedText(); text != "" {
			gtx.Execute(clipboard.WriteCmd{Type: "application/text", Data: io.NopCloser(strings.NewReader(text))})
		}
	case core.SelectableEditSelectAll:
		selectable.SetCaret(0, len([]rune(selectable.Text())))
	}
}
