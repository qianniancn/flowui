package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Name string
}

type Msg any

type NameChanged struct {
	Name string
}

func Update(m *Model, msg Msg) {
	switch msg := msg.(type) {
	case NameChanged:
		m.Name = msg.Name
	}
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	greeting := "Hello"
	if m.Name != "" {
		greeting = fmt.Sprintf("Hello, %s", m.Name)
	}

	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Form").Size(24),
				ui.Box(
					ui.Input("name", m.Name).
						Hint("Name").
						OnChange(func(text string) {
							send(NameChanged{Name: text})
						}),
				).Width(320),
				ui.Text(greeting).Size(18),
			).Gap(12),
		).Padding(24),
	)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Form"),
		ui.Size(900, 600),
	)
}
