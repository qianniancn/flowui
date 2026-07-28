package input

import "gioui.org/widget"

type inputEvents struct {
	changed    bool
	submitted  bool
	submitText string
}

func (e *inputEvents) add(event widget.EditorEvent) {
	switch event := event.(type) {
	case widget.ChangeEvent:
		e.changed = true
	case widget.SubmitEvent:
		e.submitted = true
		e.submitText = event.Text
	}
}

func (i InputWidget) dispatchEvents(state *inputState, editor *widget.Editor, events inputEvents) {
	if events.changed {
		state.requestValue(i, editor.Text())
	}
	if events.submitted && i.onSubmit != nil {
		i.onSubmit(events.submitText)
	}
}
