package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/input"
	textpkg "github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Name string
}

type Msg any

type NameChanged struct {
	Name string
}

func Update(m *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case NameChanged:
		m.Name = msg.Name
	}
	return nil
}

func View(ctx *ui.Context, m Model, send ui.Send[Msg]) ui.Widget {
	greeting := "Hello"
	if m.Name != "" {
		greeting = fmt.Sprintf("Hello, %s", m.Name)
	}

	return layout.Center(
		layout.Box(
			layout.Column(
				textpkg.New("FlowUI Form").Size(24),
				layout.Box(
					input.New(m.Name).Key("name").
						Hint("Name").
						OnChange(func(text string) {
							send(NameChanged{Name: text})
						}),
				).Style(style.New().Width(320)),
				textpkg.New(greeting).Size(18),
			).Gap(12),
		).Style(style.New().Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Form"),
		window.Size(900, 600),
		window.Theme(base.Light()),
	)
}
