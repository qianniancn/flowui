package main

import (
	"github.com/qianniancn/flowui/examples/modules/counter"
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Counter counter.Model
}

type Msg any

type CounterMsg struct {
	Value counter.Msg
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	switch msg := msg.(type) {
	case CounterMsg:
		cmd := counter.Update(&model.Counter, msg.Value)
		return ui.MapCmd(cmd, func(child counter.Msg) Msg {
			return CounterMsg{Value: child}
		})
	}
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Key("counter", counter.View(model.Counter, func(msg counter.Msg) {
				send(CounterMsg{Value: msg})
			})),
		).Style(style.New().Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI MVU Modules"),
		window.Size(640, 420),
		window.Theme(base.Light()),
	)
}
