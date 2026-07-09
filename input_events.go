package flowui

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

func (i InputWidget) dispatchEvents(editor *widget.Editor, events inputEvents) {
	if events.changed && i.onChange != nil {
		i.onChange(editor.Text())
	}
	if events.submitted && i.onSubmit != nil {
		i.onSubmit(events.submitText)
	}
}
