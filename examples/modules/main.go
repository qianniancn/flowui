package main

import (
	"github.com/qianniancn/flowui/examples/modules/counter"
	"github.com/qianniancn/flowui/ui"
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
	return ui.Center(
		ui.Box(
			ui.Key("counter", counter.View(model.Counter, func(msg counter.Msg) {
				send(CounterMsg{Value: msg})
			})),
		).Style(ui.Padding(24)),
	)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), ui.Title("FlowUI MVU Modules"),
		ui.Size(640, 420),
	)
}
