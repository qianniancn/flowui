package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct {
	Value float64
}

type Msg struct {
	Delta float64
	Reset bool
}

func Update(model *Model, msg Msg) {
	if msg.Reset {
		model.Value = 0
		return
	}
	model.Value = min(max(model.Value+msg.Delta, 0), 100)
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return ui.Center(
		ui.Box(
			ui.Scroll("progress-circles",
				ui.Column(
					ui.Text("FlowUI ProgressCircle").Size(24),
					ui.Divider(),
					section("Controlled",
						ui.Column(
							ui.Row(
								ui.ProgressCircle("controlled", model.Value).Label("Upload"),
								ui.Text(fmt.Sprintf("%.0f%% complete", model.Value)),
							).Gap(12).AlignMiddle(),
							ui.Row(
								ui.Button("advance", ui.Text("Advance")).OnClick(func() { send(Msg{Delta: 10}) }),
								ui.Button("reset", ui.Text("Reset")).Variant(ui.ButtonSecondary).OnClick(func() { send(Msg{Reset: true}) }),
							).Gap(8).AlignMiddle(),
						).Gap(12),
					),
					section("Sizes",
						ui.Row(
							labeledCircle("small", "Small", 40, ui.ProgressCircleSmall, ui.ProgressCircleAccent),
							labeledCircle("medium", "Medium", 60, ui.ProgressCircleMedium, ui.ProgressCircleAccent),
							labeledCircle("large", "Large", 80, ui.ProgressCircleLarge, ui.ProgressCircleAccent),
						).Gap(24).AlignMiddle(),
					),
					section("Colors",
						ui.Wrap(
							labeledCircle("default", "Default", 60, ui.ProgressCircleMedium, ui.ProgressCircleDefault),
							labeledCircle("accent", "Accent", 60, ui.ProgressCircleMedium, ui.ProgressCircleAccent),
							labeledCircle("success", "Success", 60, ui.ProgressCircleMedium, ui.ProgressCircleSuccess),
							labeledCircle("warning", "Warning", 60, ui.ProgressCircleMedium, ui.ProgressCircleWarning),
							labeledCircle("danger", "Danger", 60, ui.ProgressCircleMedium, ui.ProgressCircleDanger),
						).Gap(24).AlignMiddle(),
					),
					section("Indeterminate",
						ui.Row(
							ui.ProgressCircle("indeterminate", 0).Indeterminate().Label("Loading"),
							ui.Text("Loading"),
						).Gap(12).AlignMiddle(),
					),
					section("Disabled",
						ui.Row(ui.ProgressCircle("disabled", 65).Disabled(true).Label("Disabled progress")),
					),
				).Gap(20),
			).Vertical(),
		).Style(ui.FillWidth()).Style(ui.MaxWidth(760)).Style(ui.Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return ui.Column(ui.Text(title).Size(18), child).Gap(10)
}

func labeledCircle(key, label string, value float64, size ui.ProgressCircleSize, circleColor ui.ProgressCircleColor) ui.Widget {
	return ui.Column(
		ui.ProgressCircle(key, value).Size(size).Color(circleColor).Label(label),
		ui.Text(label).Size(12),
	).Gap(6).AlignMiddle()
}

func main() {
	ui.Run(Model{Value: 60}, Update, View,
		ui.Title("FlowUI ProgressCircle"),
		ui.Size(860, 680),
	)
}
