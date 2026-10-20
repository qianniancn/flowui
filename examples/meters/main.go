package main

import (
	"fmt"
	"image/color"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/progress"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct{}
type Msg any

func Update(*Model, Msg) ui.Cmd[Msg] { return nil }

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Column(
				text.New("Meter").Size(24),
				section("Basic",
					progress.Meter("storage", 60).
						Label("Storage").
						ShowValue(),
				),
				section("Sizes",
					layout.Column(
						progress.Meter("small", 40).Label("Small").ShowValue().Color(progress.MeterSuccess).Size(progress.MeterSmall),
						progress.Meter("medium", 60).Label("Medium").ShowValue().Color(progress.MeterAccent),
						progress.Meter("large", 80).Label("Large").ShowValue().Color(progress.MeterWarning).Size(progress.MeterLarge),
					).Gap(14),
				),
				section("Colors",
					layout.Column(
						progress.Meter("default", 50).Label("Default").ShowValue().Color(progress.MeterDefault),
						progress.Meter("accent", 50).Label("Accent").ShowValue(),
						progress.Meter("success", 50).Label("Success").ShowValue().Color(progress.MeterSuccess),
						progress.Meter("warning", 50).Label("Warning").ShowValue().Color(progress.MeterWarning),
						progress.Meter("danger", 50).Label("Danger").ShowValue().Color(progress.MeterDanger),
					).Gap(14),
				),
				section("Custom value",
					progress.Meter("revenue", 750).
						Label("Revenue").
						Range(0, 1000).
						ValueFormatter(func(value float64) string {
							return fmt.Sprintf("$%.0f", value)
						}).
						Color(progress.MeterSuccess),
				),
				section("Custom fill",
					progress.Meter("custom-fill", 64).
						Label("Custom").
						ShowValue().
						Fill(color.NRGBA{R: 0x60, G: 0xa5, B: 0xfa, A: 0xff}),
				),
				section("Without visible label",
					progress.Meter("battery", 45).
						Alt("Battery level").
						Color(progress.MeterWarning),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(560).Padding(24)),
	)
}

func section(title string, content ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(16), content).Gap(10)
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Meter"), window.Size(720, 720), window.Theme(base.Light()))
}
