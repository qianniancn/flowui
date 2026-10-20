package main

import (
	"fmt"

	"github.com/qianniancn/flowui/layout"
	"github.com/qianniancn/flowui/style"
	"github.com/qianniancn/flowui/theme/base"
	"github.com/qianniancn/flowui/ui"
	"github.com/qianniancn/flowui/widget/button"
	"github.com/qianniancn/flowui/widget/progress"
	"github.com/qianniancn/flowui/widget/text"
	"github.com/qianniancn/flowui/window"
)

type Model struct {
	Value float64
}

type Msg struct {
	Delta float64
	Reset bool
}

func Update(model *Model, msg Msg) ui.Cmd[Msg] {
	if msg.Reset {
		model.Value = 0
		return nil
	}
	model.Value = min(max(model.Value+msg.Delta, 0), 100)
	return nil
}

func View(_ *ui.Context, model Model, send ui.Send[Msg]) ui.Widget {
	return layout.Center(
		layout.Box(
			layout.Scroll(layout.Column(
				text.New("FlowUI ProgressCircle").Size(24),
				layout.Divider(),
				section("Controlled",
					layout.Column(
						layout.Row(
							progress.Circle("controlled", model.Value).Label("Upload"),
							text.New(fmt.Sprintf("%.0f%% complete", model.Value)),
						).Gap(12).AlignMiddle(),
						layout.Row(
							button.New("Advance").OnClick(func() { send(Msg{Delta: 10}) }),
							button.New("Reset").Variant(button.Secondary).OnClick(func() { send(Msg{Reset: true}) }),
						).Gap(8).AlignMiddle(),
					).Gap(12),
				),
				section("Sizes",
					layout.Row(
						labeledCircle("small", "Small", 40, progress.CircleSmall, progress.CircleAccent),
						labeledCircle("medium", "Medium", 60, progress.CircleMedium, progress.CircleAccent),
						labeledCircle("large", "Large", 80, progress.CircleLarge, progress.CircleAccent),
					).Gap(24).AlignMiddle(),
				),
				section("Colors",
					layout.Wrap(
						labeledCircle("default", "Default", 60, progress.CircleMedium, progress.CircleDefault),
						labeledCircle("accent", "Accent", 60, progress.CircleMedium, progress.CircleAccent),
						labeledCircle("success", "Success", 60, progress.CircleMedium, progress.CircleSuccess),
						labeledCircle("warning", "Warning", 60, progress.CircleMedium, progress.CircleWarning),
						labeledCircle("danger", "Danger", 60, progress.CircleMedium, progress.CircleDanger),
					).Gap(24).AlignMiddle(),
				),
				section("Indeterminate",
					layout.Row(
						progress.Circle("indeterminate", 0).Indeterminate().Label("Loading"),
						text.New("Loading"),
					).Gap(12).AlignMiddle(),
				),
				section("Disabled",
					layout.Row(progress.Circle("disabled", 65).Disabled(true).Label("Disabled progress")),
				),
			).Gap(20),
			).Vertical(),
		).Style(style.New().FillWidth().MaxWidth(760).Padding(24)),
	)
}

func section(title string, child ui.Widget) ui.Widget {
	return layout.Column(text.New(title).Size(18), child).Gap(10)
}

func labeledCircle(key, label string, value float64, size progress.CircleSize, circleColor progress.CircleColor) ui.Widget {
	return layout.Column(
		progress.Circle(key, value).Size(size).Color(circleColor).Label(label),
		text.New(label).Size(12),
	).Gap(6).AlignMiddle()
}

func main() {
	ui.Run(ui.NewProgram(Model{Value: 60},
		Update, View), window.Title("FlowUI ProgressCircle"),
		window.Size(860, 680),
		window.Theme(base.Light()),
	)
}
