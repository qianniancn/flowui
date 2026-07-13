package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct{}
type Msg any

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("Meter").Size(24),
				section("Basic",
					ui.Meter("storage", 60).
						Label("Storage").
						ShowValue(),
				),
				section("Sizes",
					ui.Column(
						ui.Meter("small", 40).Label("Small").ShowValue().Color(ui.MeterSuccess).Size(ui.MeterSmall),
						ui.Meter("medium", 60).Label("Medium").ShowValue().Color(ui.MeterAccent),
						ui.Meter("large", 80).Label("Large").ShowValue().Color(ui.MeterWarning).Size(ui.MeterLarge),
					).Gap(14),
				),
				section("Colors",
					ui.Column(
						ui.Meter("default", 50).Label("Default").ShowValue().Color(ui.MeterDefault),
						ui.Meter("accent", 50).Label("Accent").ShowValue(),
						ui.Meter("success", 50).Label("Success").ShowValue().Color(ui.MeterSuccess),
						ui.Meter("warning", 50).Label("Warning").ShowValue().Color(ui.MeterWarning),
						ui.Meter("danger", 50).Label("Danger").ShowValue().Color(ui.MeterDanger),
					).Gap(14),
				),
				section("Custom value",
					ui.Meter("revenue", 750).
						Label("Revenue").
						Range(0, 1000).
						ValueFormatter(func(value float64) string {
							return fmt.Sprintf("$%.0f", value)
						}).
						Color(ui.MeterSuccess),
				),
				section("Without visible label",
					ui.Meter("battery", 45).
						Alt("Battery level").
						Color(ui.MeterWarning),
				),
			).Gap(20),
		).FillWidth().MaxWidth(560).Padding(24),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(16), content).Gap(10)
}

func main() {
	ui.Run(Model{}, Update, View, ui.Title("FlowUI Meter"), ui.Size(720, 720))
}
