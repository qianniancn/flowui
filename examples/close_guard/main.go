// Command close_guard demonstrates a native close-request guard.
//
// The native window close button (or RequestClose from an application control)
// first asks the MVU program instead of closing the window. When the model is
// dirty, Update opens an AlertDialog; confirming discards the changes and
// closes the window, while cancelling keeps it open.
package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/alertdialog"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Dirty        bool
	ConfirmClose bool
}

type Msg any

type ToggleDirty struct{}
type Save struct{}
type CloseRequested struct{}
type DiscardAndClose struct{}
type CancelClose struct{}

// application is assigned before Run starts the event loop and is read only
// from Update.
var application *ui.Application

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.(type) {
	case ToggleDirty:
		m.Dirty = !m.Dirty
	case Save:
		m.Dirty = false
	case CloseRequested:
		if !m.Dirty {
			application.Close("main")
			return nil
		}
		m.ConfirmClose = true
	case DiscardAndClose:
		m.ConfirmClose = false
		application.Close("main")
	case CancelClose:
		m.ConfirmClose = false
	}
	return nil
}

func View(_ *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return layout.Stack(
		layout.Stacked(
			layout.Center(
				layout.Column(
					text.New("Close guard").Size(24),
					text.New("The native close button now asks the model before closing."),
					text.New(fmt.Sprintf("Unsaved changes: %v", m.Dirty)),
					layout.Row(
						button.New("Toggle dirty").OnClick(func() { send(ToggleDirty{}) }),
						button.New("Save").Disabled(!m.Dirty).OnClick(func() { send(Save{}) }),
						button.New("Request close").Variant(button.Secondary).OnClick(func() {
							send(CloseRequested{})
						}),
					).Gap(8),
				).Gap(12),
			),
		),
		layout.Overlay(confirmDialog(m, send)).Expanded(),
	)
}

func confirmDialog(m Model, send ui.Send[Msg]) ui.Widget {
	return alertdialog.New(
		"Discard unsaved changes?",
		"Your changes will be lost when the window closes.",
	).
		Key("close-guard").
		Open(m.ConfirmClose).
		Status(alertdialog.StatusWarning).
		Footer(layout.Row(
			button.New("Keep editing").Variant(button.Tertiary).OnClick(func() { send(CancelClose{}) }),
			button.New("Discard and close").Variant(button.Danger).OnClick(func() {
				send(DiscardAndClose{})
			}),
		).Gap(8)).
		OnOpenChange(func(open bool) {
			if !open {
				send(CancelClose{})
			}
		})
}

func main() {
	program := ui.Program[Model, Msg]{
		Init: func() (Model, ui.Cmd[Msg]) {
			return Model{Dirty: true}, nil
		},
		Update: Update,
		View:   View,
		// A native close request aborts the close and becomes a message. The
		// window stays open until Update calls Application.Close.
		CloseRequestMessage: func() Msg { return CloseRequested{} },
	}

	application = ui.NewApplication()
	application.Run(ui.NewWindow("main", program,
		window.Theme(base.Light()),
		window.Title("FlowUI Close Guard"),
		window.Size(720, 480),
	))
}
