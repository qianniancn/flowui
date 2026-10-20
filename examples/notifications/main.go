package main

import (
	"context"
	"errors"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/notify"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/description"
	"github.com/qianniancn/flowui/widget/input"
	"github.com/qianniancn/flowui/widget/label"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Title        string
	Body         string
	Status       string
	Operation    string
	Notification *notify.Notification
}

type Msg interface{ msg() }

type EditTitle string
type EditBody string
type SendNotification struct{}
type CancelNotification struct{}

type NotificationSent struct {
	Notification *notify.Notification
	Err          error
}

type NotificationCanceled struct {
	Err error
}

func (EditTitle) msg()            {}
func (EditBody) msg()             {}
func (SendNotification) msg()     {}
func (CancelNotification) msg()   {}
func (NotificationSent) msg()     {}
func (NotificationCanceled) msg() {}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case EditTitle:
		model.Title = string(msg)
	case EditBody:
		model.Body = string(msg)
	case SendNotification:
		if model.Operation != "" {
			return nil
		}
		model.Operation = "send"
		model.Status = "Sending native notification..."
		return sendNotification(model.Title, model.Body)
	case NotificationSent:
		model.Operation = ""
		if msg.Err != nil {
			model.Status = "Notification failed: " + msg.Err.Error()
			return nil
		}
		model.Notification = msg.Notification
		if msg.Notification.Cancelable() {
			model.Status = "Notification sent; this platform supports cancellation"
		} else {
			model.Status = "Notification sent; this platform cannot cancel it"
		}
	case CancelNotification:
		if model.Operation != "" || model.Notification == nil || !model.Notification.Cancelable() {
			return nil
		}
		model.Operation = "cancel"
		model.Status = "Canceling notification..."
		handle := model.Notification
		return cancelNotification(handle)
	case NotificationCanceled:
		model.Operation = ""
		if msg.Err != nil {
			model.Status = "Cancel failed: " + msg.Err.Error()
			return nil
		}
		model.Notification = nil
		model.Status = "Notification canceled"
	}
	return nil
}

func sendNotification(title, body string) ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		handle, err := notify.Push(title, body)
		if err == nil {
			if canceled := ctx.Err(); canceled != nil {
				if handle.Cancelable() {
					_ = handle.Cancel()
				}
				return canceled
			}
		}
		send(NotificationSent{Notification: handle, Err: err})
		return nil
	})
}

func cancelNotification(handle *notify.Notification) ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		if err := ctx.Err(); err != nil {
			return err
		}
		err := handle.Cancel()
		if errors.Is(err, notify.ErrUnsupported) {
			err = nil
		}
		send(NotificationCanceled{Err: err})
		return nil
	})
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	canCancel := model.Notification != nil && model.Notification.Cancelable()
	busy := model.Operation != ""
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Native Notifications").Size(24),
				description.New(model.Status),
				label.New("Title").For("title"),
				input.New(model.Title).Key("title").
					FullWidth().
					OnChange(func(value string) { send(EditTitle(value)) }),
				label.New("Body").For("body"),
				input.NewTextArea(model.Body).Key("body").
					Rows(5).
					FullWidth().
					OnChange(func(value string) { send(EditBody(value)) }),
				layout.Row(
					button.New("Send notification").
						Disabled(busy).
						Loading(model.Operation == "send").
						OnClick(func() { send(SendNotification{}) }),
					button.New("Cancel notification").
						Variant(button.Secondary).
						Disabled(busy || !canCancel).
						Loading(model.Operation == "cancel").
						OnClick(func() { send(CancelNotification{}) }),
				).Gap(10),
			).Gap(12),
		).Style(style.New().Width(560).Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{
		Title:  "FlowUI",
		Body:   "Native notifications are available.",
		Status: "Ready",
	},

		Update,
		View), window.Title("FlowUI Notifications"),
		window.Size(700, 560),
		window.Theme(base.Light()),
	)
}
