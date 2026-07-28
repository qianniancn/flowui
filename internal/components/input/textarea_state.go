package input

import (
	"github.com/qianniancn/flowui/internal/components/disclosure"
	"github.com/qianniancn/flowui/internal/field"
	"github.com/qianniancn/flowui/internal/frame"
)

type textAreaState struct {
	field.State
	disclosure disclosure.Binding[string]
	value      string
}

func textAreaStateFor(ctx *frame.Context, key string) *textAreaState {
	return frame.UseStateWith(ctx, key, "textAreaState", func() *textAreaState {
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
