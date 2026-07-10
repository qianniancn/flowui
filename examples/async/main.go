package main

import (
	"context"
	"time"

	"github.com/qianniancn/FlowUI/ui"
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

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case Load:
		if m.Loading {
			return nil
		}
		m.Loading = true
		m.Result = ""
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
	case Loaded:
		m.Loading = false
		m.Result = msg.Text
	}
	return nil
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	status := m.Result
	if m.Loading {
		status = "Waiting for async work..."
	}
	if status == "" {
		status = "Click Load to start a command."
	}

	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Async").Size(24),
				ui.Box(
					ui.Button("load", ui.Text("Load")).
						Variant(ui.ButtonSecondary).
						Loading(m.Loading).
						OnClick(func() {
							send(Load{})
						}),
				).Width(160),
				ui.Text(status).Size(18),
			).Gap(12),
		).Padding(24),
	)
}

func main() {
	ui.RunCmd(Model{}, Update, View,
		ui.Title("FlowUI Async"),
		ui.Size(900, 600),
	)
}
