package main

import (
	"fmt"

	"github.com/qianniancn/flowui/ui"
)

type Model struct {
	Count int
}

type Msg any

type Inc struct{}
type Dec struct{}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg.(type) {
	case Inc:
		m.Count++
	case Dec:
		m.Count--
	}
	return nil
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Counter").Size(24),
				ui.Text(fmt.Sprintf("count: %d", m.Count)).Size(18),
				ui.Row(
					ui.Box(
						ui.Button("inc", ui.Text("Click +1")).OnClick(func() {
							send(Inc{})
						}),
					),
					ui.Box(
						ui.Button("dec", ui.Text("Click -1")).OnClick(func() {
							send(Dec{})
						}).Variant(ui.ButtonSecondary),
					),
				).Gap(12),
			).Gap(12),
		).Style(ui.Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), ui.Title("FlowUI Counter"),
		ui.Size(900, 600),
	)
}
