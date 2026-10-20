package input

import (
	"gioui.org/layout"
	"gioui.org/widget"
	"github.com/qianniancn/flowui/internal/computed"
	"github.com/qianniancn/flowui/internal/core"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/theme"
	"github.com/qianniancn/flowui/widget/disclosure"
)

type textAreaState struct {
	field.State
	disclosure  disclosure.Binding[string]
	value       string
	styleCache  computed.ProviderCacheMap[inputVisualKey, theme.InputVisual]
	pendingEdit EditAction
}

func textAreaStateFor(ctx *core.Context, key string) *textAreaState {
	return core.UseStateWith(ctx, key, "textAreaState", func() *textAreaState {
		return &textAreaState{}
	})
}

func textAreaDisclosureCfg(widget TextAreaWidget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *textAreaState) currentValue(widget TextAreaWidget) string {
	s.value = s.disclosure.Current(textAreaDisclosureCfg(widget))
	return s.value
}

func (s *textAreaState) bind(widget TextAreaWidget) {
	s.disclosure.Bind(textAreaDisclosureCfg(widget))
}

func (s *textAreaState) requestValue(widget TextAreaWidget, value string) string {
	s.value, _ = s.disclosure.Request(textAreaDisclosureCfg(widget), value)
	return s.value
}

// queueEdit stores an edit action activated from the context menu; the next
// Layout consumes it.
func (s *textAreaState) queueEdit(action EditAction) {
	s.pendingEdit = action
}

func (s *textAreaState) takePendingEdit(gtx layout.Context, editor *widget.Editor, readOnly bool) {
	action := s.pendingEdit
	if action == editNone {
		return
	}
	s.pendingEdit = editNone
	RunEditAction(gtx, editor, action, readOnly)
}
