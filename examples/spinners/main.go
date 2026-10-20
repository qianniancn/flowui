package main

import (
	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	spinnerpkg "github.com/qianniancn/flowui/widget/spinner"
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
				text.New("FlowUI Spinners").Size(24),
				text.New("Loading indicators for pending work and background activity."),
				layout.Divider(),
				section("Default",
					layout.Center(spinnerpkg.New()),
				),
				section("Colors",
					layout.Wrap(
						labeledSpinner("Accent", spinnerpkg.New().Color(spinnerpkg.Accent)),
						labeledSpinner("Current", spinnerpkg.New().Color(spinnerpkg.Current)),
						labeledSpinner("Success", spinnerpkg.New().Color(spinnerpkg.Success)),
						labeledSpinner("Warning", spinnerpkg.New().Color(spinnerpkg.Warning)),
						labeledSpinner("Danger", spinnerpkg.New().Color(spinnerpkg.Danger)),
					).Gap(32).AlignMiddle(),
				),
				section("Sizes",
					layout.Wrap(
						labeledSpinner("Small", spinnerpkg.New().Size(spinnerpkg.Small)),
						labeledSpinner("Medium", spinnerpkg.New()),
						labeledSpinner("Large", spinnerpkg.New().Size(spinnerpkg.Large)),
						labeledSpinner("Extra large", spinnerpkg.New().Size(spinnerpkg.ExtraLarge)),
					).Gap(36).AlignMiddle(),
				),
				section("In context",
					layout.Row(
						spinnerpkg.New().Size(spinnerpkg.Small).Label("Saving settings"),
						text.New("Saving settings..."),
					).Gap(10).AlignMiddle(),
				),
			).Gap(20),
		).Style(style.New().FillWidth().MaxWidth(720).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(
		text.New(title).Size(18),
		child,
	).Gap(12)
}

func labeledSpinner(label string, spinner ui.Widget) ui.Widget {
	return layout.Column(
		layout.Center(spinner),
		text.New(label).Size(12),
	).Gap(8).AlignMiddle()
}

func main() {
	ui.Run(ui.NewProgram(Model{},
		Update, View), window.Title("FlowUI Spinners"),
		window.Size(820, 560),
		window.Theme(base.Light()),
	)
}
