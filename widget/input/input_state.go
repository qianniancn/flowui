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

type inputState struct {
	field.State
	disclosure  disclosure.Binding[string]
	value       string
	styleCache  computed.ProviderCache[inputVisualKey, theme.InputVisual]
	pendingEdit EditAction
}

func inputStateFor(ctx *core.Context, key string) *inputState {
	return core.UseState[inputState](ctx, key, stateSlotInput)
}

// inputDisclosureCfg builds a disclosure.Config from the widget's value fields.
func inputDisclosureCfg(widget Widget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *inputState) currentValue(widget Widget) string {
	s.value = s.disclosure.Current(inputDisclosureCfg(widget))
	return s.value
}

func (s *inputState) bind(widget Widget) {
	s.disclosure.Bind(inputDisclosureCfg(widget))
}

func (s *inputState) requestValue(widget Widget, value string) string {
	s.value, _ = s.disclosure.Request(inputDisclosureCfg(widget), value)
	return s.value
}

// queueEdit stores an edit action activated from the context menu; the next
// Layout consumes it.
func (s *inputState) queueEdit(action EditAction) {
	s.pendingEdit = action
}

func (s *inputState) takePendingEdit(gtx layout.Context, editor *widget.Editor, readOnly bool) {
	action := s.pendingEdit
	if action == editNone {
		return
	}
	s.pendingEdit = editNone
	RunEditAction(gtx, editor, action, readOnly)
}
