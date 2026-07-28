package main

import "github.com/qianniancn/FlowUI/ui"

type Model struct{}
type Msg any

func Update(*Model, Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Column(
				ui.Text("FlowUI Spinners").Size(24),
				ui.Text("Loading indicators for pending work and background activity."),
				ui.Divider(),
				section("Default",
					ui.Center(ui.Spinner()),
				),
				section("Colors",
					ui.Wrap(
						labeledSpinner("Accent", ui.Spinner().Color(ui.SpinnerAccent)),
						labeledSpinner("Current", ui.Spinner().Color(ui.SpinnerCurrent)),
						labeledSpinner("Success", ui.Spinner().Color(ui.SpinnerSuccess)),
						labeledSpinner("Warning", ui.Spinner().Color(ui.SpinnerWarning)),
						labeledSpinner("Danger", ui.Spinner().Color(ui.SpinnerDanger)),
					).Gap(32).AlignMiddle(),
				),
				section("Sizes",
					ui.Wrap(
						labeledSpinner("Small", ui.Spinner().Size(ui.SpinnerSmall)),
						labeledSpinner("Medium", ui.Spinner()),
						labeledSpinner("Large", ui.Spinner().Size(ui.SpinnerLarge)),
						labeledSpinner("Extra large", ui.Spinner().Size(ui.SpinnerExtraLarge)),
					).Gap(36).AlignMiddle(),
				),
				section("In context",
					ui.Row(
						ui.Spinner().Size(ui.SpinnerSmall).Label("Saving settings"),
						ui.Text("Saving settings..."),
					).Gap(10).AlignMiddle(),
				),
			).Gap(20),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(720)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(
		ui.Text(title).Size(18),
		child,
	).Gap(12)
}

func labeledSpinner(label string, spinner ui.Widget) ui.Widget {
	return ui.Column(
		ui.Center(spinner),
		ui.Text(label).Size(12),
	).Gap(8).AlignMiddle()
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Spinners"),
		ui.Size(820, 560),
	)
}
