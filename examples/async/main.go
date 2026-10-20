package main

import (
	"context"
	"time"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Loading bool
	Result  string
}

type Msg any

type Load struct{}

type Loaded struct {
	Text string
}

func Init() (Model, ui.Cmd[Msg]) {
	return Model{Loading: true}, load()
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Load:
		if m.Loading {
			return nil
		}
		m.Loading = true
		m.Result = ""
		return load()
	case Loaded:
		m.Loading = false
		m.Result = msg.Text
	}
	return nil
}

func load() ui.Cmd[Msg] {
	return ui.DoContext(func(ctx context.Context, send ui.Send[Msg]) error {
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
			send(Loaded{Text: "Loaded after one second."})
			return nil
		}
	})
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := m.Result
	if m.Loading {
		status = "Waiting for async work..."
	}
	if status == "" {
		status = "Click Load to start a command."
	}

	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Async").Size(24),
				layout.Box(
					button.New("Load").
						Variant(button.Secondary).
						Loading(m.Loading).
						OnClick(func() {
							send(Load{})
						}),
				).Style(style.New().Width(160)),
				text.New(status).Size(18),
			).Gap(12),
		).Style(style.New().Padding(24)),
	)
}

func main() {
	ui.Run(ui.Program[Model, Msg]{
		Init:   Init,
		Update: Update,
		View:   View,
	},
		window.Title("FlowUI Async"),
		window.Size(900, 600),
		window.Theme(base.Light()),
	)
}
