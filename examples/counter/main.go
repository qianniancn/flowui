package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
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
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("FlowUI Counter").Size(24),
				text.New(fmt.Sprintf("count: %d", m.Count)).Size(18),
				layout.Row(
					layout.Box(
						button.New("Click +1").OnClick(func() {
							send(Inc{})
						}),
					),
					layout.Box(
						button.New("Click -1").OnClick(func() {
							send(Dec{})
						}).Variant(button.Secondary),
					),
				).Gap(12),
			).Gap(12),
		).Style(style.New().Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Counter"),
		window.Size(900, 600),
		window.Theme(base.Light()),
	)
}
