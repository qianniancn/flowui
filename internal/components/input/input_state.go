package input

import (
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/frame"
)

type inputState struct {
	field.State
	disclosure disclosure.Binding[string]
	value      string
}

func inputStateFor(ctx *frame.Context, key string) *inputState {
	return frame.UseState[inputState](ctx, key, stateSlotInput)
}

// inputDisclosureCfg builds a disclosure.Config from the widget's value fields.
func inputDisclosureCfg(widget InputWidget) disclosure.Config[string] {
	return disclosure.Config[string]{
		Controlled: widget.hasValue,
		Value:      widget.value,
		HasDefault: widget.hasDefault,
		Default:    widget.defaultValue,
		OnChange:   widget.onChange,
	}
}

func (s *inputState) currentValue(widget InputWidget) string {
	s.value = s.disclosure.Current(inputDisclosureCfg(widget))
	return s.value
}

func (s *inputState) bind(widget InputWidget) {
	s.disclosure.Bind(inputDisclosureCfg(widget))
}

func (s *inputState) requestValue(widget InputWidget, value string) string {
	s.value, _ = s.disclosure.Request(inputDisclosureCfg(widget), value)
	return s.value
}
