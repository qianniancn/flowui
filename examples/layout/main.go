package main

import (
	"fmt"

	"github.com/qianniancn/FlowUI/ui"
)

type Model struct{}

type Msg any

func Update(_ *Model, _ Msg) {}

func View(_ *ui.Context, _ Model, _ ui.Send[Msg]) ui.Widget {
	cards := []ui.Widget{
		statCard("Open", "18"),
		statCard("Done", "42"),
		statCard("Late", "3"),
		statCard("Queued", "9"),
	}

	tags := make([]ui.Widget, 0, 8)
	for i := 1; i <= 8; i++ {
		tags = append(tags,
			ui.Box(ui.Text(fmt.Sprintf("Tag %d", i))).
				PaddingTop(6).
				PaddingRight(10).
				PaddingBottom(6).
				PaddingLeft(10).
				Margin(2),
		)
	}

	return ui.Center(
		ui.Box(
			ui.Scroll("page",
				ui.Column(
					ui.Text("FlowUI Layout").Size(24),
					ui.Divider(),
					ui.Row(
						ui.Text("Overview").Size(18),
						ui.Box(ui.Separator()).Height(24),
						ui.Text("Responsive cards").Size(18),
					).Gap(12).AlignMiddle(),
					ui.AutoGrid(160, cards...).Gap(12),
					ui.Box(
						ui.AspectRatio(16.0/9.0,
							ui.Stack(
								ui.Stacked(
									ui.Box(ui.Text("Aspect preview")).
										FillWidth().
										FillHeight().
										Align(ui.AlignCenter),
								),
								ui.Overlay(
									ui.Box(ui.Text("Overlay")).
										PaddingTop(6).
										PaddingRight(10).
										PaddingBottom(6).
										PaddingLeft(10),
								).Align(ui.AlignTopEnd),
							),
						),
					).Clip(),
					ui.Wrap(tags...).Gap(8).AlignMiddle(),
				).Gap(16),
			).Vertical(),
		).FillWidth().MaxWidth(760).Padding(24),
	)
}

func statCard(label, value string) ui.Widget {
	return ui.Box(
		ui.Column(
			ui.Text(label).Size(14),
			ui.Text(value).Size(28),
		).Gap(6),
	).FillWidth().Padding(16)
}

func main() {
	ui.Run(Model{}, Update, View,
		ui.Title("FlowUI Layout"),
		ui.Size(900, 600),
	)
}
