package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
	"github.com/qianniancn/flowui-icons-lucide"
)

type CounterModel struct {
	Value int
}

type CounterMsg int

func counterUpdate(model *CounterModel, msg CounterMsg) {
	model.Value += int(msg)
}

func counterView(application *ui.Application) ui.View[CounterModel, CounterMsg] {
	return func(_ *ui.Context, model CounterModel, send ui.Send[CounterMsg]) ui.Widget {
		return ui.Center(
			ui.Column(
				ui.Text("Independent window").Size(22),
				ui.Text(fmt.Sprintf("Count: %d", model.Value)).Size(18),
				ui.Row(
					ui.Button("decrease", ui.Icon(lucide.Minus).Size(16)).OnClick(func() { send(-1) }),
					ui.Button("increase", ui.Icon(lucide.Plus).Size(16)).OnClick(func() { send(1) }),
					ui.Button("close", ui.Text("Close")).Variant(ui.ButtonSecondary).OnClick(func() { application.Close("counter") }),
				).Gap(8).AlignMiddle(),
			).Gap(16).AlignMiddle(),
		)
	}
}

type MainModel struct{}
type MainMsg struct{}

func main() {
	application := ui.NewApplication()
	counter := ui.NewWindow(
		"counter",
		CounterModel{},
		counterUpdate,
		counterView(application),
		ui.Title("Counter"),
		ui.Size(420, 260),
	)
	mainWindow := ui.NewWindow(
		"main",
		MainModel{},
		func(*MainModel, MainMsg) {},
		func(_ *ui.Context, _ MainModel, _ ui.Send[MainMsg]) ui.Widget {
			return ui.Center(
				ui.Column(
					ui.Text("FlowUI Multi-window").Size(24),
					ui.Button("open-counter", ui.Text("Open counter")).OnClick(func() { application.Open(counter) }),
					ui.Button("close-all", ui.Text("Close all")).Variant(ui.ButtonSecondary).OnClick(application.CloseAll),
				).Gap(16).AlignMiddle(),
			)
		},
		ui.Title("Multi-window"),
		ui.Size(520, 320),
	)
	application.Run(mainWindow)
}
